package mydocx

import (
	"strings"
	"text/template"
)

// Assume the source word document contains valid template in each paragraph
// (CAUTION template logic cannot extend beyond paragraph boundaries !)
// NewTplReplacer creates a replacer that will apply the provided content struct to the template in each paragraph.
// If an error occurs during template conversion, the text of the error is returned, together with the original text that triggered the error.
func NewTplReplacer(content any) Replacer {
	return func(para string) (string, bool) {

		if para == "" {
			return "", true // leave epty original paragraph untouched.
		}

		var res = new(strings.Builder)

		tpl := template.Must(template.New("docx").Parse(para))
		err := tpl.Execute(res, content)
		if err != nil {
			return para + " ***ERROR*** " + err.Error(), false
		}
		rs := res.String()
		return rs, rs == "" // discard if result string is empty string.
	}
}
