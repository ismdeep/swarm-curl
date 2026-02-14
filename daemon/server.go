package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Server struct {
	token string
}

func NewServer(token string) *Server {
	return &Server{token: token}
}

func (s *Server) auth(r *http.Request) bool {
	return r.Header.Get("Authorization") == "Bearer "+s.token
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if !s.auth(r) {
		log.Printf("[UNAUTHORIZED] %s from %s", r.URL.Path, r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		URL     string   `json:"url"`
		Start   int64    `json:"start"`
		End     int64    `json:"end"`
		Headers []string `json:"headers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request from %s: %v", r.RemoteAddr, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[REQUEST] %s | Range: %d-%d | Client: %s", req.URL, req.Start, req.End, r.RemoteAddr)

	httpReq, err := http.NewRequest("GET", req.URL, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to create request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpReq.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", req.Start, req.End))
	for _, h := range req.Headers {
		for i := 0; i < len(h); i++ {
			if h[i] == ':' {
				httpReq.Header.Set(h[:i], h[i+1:])
				break
			}
		}
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Printf("[ERROR] Download failed for %s: %v", req.URL, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		log.Printf("[ERROR] Upstream returned status %d for %s", resp.StatusCode, req.URL)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	written, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to send data: %v", err)
		return
	}
	log.Printf("[SUCCESS] Sent %d bytes | Range: %d-%d | Client: %s", written, req.Start, req.End, r.RemoteAddr)
}

func (s *Server) Start(addr string) error {
	http.HandleFunc("/download", s.handleDownload)
	log.Printf("[INFO] Daemon server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}
