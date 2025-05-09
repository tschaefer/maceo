package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type AnalyzeRequest struct {
	Text           string   `json:"text"`
	Language       string   `json:"language"`
	Entities       []string `json:"entities"`
	ScoreThreshold float64  `json:"score_threshold"`
}

type AnalyzeResponse struct {
	Start      uint64  `json:"start"`
	End        uint64  `json:"end"`
	Score      float64 `json:"score"`
	EntityType string  `json:"entity_type"`
}

type AnonymizeRequest struct {
	Text            string            `json:"text"`
	AnalyzerResults []AnalyzeResponse `json:"analyzer_results"`
}

type AnonymizeResponse struct {
	Text string `json:"text"`
}

type FunctionConfig struct {
	Upstreams      map[string]string `json:"upstreams"`
	Entities       []string          `json:"entities"`
	Language       string            `json:"language"`
	ScoreThreshold float64           `json:"score_threshold"`
}

const ConfigFilePath = "/var/openfaas/secrets/maceo"

var Config FunctionConfig

func Handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/health") {
		checkHealth(w, r)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	input, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(input) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	config := readConfig()
	if config == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	analyzeResponse := performAnalysis(input)
	if analyzeResponse == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	anonymizeResponse := performAnonymization(input, *analyzeResponse)
	if anonymizeResponse == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(anonymizeResponse.Text))
	if err != nil {
		log.Printf("%s", err)
	}
}

func checkHealth(w http.ResponseWriter, r *http.Request) {
	config := readConfig()
	if config == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	endpoints := []string{
		"analyze",
		"anonymize",
	}
	for _, endpoint := range endpoints {
		baseURL := Config.Upstreams[endpoint]
		health := fmt.Sprintf("%s/health", baseURL)
		_, err := http.Get(health)
		if err != nil {
			log.Printf("Error checking health: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func readConfig() *FunctionConfig {
	Config.Upstreams = map[string]string{
		"analyze":   "http://10.62.0.1:5001",
		"anonymize": "http://10.62.0.1:5002",
	}
	Config.Entities = nil
	Config.Language = "en"
	Config.ScoreThreshold = 0.0

	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		return &Config
	}

	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		log.Printf("Error reading config: %s", err)
		return nil
	}

	err = json.Unmarshal(data, &Config)
	if err != nil {
		log.Printf("Error reading config: %s", err)
		return nil
	}

	return &Config
}

func performAnalysis(input []byte) *[]AnalyzeResponse {
	analyze := AnalyzeRequest{
		Text:           string(input),
		Language:       Config.Language,
		Entities:       Config.Entities,
		ScoreThreshold: Config.ScoreThreshold,
	}
	jsonData, err := json.Marshal(analyze)
	if err != nil {
		log.Printf("Error performing analysis: %s", err)
		return nil
	}

	url := fmt.Sprintf("%s/analyze", Config.Upstreams["analyze"])
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error performing analysis: %s", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error performing analysis: %s", err)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error performing analysis: %s", body)
		return nil
	}

	var analyzeResponse []AnalyzeResponse
	if err := json.Unmarshal(body, &analyzeResponse); err != nil {
		log.Printf("Error performing analysis: %s", err)
		return nil
	}
	return &analyzeResponse
}

func performAnonymization(input []byte, analyzeResponse []AnalyzeResponse) *AnonymizeResponse {
	anonymize := AnonymizeRequest{
		Text:            string(input),
		AnalyzerResults: analyzeResponse,
	}
	jsonData, err := json.Marshal(anonymize)
	if err != nil {
		return nil
	}

	url := fmt.Sprintf("%s/anonymize", Config.Upstreams["anonymize"])
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error performing anonymization: %s", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error performing anonymization: %s", err)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error performing anonymization: %s", body)
		return nil
	}

	var anonymizeResponse AnonymizeResponse
	if err := json.Unmarshal(body, &anonymizeResponse); err != nil {
		log.Printf("Error performing anonymization: %s", err)
		return nil
	}
	return &anonymizeResponse
}
