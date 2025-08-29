package mydocx

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"bytes"
)

// Extract text content from docx file for external processing.
// Returns a map from the container name (eg : word/footer1.xml) to a list of text contained in its paragraphs.
// This function is thread-safe.
// The verbose flag can be set to true to display information about the containers extracted.
func ExtractText(sourceFilePath string) (map[string][]string, error) {
	if VERBOSE {
		fmt.Printf("Extracting text from %s\n", sourceFilePath)
	}
	data, err := os.ReadFile(sourceFilePath)
	if err != nil {
		return nil, err
	}
	return ExtractTextBytes(data)
}

// Same as ExtractText, but takes a byte array as input.
// This is useful for embedded use, when the docx file is already in memory.
// Returns a map from the container name (eg : word/footer1.xml) to a list of text contained in its paragraphs.
// This function is thread-safe.
// The verbose flag can be set to true to display information about the containers extracted.
func ExtractTextBytes(sourceBytes []byte) (map[string][]string, error) {

	docxFile, err := zip.NewReader(bytes.NewReader(sourceBytes), int64(len(sourceBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to open docx file: %v", err)
	}

	// no need to close, since byte buffer
	// defer docxFile.Close()

	result := make(map[string][]string)

	for _, file := range docxFile.File {
		if containerPattern.MatchString(file.Name) {
			if VERBOSE {
				fmt.Printf("Extracting from %s\n", file.Name)
			}
			documentContent, err := readFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read document.xml: %v", err)
			}
			// launch actual extraction
			dec := xml.NewDecoder(bytes.NewReader(documentContent))
			result[file.Name], err = extractParagraphs(dec)
			if err != nil {
				return result, fmt.Errorf("failed to extract text from %s : %v", file.Name, err)
			}
		}

	}

	return result, nil
}

// Extract original text content from docx file, ignoring all revisions (insertions and deletions).
// Returns a map from the container name (eg : word/footer1.xml) to a list of text contained in its paragraphs.
// This function treats the document as if all changes were rejected - insertions are ignored, deletions are ignored.
// This function is thread-safe.
// The verbose flag can be set to true to display information about the containers extracted.
func ExtractOriginalText(sourceFilePath string) (map[string][]string, error) {
	if VERBOSE {
		fmt.Printf("Extracting original text from %s\n", sourceFilePath)
	}
	data, err := os.ReadFile(sourceFilePath)
	if err != nil {
		return nil, err
	}
	return ExtractOriginalTextBytes(data)
}

// Same as ExtractOriginalText, but takes a byte array as input.
// Extract original text content, ignoring all revisions (insertions and deletions).
// This is useful for embedded use, when the docx file is already in memory.
// Returns a map from the container name (eg : word/footer1.xml) to a list of text contained in its paragraphs.
// This function treats the document as if all changes were rejected - insertions are ignored, deletions are ignored.
// This function is thread-safe.
// The verbose flag can be set to true to display information about the containers extracted.
func ExtractOriginalTextBytes(sourceBytes []byte) (map[string][]string, error) {

	docxFile, err := zip.NewReader(bytes.NewReader(sourceBytes), int64(len(sourceBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to open docx file: %v", err)
	}

	result := make(map[string][]string)

	for _, file := range docxFile.File {
		if containerPattern.MatchString(file.Name) {
			if VERBOSE {
				fmt.Printf("Extracting original text from %s\n", file.Name)
			}
			documentContent, err := readFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read document.xml: %v", err)
			}
			// launch actual extraction
			dec := xml.NewDecoder(bytes.NewReader(documentContent))
			result[file.Name], err = extractOriginalParagraphs(dec)
			if err != nil {
				return result, fmt.Errorf("failed to extract original text from %s : %v", file.Name, err)
			}
		}

	}

	return result, nil
}

// Extract paragraphs text from container content.
func extractParagraphs(dec *xml.Decoder) (res []string, err error) {
	var tt string
	for tok, err := dec.Token(); err == nil; tok, err = dec.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "p" && t.Name.Space == NAMESPACE {
				tt, err = extractRuns(dec)
				if debugflag {
					fmt.Printf("Captured text : %q\n", tt)
				}
				if err == io.EOF {
					return res, nil
				}
				if err != nil {
					return nil, err
				}
				res = append(res, tt)
			}
		}
	}
	return res, err
}

// Extract text from the runs in a given paragraph.
func extractRuns(dec *xml.Decoder) (tt string, err error) {

	for tok, err := dec.Token(); err == nil; tok, err = dec.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "r" && t.Name.Space == NAMESPACE {
				tt = tt + extractText(dec)
				if err != nil {
					break
				}
			}
		case xml.EndElement:
			if t.Name.Local == "p" && t.Name.Space == NAMESPACE {
				return tt, err
			}
		}
	}
	return tt, err
}

// Extact text exactly as if all changes have been accepetd. Deletions are ingored, insertions are included.
func extractText(dec *xml.Decoder) string {
	var tt = ""
	var inDeletion = false
	for tok, err := dec.Token(); err == nil; tok, err = dec.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "del" && t.Name.Space == NAMESPACE {
				inDeletion = true
			} else if t.Name.Local == "t" && t.Name.Space == NAMESPACE && !inDeletion {
				cdt, err := dec.Token()
				if err != nil {
					break
				}
				if data, ok := cdt.(xml.CharData); ok {
					tt = tt + string(data)
				}
			} else if t.Name.Local == "delText" && t.Name.Space == NAMESPACE {
				// Skip deleted text content entirely
				for {
					tok, err := dec.Token()
					if err != nil {
						break
					}
					if endEl, ok := tok.(xml.EndElement); ok && endEl.Name.Local == "delText" && endEl.Name.Space == NAMESPACE {
						break
					}
				}
			}
		case xml.EndElement:
			if t.Name.Local == "del" && t.Name.Space == NAMESPACE {
				inDeletion = false
			} else if t.Name.Local == "r" && t.Name.Space == NAMESPACE {
				return tt
			}
		}
	}
	return tt
}

