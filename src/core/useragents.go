package core

import "regexp"

type userAgentPattern struct {
	regex     *regexp.Regexp
	canonical string
}

var userAgentPatterns []*userAgentPattern

func init() {
	patterns := []struct {
		pattern   string
		canonical string
	}{
		// Mobile browser patterns (more specific - must come before desktop patterns)
		{`(?i)mobile.*safari`, "Mobile Safari"},
		{`(?i)android.*chrome`, "Chrome Mobile"},
		{`(?i)android.*firefox`, "Firefox Mobile"},
		{`(?i)iphone.*safari`, "Mobile Safari"},

		// Browser patterns (order matters - more specific patterns first)
		{`(?i)edg/[\d.]+`, "Edge"},
		{`(?i)chrome/[\d.]+`, "Chrome"},
		{`(?i)firefox/[\d.]+`, "Firefox"},
		{`(?i)safari/[\d.]+`, "Safari"},
		{`(?i)opera/[\d.]+`, "Opera"},

		// Bot patterns
		{`(?i)googlebot`, "Googlebot"},
		{`(?i)bingbot`, "Bingbot"},
		{`(?i)slurp`, "Slurp"},
		{`(?i)duckduckbot`, "DuckDuckBot"},
		{`(?i)baiduspider`, "Baiduspider"},
		{`(?i)yandexbot`, "Yandexbot"},
		{`(?i)applebot`, "Applebot"},
		{`(?i)facebookexternalhit`, "FacebookBot"},
		{`(?i)twitterbot`, "TwitterBot"},
		{`(?i)linkedinbot`, "LinkedInBot"},
		{`(?i)whatsapp`, "WhatsApp"},
		{`(?i)telegram`, "Telegram"},

		// Crawler patterns
		{`(?i)curl`, "curl"},
		{`(?i)wget`, "wget"},
		{`(?i)ruby`, "Ruby"},
		{`(?i)python`, "Python"},
		{`(?i)java`, "Java"},
		{`(?i)golang`, "Go"},
		{`(?i)node.js`, "Node.js"},

		// Generic fallback
		{`.*`, "Unknown"},
	}

	userAgentPatterns = make([]*userAgentPattern, 0, len(patterns))
	for _, p := range patterns {
		if regex, err := regexp.Compile(p.pattern); err == nil {
			userAgentPatterns = append(userAgentPatterns, &userAgentPattern{
				regex:     regex,
				canonical: p.canonical,
			})
		}
	}
}

// GetCanonicalUserAgent returns the canonical name for a user agent string.
func GetCanonicalUserAgent(userAgent string) string {
	for _, pattern := range userAgentPatterns {
		if pattern.regex.MatchString(userAgent) {
			return pattern.canonical
		}
	}
	return "Unknown"
}
