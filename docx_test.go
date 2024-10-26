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
	VERBOSE = true
	debugflag = false
}
func TestDocExtract0(t *testing.T) {

	pp, err := ExtractText(source)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}
	fmt.Println()
	for k, v := range pp {
		fmt.Printf("=== %q ===\n", k)
		for i, p := range v {
			fmt.Printf("%d: %q\n", i, p)
		}
	}

}

func TestDocModify1(t *testing.T) {

	t.Log(t.Name(), "is using a nil replacer")

	err := ModifyText(source, nil, target1)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}

}

func TestDocModifyTpl2(t *testing.T) {

	t.Log(t.Name(), "is using a template replacer, with invalid fields")
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

	t.Log(t.Name(), "is using a template replacer, with valid but empty fields")
	c := struct {
		Bullet string
		Cell   string
		Header string
		Footer string
		Skip   bool
		Title  string
	}{
		Bullet: "",
		Cell:   " ", // a known issue : if a template returns "" in a cell, the paragraph is removed, leaving a cell possibly with no pragraph at all, and word will complain (not open office, since the standard accepts that)
		Header: "",
		Footer: "",
		Skip:   false,
		Title:  "",
	}

	err := ModifyText(source, NewTplReplacer(c), target3)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}

}

// This will use in paragraph formating and special chars
func TestDocModifyTpl4(t *testing.T) {

	t.Log(t.Name(), "is using a template replacer, with valid non empty fields")

	c := struct {
		Bullet string
		Cell   string
		Header string
		Footer string
		Skip   bool
		Title  string
	}{
		Bullet: "bullet content",
		Cell:   "cell content",
		Header: "heeaaaddderrr",
		Footer: "fooooottter",
		Skip:   true,
		Title:  "MY BIG TITLE",
	}
	err := ModifyText(source, NewTplReplacer(c), target4)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}
}
