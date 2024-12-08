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
// v0.2.0 change Replacer, allow add/destroy paragraphs
// v0.2.1 add template functions
// v0.2.2 add option never to remove paragraphs that become empty
// v0.2.3 use io.Reader/Writer as an alternative to file name for extract/modify
// v0.2.4 fix versionning and release issue
// v0.2.5 remove utility function ReeaderToZip from public API

const (
	AUTHOR      = "Xavier Gandillot"
	DESCRIPTION = "A simple library to modify Microsoft Word .docx documents with go templates"
	NAME        = "mydocx"
	VERSION     = "0.2.5"
	COPYRIGHT   = "(c) Xavier Gandillot 2024"
	NAMESPACE   = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
)

var (
	// if true, verbose information will be printed to stdout
	VERBOSE = false

	// If set to true, paragraphs that become empty after replacement are removed (paragraphs that were initially empty, before replacement, are never removed).
	// If false, paragraphs that become empty are kept.
	// Use the functions {{removeEmpty}}  or {{keepEmpty}} in the source word document to set this value.
	// You may also set this variable directly from code.
	// Default is true.
	REMOVE_EMPTY_PARAGRAPH bool = true

	// pattern to select which xml container will be transformed
	containerPattern = regexp.MustCompile(`^(word/document\.xml)|(word/footer[0-9]+\.xml)|(word/header[0-9]+\.xml)$`)

	// set to true for detailed debugging information
	debugflag = false
)
