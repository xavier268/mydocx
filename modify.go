package mydocx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

// A replacer replaces a string with a list modified string. It is provide the container name where replacement will occur ("word/document.xm", "word/footer1.xml", ...).
// Only documents, headers and footers will be submitted.
// If the returned slice is empty, the paragraph is removed.
// If the returned slice contains more than 1 element, new paragraphs are added, duplicated from the original paragraph.
type Replacer func(container string, original string) (replaced []string)

// All text from the sourceFile is modified by applying the replace function to it.
// Before applying the function, the whole paragraph is collected as a single text, even if split on multiple runs.
// Replace function is called paragraph by paragraph. No special assumption is made for empty paragraph.
// If the replace function is nil, text will be copied unmodified (but paragraph format WILL be extended from the start of paragraph, removing internal paragraph formatting !).
// If the targetFile name is empty, the sourceFile will be used, modification will be done in place.
func ModifyText(sourceFilePath string, replace Replacer, targetFilePath string) error {

	if targetFilePath == "" {
		targetFilePath = sourceFilePath
	}
	if VERBOSE {
		fmt.Println("Modifying : ", sourceFilePath, "-->", targetFilePath)
	}

	// Open the .docx (which is a zip file)
	docxFile, err := zip.OpenReader(sourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open docx file: %v", err)
	}
	defer docxFile.Close()

	// default replace function, no change.
	if replace == nil {
		replace = func(_, s string) []string { return []string{s} }
	}

	// Prepare a buffer to store the modified .docx content
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	// Locate the document.xml and headers/footers files
	var documentContent []byte
	for _, file := range docxFile.File {
		fname := file.Name
		if containerPattern.MatchString(fname) {
			if VERBOSE {
				fmt.Println("Processing", fname)
			}
			documentContent, err = readFile(file)
			if err != nil {
				return fmt.Errorf("failed to read document.xml: %v", err)
			}
			err = processContent(fname, documentContent, replace, zipWriter)
			if err != nil {
				return err
			}
			continue // to next file container ...
		}
		// Copy other files unmodified into the new .docx
		if err := copyFileToZip(zipWriter, file); err != nil {
			return fmt.Errorf("failed to copy file: %v", err)
		}
	}

	// Close the zip writer
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %v", err)
	}

	// Save the modified .docx
	return os.WriteFile(targetFilePath, buffer.Bytes(), 0644)
}

// process either the actual document.xml or the footer/header(s)
func processContent(filename string, documentContent []byte, replace Replacer, zipWriter *zip.Writer) error {

	if documentContent == nil {
		return fmt.Errorf("%s not found in the docx file", filename)
	}

	cd := newCustDecoder(documentContent, replace)
	cd.container = filename
	cd.processParagraphs()
	if VERBOSE {
		cd.debug("Finished processing ...", filename)
	}
	modifiedXML, err := cd.result()
	if err != nil {
		return fmt.Errorf("failed to process %s: %v", filename, err)
	}

	// Add the modified xxx.xml back into the new .docx archive
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to add modified %s to docx: %v", filename, err)
	}
	_, err = writer.Write(modifiedXML)
	if err != nil {
		return fmt.Errorf("failed to write modified %s: %v", filename, err)
	}
	return nil
}

type custDecoder struct {
	dec          *xml.Decoder
	input        []byte   // initial doc content, unchanged
	container    string   // current container being processed ("word/document.xm", "word/footer1.xml", ...)
	res          [][]byte // result afeter processing
	replace      Replacer // replacer function
	lastSaved    int64    // index of last saved byte, index from input byte slice
	err          error    // last error
	rcontent     []byte   // agrregated text content of all runs from the same paragraph
	curPara      int      // index of the the current paragraph start within res. Used to destroy entire paragraph upon request.
	firstRunText int      // contains res index of first run text placeholder

}

func newCustDecoder(documentContent []byte, replacer Replacer) *custDecoder {
	return &custDecoder{
		input:        documentContent,
		dec:          xml.NewDecoder(bytes.NewReader(documentContent)),
		res:          make([][]byte, 1, 200), // ensure starts with empty string ...
		replace:      replacer,
		lastSaved:    -1,
		err:          nil,
		firstRunText: -1,
		rcontent:     nil,
		curPara:      -1,
	}
}

// Get transformed result as a byte slice
func (cd *custDecoder) result() ([]byte, error) {
	cd.copy()
	fr := bytes.Join(cd.res, nil)
	//fmt.Println("Final result \n", (string)(fr))
	return fr, cd.err
}

