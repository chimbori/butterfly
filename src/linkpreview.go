package main

import (
	"errors"
	"net/url"
	"strings"

	"go.chimbori.app/butterfly/conf"
)

// Validates a URL provided by the user, and returns a formatted URL string.
func validateUrl(userUrl string) (string, error) {
	if userUrl == "" {
		return "", errors.New("missing url")
	}

	if !strings.HasPrefix(userUrl, "https://") && !strings.HasPrefix(userUrl, "http://") {
		userUrl = "https://" + userUrl
	}

	u, err := url.Parse(userUrl)
	if err != nil {
		return "", errors.New("invalid url")
	}

	if !isAuthorized(u) {
		return "", errors.New("domain " + u.Host + " not authorized")
	}

	return u.String(), nil
}

// isAuthorized returns true if the given URL's domain is in the list of authorized domains.
func isAuthorized(u *url.URL) bool {
	hostname := u.Hostname()
	for _, domain := range conf.Config.LinkPreview.Domains {
		if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
			return true
		}
	}
	return false
}
