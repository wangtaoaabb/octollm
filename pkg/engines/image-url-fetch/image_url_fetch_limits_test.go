package image_url_fetch

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/infinigence/octollm/pkg/engines/image-url-fetch/limits"
	"github.com/infinigence/octollm/pkg/octollm"
	"github.com/stretchr/testify/require"
)

func TestLimits_perURL_contentLength(t *testing.T) {
	t.Parallel()
	var notif *limits.ImageLimitEvent
	notifier := notifierFunc(func(ctx context.Context, ev limits.ImageLimitEvent) error {
		notif = &ev
		return nil
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "9999")
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		// body not sent; client should reject before reading
	}))
	t.Cleanup(srv.Close)

	raw := []byte(fmt.Sprintf(`{"model":"m","messages":[{"role":"user","content":[{"type":"image_url","image_url":{"url":%q}}]}]}`, srv.URL+"/x"))
	next := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
		t.Fatal("next must not run")
		return nil, nil
	})

	eng, err := NewImageURLFetchEngine(ImageURLFetchConfig{
		Next:       next,
		HTTPClient: srv.Client(),
		Timeout:    5 * time.Second,
		Limits: limits.ImageURLFetchLimits{
			MaxBytesPerURL: 100,
			Notifier:       notifier,
		},
	})
	require.NoError(t, err)
	u := octollm.NewRequest(httptest.NewRequest(http.MethodPost, "/", nil), octollm.APIFormatChatCompletions)
	u.Body = octollm.NewBodyFromBytes(raw, testChatBodyParser)

	_, perr := eng.Process(u)
	require.Error(t, perr)
	require.True(t, errors.Is(perr, limits.ErrPerImageSizeExceeded), perr)
	require.NotNil(t, notif)
	require.Equal(t, limits.ImageLimitPerURL, notif.Kind)
	require.Equal(t, int64(100), notif.LimitBytes)
	require.Equal(t, int64(9999), notif.ActualBytes)
	require.Contains(t, notif.ImageURL, srv.URL)
}

func TestLimits_perURL_limitReader(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		// no Content-Length — chunked; body larger than limit
		for i := 0; i < 50; i++ {
			_, _ = w.Write([]byte("x"))
		}
	}))
	t.Cleanup(srv.Close)

	raw := []byte(fmt.Sprintf(`{"model":"m","messages":[{"role":"user","content":[{"type":"image_url","image_url":{"url":%q}}]}]}`, srv.URL+"/x"))
	next := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
		t.Fatal("next must not run")
		return nil, nil
	})

	eng, err := NewImageURLFetchEngine(ImageURLFetchConfig{
		Next:       next,
		HTTPClient: srv.Client(),
		Timeout:    5 * time.Second,
		Limits: limits.ImageURLFetchLimits{
			MaxBytesPerURL: 10,
		},
	})
	require.NoError(t, err)
	u := octollm.NewRequest(httptest.NewRequest(http.MethodPost, "/", nil), octollm.APIFormatChatCompletions)
	u.Body = octollm.NewBodyFromBytes(raw, testChatBodyParser)

	_, err = eng.Process(u)
	require.Error(t, err)
	require.True(t, errors.Is(err, limits.ErrPerImageSizeExceeded), err)
}

func TestLimits_perRequest(t *testing.T) {
	t.Parallel()
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(png)
	}))
	t.Cleanup(srv.Close)

	u1 := srv.URL + "/a.png"
	u2 := srv.URL + "/b.png"
	raw := []byte(fmt.Sprintf(
		`{"model":"m","messages":[{"role":"user","content":[{"type":"image_url","image_url":{"url":%q}},{"type":"image_url","image_url":{"url":%q}}]}]}`,
		u1, u2,
	))

	var notif *limits.ImageLimitEvent
	next := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
		t.Fatal("next must not run")
		return nil, nil
	})

	sum := int64(len(png) * 2)
	limit := sum - 1

	eng, err := NewImageURLFetchEngine(ImageURLFetchConfig{
		Next:       next,
		HTTPClient: srv.Client(),
		Timeout:    5 * time.Second,
		Limits: limits.ImageURLFetchLimits{
			MaxBytesPerRequest: limit,
			Notifier: notifierFunc(func(ctx context.Context, ev limits.ImageLimitEvent) error {
				notif = &ev
				return nil
			}),
		},
	})
	require.NoError(t, err)
	u := octollm.NewRequest(httptest.NewRequest(http.MethodPost, "/", nil), octollm.APIFormatChatCompletions)
	u.Body = octollm.NewBodyFromBytes(raw, testChatBodyParser)

	_, err = eng.Process(u)
	require.Error(t, err)
	require.True(t, errors.Is(err, limits.ErrTotalImageSizeExceeded), err)
	require.NotNil(t, notif)
	require.Equal(t, limits.ImageLimitPerRequest, notif.Kind)
	require.Equal(t, limit, notif.LimitBytes)
	require.Equal(t, sum, notif.ActualBytes)
}

type notifierFunc func(context.Context, limits.ImageLimitEvent) error

func (f notifierFunc) OnLimitExceeded(ctx context.Context, ev limits.ImageLimitEvent) error {
	return f(ctx, ev)
}
