package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type ChronyTrackingResponse struct {
	ReferenceID      string `json:"reference_id"`
	Stratum          string `json:"stratum"`
	RefTime          string `json:"ref_time_utc"`
	SystemTime       string `json:"system_time"`
	LastOffset       string `json:"last_offset_sec"`
	RMSOffset        string `json:"rms_offset_sec"`
	Frequency        string `json:"frequency_ppm"`
	ResidualFreq     string `json:"residual_freq_ppm"`
	Skew             string `json:"skew_ppm"`
	RootDelay        string `json:"root_delay_sec"`
	RootDispersion   string `json:"root_dispersion_sec"`
	UpdateInterval   string `json:"update_interval_sec"`
	LeapStatus       string `json:"leap_status"`
	ErrorDescription string `json:"error,omitempty"`
}

func parseChronyOutput(output string) ChronyTrackingResponse {
	lines := strings.Split(output, "\n")
	data := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			data[key] = value
		}
	}

	return ChronyTrackingResponse{
		ReferenceID:      data["Reference ID"],
		Stratum:          data["Stratum"],
		RefTime:          data["Ref time (UTC)"],
		SystemTime:       data["System time"],
		LastOffset:       data["Last offset"],
		RMSOffset:        data["RMS offset"],
		Frequency:        data["Frequency"],
		ResidualFreq:     data["Residual freq"],
		Skew:             data["Skew"],
		RootDelay:        data["Root delay"],
		RootDispersion:   data["Root dispersion"],
		UpdateInterval:   data["Update interval"],
		LeapStatus:       data["Leap status"],
	}
}

func trackingHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("chronyc", "tracking")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: Failed to execute 'chronyc tracking': %v, output: %s", err, string(out))
		http.Error(w, `{"error": "failed to execute chronyc command"}`, http.StatusInternalServerError)
		return
	}
	response := parseChronyOutput(string(out))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/tracking", trackingHandler)
	log.Println("ntp-proxy server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}