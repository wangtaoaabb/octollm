package repeat_detector

import (
	"strings"
)

// ExtractRepeatPattern extract the repeat pattern and repeat count
// return: repeat pattern, repeat count, whether found
func ExtractRepeatPattern(text string, minRepeatLen, maxRepeatLen, repeatThreshold int) (pattern string, repeatCount int, found bool) {
	text = strings.TrimSpace(text)
	runes := []rune(text)
	n := len(runes)

	if n < minRepeatLen*repeatThreshold {
		return "", 0, false
	}

	return extractRepeatPattern(runes, minRepeatLen, maxRepeatLen, repeatThreshold)
}

// extractRepeatPattern use suffix comparison to detect repeat
func extractRepeatPattern(runes []rune, minRepeatLen, maxRepeatLen, repeatThreshold int) (pattern string, repeatCount int, found bool) {
	n := len(runes)
	if n < minRepeatLen*repeatThreshold {
		return "", 0, false
	}

	// calculate the upper limit of pattern length
	maxPatternLen := n / repeatThreshold
	if maxRepeatLen > 0 && maxRepeatLen < maxPatternLen {
		maxPatternLen = maxRepeatLen
	}

	// check different lengths of pattern (from small to large)
	for patternLen := minRepeatLen; patternLen <= maxPatternLen; patternLen++ {
		total := repeatThreshold * patternLen
		if total > n {
			continue
		}

		// get the last repeatThreshold * patternLen length fragment
		tail := runes[n-total:]

		// core judgment: if the tail is equal to the tail after removing the first pattern,
		// then the entire tail is repeated repeatThreshold times
		if runeSliceEqual(tail[patternLen:], tail[:len(tail)-patternLen]) {
			patternRunes := runes[n-patternLen:]
			return string(patternRunes), repeatThreshold, true
		}
	}

	return "", 0, false
}

// runeSliceEqual exactly compare two rune slices
func runeSliceEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
