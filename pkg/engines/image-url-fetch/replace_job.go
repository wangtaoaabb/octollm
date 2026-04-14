package image_url_fetch

import (
	"fmt"
	"strings"
)

// openaiImageReplaceJob is one OpenAI Chat Completions remote image_url to inline (full path in request JSON).
type openaiImageReplaceJob struct {
	MsgIndex     int
	Field        string // "content" or "reasoning_content"
	PartIndex    int
	URL          string
	IsObjectForm bool // true when JSON had "image_url": {"url":"..."}; false when "image_url": "https://...".
}

func (j openaiImageReplaceJob) jsonParserPath() []string {
	p := []string{
		"messages",
		fmt.Sprintf("[%d]", j.MsgIndex),
		j.Field,
		fmt.Sprintf("[%d]", j.PartIndex),
	}
	if j.IsObjectForm {
		return append(p, "image_url", "url")
	}
	return append(p, "image_url")
}

func (j openaiImageReplaceJob) remoteURL() string {
	return strings.TrimSpace(j.URL)
}

// claudeImageReplaceJob is one Anthropic Messages remote image URL to inline (full path in request JSON).
// After fetch, the entire "source" object is replaced per Messages API: type base64, media_type, data.
// ContentIndices walks nested content arrays: first index is into messages[MsgIndex].content, each further
// index is into a tool_result.content array (messages[MsgIndex].content[i].content[j]...).
type claudeImageReplaceJob struct {
	MsgIndex       int
	ContentIndices []int
	URL            string
}

// jsonParserPathToSource returns buger/jsonparser keys through the image block's "source" key (value replaced in full).
func (j claudeImageReplaceJob) jsonParserPathToSource() []string {
	p := []string{"messages", fmt.Sprintf("[%d]", j.MsgIndex)}
	for _, idx := range j.ContentIndices {
		p = append(p, "content", fmt.Sprintf("[%d]", idx))
	}
	p = append(p, "source")
	return p
}

func (j claudeImageReplaceJob) remoteURL() string {
	return strings.TrimSpace(j.URL)
}
