package docxtransform

import (
	"fmt"
	"strings"
	"testing"
)

func TestDocModify(t *testing.T) {

	err := modifyParagraphsInDocx("test.docx", strings.ToUpper)
	if err != nil {
		fmt.Println("Error:", err)
	}

}

func TestDocExtract(t *testing.T) {

	pp, err := extractTextFromDocx("test.docx")
	if err != nil {
		fmt.Println("Error:", err)
	}
	for i, p := range pp {
		fmt.Printf("%d: %q\n", i, p)
	}

}
