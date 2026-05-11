package image_url_fetch

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRawOpenAIImageURLStrings_emptyAndNoop(t *testing.T) {
	t.Parallel()

	out, mod, err := normalizeRawOpenAIImageURLStrings(nil)
	require.NoError(t, err)
	require.False(t, mod)
	require.Nil(t, out)

	out, mod, err = normalizeRawOpenAIImageURLStrings([]byte{})
	require.NoError(t, err)
	require.False(t, mod)
	require.Empty(t, out)

	// No messages (embedding-style body)
	out, mod, err = normalizeRawOpenAIImageURLStrings([]byte(`{"model":"ada","input":"x"}`))
	require.NoError(t, err)
	require.False(t, mod)
	require.Contains(t, string(out), `"input"`)

	// messages exists but not an array
	out, mod, err = normalizeRawOpenAIImageURLStrings([]byte(`{"messages":"not-array"}`))
	require.NoError(t, err)
	require.False(t, mod)

	// Already object form — substring "image_url":" does not match object "image_url":{
	raw := []byte(`{"model":"m","messages":[{"role":"user","content":[{"type":"image_url","image_url":{"url":"https://a.com/p.png"}}]}]}`)
	out, mod, err = normalizeRawOpenAIImageURLStrings(raw)
	require.NoError(t, err)
	require.False(t, mod)
	require.Equal(t, raw, out)
}

func TestNormalizeRawOpenAIImageURLStrings_stringToObject_httpURL(t *testing.T) {
	in := []byte(`{"model":"m","messages":[{"role":"user","content":[{"type":"image_url","image_url":"https://example.com/a.png"}]}]}`)
	out, mod, err := normalizeRawOpenAIImageURLStrings(in)
	require.NoError(t, err)
	require.True(t, mod)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &payload))
	msgs := payload["messages"].([]interface{})
	msg0 := msgs[0].(map[string]interface{})
	content := msg0["content"].([]interface{})
	part0 := content[0].(map[string]interface{})
	img := part0["image_url"].(map[string]interface{})
	require.Equal(t, "https://example.com/a.png", img["url"])
}

// Pretty-printed JSON 常在冒号与值之间带空格；此前用 "\"image_url\":\"" 做 fast-path 会误判为「无需规范化」。
func TestNormalizeRawOpenAIImageURLStrings_stringToObject_httpURL_spacedAfterColon(t *testing.T) {
	in := []byte(`{
  "model": "m",
  "messages": [
    {
      "role": "user",
      "content": [
        { "type": "image_url", "image_url": "https://cdn.example.com/p.png" }
      ]
    }
  ]
}`)
	out, mod, err := normalizeRawOpenAIImageURLStrings(in)
	require.NoError(t, err)
	require.True(t, mod)
	require.Contains(t, string(out), `"image_url":{"url":"https://cdn.example.com/p.png"}`)
}

func TestNormalizeRawOpenAIImageURLStrings_stringToObject_dataURL(t *testing.T) {
	data := "data:image/png;base64,iVBORw0KGgo="
	in := []byte(`{"messages":[{"role":"user","content":[{"type":"image_url","image_url":"` + data + `"}]}]}`)
	out, mod, err := normalizeRawOpenAIImageURLStrings(in)
	require.NoError(t, err)
	require.True(t, mod)
	require.Contains(t, string(out), `"image_url":{"url":`)
	require.Contains(t, string(out), data)
}

func TestNormalizeRawOpenAIImageURLStrings_reasoningContent(t *testing.T) {
	in := []byte(`{"messages":[{"role":"user","content":[{"type":"text","text":"hi"}],"reasoning_content":[{"type":"image_url","image_url":"https://x/y.jpg"}]}]}`)
	out, mod, err := normalizeRawOpenAIImageURLStrings(in)
	require.NoError(t, err)
	require.True(t, mod)
	require.Contains(t, string(out), `"reasoning_content"`)
	require.Contains(t, string(out), `"image_url":{"url":"https://x/y.jpg"}`)
}

func TestNormalizeRawOpenAIImageURLStrings_skipsNonImageURLType(t *testing.T) {
	// type text should not be rewritten even if image_url were a string (invalid but defensive)
	in := []byte(`{"messages":[{"role":"user","content":[{"type":"text","text":"x","image_url":"https://bad.example/z"}]}]}`)
	out, mod, err := normalizeRawOpenAIImageURLStrings(in)
	require.NoError(t, err)
	require.False(t, mod)
	require.Equal(t, in, out)
}

func TestNormalizeRawOpenAIImageURLStrings_multipleStringParts(t *testing.T) {
	in := []byte(`{"messages":[{"role":"user","content":[
		{"type":"text","text":"a"},
		{"type":"image_url","image_url":"https://a.com/1.png"},
		{"type":"image_url","image_url":"https://b.com/2.png"}
	]}]}`)
	out, mod, err := normalizeRawOpenAIImageURLStrings(in)
	require.NoError(t, err)
	require.True(t, mod)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &payload))
	content := payload["messages"].([]interface{})[0].(map[string]interface{})["content"].([]interface{})
	u1 := content[1].(map[string]interface{})["image_url"].(map[string]interface{})["url"]
	u2 := content[2].(map[string]interface{})["image_url"].(map[string]interface{})["url"]
	require.Equal(t, "https://a.com/1.png", u1)
	require.Equal(t, "https://b.com/2.png", u2)
}

func TestCollectOpenAIStringImageURLTargets_indices(t *testing.T) {
	in := []byte(`{"messages":[
		{"role":"system","content":[]},
		{"role":"user","content":[{"type":"image_url","image_url":"https://u/0.png"}]},
		{"role":"user","content":[],"reasoning_content":[{"type":"image_url","image_url":"https://u/1.png"}]}
	]}`)
	tgts, err := collectOpenAIStringImageURLTargets(in)
	require.NoError(t, err)
	require.Len(t, tgts, 2)
	require.Equal(t, 1, tgts[0].msgIdx)
	require.Equal(t, "content", tgts[0].field)
	require.Equal(t, 0, tgts[0].partIdx)
	require.Equal(t, "https://u/0.png", tgts[0].url)

	require.Equal(t, 2, tgts[1].msgIdx)
	require.Equal(t, "reasoning_content", tgts[1].field)
	require.Equal(t, 0, tgts[1].partIdx)
	require.Equal(t, "https://u/1.png", tgts[1].url)
}
