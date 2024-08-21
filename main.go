package main

import (
	"encoding/json"
	"fmt"
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
	cmd := exec.Command("yt-dlp", "-x", "--audio-format", "mp3", "-o", "downloaded_audio.%(ext)s", url)
	err := cmd.Run()
	return err
}

func transcribeAudio(filename string) (string, error) {
	cmd := exec.Command("python3", "transcribe.py", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func summarizeText(text string) (string, error) {
	cmd := exec.Command("python3", "summarize.py", text)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func processVideo(w http.ResponseWriter, r *http.Request) {
    log.Println("Received request to process video")
    
    var requestData struct {
        URL string `json:"url"`
    }

    err := json.NewDecoder(r.Body).Decode(&requestData)
    if err != nil || requestData.URL == "" {
        log.Printf("Invalid request data: %v", err)
        http.Error(w, `{"error":"URL not provided or invalid JSON"}`, http.StatusBadRequest)
        return
    }

    log.Printf("Downloading audio from URL: %s", requestData.URL)
    err = downloadAudio(requestData.URL)
    if err != nil {
        log.Printf("Failed to download audio: %v", err)
        http.Error(w, fmt.Sprintf(`{"error":"Failed to download audio: %s"}`, err.Error()), http.StatusInternalServerError)
        return
    }

    log.Println("Transcribing audio...")
    transcription, err := transcribeAudio("downloaded_audio.mp3")
    if err != nil {
        log.Printf("Failed to transcribe audio: %v", err)
        http.Error(w, fmt.Sprintf(`{"error":"Failed to transcribe audio: %s"}`, err.Error()), http.StatusInternalServerError)
        return
    }

    log.Println("Summarizing transcription...")
    summary, err := summarizeText(transcription)
    if err != nil {
        log.Printf("Failed to summarize transcription: %v", err)
        http.Error(w, fmt.Sprintf(`{"error":"Failed to summarize transcription: %s"}`, err.Error()), http.StatusInternalServerError)
        return
    }

    log.Println("Cleaning up downloaded audio file...")
    err = os.Remove("downloaded_audio.mp3")
    if err != nil {
        log.Printf("Failed to remove audio file: %v", err)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"summary": summary})
    log.Println("Successfully processed video")
}
