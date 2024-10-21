package mydocx

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
)

// Define a simplified XML structure for a word document, with a focus on the relevant tags
// Only used for text extraction, since all other necessary fields are discarded.
// Text modification uses another strategy.
type SimplifiedDocument struct {
	XMLName xml.Name `xml:"document"`
	Body    Body     `xml:"body"`
	XMLNSw  string   `xml:"xmlns:w,attr"`
}

type Body struct {
	Paragraphs []Paragraph `xml:"p"`
}

type Paragraph struct {
	Runs []Run `xml:"r"`
}

type Run struct {
	Text Text `xml:"t"`
}

type Text struct {
	Value string `xml:",chardata"`
}

// Extract text content from docx file for external processing.
// The slice of strings contains a string, possibly empty, for each paragraph.
func ExtractText(sourceFilePath string) ([]string, error) {
	docxFile, err := zip.OpenReader(sourceFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open docx file: %v", err)
	}
	defer docxFile.Close()

	var paragraphs []string
	for _, file := range docxFile.File {
		if file.Name == "word/document.xml" {
			documentContent, err := readFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read document.xml: %v", err)
			}

			if debugFlag {
				fmt.Printf("\n%q\n", documentContent)
			}

			var doc SimplifiedDocument
			err = xml.Unmarshal(documentContent, &doc)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal document.xml: %v", err)
			}

			if debugFlag {
				fmt.Printf("\n%v\n", doc)
			}

			for _, p := range doc.Body.Paragraphs {
				paragraphText := ""
				for _, r := range p.Runs {
					paragraphText += r.Text.Value
				}
				paragraphs = append(paragraphs, paragraphText)
			}
			break
		}
	}

	return paragraphs, nil
}

// Helper function to read a file from a zip archive
func readFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

// Helper function to copy unmodified files to the new zip
func copyFileToZip(zipWriter *zip.Writer, file *zip.File) error {
	readCloser, err := file.Open()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	writer, err := zipWriter.Create(file.Name)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, readCloser)
	return err
}
