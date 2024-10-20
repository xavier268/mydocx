package docxtransform

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Define the XML structure for the relevant tags
type Document struct {
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

// Function to extract and modify paragraphs
func modifyParagraphsInDocx(filePath string, replacer func(string) string) error {
	// Open the .docx (which is a zip file)
	docxFile, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open docx file: %v", err)
	}
	defer docxFile.Close()

	// Prepare a buffer to store the modified .docx content
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	// Locate the document.xml file
	var documentContent []byte
	for _, file := range docxFile.File {
		if file.Name == "word/document.xml" {
			documentContent, err = readFile(file)
			if err != nil {
				return fmt.Errorf("failed to read document.xml: %v", err)
			}
			continue
		}
		// Copy other files unmodified into the new .docx
		if err := copyFileToZip(zipWriter, file); err != nil {
			return fmt.Errorf("failed to copy file: %v", err)
		}
	}

	if documentContent == nil {
		return fmt.Errorf("document.xml not found in the docx file")
	}

	// do the actual processing

	modifiedXML, err := processDocumentXML(documentContent, replacer)
	if err != nil {
		return fmt.Errorf("failed to process document.xml: %v", err)
	}

	// Add the modified document.xml back into the new .docx archive
	writer, err := zipWriter.Create("word/document.xml")
	if err != nil {
		return fmt.Errorf("failed to add modified document.xml to docx: %v", err)
	}
	_, err = writer.Write(modifiedXML)
	if err != nil {
		return fmt.Errorf("failed to write modified document.xml: %v", err)
	}

	// Close the zip writer
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %v", err)
	}

	// Save the modified .docx
	return os.WriteFile("modified.docx", buffer.Bytes(), 0644)
}

// Helper function to read a file from a zip archive
func readFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return ioutil.ReadAll(rc)
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

func processDocumentXML(documentContent []byte, replacer func(string) string) ([]byte, error) {
	var doc Document
	if err := xml.Unmarshal(documentContent, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document.xml: %v", err)
	}

	for i, paragraph := range doc.Body.Paragraphs {
		for j, run := range paragraph.Runs {
			run.Text.Value = replacer(run.Text.Value)
			doc.Body.Paragraphs[i].Runs[j] = run
		}
	}

	return xml.Marshal(doc)
}
