package docxtransform

import (
	"fmt"
	"strings"
	"testing"
)

func TestDocModify(t *testing.T) {

	DEBUG = true
	err := modifyParagraphsInDocx("test.docx", strings.ToUpper)
	if err != nil {
		fmt.Println("Error:", err)
	}
	DEBUG = false
}

func TestDocExtract(t *testing.T) {
	DEBUG = true
	pp, err := extractTextFromDocx("test.docx")
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println()
	for i, p := range pp {
		fmt.Printf("%d: %q\n", i, p)
	}
	DEBUG = false
}