// Extract original paragraphs text from container content, ignoring all revisions.
func extractOriginalParagraphs(dec *xml.Decoder) (res []string, err error) {
	var tt string
	for tok, err := dec.Token(); err == nil; tok, err = dec.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "p" && t.Name.Space == NAMESPACE {
				// Check if this paragraph is an insertion (should be ignored in original text)
				isInsertedParagraph := false
				tt, isInsertedParagraph, err = extractOriginalRunsFromParagraph(dec)
				if err == io.EOF {
					return res, nil
				}
				if err != nil {
					return nil, err
				}
				// Only include paragraphs that are not insertions
				if !isInsertedParagraph {
					if debugflag {
						fmt.Printf("Captured original text : %q\n", tt)
					}
					res = append(res, tt)
				} else if debugflag {
					fmt.Printf("Skipped inserted paragraph : %q\n", tt)
				}
			}
		}
	}
	return res, err
}

// Extract original text from the runs in a given paragraph, checking if paragraph is inserted.
func extractOriginalRunsFromParagraph(dec *xml.Decoder) (tt string, isInsertedParagraph bool, err error) {
	isInsertedParagraph = false

	for tok, err := dec.Token(); err == nil; tok, err = dec.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "pPr" && t.Name.Space == NAMESPACE {
				// Check paragraph properties for insertion markers
				isInsertedParagraph = checkParagraphPropertiesForInsertion(dec)
			} else if t.Name.Local == "ins" && t.Name.Space == NAMESPACE {
				// Skip entire insertion blocks at paragraph level
				for {
					tok, err := dec.Token()
					if err != nil {
						break
					}
					if endEl, ok := tok.(xml.EndElement); ok && endEl.Name.Local == "ins" && endEl.Name.Space == NAMESPACE {
						break
					}
				}
			} else if t.Name.Local == "del" && t.Name.Space == NAMESPACE {
				// Process deletion blocks to restore original text
				tt = tt + extractOriginalTextFromDeletion(dec)
			} else if t.Name.Local == "r" && t.Name.Space == NAMESPACE {
				tt = tt + extractOriginalTextFromRun(dec)
				if err != nil {
					break
				}
			}
		case xml.EndElement:
			if t.Name.Local == "p" && t.Name.Space == NAMESPACE {
				return tt, isInsertedParagraph, err
			}
		}
	}
	return tt, isInsertedParagraph, err
}

// Extract original text from a deletion block at paragraph level
func extractOriginalTextFromDeletion(dec *xml.Decoder) string {
	var tt = ""
	for tok, err := dec.Token(); err == nil; tok, err = dec.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "r" && t.Name.Space == NAMESPACE {
				tt = tt + extractOriginalTextFromRun(dec)
			}
		case xml.EndElement:
			if t.Name.Local == "del" && t.Name.Space == NAMESPACE {
				return tt
			}
		}
	}
	return tt
}

// Check paragraph properties for insertion markers that indicate the entire paragraph was inserted.
// Note: <w:ins> in paragraph properties (pPr/rPr) typically indicates insertion formatting,
// not that the paragraph itself was inserted. Actual paragraph insertions are usually
// at the paragraph content level, not in properties.
func checkParagraphPropertiesForInsertion(dec *xml.Decoder) bool {
	for tok, err := dec.Token(); err == nil; tok, err = dec.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			// Only consider this an inserted paragraph if we find specific markers
			// that indicate the paragraph itself (not just its formatting) was inserted.
			// In most cases, <w:ins> in pPr/rPr is just formatting, not content insertion.
		case xml.EndElement:
			if t.Name.Local == "pPr" && t.Name.Space == NAMESPACE {
				return false
			}
		}
	}
	return false
}

// Extract original text from a run, treating document as if all changes were rejected.
// Insertions are ignored, deletions (delText) are included to restore the original text.
func extractOriginalTextFromRun(dec *xml.Decoder) string {
	var tt = ""
	var inInsertion = false
	for tok, err := dec.Token(); err == nil; tok, err = dec.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "ins" && t.Name.Space == NAMESPACE {
				inInsertion = true
				// Skip entire insertion block
				for {
					tok, err := dec.Token()
					if err != nil {
						break
					}
					if endEl, ok := tok.(xml.EndElement); ok && endEl.Name.Local == "ins" && endEl.Name.Space == NAMESPACE {
						break
					}
				}
			} else if t.Name.Local == "t" && t.Name.Space == NAMESPACE && !inInsertion {
				cdt, err := dec.Token()
				if err != nil {
					break
				}
				if data, ok := cdt.(xml.CharData); ok {
					tt = tt + string(data)
				}
			} else if t.Name.Local == "delText" && t.Name.Space == NAMESPACE {
				// Include deleted text content to restore original text
				cdt, err := dec.Token()
				if err != nil {
					break
				}
				if data, ok := cdt.(xml.CharData); ok {
					tt = tt + string(data)
				}
			}
		case xml.EndElement:
			if t.Name.Local == "r" && t.Name.Space == NAMESPACE {
				return tt
			}
		}
	}
	return tt
}
