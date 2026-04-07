package image_url_fetch

import "fmt"

// OpenAIChatImageURLRef locates one remote image URL in OpenAI Chat Completions JSON
// (messages[].content / reasoning_content multipart arrays) for jsonparser.Set replacement.
// Anthropic / other APIs should use their own *Ref types and path helpers in this package.
type OpenAIChatImageURLRef struct {
	PartIndex int // index inside the content array part list
	URL       string
	// IsObjectForm is true when JSON had "image_url": {"url":"..."}; false when "image_url": "https://...".
	IsObjectForm bool
}

// JSONParserPath returns buger/jsonparser keys: messages, [msgIndex], field, [PartIndex], image_url(, url).
// field is typically "content" or "reasoning_content".
func (r OpenAIChatImageURLRef) JSONParserPath(msgIndex int, field string) []string {
	p := []string{
		"messages",
		fmt.Sprintf("[%d]", msgIndex),
		field,
		fmt.Sprintf("[%d]", r.PartIndex),
	}
	if r.IsObjectForm {
		return append(p, "image_url", "url")
	}
	return append(p, "image_url")
}
