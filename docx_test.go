package mydocx

import (
	"fmt"
	"path/filepath"
	"testing"
)

var source string = filepath.Join("testFiles", "test.docx")
var target1 string = filepath.Join("testFiles", "test-modified1.docx")
var target2 string = filepath.Join("testFiles", "test-modified-tpl2.docx")
var target3 string = filepath.Join("testFiles", "test-modified-tpl3.docx")
var target4 string = filepath.Join("testFiles", "test-modified-tpl4.docx")

func init() {
	// debugFlag = true
}
func TestDocExtract(t *testing.T) {

	pp, err := ExtractText(source)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}
	fmt.Println()
	for i, p := range pp {
		fmt.Printf("%d: %q\n", i, p)
	}

}

func TestDocModify1(t *testing.T) {

	//err := ModifyText(source, strings.ToUpper, target)
	err := ModifyText(source, nil, target1)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}

}

func TestDocModifyTpl2(t *testing.T) {

	c := struct {
		Name string
		Age  int
	}{
		Name: "John Doe",
		Age:  30,
	}

	err := ModifyText(source, NewTplReplacer(c), target2)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}

}

func TestDocModifyTpl3(t *testing.T) {

	c := struct {
		Name string
		Age  int
	}{
		Name: "John Doe",
		Age:  12,
	}

	err := ModifyText(source, NewTplReplacer(c), target3)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}

}

// This will use in paragraph formating and special chars
func TestDocModifyTpl4(t *testing.T) {

	c := struct {
		Name string
		Age  int
	}{
		Name: "John\n\r\tDoe>>><<", // \n\r will be treated as space, < and > will be correctly escaped.
		Age:  12,
	}
	err := ModifyText(source, NewTplReplacer(c), target4)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}
}
