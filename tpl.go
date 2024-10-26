package mydocx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// Function map for template
var functionMap template.FuncMap

// register predefined functions
func init() {

	// nl takes no argument and returns "\n"
	RegisterTplFunction("nl", func() string { return "\n" })

	// version takes no argument are returns versionning information.
	RegisterTplFunction("version", func() string { return NAME + "-" + VERSION })

	// copyright takes no argument and returns copyright information
	RegisterTplFunction("copyright", func() string { return COPYRIGHT })

	// date takes no argument and returns current date
	RegisterTplFunction("date", func() string { return time.Now().Format("2006-01-02") })

	// join takes a slice of strings and returns a single string, joined with the provided delimiter
	RegisterTplFunction("join", func(args []string, delim string) string { return strings.Join(args, "\n") })

	// allowDiscard will discard empty paragraphs.
	RegisterTplFunction("removeEmpty", func() string { REMOVE_EMPTY_PARAGRAPH = true; return "" })

	// preventDiscard will always keep empty paragraphs.
	RegisterTplFunction("keepEmpty", func() string { REMOVE_EMPTY_PARAGRAPH = false; return "" })
}

// Register a new function that will be available when parsing templates.
// Empty names or nil functions are ignored.
// Each function must have either a single return value, or two return values of which the second has type error.
// In that case, if the second (error) return value evaluates to non-nil during execution,
// execution terminates and Execute returns that error.
func RegisterTplFunction(name string, function any) {
	if functionMap == nil {
		functionMap = make(template.FuncMap)
	}
	if name != "" && function != nil {
		if VERBOSE {
			fmt.Printf("Registering template function {{%s}} : %T\n", name, function)
		}
		functionMap[name] = function
	}
}

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

		tpl, err := template.New(NAME + "_template").Funcs(functionMap).Parse(para)
		if err != nil {
			errmess = fmt.Sprintf("$$$$$$ ERROR $$$$$ : %v ", err)
			if VERBOSE {
				fmt.Println(para, errmess)
			}
			return []string{para, errmess}
		}
		err = tpl.Execute(res, content)
		if err != nil {
			errmess = fmt.Sprintf("$$$$$$ ERROR $$$$$ : %v ", err)
			if VERBOSE {
				fmt.Println(para, errmess)
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
