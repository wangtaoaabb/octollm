package repeat_detector

import (
	"strings"
	"testing"
)

func TestExtractRepeatPattern(t *testing.T) {
	tests := []struct {
		name            string
		text            string
		minRepeatLen    int
		maxRepeatLen    int
		repeatThreshold int
		wantPattern     string
		wantCount       int
		wantFound       bool
	}{
		{
			name:            "ABC repeated 60 times",
			text:            strings.Repeat("ABC", 60),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "ABC",
			wantCount:       50, // 达到阈值即返回
			wantFound:       true,
		},
		{
			name:            "Single char repeated 100 times",
			text:            strings.Repeat("A", 100),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "A",
			wantCount:       50, // 达到阈值即返回
			wantFound:       true,
		},
		{
			name:            "Long pattern repeated",
			text:            strings.Repeat("Hello", 60),
			minRepeatLen:    1,
			maxRepeatLen:    10,
			repeatThreshold: 50,
			wantPattern:     "Hello",
			wantCount:       50, // 达到阈值即返回
			wantFound:       true,
		},
		{
			name:            "Pattern at exact threshold",
			text:            strings.Repeat("XY", 50),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "XY",
			wantCount:       50,
			wantFound:       true,
		},
		{
			name:            "Pattern below threshold",
			text:            strings.Repeat("ABC", 49),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "",
			wantCount:       0,
			wantFound:       false,
		},
		{
			name:            "No repetition",
			text:            "This is a normal sentence without repetition." + strings.Repeat("ABC", 60) + "This is a normal sentence without repetition.",
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "",
			wantCount:       0,
			wantFound:       false,
		},
		{
			name:            "Repetition with prefix",
			text:            "Some prefix text " + strings.Repeat("ABC", 60),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "ABC",
			wantCount:       50, // 达到阈值即返回
			wantFound:       true,
		},
		{
			name:            "Unicode characters",
			text:            strings.Repeat("你好", 60),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "你好",
			wantCount:       50, // 达到阈值即返回
			wantFound:       true,
		},
		{
			name:            "Mixed pattern with spaces",
			text:            strings.Repeat("ABC ", 60),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     " ABC", // 周期性字符串，不严格匹配 pattern
			wantCount:       50,     // 达到阈值即返回
			wantFound:       true,
		},
		{
			name:            "Pattern longer than maxRepeatLen",
			text:            strings.Repeat("ABCDEFGH", 60),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "",
			wantCount:       0,
			wantFound:       false,
		},
		{
			name:            "Empty text",
			text:            "",
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "",
			wantCount:       0,
			wantFound:       false,
		},
		{
			name:            "Text too short for threshold",
			text:            "ABC",
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "",
			wantCount:       0,
			wantFound:       false,
		},
		{
			name:            "Interrupted repetition",
			text:            strings.Repeat("ABC", 30) + "XYZ" + strings.Repeat("ABC", 30),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "",
			wantCount:       0,
			wantFound:       false,
		},
		{
			name:            "Two char pattern repeated",
			text:            strings.Repeat("AB", 60),
			minRepeatLen:    2,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "AB",
			wantCount:       50, // 达到阈值即返回
			wantFound:       true,
		},
		{
			name:            "Pattern with newlines",
			text:            strings.Repeat("AB\n", 60),
			minRepeatLen:    1,
			maxRepeatLen:    5,
			repeatThreshold: 50,
			wantPattern:     "\nAB", // TrimSpace 后末尾变成 "AB"，所以 pattern 是 "\nAB"
			wantCount:       50,     // 达到阈值即返回
			wantFound:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPattern, gotCount, gotFound := ExtractRepeatPattern(
				tt.text,
				tt.minRepeatLen,
				tt.maxRepeatLen,
				tt.repeatThreshold,
			)

			if gotFound != tt.wantFound {
				t.Errorf("ExtractRepeatPattern() found = %v, want %v", gotFound, tt.wantFound)
			}

			// 如果 wantPattern 为空，说明不检查具体 pattern，只检查是否找到
			if tt.wantPattern != "" && gotPattern != tt.wantPattern {
				t.Errorf("ExtractRepeatPattern() pattern = %q, want %q", gotPattern, tt.wantPattern)
			}

			if gotCount != tt.wantCount {
				t.Errorf("ExtractRepeatPattern() count = %v, want %v", gotCount, tt.wantCount)
			}
		})
	}
}

func TestRuneSliceEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []rune
		b    []rune
		want bool
	}{
		{
			name: "Identical slices",
			a:    []rune("ABC"),
			b:    []rune("ABC"),
			want: true,
		},
		{
			name: "Different content",
			a:    []rune("ABC"),
			b:    []rune("XYZ"),
			want: false,
		},
		{
			name: "Different length",
			a:    []rune("ABC"),
			b:    []rune("AB"),
			want: false,
		},
		{
			name: "Empty slices",
			a:    []rune{},
			b:    []rune{},
			want: true,
		},
		{
			name: "Unicode match",
			a:    []rune("你好"),
			b:    []rune("你好"),
			want: true,
		},
		{
			name: "Unicode mismatch",
			a:    []rune("你好"),
			b:    []rune("世界"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runeSliceEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("runeSliceEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests，benchmark实测性能为微妙级别
func BenchmarkExtractRepeatPattern(b *testing.B) {
	text := strings.Repeat("ABC", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractRepeatPattern(text, 1, 5, 50)
	}
}

func BenchmarkExtractRepeatPatternLongPattern(b *testing.B) {
	text := strings.Repeat("HelloWorld", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractRepeatPattern(text, 1, 20, 50)
	}
}

func BenchmarkExtractRepeatPatternNoMatch(b *testing.B) {
	text := "This is a long text without any repetition patterns that meet the threshold requirements."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractRepeatPattern(text, 1, 5, 50)
	}
}

// Test algorithm correctness with various edge cases
func TestExtractRepeatPatternEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		minLen      int
		maxLen      int
		threshold   int
		shouldFind  bool
		description string
	}{
		{
			name:        "Exact threshold boundary",
			text:        strings.Repeat("X", 50),
			minLen:      1,
			maxLen:      1,
			threshold:   50,
			shouldFind:  true,
			description: "Should detect when count exactly equals threshold",
		},
		{
			name:        "Just below threshold",
			text:        strings.Repeat("X", 49),
			minLen:      1,
			maxLen:      1,
			threshold:   50,
			shouldFind:  false,
			description: "Should not detect when count is below threshold",
		},
		{
			name:        "Pattern at maxLen boundary",
			text:        strings.Repeat("ABCDE", 60),
			minLen:      1,
			maxLen:      5,
			threshold:   50,
			shouldFind:  true,
			description: "Should detect pattern exactly at maxLen",
		},
		{
			name:        "Pattern beyond maxLen",
			text:        strings.Repeat("ABCDEF", 60),
			minLen:      1,
			maxLen:      5,
			threshold:   50,
			shouldFind:  false,
			description: "Should not detect pattern longer than maxLen",
		},
		{
			name:        "Whitespace trimming",
			text:        "   " + strings.Repeat("ABC", 60) + "   ",
			minLen:      1,
			maxLen:      5,
			threshold:   50,
			shouldFind:  true,
			description: "Should trim whitespace before detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, found := ExtractRepeatPattern(tt.text, tt.minLen, tt.maxLen, tt.threshold)
			if found != tt.shouldFind {
				t.Errorf("%s: found = %v, want %v", tt.description, found, tt.shouldFind)
			}
		})
	}
}
