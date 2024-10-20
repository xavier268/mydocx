package docxtransform

import (
	"fmt"
	"strings"
	"testing"
)

func TestDocRead(t *testing.T) {

	err := modifyParagraphsInDocx("test.docx", strings.ToUpper)
	if err != nil {
		fmt.Println("Error:", err)
	}

}
