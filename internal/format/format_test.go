package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Format
		wantErr bool
	}{
		{"text format", "text", FormatText, false},
		{"json format", "json", FormatJSON, false},
		{"yaml format", "yaml", FormatYAML, false},
		{"empty string defaults to text", "", FormatText, false},
		{"invalid format", "invalid", FormatText, true},
		{"xml format (invalid)", "xml", FormatText, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFormatFromContext(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Format
	}{
		{"empty string", "", FormatText},
		{"text format", "text", FormatText},
		{"json format", "json", FormatJSON},
		{"yaml format", "yaml", FormatYAML},
		{"invalid format falls back to text", "invalid", FormatText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFormatFromContext(tt.input)
			if got != tt.want {
				t.Errorf("GetFormatFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrite_JSON(t *testing.T) {
	type TestData struct {
		Name    string   `json:"name"`
		Version string   `json:"version"`
		Tags    []string `json:"tags"`
	}

	data := TestData{
		Name:    "test",
		Version: "1.0.0",
		Tags:    []string{"tag1", "tag2"},
	}

	var buf bytes.Buffer
	err := Write(&buf, FormatJSON, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify it's valid JSON
	var decoded TestData
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}

	// Verify content
	if decoded.Name != data.Name {
		t.Errorf("Name = %v, want %v", decoded.Name, data.Name)
	}
	if decoded.Version != data.Version {
		t.Errorf("Version = %v, want %v", decoded.Version, data.Version)
	}
	if len(decoded.Tags) != len(data.Tags) {
		t.Errorf("Tags length = %v, want %v", len(decoded.Tags), len(data.Tags))
	}

	// Verify it's indented (contains newlines)
	if !strings.Contains(buf.String(), "\n") {
		t.Error("JSON output should be indented (contain newlines)")
	}
}

func TestWrite_YAML(t *testing.T) {
	type TestData struct {
		Name    string   `yaml:"name"`
		Version string   `yaml:"version"`
		Tags    []string `yaml:"tags"`
	}

	data := TestData{
		Name:    "test",
		Version: "1.0.0",
		Tags:    []string{"tag1", "tag2"},
	}

	var buf bytes.Buffer
	err := Write(&buf, FormatYAML, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify it's valid YAML
	var decoded TestData
	if err := yaml.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Output is not valid YAML: %v\nOutput: %s", err, buf.String())
	}

	// Verify content
	if decoded.Name != data.Name {
		t.Errorf("Name = %v, want %v", decoded.Name, data.Name)
	}
	if decoded.Version != data.Version {
		t.Errorf("Version = %v, want %v", decoded.Version, data.Version)
	}
	if len(decoded.Tags) != len(data.Tags) {
		t.Errorf("Tags length = %v, want %v", len(decoded.Tags), len(data.Tags))
	}
}

func TestWrite_Text(t *testing.T) {
	data := map[string]string{"key": "value"}

	var buf bytes.Buffer
	err := Write(&buf, FormatText, data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Text format should not write anything (handled by command)
	if buf.Len() != 0 {
		t.Errorf("Text format should not write output, got %d bytes", buf.Len())
	}
}

func TestWrite_InvalidFormat(t *testing.T) {
	data := map[string]string{"key": "value"}

	var buf bytes.Buffer
	err := Write(&buf, Format("invalid"), data)
	if err == nil {
		t.Error("Write() should return error for invalid format")
	}
}

func TestWrite_ComplexStructure(t *testing.T) {
	type Nested struct {
		Value int `json:"value" yaml:"value"`
	}

	type Complex struct {
		String  string            `json:"string" yaml:"string"`
		Number  int               `json:"number" yaml:"number"`
		Boolean bool              `json:"boolean" yaml:"boolean"`
		Array   []string          `json:"array" yaml:"array"`
		Map     map[string]string `json:"map" yaml:"map"`
		Nested  Nested            `json:"nested" yaml:"nested"`
	}

	data := Complex{
		String:  "test",
		Number:  42,
		Boolean: true,
		Array:   []string{"a", "b", "c"},
		Map:     map[string]string{"key1": "value1", "key2": "value2"},
		Nested:  Nested{Value: 100},
	}

	// Test JSON
	t.Run("JSON complex structure", func(t *testing.T) {
		var buf bytes.Buffer
		err := Write(&buf, FormatJSON, data)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		var decoded Complex
		if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
			t.Fatalf("Invalid JSON: %v", err)
		}

		if decoded.String != data.String {
			t.Errorf("String = %v, want %v", decoded.String, data.String)
		}
		if decoded.Number != data.Number {
			t.Errorf("Number = %v, want %v", decoded.Number, data.Number)
		}
		if decoded.Boolean != data.Boolean {
			t.Errorf("Boolean = %v, want %v", decoded.Boolean, data.Boolean)
		}
	})

	// Test YAML
	t.Run("YAML complex structure", func(t *testing.T) {
		var buf bytes.Buffer
		err := Write(&buf, FormatYAML, data)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		var decoded Complex
		if err := yaml.Unmarshal(buf.Bytes(), &decoded); err != nil {
			t.Fatalf("Invalid YAML: %v", err)
		}

		if decoded.String != data.String {
			t.Errorf("String = %v, want %v", decoded.String, data.String)
		}
		if decoded.Number != data.Number {
			t.Errorf("Number = %v, want %v", decoded.Number, data.Number)
		}
		if decoded.Boolean != data.Boolean {
			t.Errorf("Boolean = %v, want %v", decoded.Boolean, data.Boolean)
		}
	})
}

func TestWrite_EmptyData(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		data   interface{}
	}{
		{"empty map JSON", FormatJSON, map[string]string{}},
		{"empty map YAML", FormatYAML, map[string]string{}},
		{"empty slice JSON", FormatJSON, []string{}},
		{"empty slice YAML", FormatYAML, []string{}},
		{"nil JSON", FormatJSON, nil},
		{"nil YAML", FormatYAML, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Write(&buf, tt.format, tt.data)
			if err != nil {
				t.Errorf("Write() error = %v", err)
			}

			// Should produce valid output even for empty data
			if tt.format == FormatJSON {
				var decoded interface{}
				if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
					t.Errorf("Invalid JSON: %v\nOutput: %s", err, buf.String())
				}
			} else if tt.format == FormatYAML {
				var decoded interface{}
				if err := yaml.Unmarshal(buf.Bytes(), &decoded); err != nil {
					t.Errorf("Invalid YAML: %v\nOutput: %s", err, buf.String())
				}
			}
		})
	}
}

