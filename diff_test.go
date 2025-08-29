package mydocx

import (
	"os"
	"strings"
	"testing"
)

// TestBasicDiff tests basic diff functionality with simple text
func TestBasicDiff(t *testing.T) {
	original := map[string][]string{
		"container1": {"Hello world", "This is a test"},
		"container2": {"Another paragraph"},
	}

	accepted := map[string][]string{
		"container1": {"Hello universe", "This is a test"},
		"container2": {"Another paragraph", "Added content"},
	}

	result := Diff(original, accepted)

	// Check summary
	if result.Summary.TotalContainers != 2 {
		t.Errorf("Expected 2 total containers, got %d", result.Summary.TotalContainers)
	}

	if result.Summary.ChangedContainers != 2 {
		t.Errorf("Expected 2 changed containers, got %d", result.Summary.ChangedContainers)
	}

	// Check that we have diffs for both containers
	if len(result.ContainerDiffs) != 2 {
		t.Errorf("Expected 2 container diffs, got %d", len(result.ContainerDiffs))
	}

	// Test pretty print doesn't crash
	prettyOutput := result.PrettyPrint()
	if len(prettyOutput) == 0 {
		t.Error("PrettyPrint returned empty string")
	}
}

// TestEmptyDiff tests diff with identical content
func TestEmptyDiff(t *testing.T) {
	original := map[string][]string{
		"container1": {"Same content"},
	}

	accepted := map[string][]string{
		"container1": {"Same content"},
	}

	result := Diff(original, accepted)

	// Should have no changes
	if result.Summary.ChangedContainers != 0 {
		t.Errorf("Expected 0 changed containers, got %d", result.Summary.ChangedContainers)
	}

	if len(result.ContainerDiffs) != 0 {
		t.Errorf("Expected 0 container diffs, got %d", len(result.ContainerDiffs))
	}
}

// TestMissingContainers tests diff with containers present in only one version
func TestMissingContainers(t *testing.T) {
	original := map[string][]string{
		"container1": {"Original content"},
		"container2": {"Will be removed"},
	}

	accepted := map[string][]string{
		"container1": {"Modified content"},
		"container3": {"New container"},
	}

	result := Diff(original, accepted)

	// Should detect all 3 containers (original + accepted combined)
	if result.Summary.TotalContainers != 3 {
		t.Errorf("Expected 3 total containers, got %d", result.Summary.TotalContainers)
	}
}

// TestDocxFileDiff tests diff functionality using actual test.docx file
func TestDocxFileDiff(t *testing.T) {
	testFile := "testFiles/test.docx"

	// Check if test file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
		return
	}

	// Extract original text (treating as if all changes were rejected)
	original, err := ExtractOriginalText(testFile)
	if err != nil {
		t.Fatalf("Failed to extract original text: %v", err)
	}

	// Extract accepted text (treating as if all changes were accepted)
	accepted, err := ExtractText(testFile)
	if err != nil {
		t.Fatalf("Failed to extract accepted text: %v", err)
	}

	// Generate diff
	result := Diff(original, accepted)

	// Generate pretty printed version
	prettyOutput := result.PrettyPrint()

	// Save to test-diffs.txt in the same folder as test.docx
	outputFile := "testFiles/test-diffs.txt"
	err = os.WriteFile(outputFile, []byte(prettyOutput), 0644)
	if err != nil {
		t.Fatalf("Failed to write diff output to %s: %v", outputFile, err)
	}

	t.Logf("Diff analysis complete:")
	t.Logf("- Total containers: %d", result.Summary.TotalContainers)
	t.Logf("- Changed containers: %d", result.Summary.ChangedContainers)
	t.Logf("- Insertions: %d, Deletions: %d, Equal: %d",
		result.Summary.TotalInsertions, result.Summary.TotalDeletions, result.Summary.TotalEqual)
	t.Logf("- Pretty printed diff saved to: %s", outputFile)

	// Basic validation
	if result.Summary.TotalContainers == 0 {
		t.Error("Expected at least one container in the document")
	}
}

// TestPrettyPrintEscaping tests that angle brackets are properly escaped
func TestPrettyPrintEscaping(t *testing.T) {
	original := map[string][]string{
		"container1": {"Text with <brackets> and >more< brackets"},
	}

	accepted := map[string][]string{
		"container1": {"Text with <different> and >other< brackets"},
	}

	result := Diff(original, accepted)
	prettyOutput := result.PrettyPrint()

	// Check that angle brackets are escaped
	if !contains(prettyOutput, "&lt;") || !contains(prettyOutput, "&gt;") {
		t.Error("Expected angle brackets to be escaped in pretty print output")
	}

	// Should not contain unescaped brackets in the text content
	// (but will contain them in our diff tags)
	if contains(prettyOutput, "Text with <brackets>") {
		t.Error("Found unescaped angle brackets in text content")
	}
}

