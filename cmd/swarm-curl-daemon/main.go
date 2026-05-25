package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/ismdeep/swarm-curl/cmd/swarm-curl-daemon/config"
	"github.com/ismdeep/swarm-curl/daemon"
)

//go:embed config.default.yaml
var defaultConfig []byte

func main() {
	var (
		addr       string
		token      string
		configFile string
	)

	rootCmd := &cobra.Command{
		Use:           "swarm-curl-daemon",
		Short:         "Distributed collaborative download daemon",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("[INFO] swarm-curl-daemon")
			log.Println("[INFO] GitHub: https://github.com/ismdeep/swarm-curl")
			if info, ok := debug.ReadBuildInfo(); ok {
				var commit, time string
				for _, setting := range info.Settings {
					if setting.Key == "vcs.revision" {
						commit = setting.Value
					} else if setting.Key == "vcs.time" {
						time = setting.Value
					}
				}
				log.Printf("[INFO] GitTime: %s, GitCommit: %s\n", time, commit)
			}

			// load config
			cfg, err := config.Load(defaultConfig)
			if err != nil {
				return err
			}
			if configFile != "" {
				cfg, err = config.LoadFromFile(configFile)
				if err != nil {
					return err
				}
			}
			if addr != "" {
				cfg.Addr = addr
			}
			if token != "" {
				cfg.Token = token
			}

			return daemon.NewServer(cfg.Token).Start(cfg.Addr)
		},
	}

	rootCmd.PersistentFlags().StringVar(&addr, "addr", "", "Listen address")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Authentication token")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file")

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		os.Exit(1)
	}
}
