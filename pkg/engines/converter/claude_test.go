package converter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infinigence/octollm/pkg/octollm"
	"github.com/infinigence/octollm/pkg/types/anthropic"
	"github.com/infinigence/octollm/pkg/types/openai"
)

func TestApiChatCompletionsToApiMessages_convertRequestBody_SimpleText(t *testing.T) {
	converter := NewChatCompletionToClaudeMessages(nil)
	ctx := context.Background()

	// Create Anthropic request
	text := "Hello, Claude!"
	anthropicReq := &anthropic.ClaudeMessagesRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Messages: []*anthropic.MessageParam{
			{
				Role: "user",
				Content: []anthropic.MessageContent{
					&anthropic.MessageContentBlock{
						Type: "text",
						Text: &text,
					},
				},
			},
		},
	}

	// Create UnifiedBody
	srcBody := octollm.NewBodyFromBytes([]byte{}, &octollm.JSONParser[anthropic.ClaudeMessagesRequest]{})
	srcBody.SetParsed(anthropicReq)

	// Convert
	dstBody, err := converter.convertRequestBody(ctx, srcBody)
	require.NoError(t, err)

	// Parse result
	parsed, err := dstBody.Parsed()
	require.NoError(t, err)

	openaiReq, ok := parsed.(*openai.ChatCompletionRequest)
	require.True(t, ok, "parsed body should be *openai.ChatCompletionRequest")

	// Verify
	assert.Equal(t, "claude-3-5-sonnet-20241022", openaiReq.Model)
	assert.Equal(t, 1024, *openaiReq.MaxTokens)
	require.Len(t, openaiReq.Messages, 1)

	msg := openaiReq.Messages[0]
	assert.Equal(t, "user", msg.Role)
	assert.NotNil(t, msg.Content)

	contentArray, ok := msg.Content.(openai.MessageContentArray)
	require.True(t, ok)
	require.Len(t, contentArray, 1)
	assert.Equal(t, "text", contentArray[0].Type)
	assert.Equal(t, "Hello, Claude!", contentArray[0].Text)
}

func TestApiChatCompletionsToApiMessages_convertRequestBody_WithSystem(t *testing.T) {
	converter := NewChatCompletionToClaudeMessages(nil)
	ctx := context.Background()

	text := "Hello!"
	anthropicReq := &anthropic.ClaudeMessagesRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		System:    anthropic.SystemString("You are a helpful assistant"),
		Messages: []*anthropic.MessageParam{
			{
				Role: "user",
				Content: []anthropic.MessageContent{
					&anthropic.MessageContentBlock{
						Type: "text",
						Text: &text,
					},
				},
			},
		},
	}

	srcBody := octollm.NewBodyFromBytes([]byte{}, &octollm.JSONParser[anthropic.ClaudeMessagesRequest]{})
	srcBody.SetParsed(anthropicReq)

	dstBody, err := converter.convertRequestBody(ctx, srcBody)
	require.NoError(t, err)

	parsed, err := dstBody.Parsed()
	require.NoError(t, err)

	openaiReq, ok := parsed.(*openai.ChatCompletionRequest)
	require.True(t, ok)
	require.Len(t, openaiReq.Messages, 2)

	// First message should be system
	assert.Equal(t, "system", openaiReq.Messages[0].Role)
	assert.Equal(t, "You are a helpful assistant", openaiReq.Messages[0].Content.ExtractText())

	// Second message should be user
	assert.Equal(t, "user", openaiReq.Messages[1].Role)
}

func TestApiChatCompletionsToApiMessages_convertRequestBody_WithTools(t *testing.T) {
	converter := NewChatCompletionToClaudeMessages(nil)
	ctx := context.Background()

	text := "What's the weather?"
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{"location"},
	}
	schemaBytes, _ := json.Marshal(schema)

	anthropicReq := &anthropic.ClaudeMessagesRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Messages: []*anthropic.MessageParam{
			{
				Role: "user",
				Content: []anthropic.MessageContent{
					&anthropic.MessageContentBlock{
						Type: "text",
						Text: &text,
					},
				},
			},
		},
		Tools: []*anthropic.ToolDefinition{
			{
				Name:        "get_weather",
				Description: "Get weather information",
				InputSchema: json.RawMessage(schemaBytes),
			},
		},
	}

	srcBody := octollm.NewBodyFromBytes([]byte{}, &octollm.JSONParser[anthropic.ClaudeMessagesRequest]{})
	srcBody.SetParsed(anthropicReq)

	dstBody, err := converter.convertRequestBody(ctx, srcBody)
	require.NoError(t, err)

	parsed, err := dstBody.Parsed()
	require.NoError(t, err)

	openaiReq, ok := parsed.(*openai.ChatCompletionRequest)
	require.True(t, ok)
	require.Len(t, openaiReq.Tools, 1)

	tool := openaiReq.Tools[0]
	assert.Equal(t, "function", tool.Type)
	assert.Equal(t, "get_weather", *tool.Function.Name)
	assert.Equal(t, "Get weather information", *tool.Function.Description)
	assert.NotNil(t, tool.Function.Parameters)
}

