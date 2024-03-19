package filesystem

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_splitBase(t *testing.T) {
	tests := []struct {
		base         string
		expectedName string
		expectedExt  string
		expectedErr  error
	}{
		{"invoice.pdf", "invoice", "pdf", nil},
		{"", "", "", errors.New("input is empty")},
		{".gitignore", "", "gitignore", nil},
		{"a.xml.json", "a.xml", "json", nil},
	}
	for _, tt := range tests {
		t.Run(tt.base, func(t *testing.T) {
			actualName, actualExt, actualErr := splitBase(tt.base)
			assert.Equal(t, tt.expectedErr, actualErr)
			assert.Equal(t, tt.expectedName, actualName)
			assert.Equal(t, tt.expectedExt, actualExt)
		})
	}
}
