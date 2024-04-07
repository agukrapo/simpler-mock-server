package mime

import (
	"fmt"
	"testing"
)

func TestTypes_Extension(t *testing.T) {
	types := New(map[string]string{"foo": "bar"})
	tests := []struct {
		in  Type
		out Extension
	}{
		{"yadda", "json"},
		{"text/plain", "txt"},
		{"application/json", "json"},
		{"text/yaml", "yaml"},
		{"application/xml", "xml"},
		{"text/html", "html"},
		{"text/csv", "csv"},
		{"bar", "foo"},
		{"", "json"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s -> %s", tt.in, tt.out), func(t *testing.T) {
			if got := types.Extension(tt.in); got != tt.out {
				t.Errorf("Extension() = %v, want %v", got, tt.out)
			}
		})
	}
}
