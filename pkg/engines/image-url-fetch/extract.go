package image_url_fetch

import (
	"github.com/infinigence/octollm/pkg/types/openai"
)

// ExtractOpenAIChatImageURLRefs returns OpenAI-style remote image_url references in part order.
// MessageContentString and nil content yield nil. Unknown concrete types yield nil.
//
// For Anthropic Messages (ClaudeMessagesRequest), add a separate ExtractClaude* function
// that produces OpenAIChatImageURLRef only if shapes align, or a dedicated ref type + path method.
func ExtractOpenAIChatImageURLRefs(c openai.MessageContent) []OpenAIChatImageURLRef {
	if c == nil {
		return nil
	}
	switch v := c.(type) {
	case openai.MessageContentString:
		return nil
	case openai.MessageContentArray:
		return extractOpenAIRefsFromArray(v)
	default:
		return nil
	}
}

func extractOpenAIRefsFromArray(m openai.MessageContentArray) []OpenAIChatImageURLRef {
	var refs []OpenAIChatImageURLRef
	for i, item := range m {
		if item == nil || item.ImageURL == nil {
			continue
		}
		if item.Type != "" && item.Type != "image_url" {
			continue
		}
		u := item.ImageURL.GetImageUrl()
		if u == "" {
			continue
		}
		_, isObject := item.ImageURL.(*openai.MessageContentItemImageURL)
		refs = append(refs, OpenAIChatImageURLRef{
			PartIndex:    i,
			URL:          u,
			IsObjectForm: isObject,
		})
	}
	return refs
}
