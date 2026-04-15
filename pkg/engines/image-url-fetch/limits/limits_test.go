package limits

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()
	require.True(t, errors.Is(ErrPerImageSizeExceeded, ErrPerImageSizeExceeded))
	require.False(t, errors.Is(ErrPerImageSizeExceeded, ErrTotalImageSizeExceeded))
}
