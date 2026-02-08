package core

import "testing"

func TestGetCanonicalUserAgent(t *testing.T) {
	tests := []struct {
		userAgent string
		expected  string
	}{
		// Browsers
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Chrome",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Firefox",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
			"Safari",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
			"Edge",
		},

		// Mobile browsers
		{
			"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
			"Mobile Safari",
		},
		{
			"Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 Chrome/120.0.6099.129 Mobile Safari/537.36",
			"Mobile Safari",
		},

		// Bots
		{
			"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			"Googlebot",
		},
		{
			"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
			"Bingbot",
		},
		{
			"Mozilla/5.0 (compatible; Applebot/2.1; +http://www.apple.com/go/applebot)",
			"Applebot",
		},
		{
			"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
			"FacebookBot",
		},

		// Tools
		{
			"curl/7.88.1",
			"curl",
		},
		{
			"Mozilla/5.0 (compatible; Wget/1.21; +https://www.gnu.org/software/wget/)",
			"wget",
		},

		// Unknown
		{
			"",
			"Unknown",
		},
		{
			"WeirdBot/1.0",
			"Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.userAgent, func(t *testing.T) {
			result := GetCanonicalUserAgent(tt.userAgent)
			if result != tt.expected {
				t.Errorf("GetCanonicalUserAgent(%q) = %q, want %q", tt.userAgent, result, tt.expected)
			}
		})
	}
}
