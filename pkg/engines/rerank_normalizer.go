package engines

import (
	"encoding/json"
	"fmt"

	"github.com/infinigence/octollm/pkg/octollm"
	"github.com/infinigence/octollm/pkg/types/openai"
	"github.com/sirupsen/logrus"
)

// RerankNormalizerEngine wraps an engine to normalize rerank responses
// It converts RawRerankResponse (xinference/vllm formats) to RerankResponse (unified format)
type RerankNormalizerEngine struct {
	Next octollm.Engine
}

var _ octollm.Engine = (*RerankNormalizerEngine)(nil)

func NewRerankNormalizer(next octollm.Engine) *RerankNormalizerEngine {
	return &RerankNormalizerEngine{Next: next}
}

func (e *RerankNormalizerEngine) Process(req *octollm.Request) (*octollm.Response, error) {
	// Only process rerank requests
	if req.Format != octollm.APIFormatRerank {
		return e.Next.Process(req)
	}

	resp, err := e.Next.Process(req)
	if err != nil {
		return resp, err
	}

	// Only normalize non-streaming responses
	if resp.Stream != nil {
		return resp, nil
	}

	if resp.Body == nil {
		return resp, nil
	}

	// Parse the response as RawRerankResponse
	parsed, err := resp.Body.Parsed()
	if err != nil {
		return nil, fmt.Errorf("failed to parse rerank response: %w", err)
	}

	rawResp, ok := parsed.(*openai.RawRerankResponse)
	if !ok {
		// Not a rerank response, return as-is
		return resp, nil
	}

	// Convert to normalized RerankResponse
	normalizedResp := &openai.RerankResponse{
		Results: rawResp.Results,
		Model:   "", // Will be set from request
		Usage:   openai.RerankUsage{},
	}

	// Extract model name from request
	reqParsed, err := req.Body.Parsed()
	if err == nil {
		if rerankReq, ok := reqParsed.(*openai.RerankRequest); ok {
			normalizedResp.Model = rerankReq.Model
		}
	}

	// Normalize usage from different formats
	// Priority: xinference (meta) > vllm (usage)
	if rawResp.Meta != nil && rawResp.Meta.Tokens.InputTokens > 0 {
		// Xinference format
		normalizedResp.Usage.PromptTokens = rawResp.Meta.Tokens.InputTokens
		normalizedResp.Usage.TotalTokens = rawResp.Meta.Tokens.InputTokens
		logrus.Debugf("[RerankNormalizer] Using xinference format: input_tokens=%d", rawResp.Meta.Tokens.InputTokens)
	} else if rawResp.Usage != nil && rawResp.Usage.TotalTokens > 0 {
		// VLLM format
		normalizedResp.Usage.PromptTokens = rawResp.Usage.TotalTokens
		normalizedResp.Usage.TotalTokens = rawResp.Usage.TotalTokens
		logrus.Debugf("[RerankNormalizer] Using vllm format: total_tokens=%d", rawResp.Usage.TotalTokens)
	}

	// Serialize normalized response
	normalizedBytes, err := json.Marshal(normalizedResp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal normalized rerank response: %w", err)
	}

	// Update response body
	resp.Body.SetBytes(normalizedBytes)
	resp.Body.SetParser(&octollm.JSONParser[openai.RerankResponse]{})

	return resp, nil
}
