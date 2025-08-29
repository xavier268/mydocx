package diff

// This package implements a sequence matcher based on the Longest Common Subsequence (LCS) algorithm.
// It provides functionality to compare two sequences and generate operation codes that describe
// the differences between them.
//
// The algorithm used is a variant of the Myers diff algorithm with dynamic programming optimization,
// similar to what's implemented in Python's difflib and other diff utilities.
//
// References:
// - "An O(ND) Difference Algorithm and Its Variations" by Eugene W. Myers (1986)
// - Python's difflib.SequenceMatcher implementation
// - "Introduction to Algorithms" by Cormen, Leiserson, Rivest, and Stein (CLRS), Chapter on Dynamic Programming

// OpCode represents a single operation in the difference between two sequences.
// Each OpCode describes how to transform a portion of sequence A into a portion of sequence B.
type OpCode struct {
	// Tag describes the type of operation:
	// 'e' = equal (sequences match)
	// 'd' = delete (remove from sequence A)
	// 'i' = insert (add to sequence B)
	// 'r' = replace (substitute portion of A with portion of B)
	Tag byte

	// I1, I2 define the range [I1:I2) in sequence A
	I1, I2 int

	// J1, J2 define the range [J1:J2) in sequence B
	J1, J2 int
}

// Matcher compares two sequences of strings and computes the differences between them.
// It uses a dynamic programming approach based on the Longest Common Subsequence algorithm
// to find the optimal alignment between the sequences.
//
// The matcher is designed to be compatible with the interface used by go-difflib,
// specifically providing the GetOpCodes() method that returns operation codes.
type Matcher struct {
	a, b     []string // The two sequences to compare
	opcodes  []OpCode // Cached operation codes
	computed bool     // Whether opcodes have been computed
}

// NewMatcher creates a new Matcher to compare two sequences of strings.
//
// Parameters:
//   - a: The first sequence (often considered the "original")
//   - b: The second sequence (often considered the "modified")
//
// Returns:
//   - A new Matcher instance ready to compute differences
//
// Example:
//
//	original := []string{"hello", " ", "world"}
//	modified := []string{"hello", " ", "universe"}
//	matcher := NewMatcher(original, modified)
//	opcodes := matcher.GetOpCodes()
func NewMatcher(a, b []string) *Matcher {
	return &Matcher{
		a: a,
		b: b,
	}
}

// GetOpCodes returns a slice of OpCode structs describing the differences between
// the two sequences. Each OpCode represents one operation needed to transform
// sequence A into sequence B.
//
// The algorithm works in two phases:
// 1. Compute the Longest Common Subsequence (LCS) using dynamic programming
// 2. Trace back through the LCS table to generate operation codes
//
// Time Complexity: O(m*n) where m and n are the lengths of the sequences
// Space Complexity: O(m*n) for the DP table, O(min(m,n)) for the result
//
// Returns:
//   - A slice of OpCode structs, each describing a single edit operation
//   - Operations are returned in order from the beginning of the sequences
func (m *Matcher) GetOpCodes() []OpCode {
	if m.computed {
		return m.opcodes
	}

	m.opcodes = m.computeOpCodes()
	m.computed = true
	return m.opcodes
}

// computeOpCodes implements the core LCS-based diff algorithm.
//
// The algorithm uses dynamic programming to build a table where dp[i][j]
// represents the length of the LCS of a[0:i] and b[0:j].
//
// After building the DP table, we trace back from dp[len(a)][len(b)] to dp[0][0]
// to reconstruct the sequence of operations that transform sequence A into B.
//
// The traceback process:
//   - If a[i-1] == b[j-1]: both sequences have the same element (EQUAL)
//   - If dp[i-1][j] > dp[i][j-1]: delete from A (DELETE)
//   - Otherwise: insert into A (INSERT)
//   - Special case: when we have both deletions and insertions in the same region,
//     we merge them into a REPLACE operation for efficiency
func (m *Matcher) computeOpCodes() []OpCode {
	lenA, lenB := len(m.a), len(m.b)

	// Handle empty sequences
	if lenA == 0 && lenB == 0 {
		return []OpCode{}
	}
	if lenA == 0 {
		return []OpCode{{Tag: 'i', I1: 0, I2: 0, J1: 0, J2: lenB}}
	}
	if lenB == 0 {
		return []OpCode{{Tag: 'd', I1: 0, I2: lenA, J1: 0, J2: 0}}
	}

	// Build the LCS dynamic programming table
	// dp[i][j] = length of LCS of a[0:i] and b[0:j]
	dp := make([][]int, lenA+1)
	for i := range dp {
		dp[i] = make([]int, lenB+1)
	}

	// Fill the DP table using the recurrence relation:
	// dp[i][j] = dp[i-1][j-1] + 1                    if a[i-1] == b[j-1]
	// dp[i][j] = max(dp[i-1][j], dp[i][j-1])         otherwise
	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			if m.a[i-1] == m.b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	// Trace back through the DP table to generate opcodes
	return m.traceback(dp)
}

