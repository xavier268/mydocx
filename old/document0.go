package docxtransform

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// Define the XML structure for the relevant tags
type Document struct {
	Paragraphs []Paragraph `xml:"body>p"`
}

type Paragraph struct {
	Texts []Text `xml:"r>t"`
}

type Text struct {
	Value string `xml:",chardata"`
}

// Function to extract and print paragraphs from a DOCX file
func extractParagraphsFromDocx(filePath string) error {
	// Open the .docx (which is a zip file)
	docxFile, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open docx file: %v", err)
	}
	defer docxFile.Close()

	// Find the word/document.xml file inside the .docx archive
	var documentFile *zip.File
	for _, file := range docxFile.File {
		if file.Name == "word/document.xml" {
			documentFile = file
			break
		}
	}
	if documentFile == nil {
		return fmt.Errorf("document.xml not found in the docx file")
	}

	// Open the document.xml file for reading
	xmlFile, err := documentFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open document.xml: %v", err)
	}
	defer xmlFile.Close()

	// Read the XML content
	var buf bytes.Buffer
	_, err = io.Copy(&buf, xmlFile)
	if err != nil {
		return fmt.Errorf("failed to read document.xml: %v", err)
	}

	// Parse the XML content
	var doc Document
	err = xml.Unmarshal(buf.Bytes(), &doc)
	if err != nil {
		return fmt.Errorf("failed to unmarshal document.xml: %v", err)
	}

	// Iterate over the paragraphs and print the text
	for i, paragraph := range doc.Paragraphs {
		var paragraphText string
		for _, text := range paragraph.Texts {
			paragraphText += text.Value
		}
		fmt.Printf("Paragraph %d: %q\n", i+1, paragraphText)
	}

	return nil
}
