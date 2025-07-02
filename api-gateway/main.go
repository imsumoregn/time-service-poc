package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func getUpstreamURL() string {
	url := os.Getenv("NTP_PROXY_URL")
	if url == "" {
		log.Fatal("FATAL: NTP_PROXY_URL environment variable not set.")
	}
	return url
}

func apiStatusHandler(w http.ResponseWriter, r *http.Request) {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(getUpstreamURL())
	if err != nil {
		log.Printf("ERROR: Failed to call ntp-proxy service: %v", err)
		http.Error(w, `{"error": "upstream service unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	log.Println("api-gateway server starting on port 8080")
	http.HandleFunc("/api/v1/status", apiStatusHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}