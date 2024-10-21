package mydocx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

const NAMESPACE = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

// A replacer replaces a string with a modified string.
type Replacer func(string) string

// All text from the sourceFile is modified by applying the replace function to it.
// Before applying the function, the whole paragraph is collected as a single text, even if split on multiple runs.
// Replace function is called paragraph by paragraph. No special assupmtion is made for empty paragraph.
// If the replace function is nil, text will be copied unmodified (but paragraph format WILL be extended from the start of paragraph, removing internal paragraph formatting !).
// If the targetFile name is empty, the sourceFile will be used, modification will be done in place.
func ModifyText(sourceFilePath string, replace Replacer, targetFilePath string) error {

	// Open the .docx (which is a zip file)
	docxFile, err := zip.OpenReader(sourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open docx file: %v", err)
	}
	defer docxFile.Close()

	// default replace function
	if replace == nil {
		replace = func(s string) string { return s }
	}

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
	cd := newCustDecoder(documentContent, replace)
	cd.processBody()
	modifiedXML, err := cd.result()
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
	if targetFilePath == "" {
		targetFilePath = sourceFilePath
	}
	return os.WriteFile(targetFilePath, buffer.Bytes(), 0644)
}

type custDecoder struct {
	dec       *xml.Decoder
	input     []byte              // initial doc content, unchanged
	res       [][]byte            // result afeter processing
	replace   func(string) string // replacer string
	lastSaved int64               // index of last saved byte, index from input byte slice
	err       error               // last error
	firstRun  int                 // contains index of first run content
	rcontent  []byte              // agrregated text content of all runs from the same paragraph

}

func newCustDecoder(documentContent []byte, replacer func(string) string) *custDecoder {
	return &custDecoder{
		input:     documentContent,
		dec:       xml.NewDecoder(bytes.NewReader(documentContent)),
		res:       make([][]byte, 1, 200), // ensure starts with empty string ...
		replace:   replacer,
		lastSaved: -1,
		err:       nil,
		firstRun:  -1,
		rcontent:  nil,
	}
}

// Get transformed result as a byte slice
func (cd *custDecoder) result() ([]byte, error) {
	cd.copy()
	fr := bytes.Join(cd.res, nil)
	fmt.Println("Final result \n", (string)(fr))
	return fr, cd.err
}

// Copy the newly parsed content of the original docx to the result up to the last token parsed, included.
func (cd *custDecoder) copy() {
	last := cd.dec.InputOffset()
	cd.res = append(cd.res, cd.input[cd.lastSaved+1:last])
	cd.lastSaved = last - 1
}

// Process the body tags
func (cd *custDecoder) processBody() {
	cd.copy()
	defer cd.copy()
	var tok xml.Token
	for cd.err == nil {
		tok, cd.err = cd.dec.Token()
		if cd.err != nil {
			if cd.err == io.EOF { // normal exit
				cd.err = nil
			}
			break // in all case, stop and return err !
		}
		switch t := tok.(type) {
		default:
		case xml.StartElement:
			if t.Name.Local == "body" && t.Name.Space == NAMESPACE {
				fmt.Printf("Captured :%s\n", t.Name.Local)
				cd.copy()
				cd.processParagraphs()
			}
		}

	}
}

// process paragraphs
func (cd *custDecoder) processParagraphs() {
	cd.copy()
	defer cd.copy()
	var tok xml.Token
	for cd.err == nil {
		tok, cd.err = cd.dec.Token()
		if cd.err != nil {
			break // in all case, stop and return err - EOF is abnormal in this case.
		}

		switch t := tok.(type) {
		default:
		case xml.StartElement:
			if t.Name.Local == "p" && t.Name.Space == NAMESPACE {
				fmt.Printf("Captured :%s\n", t.Name.Local)
				cd.copy()
				cd.processRuns()
			}
		case xml.EndElement:
			if t.Name.Local == "body" && t.Name.Space == NAMESPACE {
				return
			}
		}
	}
}

// process runs
func (cd *custDecoder) processRuns() {
	cd.copy()
	defer cd.copy()
	var tok xml.Token
	cd.rcontent = nil
	cd.firstRun = -1

	for cd.err == nil {
		tok, cd.err = cd.dec.Token()
		if cd.err != nil {
			break // in all case, stop and return err - EOF is abnormal in this case.
		}

		switch t := tok.(type) {
		default:
		case xml.StartElement:
			if t.Name.Local == "r" && t.Name.Space == NAMESPACE {
				fmt.Printf("Captured :%s\n", t.Name.Local)
				cd.copy()
				cd.processText()
			}
		case xml.EndElement:
			if t.Name.Local == "p" && t.Name.Space == NAMESPACE {
				fmt.Printf("Captured :/%s\n", t.Name.Local)
				if cd.firstRun >= 0 {
					cd.res[cd.firstRun] = []byte(cd.replace((string)(cd.rcontent))) // save agg content to first run
					fmt.Println("saving rcontent to index ", cd.firstRun)
				}
				cd.rcontent = nil
				cd.firstRun = -1
				fmt.Printf("Res : %s\n", cd.res)
				return
			}
		}
	}
}

func (cd *custDecoder) processText() {
	cd.copy()
	defer cd.copy()
	var tok xml.Token
	for cd.err == nil {
		tok, cd.err = cd.dec.Token()
		if cd.err != nil {
			break // in all case, stop and return err - EOF is abnormal in this case.
		}

		switch t := tok.(type) {
		default:
		case xml.StartElement:
			if t.Name.Local == "t" && t.Name.Space == NAMESPACE {
				fmt.Printf("Captured :%s\n", t.Name.Local)
				cd.copy()
				if cd.firstRun < 0 {
					cd.res = append(cd.res, []byte{}) // place holder for future aggregated text
					cd.firstRun = len(cd.res) - 1     // remember index of place holder !
				}
				cd.processTextContent()
			}
		case xml.EndElement:
			if t.Name.Local == "r" && t.Name.Space == NAMESPACE {
				fmt.Printf("Captured :/%s\n", t.Name.Local)
				return
			}

		}
	}
}

func (cd *custDecoder) processTextContent() {
	cd.copy()
	defer cd.copy()
	var tok xml.Token
	for cd.err == nil {
		cd.copy()
		tok, cd.err = cd.dec.Token()
		if cd.err != nil {
			break // in all case, stop and return err - EOF is abnormal in this case.
		}

		switch t := tok.(type) {
		default:
			cd.copy()
		case xml.CharData:
			cd.rcontent = append(cd.rcontent, t...)
			fmt.Printf("\t -> %s\n", (string)(cd.rcontent))
			// skip copy of chardata, assuming the rest was already copied
			cd.lastSaved = cd.dec.InputOffset() - 1
		case xml.EndElement:
			if t.Name.Local == "t" && t.Name.Space == NAMESPACE {
				return
			}
		}
	}
}
