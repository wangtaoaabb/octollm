package openai

import (
	"encoding/json"
	"testing"
)

func TestMessage_UnmarshalJSON_String(t *testing.T) {
	jsonData := `{
		"role": "user",
		"content": "Hello, world!"
	}`

	var msg Message
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if msg.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", msg.Role)
	}

	contentStr, ok := msg.Content.(MessageContentString)
	if !ok {
		t.Fatalf("Expected MessageContentString, got %T", msg.Content)
	}

	if string(contentStr) != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%s'", contentStr)
	}
}

func TestMessage_UnmarshalJSON_Array(t *testing.T) {
	jsonData := `{
		"role": "user",
		"content": [
			{"type": "text", "text": "Hello"},
			{"type": "text", "text": " world!"}
		]
	}`

	var msg Message
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if msg.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", msg.Role)
	}

	contentArr, ok := msg.Content.(MessageContentArray)
	if !ok {
		t.Fatalf("Expected MessageContentArray, got %T", msg.Content)
	}

	if len(contentArr) != 2 {
		t.Errorf("Expected 2 items, got %d", len(contentArr))
	}

	if contentArr[0].Text != "Hello" {
		t.Errorf("Expected first item text 'Hello', got '%s'", contentArr[0].Text)
	}
}

func TestMessage_MarshalJSON_String(t *testing.T) {
	msg := Message{
		Role:    "assistant",
		Content: MessageContentString("Hello, world!"),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{"role":"assistant","content":"Hello, world!"}`
	var expectedMap, actualMap map[string]interface{}
	json.Unmarshal([]byte(expected), &expectedMap)
	json.Unmarshal(data, &actualMap)

	if expectedMap["role"] != actualMap["role"] {
		t.Errorf("Role mismatch")
	}
	if expectedMap["content"] != actualMap["content"] {
		t.Errorf("Content mismatch: expected '%v', got '%v'", expectedMap["content"], actualMap["content"])
	}
}

func TestMessage_MarshalJSON_Array(t *testing.T) {
	msg := Message{
		Role: "user",
		Content: MessageContentArray([]*MessageContentItem{
			{Type: "text", Text: "Hello"},
			{Type: "text", Text: " world!"},
		}),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if result["role"] != "user" {
		t.Errorf("Role mismatch")
	}

	content, ok := result["content"].([]interface{})
	if !ok {
		t.Fatalf("Expected content to be array, got %T", result["content"])
	}

	if len(content) != 2 {
		t.Errorf("Expected 2 items, got %d", len(content))
	}
}

func TestApiChatCompletionsRequest_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"messages": [
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Hello!"}
		],
		"max_tokens": 100,
		"temperature": 0.7
	}`

	var req ChatCompletionRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if req.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", req.Model)
	}

	if len(req.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(req.Messages))
	}

	if req.Messages[0].Role != "system" {
		t.Errorf("Expected first message role 'system', got '%s'", req.Messages[0].Role)
	}

	content := req.Messages[0].Content.(MessageContentString)
	if string(content) != "You are a helpful assistant." {
		t.Errorf("Expected content 'You are a helpful assistant.', got '%s'", content)
	}
}
