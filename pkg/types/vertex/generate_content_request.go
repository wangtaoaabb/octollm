package vertex

// GenerateContentRequest represents the request body for Vertex AI generateContent API
type GenerateContentRequest struct {
	CachedContent     string             `json:"cachedContent,omitempty"`
	Contents          []Content          `json:"contents"`
	SystemInstruction *Content           `json:"systemInstruction,omitempty"`
	Tools             []Tool             `json:"tools,omitempty"`
	ToolConfig        *ToolConfig        `json:"toolConfig,omitempty"`
	SafetySettings    []SafetySetting    `json:"safetySettings,omitempty"`
	GenerationConfig  *GenerationConfig  `json:"generationConfig,omitempty"`
	Labels            map[string]string  `json:"labels,omitempty"`
}

// Content represents a message with role and parts
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

// Part represents a part of content (text, inline data, file data, etc.)
type Part struct {
	Text             string            `json:"text,omitempty"`
	InlineData       *Blob             `json:"inlineData,omitempty"`
	FileData         *FileData         `json:"fileData,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	VideoMetadata    *VideoMetadata    `json:"videoMetadata,omitempty"`
	MediaResolution  string            `json:"mediaResolution,omitempty"`
}

// Blob represents inline binary data
type Blob struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

// FileData represents a file URI
type FileData struct {
	MimeType string `json:"mimeType"`
	FileURI  string `json:"fileUri"`
}

// FunctionCall represents a function call from the model
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

// FunctionResponse represents a function response to the model
type FunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// VideoMetadata represents video-specific metadata
type VideoMetadata struct {
	StartOffset *Duration `json:"startOffset,omitempty"`
	EndOffset   *Duration `json:"endOffset,omitempty"`
	FPS         float64   `json:"fps,omitempty"`
}

// Duration represents a time duration
type Duration struct {
	Seconds int64 `json:"seconds,omitempty"`
	Nanos   int   `json:"nanos,omitempty"`
}

// Tool represents a tool (function declaration)
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// FunctionDeclaration represents a function that can be called
type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"` // OpenAPI Object Schema
}

// ToolConfig represents tool configuration
type ToolConfig struct {
	FunctionCallingConfig *FunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

// FunctionCallingConfig configures function calling behavior
type FunctionCallingConfig struct {
	Mode                  string   `json:"mode,omitempty"` // AUTO, ANY, NONE
	AllowedFunctionNames  []string `json:"allowedFunctionNames,omitempty"`
}

// SafetySetting represents a safety setting
type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
	Method    string `json:"method,omitempty"`
}

// GenerationConfig represents generation configuration
type GenerationConfig struct {
	Temperature      *float64          `json:"temperature,omitempty"`
	TopP             *float64          `json:"topP,omitempty"`
	TopK             *int              `json:"topK,omitempty"`
	CandidateCount   *int              `json:"candidateCount,omitempty"`
	MaxOutputTokens  *int              `json:"maxOutputTokens,omitempty"`
	PresencePenalty  *float64          `json:"presencePenalty,omitempty"`
	FrequencyPenalty *float64          `json:"frequencyPenalty,omitempty"`
	StopSequences    []string          `json:"stopSequences,omitempty"`
	ResponseMimeType string            `json:"responseMimeType,omitempty"`
	ResponseSchema   map[string]interface{} `json:"responseSchema,omitempty"`
	Seed             *int              `json:"seed,omitempty"`
	ResponseLogprobs *bool             `json:"responseLogprobs,omitempty"`
	Logprobs         *int              `json:"logprobs,omitempty"`
	AudioTimestamp   *bool             `json:"audioTimestamp,omitempty"`
	ThinkingConfig   *ThinkingConfig   `json:"thinkingConfig,omitempty"`
	MediaResolution  string            `json:"mediaResolution,omitempty"`
}

// ThinkingConfig represents thinking configuration for Gemini 2.5+
type ThinkingConfig struct {
	ThinkingBudget *int   `json:"thinkingBudget,omitempty"`
	ThinkingLevel  string `json:"thinkingLevel,omitempty"` // LOW, HIGH
}
