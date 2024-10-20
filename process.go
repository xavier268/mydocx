package docxtransform

import "encoding/xml"

// TODO ... (no change at this stage)
// Le pb est de ne pas modifier les bytes en dehors du text des runs. Peut-^tre en travaillant directement sur les bytes, en en utilisant unmarshal pour trouver les frontireres des balises que l'on cherche ?
func processDocumentXML(documentContent []byte, replacer func(string) string) (modifiedDocument []byte, err error) {
	_ = replacer // for the compiler to not complain about an unused variable
	var doc Document

	// Unmarshal
	err = xml.Unmarshal(documentContent, &doc)
	if err != nil {
		return nil, err
	}

	// modify
	// do something here ... nothing yet.

	// marshal back
	modifiedDocument, err = xml.MarshalIndent(doc, "", " ")
	return modifiedDocument, err
}
