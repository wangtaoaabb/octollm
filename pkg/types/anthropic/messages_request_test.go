package anthropic

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiMessagesRequest_UnmarshalJSON_StringContent(t *testing.T) {
	jsonStr := `{
		"model": "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"messages": [
			{
				"role": "user",
				"content": "Hello, Claude!"
			}
		]
	}`

	var req ClaudeMessagesRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err)

	assert.Equal(t, "claude-3-5-sonnet-20241022", req.Model)
	assert.Equal(t, int64(1024), req.MaxTokens)
	require.Len(t, req.Messages, 1)

	msg := req.Messages[0]
	assert.Equal(t, "user", msg.Role)
	require.Len(t, msg.Content, 1)

	// Check if it's a MessageContentString
	if str, ok := msg.Content[0].(MessageContentString); ok {
		assert.Equal(t, "Hello, Claude!", string(str))
	} else if block, ok := msg.Content[0].(*MessageContentBlock); ok {
		assert.Equal(t, "text", block.Type)
		assert.NotNil(t, block.Text)
		assert.Equal(t, "Hello, Claude!", *block.Text)
	} else {
		t.Fatalf("unexpected content type: %T", msg.Content[0])
	}
}

func TestApiMessagesRequest_UnmarshalJSON_ArrayContent(t *testing.T) {
	jsonStr := `{
		"model": "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"type": "text",
						"text": "Hello, Claude!"
					}
				]
			}
		]
	}`

	var req ClaudeMessagesRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err)

	assert.Equal(t, "claude-3-5-sonnet-20241022", req.Model)
	assert.Equal(t, int64(1024), req.MaxTokens)
	require.Len(t, req.Messages, 1)

	msg := req.Messages[0]
	assert.Equal(t, "user", msg.Role)
	require.Len(t, msg.Content, 1)

	// Should be a MessageContentBlock
	block, ok := msg.Content[0].(*MessageContentBlock)
	require.True(t, ok, "content should be *MessageContentBlock")
	assert.Equal(t, "text", block.Type)
	assert.NotNil(t, block.Text)
	assert.Equal(t, "Hello, Claude!", *block.Text)
}

func TestApiMessagesRequest_UnmarshalJSON_ObjectContent(t *testing.T) {
	jsonStr := `{
		"model": "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"messages": [
			{
				"role": "user",
				"content": {
					"type": "text",
					"text": "Hello, Claude!"
				}
			}
		]
	}`

	var req ClaudeMessagesRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err)

	require.Len(t, req.Messages, 1)
	msg := req.Messages[0]
	require.Len(t, msg.Content, 1)

	// Should be a MessageContentBlock
	block, ok := msg.Content[0].(*MessageContentBlock)
	require.True(t, ok, "content should be *MessageContentBlock")
	assert.Equal(t, "text", block.Type)
	assert.NotNil(t, block.Text)
	assert.Equal(t, "Hello, Claude!", *block.Text)
}

func TestApiMessagesRequest_MarshalJSON_SimpleRequest(t *testing.T) {
	text := "Hello, Claude!"
	req := &ClaudeMessagesRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Messages: []*MessageParam{
			{
				Role: "user",
				Content: []MessageContent{
					&MessageContentBlock{
						Type: "text",
						Text: &text,
					},
				},
			},
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	// Verify basic structure
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "claude-3-5-sonnet-20241022", result["model"])
	assert.Equal(t, float64(1024), result["max_tokens"])

	// Verify messages
	messages, ok := result["messages"].([]interface{})
	require.True(t, ok)
	require.Len(t, messages, 1)

	msg, ok := messages[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "user", msg["role"])

	// Content is now an array
	contentArray, ok := msg["content"].([]interface{})
	require.True(t, ok)
	require.Len(t, contentArray, 1)

	content, ok := contentArray[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "text", content["type"])
	assert.Equal(t, "Hello, Claude!", content["text"])
}

