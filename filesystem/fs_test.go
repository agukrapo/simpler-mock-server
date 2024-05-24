package filesystem

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

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

func Test_parsePrefix(t *testing.T) {
	type args struct {
		name   string
		status int
	}
	tests := []struct {
		name       string
		args       args
		wantName   string
		wantStatus int
		wantDelay  time.Duration
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name:       "no prefix",
			args:       args{name: "invoice.pdf", status: http.StatusOK},
			wantName:   "invoice.pdf",
			wantStatus: http.StatusOK,
			wantDelay:  0,
			wantErr:    assert.NoError,
		},
		{
			name:       "status prefix",
			args:       args{name: "500___invoice.pdf", status: http.StatusOK},
			wantName:   "invoice.pdf",
			wantStatus: http.StatusInternalServerError,
			wantDelay:  0,
			wantErr:    assert.NoError,
		},
		{
			name:       "delay prefix",
			args:       args{name: "9m___invoice.pdf", status: http.StatusCreated},
			wantName:   "invoice.pdf",
			wantStatus: http.StatusCreated,
			wantDelay:  9 * time.Minute,
			wantErr:    assert.NoError,
		},
		{
			name:       "status and delay prefix",
			args:       args{name: "202.3s___invoice.pdf", status: http.StatusBadGateway},
			wantName:   "invoice.pdf",
			wantStatus: http.StatusAccepted,
			wantDelay:  3 * time.Second,
			wantErr:    assert.NoError,
		},
		{
			name:       "delay and status prefix",
			args:       args{name: "123h.400___invoice.pdf", status: http.StatusBadGateway},
			wantName:   "invoice.pdf",
			wantStatus: http.StatusBadRequest,
			wantDelay:  123 * time.Hour,
			wantErr:    assert.NoError,
		},
		{
			name:       "invalid prefix",
			args:       args{name: "nonsense___invoice.pdf", status: http.StatusContinue},
			wantName:   "invoice.pdf",
			wantStatus: http.StatusContinue,
			wantDelay:  0,
			wantErr:    err(`invalid prefix nonsense`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := parsePrefix(tt.args.name, tt.args.status)
			if !tt.wantErr(t, err, fmt.Sprintf("parsePrefix(%v, %v)", tt.args.name, tt.args.status)) {
				return
			}
			assert.Equalf(t, tt.wantName, got, "parsePrefix(%v, %v)", tt.args.name, tt.args.status)
			assert.Equalf(t, tt.wantStatus, got1, "parsePrefix(%v, %v)", tt.args.name, tt.args.status)
			assert.Equalf(t, tt.wantDelay, got2, "parsePrefix(%v, %v)", tt.args.name, tt.args.status)
		})
	}
}

func err(msg string) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, _ ...interface{}) bool {
		return assert.EqualError(t, err, msg)
	}
}
