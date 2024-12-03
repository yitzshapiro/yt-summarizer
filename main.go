package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	http.HandleFunc("/process_video", corsMiddleware(processVideo))
	log.Println("Starting server on :5001...")
	log.Fatal(http.ListenAndServe(":5001", nil))
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func downloadAudio(url string) error {
	// Get the absolute path to the virtual environment's Python executable
	pythonPath := ".venv/bin/python3"

	log.Printf("Using Python path: %s", pythonPath)

	// Check if Python executable exists
	if _, err := os.Stat(pythonPath); os.IsNotExist(err) {
		log.Printf("Python executable not found at: %s", pythonPath)
		return fmt.Errorf("python executable not found: %v", err)
	}

	// Check if pip is available
	pipCheckCmd := exec.Command(pythonPath, "-m", "pip", "--version")
	pipCheckOutput, err := pipCheckCmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to check pip version: %v", err)
		log.Printf("pip check output: %s", string(pipCheckOutput))
		return fmt.Errorf("pip not available: %v", err)
	}
	log.Printf("pip version: %s", string(pipCheckOutput))

	// Check if yt-dlp is installed
	checkCmd := exec.Command(pythonPath, "-m", "pip", "list")
	pipOutput, err := checkCmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to list pip packages: %v", err)
		log.Printf("pip list output: %s", string(pipOutput))
		return fmt.Errorf("failed to check installed packages: %v", err)
	}
	log.Printf("Installed packages:\n%s", string(pipOutput))

	// Use python to run yt-dlp as a module
	cmd := exec.Command(pythonPath, "-m", "yt_dlp", "-x", "--audio-format", "mp3", "-o", "downloaded_audio.%(ext)s", url)

	// Log the full command being executed
	log.Printf("Executing command: %v", cmd.String())

	// Capture the output and error streams
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("yt-dlp command failed with error: %v", err)
		log.Printf("yt-dlp output: %s", string(output))
		return fmt.Errorf("failed to download audio: %v", err)
	}

	log.Printf("yt-dlp command output: %s", string(output))
	return nil
}

func transcribeAudio(filename string) (string, error) {
	cmd := exec.Command("python3", "transcribe.py", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Transcription failed with error: %v", err)
		log.Printf("Transcription output: %s", string(output))
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func summarizeText(text string) (string, error) {
	cmd := exec.Command("python3", "summarize.py", text)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Summarization failed with error: %v", err)
		log.Printf("Summarization output: %s", string(output))
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func processVideo(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	sendEvent := func(event string, data string) {
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
		flusher.Flush()
	}

	log.Printf("Received request with method: %s", r.Method)
	log.Printf("Request headers: %v", r.Header)

	var requestData struct {
		URL string `json:"url"`
	}

	if r.Method == http.MethodGet {
		// Handle GET request
		requestData.URL = r.URL.Query().Get("url")
	} else if r.Method == http.MethodPost {
		// Handle POST request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			sendEvent("error", fmt.Sprintf("Error reading request body: %s", err.Error()))
			return
		}
		log.Printf("Request body: %s", string(body))

		err = json.Unmarshal(body, &requestData)
		if err != nil {
			log.Printf("Error unmarshaling JSON: %v", err)
			sendEvent("error", fmt.Sprintf("Invalid JSON: %s", err.Error()))
			return
		}
	} else {
		log.Printf("Unsupported HTTP method: %s", r.Method)
		sendEvent("error", "Unsupported HTTP method")
		return
	}

	if requestData.URL == "" {
		log.Printf("URL not provided in request")
		sendEvent("error", "URL not provided")
		return
	}

	log.Printf("Received URL: %s", requestData.URL)

	sendEvent("status", "Downloading audio...")
	err := downloadAudio(requestData.URL)
	if err != nil {
		sendEvent("error", fmt.Sprintf("Failed to download audio: %s", err.Error()))
		return
	}

	sendEvent("status", "Transcribing audio...")
	transcription, err := transcribeAudio("downloaded_audio.mp3")
	if err != nil {
		sendEvent("error", fmt.Sprintf("Failed to transcribe audio: %s", err.Error()))
		return
	}
	log.Printf("Transcription: %s", transcription)

	sendEvent("status", "Summarizing transcription...")
	summary, err := summarizeText(transcription)
	if err != nil {
		sendEvent("error", fmt.Sprintf("Failed to summarize transcription: %s", err.Error()))
		return
	}
	log.Printf("Summary: %s", summary)

	sendEvent("status", "Cleaning up...")
	err = os.Remove("downloaded_audio.mp3")
	if err != nil {
		log.Printf("Failed to remove audio file: %v", err)
	}

	sendEvent("result", summary)
	sendEvent("status", "Completed")
}
