package openxml

import "encoding/xml"

// Contains simplified structure for a word document.
// Made internal because external use not recommanded.

// Define a simplified XML structure for a word document, with a focus on the relevant tags
// Only used for text extraction, since all other necessary fields are discarded.
// Text modification uses another strategy.
type SimplifiedDocument struct {
	XMLName xml.Name `xml:"document"`
	Body    Body     `xml:"body"`
	XMLNSw  string   `xml:"xmlns:w,attr"`
}

type Body struct {
	Paragraphs []Paragraph `xml:"p"`
}

type Paragraph struct {
	Runs []Run `xml:"r"`
}

type Run struct {
	Text Text `xml:"t"`
}

type Text struct {
	Value string `xml:",chardata"`
}
