package engines

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/infinigence/octollm/pkg/octollm"
	"github.com/infinigence/octollm/pkg/types/openai"
)

func TestRerankNormalizer_Xinference(t *testing.T) {
	// Mock xinference response
	xinferenceResp := `{
		"results": [
			{
				"index": 0,
				"relevance_score": 0.999,
				"document": {"text": "Paris is the capital of France."}
			}
		],
		"meta": {
			"tokens": {
				"input_tokens": 35,
				"output_tokens": 0
			}
		}
	}`

	// Create mock engine that returns xinference format
	mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
		body := octollm.NewBodyFromBytes([]byte(xinferenceResp), &octollm.JSONParser[openai.RawRerankResponse]{})
		return &octollm.Response{
			StatusCode: 200,
			Body:       body,
		}, nil
	})

	// Wrap with normalizer
	normalizer := NewRerankNormalizer(mockEngine)

	// Create test request
	reqBody := `{"query": "test", "model": "bge-reranker", "documents": ["doc1"]}`
	req := &octollm.Request{
		Format: octollm.APIFormatRerank,
		Body:   octollm.NewBodyFromBytes([]byte(reqBody), &octollm.JSONParser[openai.RerankRequest]{}),
	}

	// Process
	resp, err := normalizer.Process(req)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Parse normalized response
	parsed, err := resp.Body.Parsed()
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	rerankResp, ok := parsed.(*openai.RerankResponse)
	if !ok {
		t.Fatalf("Expected *openai.RerankResponse, got %T", parsed)
	}

	// Verify normalized format
	if rerankResp.Model != "bge-reranker" {
		t.Errorf("Expected model 'bge-reranker', got '%s'", rerankResp.Model)
	}

	if rerankResp.Usage.PromptTokens != 35 {
		t.Errorf("Expected prompt_tokens 35, got %d", rerankResp.Usage.PromptTokens)
	}

	if rerankResp.Usage.TotalTokens != 35 {
		t.Errorf("Expected total_tokens 35, got %d", rerankResp.Usage.TotalTokens)
	}

	if len(rerankResp.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(rerankResp.Results))
	}

	// Verify serialized format doesn't contain meta
	bytes, err := resp.Body.Bytes()
	if err != nil {
		t.Fatalf("Failed to get bytes: %v", err)
	}

	var jsonResp map[string]interface{}
	if err := json.Unmarshal(bytes, &jsonResp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if _, hasMeta := jsonResp["meta"]; hasMeta {
		t.Error("Normalized response should not contain 'meta' field")
	}

	if _, hasUsage := jsonResp["usage"]; !hasUsage {
		t.Error("Normalized response should contain 'usage' field")
	}
}

func TestRerankNormalizer_VLLM(t *testing.T) {
	// Mock vllm response
	vllmResp := `{
		"results": [
			{
				"index": 0,
				"relevance_score": 0.999,
				"document": {"text": "Paris is the capital of France."}
			}
		],
		"usage": {
			"total_tokens": 9
		}
	}`

	// Create mock engine
	mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
		body := octollm.NewBodyFromBytes([]byte(vllmResp), &octollm.JSONParser[openai.RawRerankResponse]{})
		return &octollm.Response{
			StatusCode: 200,
			Body:       body,
		}, nil
	})

	normalizer := NewRerankNormalizer(mockEngine)

	reqBody := `{"query": "test", "model": "bge-reranker", "documents": ["doc1"]}`
	req := &octollm.Request{
		Format: octollm.APIFormatRerank,
		Body:   octollm.NewBodyFromBytes([]byte(reqBody), &octollm.JSONParser[openai.RerankRequest]{}),
	}

	resp, err := normalizer.Process(req)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	parsed, err := resp.Body.Parsed()
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	rerankResp, ok := parsed.(*openai.RerankResponse)
	if !ok {
		t.Fatalf("Expected *openai.RerankResponse, got %T", parsed)
	}

	// Verify VLLM format normalization
	if rerankResp.Usage.PromptTokens != 9 {
		t.Errorf("Expected prompt_tokens 9, got %d", rerankResp.Usage.PromptTokens)
	}

	if rerankResp.Usage.TotalTokens != 9 {
		t.Errorf("Expected total_tokens 9, got %d", rerankResp.Usage.TotalTokens)
	}
}

func TestRerankNormalizer_NonRerank(t *testing.T) {
	// Test that non-rerank requests pass through unchanged
	chatResp := `{"choices": [{"message": {"content": "hello"}}]}`

	mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
		return &octollm.Response{
			StatusCode: 200,
			Body:       octollm.NewBodyFromBytes([]byte(chatResp), nil),
		}, nil
	})

	normalizer := NewRerankNormalizer(mockEngine)

	req := &octollm.Request{
		Format: octollm.APIFormatChatCompletions,
		Body:   octollm.NewBodyFromBytes([]byte(`{"messages":[]}`), nil),
	}

	resp, err := normalizer.Process(req)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify response is unchanged
	reader, err := resp.Body.Reader()
	if err != nil {
		t.Fatalf("Failed to get reader: %v", err)
	}

	respBytes, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if string(respBytes) != chatResp {
		t.Errorf("Non-rerank response should be unchanged")
	}
}

func TestRerankNormalizer_Stream(t *testing.T) {
	// Test that streaming responses pass through unchanged
	mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
		ch := make(chan *octollm.StreamChunk, 1)
		ch <- &octollm.StreamChunk{
			Body: octollm.NewBodyFromBytes([]byte("data"), nil),
		}
		close(ch)

		return &octollm.Response{
			StatusCode: 200,
			Stream:     octollm.NewStreamChan(ch, func() {}),
		}, nil
	})

	normalizer := NewRerankNormalizer(mockEngine)

	req := &octollm.Request{
		Format: octollm.APIFormatRerank,
		Body:   octollm.NewBodyFromBytes([]byte(`{"query":"test","model":"m","documents":[]}`), nil),
	}

	resp, err := normalizer.Process(req)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify stream is present and unchanged
	if resp.Stream == nil {
		t.Error("Stream should be present")
	}
}
