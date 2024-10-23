package mydocx

// v0.1.1		first functionnal version
// v0.1.2		code cleanup, API simplification
// v0.1.3		use golang templates (optionnal)
// v0.1.4		programmatically remove paragraphs, based on their content
// v0.1.5		cleanup code, doc
// v0.1.6		reduced public API, hiding simplified xml structure.
// v0.1.7       redesign parsing for footer/header templating. Escape text replaced.

const NAME = "mydocx"

const VERSION = "0.1.7"

const COPYRIGHT = "(c) Xavier Gandillot 2024"

// if true, verbose information will be printed to stdout
var VERBOSE = false

// set to true for detailled debugging information
var debugflag = false
