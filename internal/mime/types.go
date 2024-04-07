package mime

import (
	"github.com/agukrapo/simpler-mock-server/internal/bimap"
	log "github.com/sirupsen/logrus"
)

type (
	Type      string
	Extension string
)

type Types struct {
	mapping *bimap.M[Extension, Type]
}

func New(ext2MIMEType map[string]string) *Types {
	mapping := defaults()
	for e, t := range ext2MIMEType {
		mapping.Put(Extension(e), Type(t))
	}

	return &Types{
		mapping: mapping,
	}
}

func (t *Types) Extension(in Type) Extension {
	if in == "" {
		return "json"
	}

	if ext, ok := t.mapping.GetByValue(in); ok {
		return ext
	}
	log.Warnf("Unmapped MIME type %q", in)
	return "json"
}

func (t *Types) Type(in Extension) Type {
	if in == "" {
		return "application/json"
	}

	if ext, ok := t.mapping.GetByKey(in); ok {
		return ext
	}
	log.Warnf("Unmapped extension %q", in)
	return "application/json"
}

func defaults() *bimap.M[Extension, Type] {
	out := bimap.New[Extension, Type](nil)
	out.Put("txt", "text/plain")
	out.Put("json", "application/json")
	out.Put("yaml", "text/yaml")
	out.Put("xml", "application/xml")
	out.Put("html", "text/html")
	out.Put("csv", "text/csv")

	return out
}
