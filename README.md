# Go DOCX Text Processor

A powerful Go library for extracting and manipulating text in Microsoft Word (DOCX) files, with zero external dependencies. Transform your documents using Go templates or custom replacers while maintaining document structure.

[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/xavier268/mydocx)
[![Go Report Card](https://goreportcard.com/badge/github.com/xavier268/mydocx)](https://goreportcard.com/report/github.com/xavier268/mydocx)

## ✨ Features

- **Text extraction** from DOCX files with two modes:
  - `ExtractText()` - Extract text with changes accepted (insertions included, deletions ignored)
  - `ExtractOriginalText()` - Extract original text with changes rejected (insertions ignored, deletions restored)
- **Word-level diff analysis** with readable output:
  - `Diff()` - Compare original vs accepted text with semantic word-level differences  
  - `PrettyPrint()` - Generate LLM-friendly diff output with `<delete>` and `<insert>` tags
  - Built on proven difflib algorithms for optimal readability
- **Text modification** using Go templates or custom replacers
- **Full document support**:
  - Main document body
  - Headers and footers
  - Tables and cells
  - Bullet points and numbered lists
- **Track changes handling** (insertions/deletions) for both extraction and modification
- **Memory support** with byte array functions (`ExtractTextBytes`, `ExtractOriginalTextBytes`)
- Minimal external dependencies (only difflib for diff functionality)
- Efficient (single pass processing)
- OOXML standard compliant
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
    // This extracts text as if all track changes were ACCEPTED
    content, err := mydocx.ExtractText("document.docx")
    if err != nil {
        log.Fatal(err)
    }

    // Extract original text as if all track changes were REJECTED
    original, err := mydocx.ExtractOriginalText("document.docx")
    if err != nil {
        log.Fatal(err)
    }

    // Both return a map[string][]string where:
    // - key is the container name (e.g., "word/document.xml", "word/footer1.xml")
    // - value is a slice of strings, one for each paragraph
    for container, paragraphs := range content {
        fmt.Printf("Content from %s (changes accepted):\n", container)
        for _, para := range paragraphs {
            fmt.Println(para)
        }
    }
    
    for container, paragraphs := range original {
        fmt.Printf("Original content from %s (changes rejected):\n", container)
        for _, para := range paragraphs {
            fmt.Println(para)
        }
    }
}
```

### Document Diff Analysis

#### Simple One-Line Analysis

```go
import "github.com/xavier268/mydocx"

func main() {
    // Get LLM-friendly diff analysis in one line
    analysis, err := mydocx.DiffAnalyse("document.docx")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Print(analysis) // Ready for LLM processing
}
```

#### Detailed Diff Processing

```go
import "github.com/xavier268/mydocx"

func main() {
    // Extract both original and accepted versions
    original, err := mydocx.ExtractOriginalText("document.docx")
    if err != nil {
        log.Fatal(err)
    }
    
    accepted, err := mydocx.ExtractText("document.docx")
    if err != nil {
        log.Fatal(err)
    }
    
    // Generate word-level diff analysis
    diffResult := mydocx.Diff(original, accepted)
    
    // Get readable diff output for LLM analysis
    prettyDiff := diffResult.PrettyPrint()
    fmt.Println(prettyDiff)
    
    // Access structured diff data
    fmt.Printf("Total containers: %d\n", diffResult.Summary.TotalContainers)
    fmt.Printf("Changed containers: %d\n", diffResult.Summary.ChangedContainers)
    fmt.Printf("Insertions: %d, Deletions: %d\n", 
        diffResult.Summary.TotalInsertions, 
        diffResult.Summary.TotalDeletions)
        
    // Process individual container diffs
    for containerName, containerDiff := range diffResult.ContainerDiffs {
        fmt.Printf("Changes in %s:\n", containerName)
        for _, op := range containerDiff.Operations {
            fmt.Printf("  %s: %q\n", op.Type, op.Text)
        }
    }
}
```

Example diff output:
```
=== DIFF SUMMARY ===
Total containers: 3
Changed containers: 1
Insertions: 2, Deletions: 1, Equal: 5

=== CONTAINER: word/document.xml ===
The document contains <delete>old content</delete><insert>new updated content</insert> here.
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

## 📝 Track Changes Support

This library provides comprehensive support for Microsoft Word track changes (revisions) with different extraction modes:

### Text Extraction Options

#### 1. Extract with Changes Accepted (`ExtractText`)
- **Deletions**: Text marked for deletion is **excluded** from the extracted content
- **Insertions**: Text marked as inserted is **included** in the extracted content
- **Result**: The extracted text represents the "accepted changes" version

#### 2. Extract Original Text (`ExtractOriginalText`)
- **Deletions**: Text marked for deletion is **included** to restore original content
- **Insertions**: Text marked as inserted is **excluded** from the extracted content  
- **Result**: The extracted text represents the original document before any changes

#### 3. Byte Array Support
Both extraction modes also support byte array input:
- `ExtractTextBytes([]byte)` - Extract with changes accepted
- `ExtractOriginalTextBytes([]byte)` - Extract original text with changes rejected

Example:
```
Document with track changes: "Hello [deleted: old] [inserted: new] world"

ExtractText result:         "Hello new world"      (changes accepted)
ExtractOriginalText result: "Hello old world"      (changes rejected)
```

### Text Modification

During text modification (`ModifyText`), track changes are handled differently:
- **Deletions**: Deletion markup is preserved unchanged in the output document
- **Insertions**: Insertion markup is preserved unchanged in the output document  
- **Templates/Replacers**: Only operate on the "clean" text (like extraction), but track changes markup is maintained in the final document

This means:
1. Your templates and replacers work with clean text (as if changes were accepted)
2. The original track changes markup is preserved in the template document
3. The output document maintains the same revision history as the input template

### Important Notes

- Track changes from the template document are preserved during modification
- If you need to work with documents without track changes, accept all changes in Word before using them as templates
- The extraction function gives you a preview of what text your templates will process

## 🔧 Advanced Features

### Template Functions

#### Built-in Functions

All go template functions are available. In addition, the following built-in functions are always available :

- `{{nl}}` - Inserts a new paragraph
- `{{version}}` - Returns version information
- `{{copyright}}` - Returns copyright text
- `{{date}}`- Returns the current date, as 2006-02-10
- `{{join}}` - Expects an array of strings and a delimiter string, returns a single concatenated string with the delimiter (see go function `strings.Join`)
- `{{keepEmpty}}` - From this point, will never remove a paragraph that becomes empty after modification.
- `{{removeEmpty}}`- From this point, non empty paragraphs that become empty after `Replacer` is applied are removed. **This is the default**.

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

## 🔄 Paragraph Management

### With Custom Replacer

The Replacer function controls paragraph creation and removal through its return value:

```go
type Replacer func(container string, text string) []string
```

1. **Remove Paragraph**
   ```go
   // Return empty slice to remove the paragraph (unless {{keepEmpty}} was called earlier)
   func myReplacer(container, text string) []string {
       if strings.Contains(text, "DELETE") {
           return []string{} // Paragraph will be removed
       }
       return []string{text}
   }
   ```

2. **Create Multiple Paragraphs**
   ```go
   // Return multiple strings to create multiple paragraphs
   func myReplacer(container, text string) []string {
       if strings.Contains(text, "DUPLICATE") {
           // Creates three identical paragraphs with the same formatting
           return []string{text, text, text}
       }
       return []string{text}
   }
   ```

Each new paragraph inherits the formatting of the original paragraph.

### With Go Templates

When using the template-based replacer (`NewTplReplacer`), paragraph management is controlled by newlines in the template output:

1. **Remove Paragraph**
   ```
   {{if .ShouldDelete}}{{else}}Original content{{end}}
   ```
   If `.ShouldDelete` is true, the empty output will remove the paragraph (unless {{keepEmpty}} was called before).

2. **Create Multiple Paragraphs**
   ```
   {{.Title}}
   Items:
   {{range .Items}} - {{.}}{{nl}}{{end}}
   Contact: {{.Contact}}
   ```

The `{{nl}}` function inserts a newline, and the template output is split on newlines to create new paragraphs. Each resulting paragraph inherits the formatting of the original paragraph. Notice how  {{range}} ... {{end}} fits within a single source paragraph but will create multiple paragraphs !

### Paragraph Creation Rules

1. **Initially empty paragraphs**
   - Initially empty paragraphs are *always* left unchanged

2. **Empty Result**
   - If the Replacer returns an empty slice → paragraph is removed 
   - If a template produces empty output → paragraph is removed   
*Note : this is the default, it can be changed with {{keepEmpty}}*

1. **Multiple Paragraphs**
   - Custom Replacer: Each string in the returned slice becomes a new paragraph
   - Template: Output is split on newlines (`\n`), each line becomes a new paragraph
   - All new paragraphs inherit formatting from the original paragraph

2. **Examples with Templates**

   ```
   // Template in document
   Dear {{.Name}},
   {{if .Premium}}Thank you for being a premium member!{{end}}
   Your balance is ${{.Balance}}.
   ```

   This template could produce:
   ```
   Dear John Doe,
   Thank you for being a premium member!
   Your balance is $100.
   ```
   Or (if not premium):
   ```
   Dear John Doe,
   Your balance is $100.
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


