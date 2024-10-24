package mydocx

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"

	"bytes"
)

// Extract text content from docx file for external processing.
// Returns a map from the container name (eg : word/footer1.xml) to a list of text contained in its paragraphs.
func ExtractText(sourceFilePath string) (map[string][]string, error) {

	docxFile, err := zip.OpenReader(sourceFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open docx file: %v", err)
	}
	defer docxFile.Close()

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

func extractText(dec *xml.Decoder) string {
	var tt = ""
	for tok, err := dec.Token(); err == nil; tok, err = dec.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" && t.Name.Space == NAMESPACE {
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