func TestApiMessagesRequest_MarshalJSON_WithSystemString(t *testing.T) {
	text := "Hello!"
	req := &ClaudeMessagesRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		System:    SystemString("You are a helpful assistant"),
		Messages: []*MessageParam{
			{
				Role: "user",
				Content: []MessageContent{
					&MessageContentBlock{
						Type: "text",
						Text: &text,
					},
				},
			},
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// System should be serialized as string
	system, ok := result["system"].(string)
	require.True(t, ok)
	assert.Equal(t, "You are a helpful assistant", system)
}

func TestApiMessagesRequest_MarshalJSON_WithMultiSystem(t *testing.T) {
	text := "Hello!"
	req := &ClaudeMessagesRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		System: SystemBlocks{
			{
				Type: "text",
				Text: "You are a helpful assistant",
			},
			{
				Type: "text",
				Text: "Be concise",
			},
		},
		Messages: []*MessageParam{
			{
				Role: "user",
				Content: []MessageContent{
					&MessageContentBlock{
						Type: "text",
						Text: &text,
					},
				},
			},
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// System should be serialized as array (MultiSystem takes precedence)
	system, ok := result["system"].([]interface{})
	require.True(t, ok)
	require.Len(t, system, 2)

	sysBlock1, ok := system[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "text", sysBlock1["type"])
	assert.Equal(t, "You are a helpful assistant", sysBlock1["text"])
}

func TestApiMessagesRequest_UnmarshalJSON_WithSystemString(t *testing.T) {
	jsonStr := `{
		"model": "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"system": "You are a helpful assistant",
		"messages": [
			{
				"role": "user",
				"content": "Hello, Claude!"
			}
		]
	}`

	var req ClaudeMessagesRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err)

	assert.Equal(t, "claude-3-5-sonnet-20241022", req.Model)
	assert.Equal(t, int64(1024), req.MaxTokens)

	// System should be SystemString
	systemStr, ok := req.System.(SystemString)
	require.True(t, ok, "system should be SystemString")
	assert.Equal(t, "You are a helpful assistant", string(systemStr))
}

func TestApiMessagesRequest_UnmarshalJSON_WithSystemArray(t *testing.T) {
	jsonStr := `{
		"model": "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"system": [
			{
				"type": "text",
				"text": "You are a helpful assistant"
			},
			{
				"type": "text",
				"text": "Be concise"
			}
		],
		"messages": [
			{
				"role": "user",
				"content": "Hello, Claude!"
			}
		]
	}`

	var req ClaudeMessagesRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err)

	assert.Equal(t, "claude-3-5-sonnet-20241022", req.Model)
	assert.Equal(t, int64(1024), req.MaxTokens)

	// System should be SystemBlocks
	systemBlocks, ok := req.System.(SystemBlocks)
	require.True(t, ok, "system should be SystemBlocks")
	require.Len(t, systemBlocks, 2)
	assert.Equal(t, "text", systemBlocks[0].Type)
	assert.Equal(t, "You are a helpful assistant", systemBlocks[0].Text)
	assert.Equal(t, "Be concise", systemBlocks[1].Text)
}

func TestApiMessagesRequest_RoundTrip_SystemString(t *testing.T) {
	original := &ClaudeMessagesRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		System:    SystemString("You are a helpful assistant"),
		Messages: []*MessageParam{
			{
				Role: "user",
				Content: []MessageContent{
					MessageContentString("Hello"),
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var unmarshaled ClaudeMessagesRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify system
	systemStr, ok := unmarshaled.System.(SystemString)
	require.True(t, ok)
	assert.Equal(t, string(original.System.(SystemString)), string(systemStr))
}

func TestApiMessagesRequest_RoundTrip_SystemBlocks(t *testing.T) {
	original := &ClaudeMessagesRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		System: SystemBlocks{
			{
				Type: "text",
				Text: "You are a helpful assistant",
			},
		},
		Messages: []*MessageParam{
			{
				Role: "user",
				Content: []MessageContent{
					MessageContentString("Hello"),
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var unmarshaled ClaudeMessagesRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify system
	systemBlocks, ok := unmarshaled.System.(SystemBlocks)
	require.True(t, ok)
	require.Len(t, systemBlocks, 1)
	assert.Equal(t, original.System.(SystemBlocks)[0].Text, systemBlocks[0].Text)
}

func TestApiMessagesRequest_UnmarshalJSON_WithToolResult(t *testing.T) {
	jsonStr := `{
		"model": "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"tool_use_id": "call_4cba1ad8bc8e4ba2983278ac",
						"type": "tool_result",
						"content": [
							{
								"type": "text",
								"text": "API Error: 404"
							},
							{
								"type": "text",
								"text": "agentId: a168307"
							}
						]
					},
					{
						"type": "text",
						"text": "[Request interrupted by user]"
					},
					{
						"type": "text",
						"text": "帮我看看这个项目是在做什么",
						"cache_control": {
							"type": "ephemeral"
						}
					}
				]
			}
		]
	}`

	var req ClaudeMessagesRequest
	err := json.Unmarshal([]byte(jsonStr), &req)
	require.NoError(t, err, "should successfully unmarshal tool_result with nested content")

	assert.Equal(t, "claude-3-5-sonnet-20241022", req.Model)
	assert.Equal(t, int64(1024), req.MaxTokens)
	require.Len(t, req.Messages, 1)

	msg := req.Messages[0]
	assert.Equal(t, "user", msg.Role)
	require.Len(t, msg.Content, 3)

	// First block should be tool_result
	block0, ok := msg.Content[0].(*MessageContentBlock)
	require.True(t, ok, "first content should be *MessageContentBlock")
	assert.Equal(t, "tool_result", block0.Type)
	assert.NotNil(t, block0.MessageContentToolResult)
	assert.Equal(t, "call_4cba1ad8bc8e4ba2983278ac", *block0.ToolUseID)
	require.Len(t, block0.Content, 2)

	// Second block should be text
	block1, ok := msg.Content[1].(*MessageContentBlock)
	require.True(t, ok, "second content should be *MessageContentBlock")
	assert.Equal(t, "text", block1.Type)
	assert.Equal(t, "[Request interrupted by user]", *block1.Text)

	// Third block should be text with cache_control
	block2, ok := msg.Content[2].(*MessageContentBlock)
	require.True(t, ok, "third content should be *MessageContentBlock")
	assert.Equal(t, "text", block2.Type)
	assert.Equal(t, "帮我看看这个项目是在做什么", *block2.Text)
	assert.NotNil(t, block2.CacheControl)
	assert.Equal(t, "ephemeral", block2.CacheControl.Type)
}
