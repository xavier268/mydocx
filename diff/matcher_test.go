package diff

import (
	"reflect"
	"testing"
)

// TestBasicEqual tests matching of identical sequences
func TestBasicEqual(t *testing.T) {
	a := []string{"hello", " ", "world"}
	b := []string{"hello", " ", "world"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	expected := []OpCode{
		{Tag: 'e', I1: 0, I2: 3, J1: 0, J2: 3},
	}

	if !reflect.DeepEqual(opcodes, expected) {
		t.Errorf("Expected %+v, got %+v", expected, opcodes)
	}
}

// TestBasicDelete tests deletion operations
func TestBasicDelete(t *testing.T) {
	a := []string{"hello", " ", "world", " ", "test"}
	b := []string{"hello", " ", "world"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	expected := []OpCode{
		{Tag: 'e', I1: 0, I2: 3, J1: 0, J2: 3},
		{Tag: 'd', I1: 3, I2: 5, J1: 3, J2: 3},
	}

	if !reflect.DeepEqual(opcodes, expected) {
		t.Errorf("Expected %+v, got %+v", expected, opcodes)
	}
}

// TestBasicInsert tests insertion operations
func TestBasicInsert(t *testing.T) {
	a := []string{"hello", " ", "world"}
	b := []string{"hello", " ", "world", " ", "test"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	expected := []OpCode{
		{Tag: 'e', I1: 0, I2: 3, J1: 0, J2: 3},
		{Tag: 'i', I1: 3, I2: 3, J1: 3, J2: 5},
	}

	if !reflect.DeepEqual(opcodes, expected) {
		t.Errorf("Expected %+v, got %+v", expected, opcodes)
	}
}

// TestBasicReplace tests replacement operations
func TestBasicReplace(t *testing.T) {
	a := []string{"hello", " ", "world"}
	b := []string{"hello", " ", "universe"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	expected := []OpCode{
		{Tag: 'e', I1: 0, I2: 2, J1: 0, J2: 2},
		{Tag: 'r', I1: 2, I2: 3, J1: 2, J2: 3},
	}

	if !reflect.DeepEqual(opcodes, expected) {
		t.Errorf("Expected %+v, got %+v", expected, opcodes)
	}
}

// TestEmptySequences tests handling of empty sequences
func TestEmptySequences(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []string
		expected []OpCode
	}{
		{
			name:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: []OpCode{},
		},
		{
			name: "a empty, b has content",
			a:    []string{},
			b:    []string{"hello"},
			expected: []OpCode{
				{Tag: 'i', I1: 0, I2: 0, J1: 0, J2: 1},
			},
		},
		{
			name: "a has content, b empty",
			a:    []string{"hello"},
			b:    []string{},
			expected: []OpCode{
				{Tag: 'd', I1: 0, I2: 1, J1: 0, J2: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewMatcher(tt.a, tt.b)
			opcodes := matcher.GetOpCodes()

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, opcodes)
			}
		})
	}
}

// TestComplexDiff tests a more complex diff scenario
func TestComplexDiff(t *testing.T) {
	a := []string{"The", " ", "quick", " ", "brown", " ", "fox"}
	b := []string{"The", " ", "fast", " ", "red", " ", "fox"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	// Should have:
	// - Equal: "The ",
	// - Replace: "quick" -> "fast"
	// - Equal: " "
	// - Replace: "brown" -> "red"
	// - Equal: " fox"

	if len(opcodes) == 0 {
		t.Fatal("No opcodes generated")
	}

	// Verify we have the right operations
	hasEqual := false
	hasReplace := false

	for _, op := range opcodes {
		switch op.Tag {
		case 'e':
			hasEqual = true
		case 'r':
			hasReplace = true
		}
	}

	if !hasEqual {
		t.Error("Expected at least one equal operation")
	}
	if !hasReplace {
		t.Error("Expected at least one replace operation")
	}

	t.Logf("Generated opcodes: %+v", opcodes)
}

// TestWordLevelReplacement tests the specific word replacement case
func TestWordLevelReplacement(t *testing.T) {
	a := []string{"will", " ", "be"}
	b := []string{"are"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	// Should replace "will be" with "are"
	if len(opcodes) != 1 {
		t.Fatalf("Expected 1 opcode, got %d: %+v", len(opcodes), opcodes)
	}

	op := opcodes[0]
	if op.Tag != 'r' {
		t.Errorf("Expected replace operation, got %c", op.Tag)
	}
	if op.I1 != 0 || op.I2 != 3 {
		t.Errorf("Expected I range [0:3], got [%d:%d]", op.I1, op.I2)
	}
	if op.J1 != 0 || op.J2 != 1 {
		t.Errorf("Expected J range [0:1], got [%d:%d]", op.J1, op.J2)
	}
}

// TestLongestCommonSubsequence tests LCS correctness with a known case
func TestLongestCommonSubsequence(t *testing.T) {
	// Classic LCS example: ABCDGH vs AEDFHR
	a := []string{"A", "B", "C", "D", "G", "H"}
	b := []string{"A", "E", "D", "F", "H", "R"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	// LCS should be "ADH" (length 3)
	// Verify that the operations make sense
	totalEqual := 0
	for _, op := range opcodes {
		if op.Tag == 'e' {
			totalEqual += op.I2 - op.I1
		}
	}

	if totalEqual != 3 {
		t.Errorf("Expected LCS length of 3, got equal operations totaling %d", totalEqual)
	}

	t.Logf("LCS test opcodes: %+v", opcodes)
}

// TestCaching tests that opcodes are cached after first computation
func TestCaching(t *testing.T) {
	a := []string{"hello", "world"}
	b := []string{"hello", "universe"}

	matcher := NewMatcher(a, b)

	// First call
	opcodes1 := matcher.GetOpCodes()

	// Second call should return cached result
	opcodes2 := matcher.GetOpCodes()

	// Should be identical (same memory reference for slices)
	if &opcodes1[0] != &opcodes2[0] {
		t.Error("Expected cached opcodes to be the same memory reference")
	}

	if !reflect.DeepEqual(opcodes1, opcodes2) {
		t.Error("Expected cached opcodes to be identical")
	}
}

// TestEdgeCasesSingleElements tests edge cases with single elements
func TestEdgeCasesSingleElements(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
	}{
		{"single to single different", []string{"a"}, []string{"b"}},
		{"single to single same", []string{"a"}, []string{"a"}},
		{"single to empty", []string{"a"}, []string{}},
		{"empty to single", []string{}, []string{"a"}},
		{"single to multiple", []string{"a"}, []string{"a", "b", "c"}},
		{"multiple to single", []string{"a", "b", "c"}, []string{"a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewMatcher(tt.a, tt.b)
			opcodes := matcher.GetOpCodes()

			// Basic sanity checks
			if len(opcodes) == 0 && (len(tt.a) > 0 || len(tt.b) > 0) {
				t.Error("Expected non-empty opcodes for non-identical sequences")
			}

			// Verify opcode integrity
			for i, op := range opcodes {
				if op.I1 < 0 || op.I2 < op.I1 || op.I2 > len(tt.a) {
					t.Errorf("Opcode %d has invalid I range: [%d:%d] for sequence of length %d",
						i, op.I1, op.I2, len(tt.a))
				}
				if op.J1 < 0 || op.J2 < op.J1 || op.J2 > len(tt.b) {
					t.Errorf("Opcode %d has invalid J range: [%d:%d] for sequence of length %d",
						i, op.J1, op.J2, len(tt.b))
				}
			}

			t.Logf("%s: %+v", tt.name, opcodes)
		})
	}
}

// TestOperationCoverage tests that opcodes cover the entire sequences
func TestOperationCoverage(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
	}{
		{"simple replace", []string{"old"}, []string{"new"}},
		{"complex mix", []string{"a", "b", "c", "d"}, []string{"a", "x", "c", "y", "z"}},
		{"all different", []string{"a", "b", "c"}, []string{"x", "y", "z"}},
		{"partial overlap", []string{"hello", "world", "test"}, []string{"hello", "universe", "test", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewMatcher(tt.a, tt.b)
			opcodes := matcher.GetOpCodes()

			// Check that opcodes cover the entire sequences
			lastI, lastJ := 0, 0
			for i, op := range opcodes {
				if op.I1 != lastI {
					t.Errorf("Opcode %d: I1=%d but expected %d (gap in coverage)", i, op.I1, lastI)
				}
				if op.J1 != lastJ {
					t.Errorf("Opcode %d: J1=%d but expected %d (gap in coverage)", i, op.J1, lastJ)
				}
				lastI, lastJ = op.I2, op.J2
			}

			if lastI != len(tt.a) {
				t.Errorf("Opcodes don't cover entire sequence A: ended at %d, length is %d", lastI, len(tt.a))
			}
			if lastJ != len(tt.b) {
				t.Errorf("Opcodes don't cover entire sequence B: ended at %d, length is %d", lastJ, len(tt.b))
			}
		})
	}
}

// TestRealWorldWordDiff tests with realistic word-level diffs
func TestRealWorldWordDiff(t *testing.T) {
	// Simulate the kind of word-level diff that would come from splitIntoWords
	a := []string{"The", " ", "quick", " ", "brown", " ", "fox", " ", "jumps", " ", "over", " ", "the", " ", "lazy", " ", "dog"}
	b := []string{"The", " ", "fast", " ", "red", " ", "cat", " ", "runs", " ", "under", " ", "the", " ", "sleepy", " ", "mouse"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	// Should preserve "The " and " the " as equal operations
	foundEqualThe := false
	foundEqualSpace := false

	for _, op := range opcodes {
		if op.Tag == 'e' {
			if op.I2-op.I1 > 0 && a[op.I1] == "The" {
				foundEqualThe = true
			}
			if op.I2-op.I1 > 0 && a[op.I1] == " " {
				foundEqualSpace = true
			}
		}
	}

	if !foundEqualThe {
		t.Error("Expected 'The' to be preserved as equal")
	}
	if !foundEqualSpace {
		t.Error("Expected spaces to be preserved as equal")
	}

	t.Logf("Real-world diff opcodes (%d total):", len(opcodes))
	for i, op := range opcodes {
		aSlice := ""
		bSlice := ""
		if op.I2 > op.I1 {
			aSlice = joinStrings(a[op.I1:op.I2])
		}
		if op.J2 > op.J1 {
			bSlice = joinStrings(b[op.J1:op.J2])
		}
		t.Logf("  %d: %c [%d:%d]->[%d:%d] '%s' -> '%s'",
			i, op.Tag, op.I1, op.I2, op.J1, op.J2, aSlice, bSlice)
	}
}

// BenchmarkLargeSequences benchmarks performance with larger sequences
func BenchmarkLargeSequences(b *testing.B) {
	// Create two large sequences with some differences
	size := 1000
	a := make([]string, size)
	seq_b := make([]string, size)

	for i := 0; i < size; i++ {
		if i%10 == 0 {
			a[i] = "word" + string(rune('A'+i%26))
			seq_b[i] = "different" + string(rune('A'+i%26))
		} else {
			a[i] = "common" + string(rune('A'+i%26))
			seq_b[i] = "common" + string(rune('A'+i%26))
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher := NewMatcher(a, seq_b)
		_ = matcher.GetOpCodes()
	}
}

// BenchmarkWorstCase benchmarks the worst case scenario (completely different sequences)
func BenchmarkWorstCase(b *testing.B) {
	size := 100
	a := make([]string, size)
	seq_b := make([]string, size)

	for i := 0; i < size; i++ {
		a[i] = "a" + string(rune(i))
		seq_b[i] = "b" + string(rune(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher := NewMatcher(a, seq_b)
		_ = matcher.GetOpCodes()
	}
}

// TestUnicodeAndMultibyteChars tests diff with Unicode and multibyte characters
func TestUnicodeAndMultibyteChars(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
	}{
		{
			name: "basic unicode replacement",
			a:    []string{"café", " ", "naïve"},
			b:    []string{"coffee", " ", "simple"},
		},
		{
			name: "emoji replacement",
			a:    []string{"Hello", " ", "🌍", " ", "world"},
			b:    []string{"Hello", " ", "🌎", " ", "world"},
		},
		{
			name: "mixed scripts",
			a:    []string{"English", " ", "中文", " ", "العربية"},
			b:    []string{"English", " ", "日本語", " ", "العربية"},
		},
		{
			name: "combining characters",
			a:    []string{"café"},       // é as single character
			b:    []string{"cafe\u0301"}, // e + combining acute accent
		},
		{
			name: "cyrillic text",
			a:    []string{"Привет", " ", "мир"},
			b:    []string{"Добро", " ", "пожаловать"},
		},
		{
			name: "mathematical symbols",
			a:    []string{"x", " ", "=", " ", "∑", "ᵢ", "aᵢ"},
			b:    []string{"y", " ", "=", " ", "∏", "ᵢ", "bᵢ"},
		},
		{
			name: "mixed emoji and text",
			a:    []string{"I", " ", "❤️", " ", "Go"},
			b:    []string{"I", " ", "💚", " ", "Python"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewMatcher(tt.a, tt.b)
			opcodes := matcher.GetOpCodes()

			// Basic integrity checks
			if len(opcodes) == 0 && !reflect.DeepEqual(tt.a, tt.b) {
				t.Error("Expected non-empty opcodes for different sequences")
			}

			// Verify opcode coverage
			lastI, lastJ := 0, 0
			for i, op := range opcodes {
				if op.I1 != lastI {
					t.Errorf("Opcode %d: I1=%d but expected %d (gap in coverage)", i, op.I1, lastI)
				}
				if op.J1 != lastJ {
					t.Errorf("Opcode %d: J1=%d but expected %d (gap in coverage)", i, op.J1, lastJ)
				}
				lastI, lastJ = op.I2, op.J2

				// Verify ranges are valid
				if op.I1 < 0 || op.I2 < op.I1 || op.I2 > len(tt.a) {
					t.Errorf("Opcode %d has invalid I range: [%d:%d] for sequence of length %d",
						i, op.I1, op.I2, len(tt.a))
				}
				if op.J1 < 0 || op.J2 < op.J1 || op.J2 > len(tt.b) {
					t.Errorf("Opcode %d has invalid J range: [%d:%d] for sequence of length %d",
						i, op.J1, op.J2, len(tt.b))
				}
			}

			if lastI != len(tt.a) {
				t.Errorf("Opcodes don't cover entire sequence A: ended at %d, length is %d", lastI, len(tt.a))
			}
			if lastJ != len(tt.b) {
				t.Errorf("Opcodes don't cover entire sequence B: ended at %d, length is %d", lastJ, len(tt.b))
			}

			// Log results for inspection
			t.Logf("Unicode test '%s' opcodes:", tt.name)
			for i, op := range opcodes {
				aSlice := ""
				bSlice := ""
				if op.I2 > op.I1 {
					aSlice = joinStrings(tt.a[op.I1:op.I2])
				}
				if op.J2 > op.J1 {
					bSlice = joinStrings(tt.b[op.J1:op.J2])
				}
				t.Logf("  %d: %c [%d:%d]->[%d:%d] '%s' -> '%s'",
					i, op.Tag, op.I1, op.I2, op.J1, op.J2, aSlice, bSlice)
			}
		})
	}
}

// TestUnicodeStringComparison tests that Unicode strings are compared correctly
func TestUnicodeStringComparison(t *testing.T) {
	// Test that our LCS algorithm properly handles Unicode string comparison
	a := []string{"résumé", " ", "naïve", " ", "café"}
	b := []string{"resume", " ", "naive", " ", "coffee"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	// Should preserve the spaces as equal
	foundEqualSpace := false
	for _, op := range opcodes {
		if op.Tag == 'e' && op.I2 > op.I1 && a[op.I1] == " " {
			foundEqualSpace = true
			break
		}
	}

	if !foundEqualSpace {
		t.Error("Expected spaces to be preserved as equal operations")
	}

	// Should show replacements for the accented words
	foundReplacements := 0
	for _, op := range opcodes {
		if op.Tag == 'r' {
			foundReplacements++
		}
	}

	if foundReplacements == 0 {
		t.Error("Expected at least one replacement operation for accented words")
	}

	t.Logf("Unicode comparison opcodes:")
	for i, op := range opcodes {
		aSlice := ""
		bSlice := ""
		if op.I2 > op.I1 {
			aSlice = joinStrings(a[op.I1:op.I2])
		}
		if op.J2 > op.J1 {
			bSlice = joinStrings(b[op.J1:op.J2])
		}
		t.Logf("  %d: %c '%s' -> '%s'", i, op.Tag, aSlice, bSlice)
	}
}

// TestCJKCharacters tests diff with Chinese, Japanese, and Korean characters
func TestCJKCharacters(t *testing.T) {
	a := []string{"你好", "世界", "こんにちは", "안녕하세요"}
	b := []string{"您好", "世界", "さようなら", "안녕히가세요"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	// "世界" should remain equal
	foundEqual := false
	for _, op := range opcodes {
		if op.Tag == 'e' && op.I2 > op.I1 {
			if a[op.I1] == "世界" {
				foundEqual = true
				break
			}
		}
	}

	if !foundEqual {
		t.Error("Expected '世界' to be found as equal")
	}

	t.Logf("CJK characters diff:")
	for i, op := range opcodes {
		aSlice := ""
		bSlice := ""
		if op.I2 > op.I1 {
			aSlice = joinStrings(a[op.I1:op.I2])
		}
		if op.J2 > op.J1 {
			bSlice = joinStrings(b[op.J1:op.J2])
		}
		t.Logf("  %d: %c '%s' -> '%s'", i, op.Tag, aSlice, bSlice)
	}
}

// TestEmojiSequences tests diff with complex emoji sequences
func TestEmojiSequences(t *testing.T) {
	// Test with complex emoji including skin tone modifiers and ZWJ sequences
	a := []string{"👨‍👩‍👧‍👦", " ", "🏳️‍🌈", " ", "👍🏽"}
	b := []string{"👨‍👩‍👧‍👧", " ", "🏳️‍⚧️", " ", "👎🏽"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	// Space should be preserved
	foundEqualSpace := false
	for _, op := range opcodes {
		if op.Tag == 'e' && op.I2 > op.I1 && a[op.I1] == " " {
			foundEqualSpace = true
			break
		}
	}

	if !foundEqualSpace {
		t.Error("Expected spaces to be preserved as equal")
	}

	t.Logf("Complex emoji sequences diff:")
	for i, op := range opcodes {
		aSlice := ""
		bSlice := ""
		if op.I2 > op.I1 {
			aSlice = joinStrings(a[op.I1:op.I2])
		}
		if op.J2 > op.J1 {
			bSlice = joinStrings(b[op.J1:op.J2])
		}
		t.Logf("  %d: %c '%s' -> '%s'", i, op.Tag, aSlice, bSlice)
	}
}

// TestMixedScriptsAndDirections tests mixed writing directions
func TestMixedScriptsAndDirections(t *testing.T) {
	// Mix of LTR (Latin, Cyrillic), RTL (Arabic, Hebrew), and neutral
	a := []string{"Hello", " ", "مرحبا", " ", "שלום", " ", "Привет"}
	b := []string{"Goodbye", " ", "مع السلامة", " ", "להתראות", " ", "До свидания"}

	matcher := NewMatcher(a, b)
	opcodes := matcher.GetOpCodes()

	// Spaces should be preserved
	spaceCount := 0
	for _, op := range opcodes {
		if op.Tag == 'e' && op.I2 > op.I1 {
			for i := op.I1; i < op.I2; i++ {
				if a[i] == " " {
					spaceCount++
				}
			}
		}
	}

	if spaceCount == 0 {
		t.Error("Expected at least one space to be preserved as equal")
	}

	t.Logf("Mixed scripts and directions diff:")
	for i, op := range opcodes {
		aSlice := ""
		bSlice := ""
		if op.I2 > op.I1 {
			aSlice = joinStrings(a[op.I1:op.I2])
		}
		if op.J2 > op.J1 {
			bSlice = joinStrings(b[op.J1:op.J2])
		}
		t.Logf("  %d: %c '%s' -> '%s'", i, op.Tag, aSlice, bSlice)
	}
}

// BenchmarkUnicodeSequences benchmarks performance with Unicode content
func BenchmarkUnicodeSequences(b *testing.B) {
	// Create sequences with mixed Unicode content
	size := 100
	a := make([]string, size)
	seq_b := make([]string, size)

	unicodeWords := []string{"café", "naïve", "résumé", "🌍", "你好", "世界", "مرحبا", "שלום", "Привет"}

	for i := 0; i < size; i++ {
		a[i] = unicodeWords[i%len(unicodeWords)] + string(rune(i))
		if i%7 == 0 {
			seq_b[i] = "different" + string(rune(i))
		} else {
			seq_b[i] = a[i] // Same as a[i]
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher := NewMatcher(a, seq_b)
		_ = matcher.GetOpCodes()
	}
}

// Helper function to join strings for display
func joinStrings(strs []string) string {
	result := ""
	for _, s := range strs {
		result += s
	}
	return result
}
