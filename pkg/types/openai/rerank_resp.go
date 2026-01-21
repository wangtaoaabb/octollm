package openai

// RerankResponse represents the response structure for rerank API
type RerankResponse struct {
	Results []RerankResult `json:"results"`
	Model   string         `json:"model"`
	Usage   RerankUsage    `json:"usage"`
}

// RerankUsage represents token usage in the rerank response
type RerankUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// RerankResult represents a single rerank result
type RerankResult struct {
	Index          int        `json:"index"`
	RelevanceScore float64    `json:"relevance_score"`
	Document       RerankText `json:"document,omitempty"`
}

// RerankText represents the document text in a rerank result
type RerankText struct {
	Text string `json:"text"`
}

// RawRerankResponse is used for parsing different backend response formats
// Supports both xinference (with meta) and vllm (with usage) formats
type RawRerankResponse struct {
	Results []RerankResult `json:"results"` // Common field
	Meta    *RerankMeta    `json:"meta"`    // xinference format
	Usage   *RerankUsage   `json:"usage"`   // vllm format
}

// RerankMeta represents the metadata in xinference response
type RerankMeta struct {
	Tokens RerankTokens `json:"tokens"`
}

// RerankTokens represents token information in xinference response
type RerankTokens struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
