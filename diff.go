package mydocx

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// DiffOperation represents a single diff operation
type DiffOperation struct {
	Type      string // "equal", "delete", "insert"
	Text      string
	Container string
	Paragraph int
}

// DiffOpType represents the type of diff operation
type DiffOpType string

const (
	DiffEqual  DiffOpType = "equal"
	DiffDelete DiffOpType = "delete"
	DiffInsert DiffOpType = "insert"
)

// InternalDiff represents a diff operation used internally
type InternalDiff struct {
	Type DiffOpType
	Text string
}

// ContainerDiff represents differences in a single container
type ContainerDiff struct {
	Operations []DiffOperation
}

// DiffResult represents the complete diff between original and accepted text
type DiffResult struct {
	ContainerDiffs map[string]ContainerDiff
	Summary        DiffSummary
}

// DiffSummary provides high-level statistics about the diff
type DiffSummary struct {
	TotalContainers   int
	ChangedContainers int
	TotalInsertions   int
	TotalDeletions    int
	TotalEqual        int
}

// Diff compares original and accepted extracted text and returns a structured diff
func Diff(original, accepted map[string][]string) *DiffResult {
	result := &DiffResult{
		ContainerDiffs: make(map[string]ContainerDiff),
		Summary:        DiffSummary{},
	}

	// Get all unique container names from both maps
	containerNames := make(map[string]bool)
	for name := range original {
		containerNames[name] = true
	}
	for name := range accepted {
		containerNames[name] = true
	}

	result.Summary.TotalContainers = len(containerNames)

	// Process each container
	for containerName := range containerNames {
		originalParagraphs := original[containerName]
		acceptedParagraphs := accepted[containerName]

		containerDiff := diffContainer(originalParagraphs, acceptedParagraphs)

		// Only add containers with actual changes (non-equal operations)
		hasChanges := false
		for _, op := range containerDiff.Operations {
			if op.Type != "equal" {
				hasChanges = true
				break
			}
		}

		if hasChanges {
			result.ContainerDiffs[containerName] = containerDiff
			result.Summary.ChangedContainers++
		}

		// Update summary statistics
		for _, op := range containerDiff.Operations {
			switch op.Type {
			case "insert":
				result.Summary.TotalInsertions++
			case "delete":
				result.Summary.TotalDeletions++
			case "equal":
				result.Summary.TotalEqual++
			}
		}
	}

	return result
}

// diffContainer compares paragraphs within a single container using word-level diff
func diffContainer(original, accepted []string) ContainerDiff {
	containerDiff := ContainerDiff{
		Operations: make([]DiffOperation, 0),
	}

	// Convert paragraph arrays to single strings for comparison
	originalText := joinParagraphs(original)
	acceptedText := joinParagraphs(accepted)

	// Skip if both are empty
	if originalText == "" && acceptedText == "" {
		return containerDiff
	}

	// Perform word-level diff
	diffs := diffAtWordLevel(originalText, acceptedText)

	// Convert to our DiffOperation format
	for _, diff := range diffs {
		op := DiffOperation{
			Type:      string(diff.Type),
			Text:      diff.Text,
			Container: "",
		}
		containerDiff.Operations = append(containerDiff.Operations, op)
	}

	return containerDiff
}

// diffAtWordLevel performs word-level diff comparison
func diffAtWordLevel(original, accepted string) []InternalDiff {
	// Split texts into words for word-level comparison
	originalWords := splitIntoWords(original)
	acceptedWords := splitIntoWords(accepted)

	// Use difflib for proper word-level diff
	matcher := difflib.NewMatcher(originalWords, acceptedWords)
	opcodes := matcher.GetOpCodes()

	result := make([]InternalDiff, 0)

	for _, opcode := range opcodes {
		tag := opcode.Tag
		i1, i2, j1, j2 := opcode.I1, opcode.I2, opcode.J1, opcode.J2

		switch tag {
		case 'e': // equal
			if i1 < i2 {
				text := strings.Join(originalWords[i1:i2], "")
				result = append(result, InternalDiff{
					Type: DiffEqual,
					Text: text,
				})
			}
		case 'd': // delete
			if i1 < i2 {
				text := strings.Join(originalWords[i1:i2], "")
				result = append(result, InternalDiff{
					Type: DiffDelete,
					Text: text,
				})
			}
		case 'i': // insert
			if j1 < j2 {
				text := strings.Join(acceptedWords[j1:j2], "")
				result = append(result, InternalDiff{
					Type: DiffInsert,
					Text: text,
				})
			}
		case 'r': // replace
			if i1 < i2 {
				text := strings.Join(originalWords[i1:i2], "")
				result = append(result, InternalDiff{
					Type: DiffDelete,
					Text: text,
				})
			}
			if j1 < j2 {
				text := strings.Join(acceptedWords[j1:j2], "")
				result = append(result, InternalDiff{
					Type: DiffInsert,
					Text: text,
				})
			}
		}
	}

	return result
}

// splitIntoWords splits text into words while preserving whitespace separately
func splitIntoWords(text string) []string {
	if text == "" {
		return []string{}
	}

	// Split into words and whitespace/punctuation separately for cleaner diffs
	re := regexp.MustCompile(`\S+|\s+`)
	matches := re.FindAllString(text, -1)

	// Filter out empty matches
	result := make([]string, 0)
	for _, match := range matches {
		if match != "" {
			result = append(result, match)
		}
	}

	return result
}

// joinParagraphs converts a paragraph slice to a single string with paragraph separators
func joinParagraphs(paragraphs []string) string {
	if len(paragraphs) == 0 {
		return ""
	}

	result := ""
	for i, paragraph := range paragraphs {
		if i > 0 {
			result += "\n"
		}
		result += paragraph
	}
	return result
}

// PrettyPrint returns a formatted string representation of the diff with XML-like tags
// for easy understanding by LLMs. Deleted text is wrapped in <delete> tags,
// inserted text is wrapped in <insert> tags.
func (dr *DiffResult) PrettyPrint() string {
	var result strings.Builder

	// Add summary header
	result.WriteString("=== DIFF SUMMARY ===\n")
	result.WriteString(fmt.Sprintf("Total containers: %d\n", dr.Summary.TotalContainers))
	result.WriteString(fmt.Sprintf("Changed containers: %d\n", dr.Summary.ChangedContainers))
	result.WriteString(fmt.Sprintf("Insertions: %d, Deletions: %d, Equal: %d\n\n",
		dr.Summary.TotalInsertions, dr.Summary.TotalDeletions, dr.Summary.TotalEqual))

	// Process each container with changes
	for containerName, containerDiff := range dr.ContainerDiffs {
		result.WriteString(fmt.Sprintf("=== CONTAINER: %s ===\n", containerName))

		// Reconstruct text with diff markup
		for _, op := range containerDiff.Operations {
			switch op.Type {
			case "delete":
				result.WriteString(fmt.Sprintf("<delete>%s</delete>", escapeText(op.Text)))
			case "insert":
				result.WriteString(fmt.Sprintf("<insert>%s</insert>", escapeText(op.Text)))
			case "equal":
				result.WriteString(escapeText(op.Text))
			}
		}
		result.WriteString("\n\n")
	}

	return result.String()
}

// escapeText escapes angle brackets in text to avoid confusion with diff tags
func escapeText(text string) string {
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}