func TestApiChatCompletionsToApiMessages_convertNonStreamResponseBody(t *testing.T) {
	converter := NewChatCompletionToClaudeMessages(nil)
	ctx := context.Background()

	// Create OpenAI response
	openaiResp := &openai.ChatCompletionResponse{
		ID:     "msg_123",
		Model:  "gpt-4",
		Object: "chat.completion",
		Choices: []*openai.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: "stop",
				Message: &openai.Message{
					Role:    "assistant",
					Content: openai.MessageContentString("Hello! How can I help you?"),
				},
			},
		},
		Usage: &openai.Usage{
			PromptTokens:     10,
			CompletionTokens: 8,
			TotalTokens:      18,
		},
	}

	srcBody := octollm.NewBodyFromBytes([]byte{}, &octollm.JSONParser[openai.ChatCompletionResponse]{})
	srcBody.SetParsed(openaiResp)

	// Convert
	dstBody, err := converter.convertNonStreamResponseBody(ctx, srcBody)
	require.NoError(t, err)

	// Parse result
	parsed, err := dstBody.Parsed()
	require.NoError(t, err)

	claudeResp, ok := parsed.(*anthropic.ClaudeMessagesResponse)
	require.True(t, ok, "parsed body should be *anthropic.ClaudeMessagesResponse")

	// Verify
	assert.Equal(t, "msg_123", claudeResp.ID)
	assert.Equal(t, "message", claudeResp.Type)
	assert.Equal(t, "assistant", claudeResp.Role)
	assert.Equal(t, "gpt-4", claudeResp.Model)
	assert.Equal(t, "end_turn", claudeResp.StopReason)
	assert.Equal(t, int64(10), claudeResp.Usage.InputTokens)
	assert.Equal(t, int64(8), claudeResp.Usage.OutputTokens)

	require.Len(t, claudeResp.Content, 1)
	block, ok := claudeResp.Content[0].(*anthropic.MessageContentBlock)
	require.True(t, ok)
	assert.Equal(t, "text", block.Type)
	assert.Equal(t, "Hello! How can I help you?", *block.Text)
}

func TestApiChatCompletionsToApiMessages_convertNonStreamResponseBody_WithToolCall(t *testing.T) {
	converter := NewChatCompletionToClaudeMessages(nil)
	ctx := context.Background()

	// Create OpenAI response with tool call
	openaiResp := &openai.ChatCompletionResponse{
		ID:     "msg_456",
		Model:  "gpt-4",
		Object: "chat.completion",
		Choices: []*openai.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: "tool_calls",
				Message: &openai.Message{
					Role: "assistant",
					ToolCalls: []*openai.ToolCall{
						{
							ID:    "call_123",
							Index: 0,
							Type:  "function",
							Function: &openai.ToolCallFunction{
								Name:      "get_weather",
								Arguments: `{"location":"San Francisco"}`,
							},
						},
					},
				},
			},
		},
		Usage: &openai.Usage{
			PromptTokens:     15,
			CompletionTokens: 20,
			TotalTokens:      35,
		},
	}

	srcBody := octollm.NewBodyFromBytes([]byte{}, &octollm.JSONParser[openai.ChatCompletionResponse]{})
	srcBody.SetParsed(openaiResp)

	// Convert
	dstBody, err := converter.convertNonStreamResponseBody(ctx, srcBody)
	require.NoError(t, err)

	// Parse result
	parsed, err := dstBody.Parsed()
	require.NoError(t, err)

	claudeResp, ok := parsed.(*anthropic.ClaudeMessagesResponse)
	require.True(t, ok)

	// Verify
	assert.Equal(t, "tool_use", claudeResp.StopReason)
	require.Len(t, claudeResp.Content, 1)

	block, ok := claudeResp.Content[0].(*anthropic.MessageContentBlock)
	require.True(t, ok)
	assert.Equal(t, "tool_use", block.Type)
	assert.NotNil(t, block.MessageContentToolUse)
	assert.Equal(t, "call_123", block.MessageContentToolUse.ID)
	assert.Equal(t, "get_weather", block.MessageContentToolUse.Name)
	assert.JSONEq(t, `{"location":"San Francisco"}`, string(block.MessageContentToolUse.Input))
}

func TestApiChatCompletionsToApiMessages_mapFinishReason(t *testing.T) {
	converter := NewChatCompletionToClaudeMessages(nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"stop", "end_turn"},
		{"length", "max_tokens"},
		{"tool_calls", "tool_use"},
		{"content_filter", "content_filter"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := converter.mapFinishReason(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
