
[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/xavier268/mydocx)
[![Go Report Card](https://goreportcard.com/badge/github.com/xavier268/mydocx)](https://goreportcard.com/report/github.com/xavier268/mydocx)
# mydocx

A pure Go library to extract and transform text content within Word documents, with no external dependencies or licensing fees.

## Features

* **Text Extraction**: Extract text from multiple document locations:
  - Main document body
  - Headers and footers
  - Table cells
* **Content Organization**: Retrieve text organized by paragraphs while preserving the original Word file
* **Text Transformation**: Replace text content using customizable `Replacer` interfaces
  - Includes a built-in Go template-based `Replacer`
  - Support for custom `Replacer` implementations
* MIT Licence

## Working with Paragraphs

### Replacer

* `Replacer` takes the type of container and the original string as input, and returns a replacement string and a flag to request paragraph removal. This allows to handle differently the main document or the header/footer.

### Limitations

* New paragraphs cannot be created
* Existing paragraphs can be programmatically discarded (see test files for examples)

### Template Engine Constraints

The template engine has specific boundary limitations:
* Uses syntax defined in https://pkg.go.dev/text/template@latest
* Template directives **cannot** span across paragraph boundaries
* Example of invalid usage:
  ```
  Paragraph 1: {{if condition}}
  Paragraph 2: {{end}}
  ```

## Technical Details

### Run Management

Special consideration has been given to Word's internal text structure:

* In a word document, text is internally split across multiple "runs"
* To handle this complexity, the library:
  1. Collects text pieces from various runs
  2. Consolidates them into the first run
  3. Maintains empty subsequent runs for structural integrity

### Formatting Behavior

Due to the run consolidation process:
* The resulting paragraph will have uniform text formatting
* The formatting is inherited from the style applied to the beginning of the paragraph

## Usage Notes

* When implementing custom text transformations, consider the single-format limitation
* Refer to test files for comprehensive examples of paragraph manipulation
* The library maintains document structure while allowing powerful text modifications