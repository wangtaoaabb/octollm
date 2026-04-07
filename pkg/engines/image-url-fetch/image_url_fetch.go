package image_url_fetch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/infinigence/octollm/pkg/octollm"
	"github.com/infinigence/octollm/pkg/types/openai"
)

var ErrImageURLFetch = errors.New("image url fetch engine")

// ImageURLFetchEngine downloads remote image_url targets in chat completion requests,
// replaces them with data:...;base64,... using jsonparser.Set on the raw body.
// Array indices in paths must use bracket syntax ("[0]", "[1]"); see jsonparser.searchKeys.
// URL discovery uses ExtractOpenAIChatImageURLRefs (this package).
type ImageURLFetchEngine struct {
	Next       octollm.Engine
	HTTPClient *http.Client

	retryCount int
	timeout    time.Duration
}

var _ octollm.Engine = (*ImageURLFetchEngine)(nil)

// ImageURLFetchConfig configures ImageURLFetchEngine. Zero values get defaults in NewImageURLFetchEngine.
type ImageURLFetchConfig struct {
	Next octollm.Engine
	// HTTPClient fetches remote images; nil uses http.DefaultClient.
	HTTPClient *http.Client
	// RetryCount is extra attempts after the first (e.g. 1 => up to 2 tries). Negative values are clamped to 0.
	RetryCount int
	// Timeout is per fetch attempt. Zero defaults to 10s.
	Timeout time.Duration
}

// NewImageURLFetchEngine wraps cfg.Next. Applies defaults: HTTPClient -> DefaultClient, Timeout -> 10s, RetryCount >= 0.
func NewImageURLFetchEngine(cfg ImageURLFetchConfig) *ImageURLFetchEngine {
	retry := cfg.RetryCount
	if retry < 0 {
		retry = 0
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return &ImageURLFetchEngine{
		Next:       cfg.Next,
		HTTPClient: client,
		retryCount: retry,
		timeout:    timeout,
	}
}

type imageReplaceJob struct {
	msgIndex int
	field    string // "content" or "reasoning_content"
	ref      OpenAIChatImageURLRef
}

// extractImageReplaceJobsFromBody inspects the unified body like moderator.OpenAIAdapter.ExtractTextFromBody:
// it uses Parsed() and switches on concrete types; only *openai.ChatCompletionRequest yields image jobs.
// A nil body or unsupported parsed type is logged and yields (nil, nil) so the engine passes the request through.
// The body must have a non-nil Parser (e.g. JSONParser[openai.ChatCompletionRequest]); Parsed() calls parser.Parse and a nil parser panics.
func extractImageReplaceJobsFromBody(ctx context.Context, body *octollm.UnifiedBody) ([]imageReplaceJob, error) {
	if body == nil {
		slog.WarnContext(ctx, "[ImageURLFetchEngine] request body is nil; skip image inlining")
		return nil, nil
	}

	parsed, err := body.Parsed()
	if err != nil {
		return nil, fmt.Errorf("%w: parse body error: %w", ErrImageURLFetch, err)
	}
	switch p := parsed.(type) {
	case *openai.ChatCompletionRequest:
		return collectImageReplaceJobs(p), nil
	default:
		slog.WarnContext(ctx, "[ImageURLFetchEngine] unsupported body type for image inlining; skip",
			slog.String("type", fmt.Sprintf("%T", parsed)))
		return nil, nil
	}
}

func (e *ImageURLFetchEngine) Process(req *octollm.Request) (*octollm.Response, error) {
	jobs, err := extractImageReplaceJobsFromBody(req.Context(), req.Body)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return e.Next.Process(req)
	}

	body, err := req.Body.Bytes()
	if err != nil {
		return nil, fmt.Errorf("%w: get request body bytes error: %w", ErrImageURLFetch, err)
	}

	unique := make(map[string]struct{})
	for _, j := range jobs {
		u := strings.TrimSpace(j.ref.URL)
		if u == "" || strings.HasPrefix(strings.ToLower(u), "data:") {
			continue
		}
		unique[u] = struct{}{}
	}
	if len(unique) == 0 {
		return e.Next.Process(req)
	}

	result := make(map[string]fetchImageResult, len(unique))
	var mu sync.Mutex
	var wg sync.WaitGroup
	ctx := req.Context()
	for u := range unique {
		wg.Add(1)
		go func(imageURL string) {
			defer wg.Done()
			dataURL, ferr := e.fetchAndEncode(ctx, imageURL)
			mu.Lock()
			result[imageURL] = fetchImageResult{dataURL: dataURL, err: ferr}
			mu.Unlock()
		}(u)
	}
	wg.Wait()

	for u, r := range result {
		if r.err != nil {
			return nil, fmt.Errorf("%w: fetch image %q: %w", ErrImageURLFetch, u, r.err)
		}
	}

	out := body
	for _, job := range jobs {
		u := strings.TrimSpace(job.ref.URL)
		if u == "" || strings.HasPrefix(strings.ToLower(u), "data:") {
			continue
		}
		dataURL := result[u].dataURL
		if dataURL == "" {
			continue
		}
		path := job.ref.JSONParserPath(job.msgIndex, job.field)
		encoded, mErr := json.Marshal(dataURL)
		if mErr != nil {
			return nil, fmt.Errorf("%w: marshal data url: %w", ErrImageURLFetch, mErr)
		}
		var setErr error
		out, setErr = jsonparser.Set(out, encoded, path...)
		if setErr != nil {
			return nil, fmt.Errorf("%w: json replace at %v: %w", ErrImageURLFetch, path, setErr)
		}
	}

	req.Body.SetBytes(out)
	slog.DebugContext(req.Context(), "[ImageURLFetchEngine] inlined remote image_url parts")

	resp, err := e.Next.Process(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type fetchImageResult struct {
	dataURL string
	err     error
}

func collectImageReplaceJobs(chatReq *openai.ChatCompletionRequest) []imageReplaceJob {
	var jobs []imageReplaceJob
	for i, msg := range chatReq.Messages {
		if msg == nil {
			continue
		}
		if msg.Content != nil {
			for _, ref := range ExtractOpenAIChatImageURLRefs(msg.Content) {
				jobs = append(jobs, imageReplaceJob{msgIndex: i, field: "content", ref: ref})
			}
		}
		if msg.ReasoningContent != nil {
			for _, ref := range ExtractOpenAIChatImageURLRefs(msg.ReasoningContent) {
				jobs = append(jobs, imageReplaceJob{msgIndex: i, field: "reasoning_content", ref: ref})
			}
		}
	}
	return jobs
}

func (e *ImageURLFetchEngine) fetchAndEncode(ctx context.Context, url string) (string, error) {
	maxAttempts := e.retryCount + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		dataURL, err := e.fetchAndEncodeOnce(ctx, url)
		if err == nil {
			return dataURL, nil
		}
		lastErr = err
	}
	return "", lastErr
}

func (e *ImageURLFetchEngine) fetchAndEncodeOnce(ctx context.Context, url string) (string, error) {
	attemptTimeout := e.timeout
	if attemptTimeout <= 0 {
		attemptTimeout = 10 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := e.HTTPClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		contentType = "image/jpeg"
	}
	if idx := strings.Index(contentType, ";"); idx > 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, b64), nil
}
