package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ismdeep/swarm-curl/daemon"
)

func main() {
	var (
		addr  string
		token string
	)

	rootCmd := &cobra.Command{
		Use:           "swarm-curl-daemon",
		Short:         "Distributed collaborative download daemon",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if info, ok := debug.ReadBuildInfo(); ok {
				var commit, time string
				for _, setting := range info.Settings {
					if setting.Key == "vcs.revision" {
						commit = setting.Value
					} else if setting.Key == "vcs.time" {
						time = setting.Value
					}
				}
				log.Printf("[INFO] BuildTime: %s, GitCommit: %s", time, commit)
			}

			addr = loadAddr(addr)
			token = loadToken(token)
			if token == "" {
				return errors.New("token is not set")
			}

			server := daemon.NewServer(token)
			log.Println("[INFO] Starting daemon on:", addr)
			return server.Start(addr)
		},
	}

	rootCmd.PersistentFlags().StringVar(&addr, "addr", "", "Listen address")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Authentication token")

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		os.Exit(1)
	}
}

func loadAddr(cmdAddr string) string {
	if cmdAddr != "" {
		return cmdAddr
	}

	if envAddr := os.Getenv("SWARM_CURL_ADDR"); envAddr != "" {
		return envAddr
	}

	if addr := readTokenFile("/etc/swarm-curl-daemon/addr"); addr != "" {
		return addr
	}

	if home, err := os.UserHomeDir(); err == nil {
		if addr := readTokenFile(filepath.Join(home, ".swarm-curl-daemon", "addr")); addr != "" {
			return addr
		}
	}

	return "0.0.0.0:8080"
}

func loadToken(cmdToken string) string {
	if cmdToken != "" {
		return cmdToken
	}

	if envToken := os.Getenv("SWARM_CURL_TOKEN"); envToken != "" {
		return envToken
	}

	if token := readTokenFile("/etc/swarm-curl-daemon/token"); token != "" {
		return token
	}

	if home, err := os.UserHomeDir(); err == nil {
		if token := readTokenFile(filepath.Join(home, ".swarm-curl-daemon", "token")); token != "" {
			return token
		}
	}

	return ""
}

func readTokenFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
