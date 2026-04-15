package readbody

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseContentLength(t *testing.T) {
	t.Parallel()
	t.Run("from struct", func(t *testing.T) {
		r := &http.Response{ContentLength: 42}
		require.Equal(t, int64(42), ParseContentLength(r))
	})
	t.Run("from header", func(t *testing.T) {
		r := &http.Response{
			ContentLength: -1,
			Header:        http.Header{"Content-Length": []string{" 99 "}},
		}
		require.Equal(t, int64(99), ParseContentLength(r))
	})
	t.Run("unknown", func(t *testing.T) {
		r := &http.Response{ContentLength: -1,
			Header: http.Header{},
		}
		require.Equal(t, int64(-1), ParseContentLength(r))
	})
}

func TestNormalizeImageContentType(t *testing.T) {
	t.Parallel()
	require.Equal(t, "image/png", NormalizeImageContentType("image/png; charset=binary"))
	require.Equal(t, "image/jpeg", NormalizeImageContentType("text/plain"))
}

func TestReadLimited(t *testing.T) {
	t.Parallel()
	data, err := ReadLimited(strings.NewReader("hello"), 10)
	require.NoError(t, err)
	require.Equal(t, "hello", string(data))

	large := strings.Repeat("x", 100)
	data, err = ReadLimited(strings.NewReader(large), 10)
	require.NoError(t, err)
	require.Len(t, data, 11)
}
