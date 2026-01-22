package openai

import (
	"encoding/json"
	"testing"
)

func TestCompletionRequest_UnmarshalJSON_String(t *testing.T) {
	jsonStr := `{
		"model": "gpt-3.5-turbo-instruct",
		"prompt": "Say this is a test",
		"max_tokens": 7,
		"temperature": 0
	}`

	var req CompletionRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.Model != "gpt-3.5-turbo-instruct" {
		t.Errorf("Expected model 'gpt-3.5-turbo-instruct', got '%s'", req.Model)
	}

	if req.Prompt == nil {
		t.Fatal("Prompt is nil")
	}

	promptStr, ok := (*req.Prompt).(string)
	if !ok {
		t.Fatalf("Expected string prompt, got %T", *req.Prompt)
	}

	if promptStr != "Say this is a test" {
		t.Errorf("Expected prompt 'Say this is a test', got '%s'", promptStr)
	}

	if req.MaxTokens == nil || *req.MaxTokens != 7 {
		t.Error("Expected max_tokens to be 7")
	}
}

func TestCompletionRequest_UnmarshalJSON_Array(t *testing.T) {
	jsonStr := `{
		"model": "gpt-3.5-turbo-instruct",
		"prompt": ["Hello", " ", "World"],
		"max_tokens": 100
	}`

	var req CompletionRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.Prompt == nil {
		t.Fatal("Prompt is nil")
	}

	promptArr, ok := (*req.Prompt).([]interface{})
	if !ok {
		t.Fatalf("Expected array prompt, got %T", *req.Prompt)
	}

	if len(promptArr) != 3 {
		t.Errorf("Expected 3 items, got %d", len(promptArr))
	}

	if promptArr[0].(string) != "Hello" || promptArr[1].(string) != " " || promptArr[2].(string) != "World" {
		t.Error("Prompt array values don't match")
	}
}

func TestCompletionRequest_MarshalJSON_String(t *testing.T) {
	maxTokens := 50
	temp := 0.7
	var prompt any = "Test prompt"
	req := CompletionRequest{
		Model:       "gpt-3.5-turbo-instruct",
		Prompt:      &prompt,
		MaxTokens:   &maxTokens,
		Temperature: &temp,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "gpt-3.5-turbo-instruct" {
		t.Errorf("Expected model 'gpt-3.5-turbo-instruct', got '%v'", result["model"])
	}

	if result["prompt"] != "Test prompt" {
		t.Errorf("Expected prompt 'Test prompt', got '%v'", result["prompt"])
	}
}

func TestCompletionRequest_MarshalJSON_Array(t *testing.T) {
	promptArr := []string{"First", "Second"}
	var promptAny any = promptArr
	req := CompletionRequest{
		Model:  "gpt-3.5-turbo-instruct",
		Prompt: &promptAny,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	resultPrompt, ok := result["prompt"].([]interface{})
	if !ok || len(resultPrompt) != 2 {
		t.Error("Expected prompt to be an array with 2 elements")
	}
}

func TestCompletionRequest_Stream(t *testing.T) {
	jsonStr := `{
		"model": "gpt-3.5-turbo-instruct",
		"prompt": "Test",
		"stream": true
	}`

	var req CompletionRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !req.Stream {
		t.Error("Expected stream to be true")
	}
}

func TestCompletionRequest_LogProbs(t *testing.T) {
	jsonStr := `{
		"model": "gpt-3.5-turbo-instruct",
		"prompt": "Test",
		"logprobs": true
	}`

	var req CompletionRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.LogProbs == nil || !*req.LogProbs {
		t.Error("Expected logprobs to be true")
	}
}
