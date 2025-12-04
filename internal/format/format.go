package format

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Format represents the output format
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// ParseFormat parses a format string and returns the Format
func ParseFormat(s string) (Format, error) {
	switch s {
	case "text", "":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	case "yaml":
		return FormatYAML, nil
	default:
		return FormatText, fmt.Errorf("invalid format: %s (must be text, json, or yaml)", s)
	}
}

// Write writes data to the output writer in the specified format
func Write(w io.Writer, format Format, data interface{}) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case FormatYAML:
		enc := yaml.NewEncoder(w)
		defer enc.Close()
		return enc.Encode(data)
	case FormatText:
		// Text format is handled by the command itself
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// GetFormatFromContext extracts the format from a context or returns default
func GetFormatFromContext(formatFlag string) Format {
	if formatFlag == "" {
		return FormatText
	}
	f, err := ParseFormat(formatFlag)
	if err != nil {
		// Return text as fallback
		return FormatText
	}
	return f
}
