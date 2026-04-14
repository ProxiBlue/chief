package prd

import "path/filepath"

// LoadPRD reads and parses a PRD markdown file from the given path.
func LoadPRD(path string) (*PRD, error) {
	return ParseMarkdownPRD(path)
}

// ExtractPRDName returns the PRD name from its file path.
// For .chief/prds/<name>/prd.md, returns <name>.
func ExtractPRDName(prdPath string) string {
	dir := filepath.Dir(prdPath)
	name := filepath.Base(dir)
	if name == "." || name == "/" {
		return "main"
	}
	return name
}
