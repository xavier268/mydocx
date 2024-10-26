package mydocx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"text/template"
)

// Assume the source word document contains valid go template in each paragraph
// NewTplReplacer creates a replacer that will apply the provided content struct to the template in each paragraph. Container is ignored.
// The Replacer will behave as follows :
// * If initial paragraph was empty, it is left un changed.
// * If not empty, template is executed.
// * If the template execution result is empty, the paragraph is discarded.
// * If the template execution result is not empty, it is split around \n into lines and each line is added as a separate paragraph. (you may use the function {{nl}} to gererate new lines)
// * If an error occurs during template execution, an error message is added as the last paragraph of the result.
func NewTplReplacer(content any) Replacer {
	return func(_ string, para string) []string {
		errmess := ""
		if para == "" {
			return []string{""} // leave empty original paragraph untouched.
		}

		var res = new(strings.Builder)

		tpl := template.Must(template.New(NAME + "_template").Parse(para))
		err := tpl.Execute(res, content)
		if err != nil {
			errmess = fmt.Sprintf("$$$$$$ ERROR $$$$$ : %v ", err)
			if VERBOSE {
				fmt.Println(errmess)
			}
			return []string{para, errmess}
		}
		rs := res.String()
		if rs == "" && errmess == "" {
			return nil // discard paragraph if result string is empty string and no error message.
		}
		// if not empty, split lines
		rss := strings.Split(rs, "\n")
		return rss // discard if result string is empty string.
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
