package mydocx

import "regexp"

// v0.1.1 first functional version
// v0.1.2 code cleanup, API simplification
// v0.1.3 use golang templates (optionnal)
// v0.1.4 programmatically remove paragraphs, based on their content
// v0.1.5 cleanup code, doc
// v0.1.6 reduced public API, hiding simplified xml structure.
// v0.1.7 redesign modify parsing for footer/header templating. Escape text replaced.
// v0.1.8 redesign extract parsing for simplicity and robustness. Extends to footer/header. Remove internal package.
// v0.1.9 change replacer interface to access container name
// v0.1.10 typos, readme, documentation ...

const NAME = "mydocx"

const VERSION = "0.1.10"

const COPYRIGHT = "(c) Xavier Gandillot 2024"

// if true, verbose information will be printed to stdout
var VERBOSE = false

// set to true for detailed debugging information
var debugflag = false

const NAMESPACE = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

var containerPattern = regexp.MustCompile(`^(word/document\.xml)|(word/footer[0-9]+\.xml)|(word/header[0-9]+\.xml)$`)
