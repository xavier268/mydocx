package mydocx

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var source string = filepath.Join("testFiles", "test.docx")
var targettxt string = filepath.Join("testFiles", "test-modified.txt")
var target1 string = filepath.Join("testFiles", "test-modified1.docx")
var target2 string = filepath.Join("testFiles", "test-modified-tpl2.docx")
var target3 string = filepath.Join("testFiles", "test-modified-tpl3.docx")
var target4 string = filepath.Join("testFiles", "test-modified-tpl4.docx")

func init() {
	VERBOSE = false
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

	f, err := os.OpenFile(targettxt, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "Extracted content for debugging")
	
	keys := make([]string, 0, len(pp))
	for k := range pp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for _, k := range keys {
		v := pp[k]
		fmt.Fprintf(f, "=== %q ===\n", k)
		for i, p := range v {
			fmt.Fprintf(f, "%d: %q\n", i, p)
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
		List   []string
	}{}

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
		List   []string
	}{
		Bullet: "bullet content",
		Cell:   "cell content",
		Header: "heeaaaddderrr",
		Footer: "fooooottter",
		Skip:   true,
		Title:  "MY BIG TITLE",
		List:   []string{"item 1", "item 2", "item 3", "item 23", "item 99"},
	}
	err := ModifyText(source, NewTplReplacer(c), target4)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}
}

func TestExtractXMLFiles(t *testing.T) {
	t.Log(t.Name(), "is extracting XML files and folders from test.docx")

	docxPath := source
	extractDir := filepath.Join("testFiles", "test-extracted")

	err := extractDocxXML(docxPath, extractDir)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatal(err)
	}

	t.Logf("Successfully extracted XML files to: %s", extractDir)
}

func extractDocxXML(docxPath, extractDir string) error {
	r, err := zip.OpenReader(docxPath)
	if err != nil {
		return fmt.Errorf("failed to open docx file: %w", err)
	}
	defer r.Close()

	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xml") || f.FileInfo().IsDir() {
			err := extractFile(f, extractDir)
			if err != nil {
				return fmt.Errorf("failed to extract %s: %w", f.Name, err)
			}
		}
	}

	return nil
}

func extractFile(f *zip.File, destDir string) error {
	filePath := filepath.Join(destDir, f.Name)

	if f.FileInfo().IsDir() {
		return os.MkdirAll(filePath, f.FileInfo().Mode())
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}
