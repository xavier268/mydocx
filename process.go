package docxtransform

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

const NAMESPACE = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

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
