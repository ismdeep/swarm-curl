package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/ismdeep/swarm-curl/config"
)

var (
	output     string
	remoteName bool
	headers    []string
	verbose    bool
)

type endpointStats struct {
	bytes    int64
	duration time.Duration
	mu       sync.Mutex
}

var globalStats = struct {
	endpoints map[string]*endpointStats
	mu        sync.RWMutex
}{
	endpoints: make(map[string]*endpointStats),
}

func getFileSize(url string, headers []string) (int64, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}
	for _, h := range headers {
		if idx := bytes.IndexByte([]byte(h), ':'); idx > 0 {
			req.Header.Set(h[:idx], h[idx+1:])
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.ContentLength, nil
}

func downloadChunk(endpoint config.Endpoint, url string, start, end int64, headers []string, bar *progressbar.ProgressBar) ([]byte, error) {
	startTime := time.Now()
	reqBody, _ := json.Marshal(map[string]interface{}{
		"url":     url,
		"start":   start,
		"end":     end,
		"headers": headers,
	})

	req, err := http.NewRequest("POST", endpoint.Address+"/download", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+endpoint.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			buf.Write(buffer[:n])
			bar.Add(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	duration := time.Since(startTime)
	globalStats.mu.Lock()
	if globalStats.endpoints[endpoint.Address] == nil {
		globalStats.endpoints[endpoint.Address] = &endpointStats{}
	}
	globalStats.endpoints[endpoint.Address].mu.Lock()
	globalStats.endpoints[endpoint.Address].bytes += int64(buf.Len())
	globalStats.endpoints[endpoint.Address].duration += duration
	if verbose && duration > 0 {
		speed := float64(buf.Len()) / duration.Seconds() / 1024 / 1024
		fmt.Fprintf(os.Stderr, "\n[%s] %.2f MB/s\n", endpoint.Address, speed)
	}
	globalStats.endpoints[endpoint.Address].mu.Unlock()
	globalStats.mu.Unlock()

	return buf.Bytes(), nil
}

type resumeState struct {
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	ChunkSize int64  `json:"chunk_size"`
	Completed []bool `json:"completed"`
}

func loadResumeState(path string) (*resumeState, error) {
	data, err := os.ReadFile(path + ".resume")
	if err != nil {
		return nil, err
	}
	var state resumeState
	return &state, json.Unmarshal(data, &state)
}

func (s *resumeState) save(path string) error {
	data, _ := json.Marshal(s)
	tmpFile := path + ".resume.tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, path+".resume")
}

func run(cmd *cobra.Command, args []string) error {
	url := args[0]

	if output == "" && !remoteName {
		return fmt.Errorf("must specify -o or -O")
	}
	if remoteName {
		if idx := bytes.LastIndexByte([]byte(url), '/'); idx >= 0 {
			output = url[idx+1:]
		} else {
			output = "download"
		}
		if output == "" {
			output = "download"
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	size, err := getFileSize(url, headers)
	if err != nil {
		return fmt.Errorf("error getting file size: %w", err)
	}

	const maxChunkSize = 1 * 1024 * 1024 // 1MB
	numEndpoints := len(cfg.Endpoints)
	numChunks := int((size + maxChunkSize - 1) / maxChunkSize)

	state, err := loadResumeState(output)
	if err != nil || state.URL != url || state.Size != size || len(state.Completed) != numChunks {
		state = &resumeState{URL: url, Size: size, ChunkSize: maxChunkSize, Completed: make([]bool, numChunks)}
	}

	var alreadyDownloaded int64
	for _, done := range state.Completed {
		if done {
			alreadyDownloaded += maxChunkSize
		}
	}
	if alreadyDownloaded > size {
		alreadyDownloaded = size
	}

	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	if alreadyDownloaded > 0 {
		bar.Set64(alreadyDownloaded)
	}

	file, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Handle Ctrl+C signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nSaving progress...")
		os.Exit(0)
	}()

	var wg sync.WaitGroup
	failedEndpoints := make(map[int]bool)
	var mu sync.Mutex

	semaphores := make([]chan struct{}, numEndpoints)
	for i := range semaphores {
		semaphores[i] = make(chan struct{}, 8)
	}

	for i := 0; i < numChunks; i++ {
		if state.Completed[i] {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			start := int64(idx) * maxChunkSize
			end := start + maxChunkSize - 1
			if end >= size {
				end = size - 1
			}

			for attempt := 0; attempt < numEndpoints; attempt++ {
				epIdx := (idx + attempt) % numEndpoints
				mu.Lock()
				if failedEndpoints[epIdx] {
					mu.Unlock()
					continue
				}
				mu.Unlock()

				semaphores[epIdx] <- struct{}{}
				ep := cfg.Endpoints[epIdx]
				data, err := downloadChunk(ep, url, start, end, headers, bar)
				<-semaphores[epIdx]
				if err != nil {
					mu.Lock()
					failedEndpoints[epIdx] = true
					mu.Unlock()
					continue
				}
				// Write data to file first
				if _, err := file.WriteAt(data, int64(idx)*maxChunkSize); err != nil {
					fmt.Printf("\nError writing chunk %d: %v\n", idx, err)
					continue
				}
				if err := file.Sync(); err != nil {
					fmt.Printf("\nError syncing chunk %d: %v\n", idx, err)
					continue
				}
				// Update state atomically after successful write
				mu.Lock()
				state.Completed[idx] = true
				_ = state.save(output)
				mu.Unlock()
				return
			}
			fmt.Printf("\nError: all endpoints failed for chunk %d\n", idx)
		}(i)
	}

	wg.Wait()
	bar.Finish()

	if verbose {
		fmt.Fprintln(os.Stderr, "\nDaemon Speed Summary:")
		globalStats.mu.RLock()
		for addr, stats := range globalStats.endpoints {
			stats.mu.Lock()
			if stats.duration > 0 {
				speed := float64(stats.bytes) / stats.duration.Seconds() / 1024 / 1024
				fmt.Fprintf(os.Stderr, "  %s: %.2f MB/s (%.2f MB)\n", addr, speed, float64(stats.bytes)/1024/1024)
			}
			stats.mu.Unlock()
		}
		globalStats.mu.RUnlock()
	}

	// Check if all chunks are completed
	for i := 0; i < numChunks; i++ {
		if !state.Completed[i] {
			return fmt.Errorf("download failed: chunk %d missing", i)
		}
	}

	_ = os.Remove(output + ".resume")

	fmt.Println("Download complete:", output)
	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:           "swarm-curl <url>",
		Short:         "Distributed collaborative download tool",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE:          run,
	}
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Write to file instead of stdout")
	rootCmd.Flags().BoolVarP(&remoteName, "remote-name", "O", false, "Write output to file named as remote file")
	rootCmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "Custom HTTP headers")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show speed for each daemon")

	endpointCmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Manage endpoints",
	}

	endpointCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			for i, ep := range cfg.Endpoints {
				fmt.Printf("%d. %s (token: %s)\n", i+1, ep.Address, ep.Token)
			}
			return nil
		},
	})

	var addEndpoint, addToken string
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				cfg = &config.Config{}
			}
			cfg.Endpoints = append(cfg.Endpoints, config.Endpoint{Address: addEndpoint, Token: addToken})
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Println("Endpoint added successfully")
			return nil
		},
	}
	addCmd.Flags().StringVar(&addEndpoint, "endpoint", "", "Endpoint address")
	addCmd.Flags().StringVar(&addToken, "token", "", "Authentication token")
	_ = addCmd.MarkFlagRequired("endpoint")
	_ = addCmd.MarkFlagRequired("token")
	endpointCmd.AddCommand(addCmd)

	endpointCmd.AddCommand(&cobra.Command{
		Use:   "delete <address>",
		Short: "Delete endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := args[0]
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			found := false
			for i, ep := range cfg.Endpoints {
				if ep.Address == address {
					cfg.Endpoints = append(cfg.Endpoints[:i], cfg.Endpoints[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("endpoint not found")
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Println("Endpoint deleted successfully")
			return nil
		},
	})

	rootCmd.AddCommand(endpointCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
