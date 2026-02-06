package core

import (
	"crypto/sha256"
	"fmt"
	"html"
	"strings"
)

func SHA256(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// SafeWordBreakUrl HTML-escapes a URL first, then inserts <wbr> before each "/".
// This preserves the line-breaking behavior without risking XSS, because
// the escaping happens before the raw HTML insertion.
func SafeWordBreakUrl(rawUrl string) string {
	escaped := html.EscapeString(rawUrl)
	return strings.ReplaceAll(escaped, "/", "<wbr>/")
}
