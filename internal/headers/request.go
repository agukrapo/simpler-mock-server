package headers

import (
	"net/http"
	"strings"

	"github.com/agukrapo/simpler-mock-server/internal/mime"
)

func Accept(req *http.Request) mime.Type {
	if ct, _, _ := strings.Cut(req.Header.Get("Accept"), ","); ct != "*/*" {
		return mime.Type(ct)
	}

	return ""
}
