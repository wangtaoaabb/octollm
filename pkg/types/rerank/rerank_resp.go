package rerank

// RerankResponse represents the standard rerank response
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