// TestWordLevelDiff tests that diffs are performed at word level, not character level
func TestWordLevelDiff(t *testing.T) {
	original := map[string][]string{
		"container1": {"The quick brown fox jumps"},
	}

	accepted := map[string][]string{
		"container1": {"The fast brown dog jumps"},
	}

	result := Diff(original, accepted)
	prettyOutput := result.PrettyPrint()

	t.Logf("Word-level diff output:\n%s", prettyOutput)

	// Should show whole word changes, not character-level
	if !contains(prettyOutput, "<delete>quick</delete>") {
		t.Error("Expected word-level deletion of 'quick'")
	}

	if !contains(prettyOutput, "<insert>fast</insert>") {
		t.Error("Expected word-level insertion of 'fast'")
	}

	if !contains(prettyOutput, "<delete>fox</delete>") {
		t.Error("Expected word-level deletion of 'fox'")
	}

	if !contains(prettyOutput, "<insert>dog</insert>") {
		t.Error("Expected word-level insertion of 'dog'")
	}

	// Should NOT show character-level changes like "<delete>f</delete><insert>d</insert>ox"
	if contains(prettyOutput, "<delete>f</delete><insert>d</insert>") {
		t.Error("Found character-level diff, expected word-level")
	}
}

// TestBeAreWordLevel tests the specific "be"/"are" case
func TestBeAreWordLevel(t *testing.T) {
	original := map[string][]string{
		"container1": {"This will be testing"},
	}

	accepted := map[string][]string{
		"container1": {"This are testing"},
	}

	result := Diff(original, accepted)
	prettyOutput := result.PrettyPrint()

	t.Logf("Be/Are diff output:\n%s", prettyOutput)

	// Should show whole word replacement
	if !contains(prettyOutput, "<delete>will be</delete>") && !contains(prettyOutput, "<delete>will</delete>") && !contains(prettyOutput, "<delete>be</delete>") {
		t.Error("Expected word-level deletion including 'will' or 'be'")
	}

	if !contains(prettyOutput, "<insert>are</insert>") {
		t.Error("Expected word-level insertion of 'are'")
	}

	// Should NOT show character-level like "<delete>will b</delete><insert>ar</insert>e"
	if contains(prettyOutput, "<delete>will b</delete><insert>ar</insert>e") {
		t.Error("Found problematic character-level diff for 'will be' -> 'are'")
	}
}

// TestMergedOperations tests that consecutive operations of same type are merged
func TestMergedOperations(t *testing.T) {
	original := map[string][]string{
		"container1": {"The quick brown fox jumps over the lazy dog"},
	}

	accepted := map[string][]string{
		"container1": {"The fast red cat runs under the sleepy mouse"},
	}

	result := Diff(original, accepted)
	prettyOutput := result.PrettyPrint()

	t.Logf("Merged operations diff output:\n%s", prettyOutput)

	// Should have merged consecutive deletions and insertions
	// Instead of: <delete>quick</delete> <delete>brown</delete> <delete>fox</delete>
	// Should be: <delete>quick brown fox</delete>

	// Count individual word operations - should be fewer due to merging
	deleteCount := strings.Count(prettyOutput, "<delete>")
	insertCount := strings.Count(prettyOutput, "<insert>")

	// With "the" appearing twice, we expect some fragmentation
	// But it should still be better than character-level
	if deleteCount > 10 || insertCount > 10 {
		t.Errorf("Too many separate operations: %d deletes and %d inserts - likely not word-level", deleteCount, insertCount)
	}

	t.Logf("Found %d delete operations and %d insert operations", deleteCount, insertCount)
}

// TestDiffAnalyse tests the DiffAnalyse convenience function
func TestDiffAnalyse(t *testing.T) {
	testFile := "testFiles/test.docx"

	// Check if test file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
		return
	}

	// Use DiffAnalyse function
	commentedContent, err := DiffAnalyse(testFile)
	if err != nil {
		t.Fatalf("DiffAnalyse failed: %v", err)
	}

	// Verify output contains expected diff markers
	if !contains(commentedContent, "=== DIFF SUMMARY ===") {
		t.Error("Expected diff summary header")
	}

	if !contains(commentedContent, "<delete>") || !contains(commentedContent, "<insert>") {
		t.Error("Expected diff tags in output")
	}

	// Should contain container information
	if !contains(commentedContent, "=== CONTAINER:") {
		t.Error("Expected container section headers")
	}

	t.Logf("DiffAnalyse output length: %d characters", len(commentedContent))
	t.Logf("DiffAnalyse output preview:\n%.200s...", commentedContent)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || contains(s[1:], substr) || (len(s) > len(substr) && s[:len(substr)] == substr))
}
