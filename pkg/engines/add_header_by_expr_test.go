package engines

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/infinigence/octollm/pkg/exprenv"
	"github.com/infinigence/octollm/pkg/octollm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAddHeaderByExprEngine(t *testing.T) {
	mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
		return &octollm.Response{
			StatusCode: 200,
			Header:     make(http.Header),
		}, nil
	})

	t.Run("empty expressions", func(t *testing.T) {
		engine, err := NewAddHeaderByExprEngine(map[string]string{}, mockEngine)
		require.NoError(t, err)
		assert.NotNil(t, engine)
		assert.Len(t, engine.HeaderExprs, 0)
	})

	t.Run("valid expressions", func(t *testing.T) {
		exprMap := map[string]string{
			"X-Test-Header": `"test-value"`,
		}
		engine, err := NewAddHeaderByExprEngine(exprMap, mockEngine)
		require.NoError(t, err)
		assert.NotNil(t, engine)
		assert.Len(t, engine.HeaderExprs, 1)
	})

	t.Run("invalid expression", func(t *testing.T) {
		exprMap := map[string]string{
			"X-Test-Header": `invalid syntax here ((`,
		}
		engine, err := NewAddHeaderByExprEngine(exprMap, mockEngine)
		assert.Error(t, err)
		assert.Nil(t, engine)
		assert.Contains(t, err.Error(), "failed to compile expression")
	})
}

func TestAddHeaderByExprEngine_Process(t *testing.T) {
	t.Run("add header from literal", func(t *testing.T) {
		mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
			// Verify header was added
			assert.Equal(t, "test-value", req.Header.Get("X-Test-Header"))
			return &octollm.Response{
				StatusCode: 200,
				Header:     make(http.Header),
			}, nil
		})

		exprMap := map[string]string{
			"X-Test-Header": `"test-value"`,
		}
		engine, err := NewAddHeaderByExprEngine(exprMap, mockEngine)
		require.NoError(t, err)

		httpReq, _ := http.NewRequest("POST", "http://test", nil)
		req := octollm.NewRequest(httpReq, octollm.APIFormatChatCompletions)

		_, err = engine.Process(req)
		x_test_header := req.Header.Get("X-Test-Header")
		fmt.Println("x_test_header is:", x_test_header)
		assert.NoError(t, err)
		assert.Equal(t, "test-value", x_test_header)
	})

	t.Run("add header from incoming header", func(t *testing.T) {
		mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
			assert.Equal(t, "tenant-123", req.Header.Get("X-Forwarded-Tenant"))
			return &octollm.Response{
				StatusCode: 200,
				Header:     make(http.Header),
			}, nil
		})

		exprMap := map[string]string{
			"X-Forwarded-Tenant": `req.Header("X-Tenant-ID")`,
		}
		engine, err := NewAddHeaderByExprEngine(exprMap, mockEngine)
		require.NoError(t, err)

		httpReq, _ := http.NewRequest("POST", "http://test", nil)
		httpReq.Header.Set("X-Tenant-ID", "tenant-123")

		ctx := context.WithValue(context.Background(), octollm.ContextKeyReceivedHeader, httpReq.Header)
		httpReq = httpReq.WithContext(ctx)

		req := octollm.NewRequest(httpReq, octollm.APIFormatChatCompletions)

		_, err = engine.Process(req)
		assert.NoError(t, err)
	})

	t.Run("add header from feature extractor", func(t *testing.T) {
		mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
			return &octollm.Response{
				StatusCode: 200,
				Header:     make(http.Header),
			}, nil
		})

		exprMap := map[string]string{
			"X-Feature": `req.Feature("prefix20")`,
		}
		engine, err := NewAddHeaderByExprEngine(exprMap, mockEngine)
		require.NoError(t, err)

		httpReq, _ := http.NewRequest("POST", "http://test", nil)
		req := octollm.NewRequest(httpReq, octollm.APIFormatChatCompletions)

		_, err = engine.Process(req)
		assert.NoError(t, err)
		// When feature extractor returns nil, it gets formatted as "<nil>"
		assert.Equal(t, "<nil>", req.Header.Get("X-Feature"))

		exprMap = map[string]string{
			"X-Feature": `req.Feature("prefix20") != nil ? req.Feature("prefix20") : "default-value"`,
		}
		engine, err = NewAddHeaderByExprEngine(exprMap, mockEngine)
		require.NoError(t, err)

		_, err = engine.Process(req)
		assert.NoError(t, err)
		assert.Equal(t, "default-value", req.Header.Get("X-Feature"))
	})

	t.Run("add header from feature extractor - with extractor", func(t *testing.T) {
		exprenv.RegisterDefaultExtractor("user_id", exprenv.FeatureExtractorFunc(func(*octollm.Request) (any, error) {
			return "user-abc-123", nil
		}))
		defer exprenv.UnregisterDefaultExtractor("user_id")

		mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
			assert.Equal(t, "user-abc-123", req.Header.Get("X-User-ID"))
			return &octollm.Response{
				StatusCode: 200,
				Header:     make(http.Header),
			}, nil
		})

		exprMap := map[string]string{
			"X-User-ID": `req.Feature("user_id")`,
		}
		engine, err := NewAddHeaderByExprEngine(exprMap, mockEngine)
		require.NoError(t, err)

		httpReq, _ := http.NewRequest("POST", "http://test", nil)
		req := octollm.NewRequest(httpReq, octollm.APIFormatChatCompletions)

		_, err = engine.Process(req)
		assert.NoError(t, err)
		assert.Equal(t, "user-abc-123", req.Header.Get("X-User-ID"))
	})

	t.Run("override existing header", func(t *testing.T) {
		mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
			assert.Equal(t, "new-value", req.Header.Get("X-Override"))
			return &octollm.Response{
				StatusCode: 200,
				Header:     make(http.Header),
			}, nil
		})

		exprMap := map[string]string{
			"X-Override": `"new-value"`,
		}
		engine, err := NewAddHeaderByExprEngine(exprMap, mockEngine)
		require.NoError(t, err)

		httpReq, _ := http.NewRequest("POST", "http://test", nil)
		httpReq.Header.Set("X-Override", "old-value")
		req := octollm.NewRequest(httpReq, octollm.APIFormatChatCompletions)

		_, err = engine.Process(req)
		assert.NoError(t, err)
	})

	t.Run("access context value", func(t *testing.T) {
		mockEngine := octollm.EngineFunc(func(req *octollm.Request) (*octollm.Response, error) {
			assert.Equal(t, "org-123", req.Header.Get("X-Org-ID"))
			return &octollm.Response{
				StatusCode: 200,
				Header:     make(http.Header),
			}, nil
		})

		exprMap := map[string]string{
			"X-Org-ID": `req.Context("org_id")`,
		}
		engine, err := NewAddHeaderByExprEngine(exprMap, mockEngine)
		require.NoError(t, err)

		httpReq, _ := http.NewRequest("POST", "http://test", nil)
		ctx := context.WithValue(context.Background(), "org_id", "org-123")
		httpReq = httpReq.WithContext(ctx)
		req := octollm.NewRequest(httpReq, octollm.APIFormatChatCompletions)

		_, err = engine.Process(req)
		assert.NoError(t, err)
	})
}
