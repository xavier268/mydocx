package mydocx

import "fmt"

// debug helper
func (cd *custDecoder) debug(message ...any) {

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