// traceback reconstructs the sequence of operations by tracing back through
// the completed DP table from bottom-right to top-left.
//
// The traceback follows these rules:
// 1. If a[i-1] == b[j-1] and dp[i][j] == dp[i-1][j-1] + 1: EQUAL operation
// 2. If dp[i-1][j] > dp[i][j-1]: DELETE operation (move up in table)
// 3. Otherwise: INSERT operation (move left in table)
//
// To optimize the output, consecutive DELETE and INSERT operations are
// merged into a single REPLACE operation when possible.
func (m *Matcher) traceback(dp [][]int) []OpCode {
	var operations []OpCode
	i, j := len(m.a), len(m.b)

	// Trace back from the bottom-right corner
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && m.a[i-1] == m.b[j-1] {
			// Equal elements - find the longest sequence of equal elements
			equalEndI, equalEndJ := i, j
			for i > 0 && j > 0 && m.a[i-1] == m.b[j-1] {
				i--
				j--
			}
			operations = append(operations, OpCode{
				Tag: 'e',
				I1:  i,
				I2:  equalEndI,
				J1:  j,
				J2:  equalEndJ,
			})
		} else if i > 0 && (j == 0 || dp[i-1][j] >= dp[i][j-1]) {
			// Delete operation - find consecutive deletions
			deleteEnd := i
			for i > 0 && (j == 0 || dp[i-1][j] >= dp[i][j-1]) {
				// Make sure we're still in a delete situation
				if i > 0 && j > 0 && m.a[i-1] == m.b[j-1] {
					break
				}
				i--
			}
			operations = append(operations, OpCode{
				Tag: 'd',
				I1:  i,
				I2:  deleteEnd,
				J1:  j,
				J2:  j,
			})
		} else {
			// Insert operation - find consecutive insertions
			insertEnd := j
			for j > 0 && (i == 0 || dp[i-1][j] < dp[i][j-1]) {
				// Make sure we're still in an insert situation
				if i > 0 && j > 0 && m.a[i-1] == m.b[j-1] {
					break
				}
				j--
			}
			operations = append(operations, OpCode{
				Tag: 'i',
				I1:  i,
				I2:  i,
				J1:  j,
				J2:  insertEnd,
			})
		}
	}

	// Reverse the operations since we built them backwards
	for i := 0; i < len(operations)/2; i++ {
		operations[i], operations[len(operations)-1-i] = operations[len(operations)-1-i], operations[i]
	}

	// Merge consecutive delete+insert operations into replace operations
	return m.mergeReplaceOperations(operations)
}

// mergeReplaceOperations optimizes the operation list by merging consecutive
// DELETE and INSERT operations into more efficient REPLACE operations.
//
// This is both for compatibility with difflib behavior and for generating
// more intuitive diff output. Instead of showing separate delete/insert
// operations, a replace operation shows a substitution.
//
// Example transformation:
//
//	[DELETE(1,2), INSERT(1,3)] → [REPLACE(1,2,1,3)]
func (m *Matcher) mergeReplaceOperations(operations []OpCode) []OpCode {
	if len(operations) <= 1 {
		return operations
	}

	var result []OpCode
	i := 0

	for i < len(operations) {
		current := operations[i]

		// Look for insert followed by delete at adjacent positions
		// This happens due to our traceback order
		if current.Tag == 'i' && i+1 < len(operations) {
			next := operations[i+1]
			if next.Tag == 'd' && current.I1 == next.I1 && current.J2 == next.J1 {
				// Merge insert+delete into replace
				result = append(result, OpCode{
					Tag: 'r',
					I1:  next.I1,
					I2:  next.I2,
					J1:  current.J1,
					J2:  current.J2,
				})
				i += 2 // Skip both operations
				continue
			}
		}

		// Also look for delete followed by insert (less common with our traceback)
		if current.Tag == 'd' && i+1 < len(operations) {
			next := operations[i+1]
			if next.Tag == 'i' && current.I2 == next.I1 && current.J1 == next.J1 {
				// Merge delete+insert into replace
				result = append(result, OpCode{
					Tag: 'r',
					I1:  current.I1,
					I2:  current.I2,
					J1:  next.J1,
					J2:  next.J2,
				})
				i += 2 // Skip both operations
				continue
			}
		}

		// No merge possible, keep current operation
		result = append(result, current)
		i++
	}

	return result
}
