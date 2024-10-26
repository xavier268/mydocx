# Go DOCX Text Processor

A powerful Go library for extracting and manipulating text in Microsoft Word (DOCX) files, with zero external dependencies. Transform your documents using Go templates or custom replacers while maintaining document structure.

[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/xavier268/mydocx)
[![Go Report Card](https://goreportcard.com/badge/github.com/xavier268/mydocx)](https://goreportcard.com/report/github.com/xavier268/mydocx)

## ✨ Features

- Text extraction from DOCX files
- Text modification using Go templates or custom replacers
- Support for:
  - Main document body
  - Headers and footers
  - Tables and cells
  - Bullet points and numbered lists
- Zero external dependencies
- MIT License

## 🚀 Installation

```bash
go get github.com/xavier268/mydocx
```

## 📖 Quick Start

### Text Extraction

```go
import "github.com/xavier268/mydocx"

func main() {
    // Extract text from all document parts (main body, headers, footers)
    content, err := mydocx.ExtractText("document.docx")
    if err != nil {
        log.Fatal(err)
    }

    // content is a map[string][]string where:
    // - key is the container name (e.g., "word/document.xml", "word/footer1.xml")
    // - value is a slice of strings, one for each paragraph
    for container, paragraphs := range content {
        fmt.Printf("Content from %s:\n", container)
        for _, para := range paragraphs {
            fmt.Println(para)
        }
    }
}
```

### Using Go Templates

```go
import "github.com/xavier268/mydocx"

func main() {
    // Define template data
    data := struct {
        Name    string
        Company string
        Date    string
    }{
        Name:    "John Doe",
        Company: "ACME Corp",
        Date:    time.Now().Format("2006-01-02"),
    }

    // Create a template-based replacer
    replacer := mydocx.NewTplReplacer(data)

    // Modify the document
    err := mydocx.ModifyText("template.docx", replacer, "output.docx")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Custom Replacer

```go
// Define your custom replacer
func myReplacer(container, text string) []string {
    // container: "word/document.xml", "word/footer1.xml", etc.
    // text: original paragraph text
    // Return:
    // - empty slice to remove the paragraph
    // - slice with multiple strings to create multiple paragraphs
    // - slice with one string to replace paragraph content
    
    switch {
    case strings.Contains(text, "DELETE"):
        return []string{} // Remove paragraph
    case strings.Contains(text, "DUPLICATE"):
        return []string{text, text} // Duplicate paragraph
    default:
        return []string{strings.ToUpper(text)} // Convert to uppercase
    }
}

// Use your replacer
err := mydocx.ModifyText("input.docx", myReplacer, "output.docx")
```

## 🔧 Advanced Features

### Template Functions

#### Built-in Functions
- `{{nl}}` - Inserts a new paragraph
- `{{version}}` - Returns version information
- `{{copyright}}` - Returns copyright text

#### Register Custom Functions

```go
// Register a custom function
mydocx.RegisterTplFunction("upper", strings.ToUpper)

// Use in template
// {{upper .Name}}
```

### Template Guidelines

1. Each paragraph is an independent template
2. Templates cannot span across paragraphs
3. Valid example:
   ```
   Hello {{.Name}}!
   Your order #{{.OrderID}} has been processed.
   ```

4. Invalid example:
   ```
   Hello {{if .Premium}}
   Premium customer {{.Name}}!
   {{else}}
   Valued customer {{.Name}}!
   {{end}}
   ```

## ⚙️ Technical Details

### Word Run Management

Microsoft Word splits text into "runs" - segments sharing the same formatting. This creates challenges for text replacement:

```
Example: "Hello {{.Name}}!" might be split into:
Run 1: "Hello "
Run 2: "{{.Name"
Run 3: "}}!"
```

Our solution:
1. Consolidates all runs in a paragraph into the first run
2. Processes the complete text with a `Replacer`
3. Creates new paragraphs for each line in the result

⚠️ **Important**: Due to this approach, the entire paragraph will inherit the formatting from its beginning.

### Tables and Lists

- Tables and lists are fully supported
- Each cell must contain at least one paragraph (even if empty)
- Word will show an error when opening files with empty cells but can recover

## 🚨 Limitations

1. **Formatting**
   - Paragraph formatting is unified based on the first run
   - In-paragraph formatting variations are lost

2. **Template Boundaries**
   - Templates must be contained within a single paragraph
   - Cross-paragraph templates are not supported

3. **Table Cells**
   - Avoid creating completely empty cells
   - Always include at least one (empty) paragraph in cells

## 📚 Resources

- [Go Template Documentation](https://pkg.go.dev/text/template@latest)
- [Example Files](./testFiles/)
- [API Documentation](https://pkg.go.dev/github.com/xavier268/mydocx)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📝 License

MIT License - See LICENSE file for details

