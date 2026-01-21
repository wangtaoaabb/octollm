package openai

import (
	"encoding/json"
	"testing"
)

func TestRerankRequest_UnmarshalJSON(t *testing.T) {
	jsonStr := `{
		"query": "What is the capital of France?",
		"model": "bge-reranker-v2-m3",
		"documents": ["Paris is the capital of France.", "London is the capital of England."],
		"return_documents": true,
		"top_n": 5
	}`

	var req RerankRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.Model != "bge-reranker-v2-m3" {
		t.Errorf("Expected model 'bge-reranker-v2-m3', got '%s'", req.Model)
	}

	if req.Query != "What is the capital of France?" {
		t.Errorf("Expected query 'What is the capital of France?', got '%s'", req.Query)
	}

	if len(req.Documents) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(req.Documents))
	}

	if req.Documents[0] != "Paris is the capital of France." {
		t.Errorf("Expected first document 'Paris is the capital of France.', got '%s'", req.Documents[0])
	}

	if req.ReturnDocuments == nil || *req.ReturnDocuments != true {
		t.Error("Expected return_documents to be true")
	}

	if req.TopN == nil || *req.TopN != 5 {
		t.Error("Expected top_n to be 5")
	}
}

func TestRerankRequest_MarshalJSON(t *testing.T) {
	returnDocs := true
	topN := 3
	req := RerankRequest{
		Query:           "test query",
		Model:           "bge-reranker-v2-m3",
		Documents:       []string{"doc1", "doc2", "doc3"},
		ReturnDocuments: &returnDocs,
		TopN:            &topN,
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

	if result["model"] != "bge-reranker-v2-m3" {
		t.Errorf("Expected model 'bge-reranker-v2-m3', got '%v'", result["model"])
	}

	if result["query"] != "test query" {
		t.Errorf("Expected query 'test query', got '%v'", result["query"])
	}

	docs, ok := result["documents"].([]interface{})
	if !ok || len(docs) != 3 {
		t.Errorf("Expected 3 documents")
	}

	if result["return_documents"] != true {
		t.Error("Expected return_documents to be true")
	}

	topNResult, ok := result["top_n"].(float64)
	if !ok || int(topNResult) != 3 {
		t.Error("Expected top_n to be 3")
	}
}

func TestRawRerankResponse_Xinference(t *testing.T) {
	// Test xinference format (with meta)
	jsonStr := `{
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

	var resp RawRerankResponse
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(resp.Results))
	}

	if resp.Results[0].RelevanceScore != 0.999 {
		t.Errorf("Expected relevance_score 0.999, got %f", resp.Results[0].RelevanceScore)
	}

	if resp.Meta == nil {
		t.Fatal("Expected meta to be present")
	}

	if resp.Meta.Tokens.InputTokens != 35 {
		t.Errorf("Expected input_tokens 35, got %d", resp.Meta.Tokens.InputTokens)
	}
}

func TestRawRerankResponse_VLLM(t *testing.T) {
	// Test vllm format (with usage)
	jsonStr := `{
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

	var resp RawRerankResponse
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(resp.Results))
	}

	if resp.Usage == nil {
		t.Fatal("Expected usage to be present")
	}

	if resp.Usage.TotalTokens != 9 {
		t.Errorf("Expected total_tokens 9, got %d", resp.Usage.TotalTokens)
	}
}
