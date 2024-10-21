
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/xavier268/mydocx)

# mydocx
Extract or Transform text content within a word document. Pure go, no external dependencies, no licensing fees.

## Features

* **Extract** paragraph text without modifying original word file
* **Replace** paragraph text using a *Replacer* to transform paragraph content. A go template based *Replacer* is provided, but you may design your own.

## Paragraphs

This library is designed to manage text within a paragraph, from a word docx model. **It cannot create new paragraphs.**

The tempating engine **CANNOT** be used across paragraph boundaries. You may **NOT** start an *{{if ...}}* in one paragraph and the balancing *{{end}}* in the next paragraph.

However, the Replacer **CAN**  programatically request the suppression of its containing paragraph (useful if some paragraph only occur in certain cases, but you don't want empty lines ).

## Examples

See the *testFiles* directory.
