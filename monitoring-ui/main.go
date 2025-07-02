package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PodStatus struct {
	PodName    string `json:"pod_name"`
	PodIP      string `json:"pod_ip"`
	IsHealthy  bool   `json:"is_healthy"`
	Error      string `json:"error,omitempty"`
	Stratum    string `json:"stratum,omitempty"`
	LastOffset string `json:"last_offset_sec,omitempty"`
	RootDelay  string `json:"root_delay_sec,omitempty"`
}

var clientset *kubernetes.Clientset

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("FATAL: Failed to create in-cluster config: %v", err)
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("FATAL: Failed to create clientset: %v", err)
	}

	http.HandleFunc("/", serveUI)
	http.HandleFunc("/api/status", statusHandler)

	log.Println("monitoring-ui server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=ntp-proxy",
	})
	if err != nil {
		http.Error(w, "Failed to list pods", http.StatusInternalServerError)
		log.Printf("ERROR: listing pods: %v", err)
		return
	}

	var wg sync.WaitGroup
	statusChan := make(chan PodStatus, len(pods.Items))

	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning && pod.DeletionTimestamp == nil {
			wg.Add(1)
			go fetchPodStatus(pod, statusChan, &wg)
		}
	}

	wg.Wait()
	close(statusChan)

	var statuses []PodStatus
	for status := range statusChan {
		statuses = append(statuses, status)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

func fetchPodStatus(pod v1.Pod, statusChan chan<- PodStatus, wg *sync.WaitGroup) {
	defer wg.Done()
	status := PodStatus{PodName: pod.Name, PodIP: pod.Status.PodIP}
	url := fmt.Sprintf("http://%s:8080/tracking", pod.Status.PodIP)
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		status.IsHealthy, status.Error = false, "Connection failed"
		statusChan <- status
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status.IsHealthy, status.Error = false, fmt.Sprintf("HTTP %d", resp.StatusCode)
		statusChan <- status
		return
	}

	body, _ := io.ReadAll(resp.Body)
	var trackingData map[string]interface{}
	if json.Unmarshal(body, &trackingData) != nil {
		status.IsHealthy, status.Error = false, "Invalid JSON"
		statusChan <- status
		return
	}
	
	status.IsHealthy = true
	status.Stratum = fmt.Sprintf("%v", trackingData["stratum"])
	status.LastOffset = fmt.Sprintf("%v", trackingData["last_offset_sec"])
	status.RootDelay = fmt.Sprintf("%v", trackingData["root_delay_sec"])
	statusChan <- status
}