package openai

// RerankRequest represents the request structure for rerank API
type RerankRequest struct {
	Query           string   `json:"query" binding:"required"`
	Model           string   `json:"model" binding:"required"`
	Documents       []string `json:"documents" binding:"required"`
	ReturnDocuments *bool    `json:"return_documents,omitempty"`
	ReturnLen       *bool    `json:"return_len,omitempty"`
	TopN            *int     `json:"top_n,omitempty"`
}
