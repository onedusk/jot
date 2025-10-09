// Package export provides functionality for exporting documents to various formats like JSON, YAML,
// and a special format optimized for Large Language Models (LLMs).
package export

// LLMExport represents the complete data structure for an export optimized
// for Large Language Model consumption. It includes documents, metadata, and a semantic index.
type LLMExport struct {
	Version   string         `json:"version" yaml:"version"`
	Generated string         `json:"generated" yaml:"generated"`
	Documents []LLMDocument  `json:"documents" yaml:"documents"`
	Index     *SemanticIndex `json:"index" yaml:"index"`
}

// LLMDocument represents a single document structured for LLM processing.
// It contains the core content, as well as extracted components like sections,
// code blocks, and links.
type LLMDocument struct {
	ID         string                 `json:"id" yaml:"id"`
	Title      string                 `json:"title" yaml:"title"`
	Path       string                 `json:"path" yaml:"path"`
	Content    string                 `json:"content" yaml:"content"`
	HTML       string                 `json:"html,omitempty" yaml:"html,omitempty"`
	Chunks     []Chunk                `json:"chunks" yaml:"chunks"`
	Sections   []LLMSection           `json:"sections" yaml:"sections"`
	CodeBlocks []LLMCodeBlock         `json:"code_blocks" yaml:"code_blocks"`
	Links      Links                  `json:"links" yaml:"links"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Chunk represents a segment of text from a document, typically sized for
// tasks like vector embedding or processing within a model's context window.
type Chunk struct {
	ID       string    `json:"id" yaml:"id"`
	Text     string    `json:"text" yaml:"text"`
	StartPos int       `json:"start_pos" yaml:"start_pos"`
	EndPos   int       `json:"end_pos" yaml:"end_pos"`
	Vector   []float32 `json:"vector,omitempty" yaml:"vector,omitempty"`
}

// LLMSection represents a distinct section within a document, such as a
// chapter or a part introduced by a header.
type LLMSection struct {
	ID        string `json:"id" yaml:"id"`
	Title     string `json:"title" yaml:"title"`
	Level     int    `json:"level" yaml:"level"`
	Content   string `json:"content" yaml:"content"`
	StartLine int    `json:"start_line" yaml:"start_line"`
	EndLine   int    `json:"end_line" yaml:"end_line"`
}

// LLMCodeBlock represents a block of source code extracted from a document.
type LLMCodeBlock struct {
	Language  string `json:"language" yaml:"language"`
	Content   string `json:"content" yaml:"content"`
	StartLine int    `json:"start_line" yaml:"start_line"`
}

// Links categorizes hyperlinks found within a document as either internal
// (pointing to other documents within the same collection) or external.
type Links struct {
	Internal []string `json:"internal" yaml:"internal"`
	External []string `json:"external" yaml:"external"`
}

// SemanticIndex provides a simple mechanism for semantic search, mapping
// keywords to document IDs and listing conceptual topics.
type SemanticIndex struct {
	Keywords map[string][]string `json:"keywords" yaml:"keywords"`
	Concepts []string            `json:"concepts" yaml:"concepts"`
}
