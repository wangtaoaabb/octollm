package client

import (
	"fmt"
	"net/http"
	"os"

	"github.com/infinigence/octollm/pkg/octollm"
	"github.com/infinigence/octollm/pkg/types/anthropic"
	"github.com/infinigence/octollm/pkg/types/openai"
	"github.com/infinigence/octollm/pkg/types/rerank"
)

type GeneralEndpoint struct {
	*HTTPEndpoint
}

// GeneralEndpoint implements octollm.Endpoint
var _ octollm.Engine = (*GeneralEndpoint)(nil)

type GeneralEndpointConfig struct {
	BaseURL   string
	Endpoints map[octollm.APIFormat]string
	APIKey    string

	AnthropicAPIKeyAsBearer bool
}

var DefaultURLPathChatCompletions = "/v1/chat/completions"
var DefaultURLPathCompletions = "/v1/completions"
var DefaultURLPathClaudeMessages = "/v1/messages"
var DefaultURLPathEmbeddings = "/v1/embeddings"
var DefaultURLPathRerank = "/v1/rerank"

func NewGeneralEndpoint(conf GeneralEndpointConfig) *GeneralEndpoint {
	apiKey := conf.APIKey
	if apiKey == "" {
		// read from env
		apiKey = os.Getenv("OCTOLLM_API_KEY")
	}

	httpEndpoint := NewHTTPEndpoint().
		WithURLGetter(func(req *octollm.Request) (string, error) {
			endpoint, ok := conf.Endpoints[req.Format]
			if !ok {
				return "", fmt.Errorf("invalid format: %s", req.Format)
			}
			if endpoint == "" {
				switch req.Format {
				case octollm.APIFormatClaudeMessages:
					endpoint = DefaultURLPathClaudeMessages
				case octollm.APIFormatChatCompletions:
					endpoint = DefaultURLPathChatCompletions
				case octollm.APIFormatCompletions:
					endpoint = DefaultURLPathCompletions
				case octollm.APIFormatEmbeddings:
					endpoint = DefaultURLPathEmbeddings
				case octollm.APIFormatRerank:
					endpoint = DefaultURLPathRerank
				default:
					return "", fmt.Errorf("invalid format: %s", req.Format)
				}
			}
			return conf.BaseURL + endpoint, nil
		}).
		WithParser(
			func(req *octollm.Request) octollm.Parser {
				switch req.Format {
				case octollm.APIFormatClaudeMessages:
					return &octollm.JSONParser[anthropic.ClaudeMessagesResponse]{}
				case octollm.APIFormatCompletions:
					return &octollm.JSONParser[openai.CompletionResponse]{}
				case octollm.APIFormatEmbeddings:
					return &octollm.JSONParser[openai.EmbeddingResponse]{}
				case octollm.APIFormatRerank:
					return &octollm.JSONParser[rerank.RerankResponse]{}
				default:
					return &octollm.JSONParser[openai.ChatCompletionResponse]{}
				}
			},
			func(req *octollm.Request) octollm.Parser {
				switch req.Format {
				case octollm.APIFormatClaudeMessages:
					return &octollm.JSONParser[anthropic.ClaudeMessagesStreamEvent]{}
				case octollm.APIFormatCompletions:
					return &octollm.JSONParser[openai.CompletionStreamChunk]{}
				case octollm.APIFormatEmbeddings:
					// Embeddings don't support streaming
					return &octollm.JSONParser[openai.EmbeddingResponse]{}
				case octollm.APIFormatRerank:
					// Rerank doesn't support streaming
					return &octollm.JSONParser[rerank.RerankResponse]{}
				default:
					return &octollm.JSONParser[openai.ChatCompletionStreamChunk]{}
				}
			},
		)

	if apiKey != "" {
		httpEndpoint = httpEndpoint.WithRequestModifier(func(req *octollm.Request, httpReq *http.Request) *http.Request {

			if req.Format == octollm.APIFormatClaudeMessages && !conf.AnthropicAPIKeyAsBearer {
				httpReq.Header.Set("x-api-key", apiKey)
			} else {
				httpReq.Header.Set("Authorization", "Bearer "+apiKey)
			}
			return httpReq
		})
	}

	return &GeneralEndpoint{
		HTTPEndpoint: httpEndpoint,
	}
}
