// Package tools offers helpers and type for tools related operations
package tools

// ReadParameters is struct used in read tool
type ReadParameters struct {
	FilePath string `json:"file_path"`
}

type WriteParameters struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}
