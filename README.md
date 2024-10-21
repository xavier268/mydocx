# mydocx
Transform text content within a word document. Pure go, no dependencies.

## Features

* **Extract** paragraph text without modifying original word file
* **Replace** paragraph text, keeping the formatting, creating a new word file.
    * you may use the standard go templating engine, but only inside each paragraph.

## Paragraphs

This librairy can only manage text within a paragraph, from a word model. **It cannot create new paragraphs.**

The tempating engine may **NOT** be used across paragraph boundaries. You may **NOT** start an {{if ...}} in one paragraph and leaveve the ... {{end}} in the next paragraph.

However, the Replacer function may now programatically trigger the suppression of its containing paragraph.
