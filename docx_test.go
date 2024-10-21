package mydocx

import (
	"fmt"
	"path/filepath"
	"testing"
)

var source string = filepath.Join("testFiles", "test.docx")
var target1 string = filepath.Join("testFiles", "test-modified.docx")
var target2 string = filepath.Join("testFiles", "test-modified-tpl.docx")

func TestDocExtract(t *testing.T) {

	debugFlag = true
	pp, err := ExtractText(source)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println()
	for i, p := range pp {
		fmt.Printf("%d: %q\n", i, p)
	}
	debugFlag = false
}

func TestDocModify(t *testing.T) {

	debugFlag = true
	//err := ModifyText(source, strings.ToUpper, target)
	err := ModifyText(source, nil, target1)
	if err != nil {
		fmt.Println("Error:", err)
	}
	debugFlag = false
}

func TestDocModifyTpl(t *testing.T) {

	c := struct {
		Name string
		Age  int
	}{
		Name: "John Doe",
		Age:  30,
	}

	debugFlag = true
	err := ModifyText(source, NewTplReplacer(c), target2)
	if err != nil {
		fmt.Println("Error:", err)
	}
	debugFlag = false
}