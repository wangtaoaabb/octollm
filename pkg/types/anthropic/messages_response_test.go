package anthropic

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeMessagesStreamEvent_UnmarshalJSON_ContentBlockStart(t *testing.T) {
	// 模拟真实的 content_block_start 事件
	jsonStr := `{
		"type": "content_block_start",
		"index": 0,
		"content_block": {
			"type": "text",
			"text": ""
		}
	}`

	var event ClaudeMessagesStreamEvent
	err := json.Unmarshal([]byte(jsonStr), &event)
	require.NoError(t, err)

	assert.Equal(t, "content_block_start", event.Type)
	assert.NotNil(t, event.Index)
	assert.Equal(t, 0, *event.Index)
	assert.NotNil(t, event.ContentBlock)

	block, ok := event.ContentBlock.(*MessageContentBlock)
	require.True(t, ok)
	assert.Equal(t, "text", block.Type)
	assert.NotNil(t, block.Text)
	assert.Equal(t, "", *block.Text)
}

func TestClaudeMessagesStreamEvent_UnmarshalJSON_TopLevelNull(t *testing.T) {
	var event ClaudeMessagesStreamEvent
	err := json.Unmarshal([]byte(`null`), &event)
	require.NoError(t, err)

	assert.Equal(t, "", event.Type)
	assert.Nil(t, event.Index)
	assert.Nil(t, event.Message)
	assert.Nil(t, event.ContentBlock)
	assert.Nil(t, event.Usage)
	assert.Nil(t, event.Error)
	assert.Nil(t, event.DeltaRaw)
}
func TestClaudeMessagesStreamEvent_UnmarshalJSON_ContentBlockStartWithToolUse(t *testing.T) {
	// 模拟 tool_use 类型的 content_block_start 事件
	jsonStr := `{
		"type": "content_block_start",
		"index": 1,
		"content_block": {
			"type": "tool_use",
			"id": "toolu_123",
			"name": "get_weather",
			"input": {}
		}
	}`

	var event ClaudeMessagesStreamEvent
	err := json.Unmarshal([]byte(jsonStr), &event)
	require.NoError(t, err)

	assert.Equal(t, "content_block_start", event.Type)
	assert.NotNil(t, event.Index)
	assert.Equal(t, 1, *event.Index)
	assert.NotNil(t, event.ContentBlock)

	block, ok := event.ContentBlock.(*MessageContentBlock)
	require.True(t, ok)
	assert.Equal(t, "tool_use", block.Type)
	assert.NotNil(t, block.MessageContentToolUse)
	assert.Equal(t, "toolu_123", block.MessageContentToolUse.ID)
	assert.Equal(t, "get_weather", block.MessageContentToolUse.Name)
}

func TestClaudeMessagesStreamEvent_UnmarshalJSON_ContentBlockDelta(t *testing.T) {
	// 模拟 content_block_delta 事件
	jsonStr := `{
		"type": "content_block_delta",
		"index": 0,
		"delta": {
			"type": "text_delta",
			"text": "Hello"
		}
	}`

	var event ClaudeMessagesStreamEvent
	err := json.Unmarshal([]byte(jsonStr), &event)
	require.NoError(t, err)

	assert.Equal(t, "content_block_delta", event.Type)
	assert.NotNil(t, event.Index)
	assert.Equal(t, 0, *event.Index)

	delta, err := event.GetContentBlockDelta()
	require.NoError(t, err)
	require.NotNil(t, delta)
	assert.Equal(t, "text_delta", delta.Type)
	assert.NotNil(t, delta.Text)
	assert.Equal(t, "Hello", *delta.Text)
}

func TestClaudeMessagesStreamEvent_UnmarshalJSON_MessageStart(t *testing.T) {
	// 模拟 message_start 事件
	jsonStr := `{
		"type": "message_start",
		"message": {
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": [],
			"model": "claude-3-5-sonnet-20241022",
			"stop_reason": null,
			"usage": {
				"input_tokens": 10,
				"output_tokens": 0
			}
		}
	}`

	var event ClaudeMessagesStreamEvent
	err := json.Unmarshal([]byte(jsonStr), &event)
	require.NoError(t, err)

	assert.Equal(t, "message_start", event.Type)
	assert.NotNil(t, event.Message)
	assert.Equal(t, "msg_123", event.Message.ID)
	assert.Equal(t, "assistant", event.Message.Role)
	assert.Equal(t, "claude-3-5-sonnet-20241022", event.Message.Model)
}

func TestClaudeMessagesStreamEvent_UnmarshalJSON_MessageStop(t *testing.T) {
	// 模拟 message_stop 事件
	jsonStr := `{
		"type": "message_stop"
	}`

	var event ClaudeMessagesStreamEvent
	err := json.Unmarshal([]byte(jsonStr), &event)
	require.NoError(t, err)

	assert.Equal(t, "message_stop", event.Type)
	assert.True(t, event.IsMessageStop())
}

func TestClaudeMessagesStreamEvent_UnmarshalJSON_NullContentBlock(t *testing.T) {
	// 测试 content_block 为 null 的情况
	jsonStr := `{
		"type": "content_block_stop",
		"index": 0,
		"content_block": null
	}`

	var event ClaudeMessagesStreamEvent
	err := json.Unmarshal([]byte(jsonStr), &event)
	require.NoError(t, err)

	assert.Equal(t, "content_block_stop", event.Type)
	assert.Nil(t, event.ContentBlock)
}

func TestClaudeMessagesStreamEvent_CheckMethods(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		checkFn  func(*ClaudeMessagesStreamEvent) bool
		expected bool
	}{
		{
			name:     "IsMessageStart",
			jsonStr:  `{"type": "message_start"}`,
			checkFn:  func(e *ClaudeMessagesStreamEvent) bool { return e.IsMessageStart() },
			expected: true,
		},
		{
			name:     "IsMessageStop",
			jsonStr:  `{"type": "message_stop"}`,
			checkFn:  func(e *ClaudeMessagesStreamEvent) bool { return e.IsMessageStop() },
			expected: true,
		},
		{
			name:     "IsContentBlockDelta",
			jsonStr:  `{"type": "content_block_delta"}`,
			checkFn:  func(e *ClaudeMessagesStreamEvent) bool { return e.IsContentBlockDelta() },
			expected: true,
		},
		{
			name:     "IsContentBlockStart",
			jsonStr:  `{"type": "content_block_start"}`,
			checkFn:  func(e *ClaudeMessagesStreamEvent) bool { return e.IsContentBlockStart() },
			expected: true,
		},
		{
			name:     "IsContentBlockStop",
			jsonStr:  `{"type": "content_block_stop"}`,
			checkFn:  func(e *ClaudeMessagesStreamEvent) bool { return e.IsContentBlockStop() },
			expected: true,
		},
		{
			name:     "IsError",
			jsonStr:  `{"type": "error"}`,
			checkFn:  func(e *ClaudeMessagesStreamEvent) bool { return e.IsError() },
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event ClaudeMessagesStreamEvent
			err := json.Unmarshal([]byte(tt.jsonStr), &event)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.checkFn(&event))
		})
	}
}

func TestClaudeMessagesResponse_UnmarshalJSON_TopLevelNull(t *testing.T) {
	var event ClaudeMessagesResponse
	err := json.Unmarshal([]byte(`null`), &event)
	require.NoError(t, err)

	assert.Equal(t, "", event.Type)
	assert.Equal(t, "", event.ID)
	assert.Nil(t, event.StopSequence)
}
