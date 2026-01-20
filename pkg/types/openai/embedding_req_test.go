package openai

import (
	"encoding/json"
	"testing"
)

func TestEmbeddingRequest_UnmarshalJSON_String(t *testing.T) {
	jsonStr := `{
		"input": "The quick brown fox jumps over the lazy dog",
		"model": "text-embedding-ada-002"
	}`

	var req EmbeddingRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.Model != "text-embedding-ada-002" {
		t.Errorf("Expected model 'text-embedding-ada-002', got '%s'", req.Model)
	}

	if req.Input.IsArray() {
		t.Error("Expected Input to be a string, but it's an array")
	}

	if str, ok := req.Input.(RequestContentString); ok {
		if string(str) != "The quick brown fox jumps over the lazy dog" {
			t.Errorf("Expected input 'The quick brown fox jumps over the lazy dog', got '%s'", string(str))
		}
	} else {
		t.Error("Input is not of type RequestContentString")
	}
}

func TestEmbeddingRequest_UnmarshalJSON_Array(t *testing.T) {
	jsonStr := `{
		"input": ["Hello world", "How are you?", "I am fine"],
		"model": "text-embedding-ada-002"
	}`

	var req EmbeddingRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.Model != "text-embedding-ada-002" {
		t.Errorf("Expected model 'text-embedding-ada-002', got '%s'", req.Model)
	}

	if !req.Input.IsArray() {
		t.Error("Expected Input to be an array, but it's not")
	}

	if arr, ok := req.Input.(RequestContentStringArray); ok {
		if len(arr) != 3 {
			t.Errorf("Expected 3 items in array, got %d", len(arr))
		}
		if arr[0] != "Hello world" {
			t.Errorf("Expected first item 'Hello world', got '%s'", arr[0])
		}
		if arr[1] != "How are you?" {
			t.Errorf("Expected second item 'How are you?', got '%s'", arr[1])
		}
		if arr[2] != "I am fine" {
			t.Errorf("Expected third item 'I am fine', got '%s'", arr[2])
		}
	} else {
		t.Error("Input is not of type RequestContentStringArray")
	}
}

func TestEmbeddingRequest_MarshalJSON_String(t *testing.T) {
	req := EmbeddingRequest{
		Input: RequestContentString("Test input"),
		Model: "text-embedding-ada-002",
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

	if result["model"] != "text-embedding-ada-002" {
		t.Errorf("Expected model 'text-embedding-ada-002', got '%v'", result["model"])
	}

	if result["input"] != "Test input" {
		t.Errorf("Expected input 'Test input', got '%v'", result["input"])
	}
}

func TestEmbeddingRequest_MarshalJSON_Array(t *testing.T) {
	req := EmbeddingRequest{
		Input: RequestContentStringArray{"First", "Second", "Third"},
		Model: "text-embedding-ada-002",
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

	if result["model"] != "text-embedding-ada-002" {
		t.Errorf("Expected model 'text-embedding-ada-002', got '%v'", result["model"])
	}

	inputArr, ok := result["input"].([]interface{})
	if !ok {
		t.Fatalf("Expected input to be an array, got %T", result["input"])
	}

	if len(inputArr) != 3 {
		t.Errorf("Expected 3 items in array, got %d", len(inputArr))
	}

	if inputArr[0] != "First" || inputArr[1] != "Second" || inputArr[2] != "Third" {
		t.Errorf("Array values don't match expected values")
	}
}

func TestRequestContentString_GetDataLength(t *testing.T) {
	input := RequestContentString("Hello")
	if input.GetDataLength() != 5 {
		t.Errorf("Expected length 5, got %d", input.GetDataLength())
	}
}

func TestRequestContentStringArray_GetDataLength(t *testing.T) {
	input := RequestContentStringArray{"Hello", "World"}
	expected := len("Hello") + len("World")
	if input.GetDataLength() != expected {
		t.Errorf("Expected length %d, got %d", expected, input.GetDataLength())
	}
}
