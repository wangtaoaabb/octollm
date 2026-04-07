package image_url_fetch

import (
	"encoding/json"
	"testing"

	"github.com/infinigence/octollm/pkg/types/openai"
)

func TestExtractOpenAIChatImageURLRefs(t *testing.T) {
	t.Run("string content", func(t *testing.T) {
		var c openai.MessageContent = openai.MessageContentString("hello")
		if got := ExtractOpenAIChatImageURLRefs(c); got != nil {
			t.Fatalf("ExtractOpenAIChatImageURLRefs() = %v, want nil", got)
		}
	})

	t.Run("array text only", func(t *testing.T) {
		var c openai.MessageContent = openai.MessageContentArray([]*openai.MessageContentItem{
			{Type: "text", Text: "hi"},
		})
		if got := ExtractOpenAIChatImageURLRefs(c); len(got) != 0 {
			t.Fatalf("ExtractOpenAIChatImageURLRefs() = %v, want empty", got)
		}
	})

	t.Run("array with image_url string and struct", func(t *testing.T) {
		var c openai.MessageContent = openai.MessageContentArray([]*openai.MessageContentItem{
			{Type: "text", Text: "x"},
			{Type: "image_url", ImageURL: openai.ImageURLString("https://a.example/1.png")},
			{Type: "image_url", ImageURL: &openai.MessageContentItemImageURL{URL: "https://b.example/2.png", Detail: "high"}},
		})
		got := ExtractOpenAIChatImageURLRefs(c)
		want := []OpenAIChatImageURLRef{
			{PartIndex: 1, URL: "https://a.example/1.png", IsObjectForm: false},
			{PartIndex: 2, URL: "https://b.example/2.png", IsObjectForm: true},
		}
		if len(got) != len(want) {
			t.Fatalf("ExtractOpenAIChatImageURLRefs() = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("ExtractOpenAIChatImageURLRefs()[%d] = %#v, want %#v", i, got[i], want[i])
			}
		}
	})

	t.Run("from JSON unmarshal string image_url", func(t *testing.T) {
		jsonData := `{"role":"user","content":[
			{"type":"text","text":"see"},
			{"type":"image_url","image_url":"https://u.example/x.png"}
		]}`
		var msg openai.Message
		if err := json.Unmarshal([]byte(jsonData), &msg); err != nil {
			t.Fatal(err)
		}
		got := ExtractOpenAIChatImageURLRefs(msg.Content)
		if len(got) != 1 {
			t.Fatalf("ExtractOpenAIChatImageURLRefs() = %v", got)
		}
		if got[0].URL != "https://u.example/x.png" || got[0].PartIndex != 1 || got[0].IsObjectForm {
			t.Fatalf("got %#v", got[0])
		}
	})
}

func TestExtractOpenAIChatImageURLRefs_nil(t *testing.T) {
	if got := ExtractOpenAIChatImageURLRefs(nil); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestOpenAIChatImageURLRef_JSONParserPath(t *testing.T) {
	ref := OpenAIChatImageURLRef{PartIndex: 2, URL: "https://x", IsObjectForm: true}
	got := ref.JSONParserPath(1, "content")
	want := []string{"messages", "[1]", "content", "[2]", "image_url", "url"}
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Fatalf("JSONParserPath() = %v, want %v", got, want)
		}
	}
	ref2 := OpenAIChatImageURLRef{PartIndex: 0, IsObjectForm: false}
	got2 := ref2.JSONParserPath(0, "reasoning_content")
	want2 := []string{"messages", "[0]", "reasoning_content", "[0]", "image_url"}
	for i := range want2 {
		if i >= len(got2) || got2[i] != want2[i] {
			t.Fatalf("JSONParserPath() string form = %v, want %v", got2, want2)
		}
	}
}