// Copy the newly parsed content of the original docx to the result up to the last token parsed, included.
func (cd *custDecoder) copy() {
	if next := cd.dec.InputOffset(); cd.lastSaved+1 < next { // next points to the start of the next token never parsed ...
		cd.res = append(cd.res, cd.input[cd.lastSaved+1:next])
		cd.lastSaved = next - 1
	}
}

// look for paragraphs
func (cd *custDecoder) processParagraphs() {

	var tok xml.Token

	for tok, cd.err = cd.dec.Token(); cd.err == nil; tok, cd.err = cd.dec.Token() {
		cd.copy() // immediately copy token in a separate res element
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "p" && t.Name.Space == NAMESPACE {
				cd.curPara = len(cd.res) - 1 // mark para start, used to truncate later the current paragraph if so desired
				cd.processRuns()
			}
		}
	}

	if cd.err == io.EOF { // ignore EOF, it's a normal ending here.
		cd.err = nil
	}
}

// process runs, until end of paragraph
// starts with para on top of res.
func (cd *custDecoder) processRuns() {

	var tok xml.Token

	// reset run text capture, since we are starting a new paragraph ...
	cd.rcontent = nil
	cd.firstRunText = -1

	for tok, cd.err = cd.dec.Token(); cd.err == nil; tok, cd.err = cd.dec.Token() {
		cd.copy() // immediately copy current element
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "r" && t.Name.Space == NAMESPACE {
				cd.processText()
			}
		case xml.EndElement:
			if t.Name.Local == "p" && t.Name.Space == NAMESPACE {
				if cd.firstRunText >= 0 { // make sure we saw at least a run !
					cd.insert(cd.replace(cd.container, (string)(cd.rcontent)))
				}
				return
			}
		}
	}
}

// Insert provided text in paragraph.
// If slice is empty, current paragraph is discarded.
// If slice has more than 1 element, current paragraph is duplicated as needed.
// When the function is called, an entire paraggraph should be available in res.
func (cd *custDecoder) insert(paras []string) {
	defer cd.debug("after paragragrph insertions")
	if len(paras) == 0 {
		cd.res = cd.res[:cd.curPara]            // destroy the paragraph, the last copy was made for </p>
		cd.lastSaved = cd.dec.InputOffset() - 1 // saving will resume at the tag following the paraggraph
		return
	}
	cd.res[cd.firstRunText] = xmlEscape([]byte(paras[0])) // save escapes 1st content to first run
	if len(paras) == 1 {
		return // we're done
	}
	// else, duplicate paragph
	dup := cd.res[cd.curPara:]
	cd.res = append(cd.res, dup...)
	// update indexes
	cd.curPara = cd.curPara + len(dup)
	cd.firstRunText = cd.firstRunText + len(dup)
	// recurse
	cd.insert(paras[1:])
}

// process text within a run, until end of run
func (cd *custDecoder) processText() {
	var tok xml.Token
	for tok, cd.err = cd.dec.Token(); cd.err == nil; tok, cd.err = cd.dec.Token() {
		cd.copy() // copy captured element
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" && t.Name.Space == NAMESPACE {
				if cd.firstRunText < 0 { // no run was seen in this paragraph yet, prepare this run for saving aggregated text.
					cd.res = append(cd.res, []byte{}) // add empty place holder for future aggregated text
					cd.firstRunText = len(cd.res) - 1 // remember index of empty place holder !
				}
				cd.processTextContent()
			}
		case xml.EndElement:
			if t.Name.Local == "r" && t.Name.Space == NAMESPACE {
				return
			}
		}
	}
}

// process text. We just read the <t> tag ...
func (cd *custDecoder) processTextContent() {
	var tok xml.Token
	for tok, cd.err = cd.dec.Token(); cd.err == nil; tok, cd.err = cd.dec.Token() {
		switch t := tok.(type) {
		case xml.CharData: // that will not be copied, only aggregated, to be saved later in the placeholder.
			cd.rcontent = append(cd.rcontent, t...)
			cd.lastSaved = cd.dec.InputOffset() - 1
		case xml.EndElement:
			cd.copy() // copy the end tag, whatever it is
			if t.Name.Local == "t" && t.Name.Space == NAMESPACE {
				return
			}
		default:
			cd.copy() // by default, we copy everything, except chardata !
		}
	}
}
