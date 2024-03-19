package headers

import (
	"net/http"
	"strings"
)

func Accept(req *http.Request) string {
	if ct, _, _ := strings.Cut(req.Header.Get("Accept"), ","); ct != "*/*" {
		return ct
	}

	return ""
}
