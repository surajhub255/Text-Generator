package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"errors"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)
var apiURL string = "https://api.openai.com/v1/chat/completions"

var apiKey string = "Your apikey"


type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}


type Response struct {
	Id                string `json:"id"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func generateText(message string) (string, error) {
	messages := make([]Message, 0)

	messages = append(messages, Message{Role: "user", Content: message})

	body := Request{Model: "gpt-3.5-turbo", Messages: messages}

	bodyBytes, _ := json.Marshal(body)

	r, _ := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return "", errors.New("could not generate text")
	}

	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", errors.New("could not generate text")
	}

	output := Response{}

	err = json.Unmarshal(responseBody, &output)
	if err != nil {
		return "", errors.New("could not generate text")
	}

	if len(output.Choices) == 0 {
		return "", errors.New("could not generate text")
	}

	return output.Choices[0].Message.Content, nil
}

func handleGenerateText(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	output, err := generateText(requestData["message"])
	if err != nil {
		http.Error(w, "Error generating text", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"result": output}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func main() {
	r := mux.NewRouter()
	r.HandleFunc("/generate-text", handleGenerateText).Methods("POST")

	http.Handle("/", r)
	// http.ListenAndServe(":3001", nil)

	port := 8080
	fmt.Printf("Server running on :%d...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port),
		handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTION"}),
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		)(r))
}