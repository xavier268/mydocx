package mydocx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"text/template"
)

// Assume the source word document contains valid template in each paragraph
// (CAUTION template logic cannot extend beyond paragraph boundaries !)
// NewTplReplacer creates a replacer that will apply the provided content struct to the template in each paragraph. Container is ignored.
// By default, this Replacer will discard the entire paragraph if it was not initially empty but becomes empty when executing the template.
// If an error occurs during template conversion, the text of the error is returned, together with the original text that triggered the error.
func NewTplReplacer(content any) Replacer {
	return func(_ string, para string) (string, bool) {

		if para == "" {
			return "", true // leave epty original paragraph untouched.
		}

		var res = new(strings.Builder)

		tpl := template.Must(template.New(NAME).Parse(para))
		err := tpl.Execute(res, content)
		if err != nil {
			mess := para + " ***ERROR*** " + err.Error()
			if VERBOSE {
				fmt.Println(mess)
			}
			return mess, false
		}
		rs := res.String()
		return rs, rs == "" // discard if result string is empty string.
	}
}

// same as NewTplReplacer but will never discard empty paragraphs.
func NewTplReplacerNoDiscard(content any) Replacer {
	return func(_ string, para string) (string, bool) {

		if para == "" {
			return "", true // leave epty original paragraph untouched.
		}
		var res = new(strings.Builder)

		tpl := template.Must(template.New("docx").Parse(para))
		err := tpl.Execute(res, content)
		if err != nil {
			mess := para + " ***ERROR*** " + err.Error()
			if VERBOSE {
				fmt.Println(mess)
			}
			return mess, false
		}
		return res.String(), false
	}
}

// Escape text for inclusion in xml.
// Panic on error - should never happen ;-)
func xmlEscape(source []byte) []byte {
	escmess := new(bytes.Buffer)
	if err := xml.EscapeText(escmess, source); err != nil {
		panic("cannot escape  : " + err.Error())
	}
	return escmess.Bytes()
}
