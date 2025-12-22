package renderer

import "strings"

// xmlReplacer performs efficient single-pass XML escaping
var xmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\"", "&quot;",
	"'", "&apos;",
)

func escapeXML(s string) string {
	return xmlReplacer.Replace(s)
}
