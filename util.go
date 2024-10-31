package mydocx

import (
	"archive/zip"
	"fmt"
	"io"
)

// Helper function to read a file from a zip archive
func readFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

// Helper function to copy unmodified files to the new zip
func copyFileToZip(zipWriter *zip.Writer, file *zip.File) error {
	readCloser, err := file.Open()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	writer, err := zipWriter.Create(file.Name)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, readCloser)
	return err
}

// debug helper
func (cd *custDecoder) debug(message ...any) {

	if !debugflag {
		return
	}

	fmt.Print("Debug : ")
	for _, m := range message {
		fmt.Print(m)
		fmt.Print(" ")
	}
	fmt.Println("rcontent = ", (string)(cd.rcontent))
	cd.dumpRes()

}

// debug helper
func (cd *custDecoder) dumpRes() {

	if !debugflag {
		return
	}

	fmt.Println("\nRes =")
	for i, s := range cd.res {
		h := ""
		if i == cd.curPara {
			h = h + "p "
		}
		if i == cd.firstRunText {
			h = h + "r0 "
		}
		h = (h + "                ")[:8]
		fmt.Printf("%d:%s%q\n", i, h, (string)(s))
	}
	fmt.Println()
}
