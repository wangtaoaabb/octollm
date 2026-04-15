// Package readbody holds shared helpers for HTTP image responses: Content-Length parsing,
// image/* MIME normalization, and size-limited body reads.
package readbody

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

// ParseContentLength returns the declared body size when known; otherwise -1.
func ParseContentLength(resp *http.Response) int64 {
	if resp == nil {
		return -1
	}
	if resp.ContentLength >= 0 {
		return resp.ContentLength
	}
	if hdr := resp.Header.Get("Content-Length"); hdr != "" {
		if n, err := strconv.ParseInt(strings.TrimSpace(hdr), 10, 64); err == nil {
			return n
		}
	}
	return -1
}

// NormalizeImageContentType returns a primary MIME type for inline image data (defaults to image/jpeg).
func NormalizeImageContentType(raw string) string {
	ct := raw
	if !strings.HasPrefix(ct, "image/") {
		ct = "image/jpeg"
	}
	if idx := strings.Index(ct, ";"); idx > 0 {
		ct = strings.TrimSpace(ct[:idx])
	}
	return ct
}

// ReadLimited reads r. When max > 0, at most max+1 bytes are read so callers can detect overflow.
func ReadLimited(r io.Reader, max int64) ([]byte, error) {
	if max > 0 {
		return io.ReadAll(io.LimitReader(r, max+1))
	}
	return io.ReadAll(r)
}
