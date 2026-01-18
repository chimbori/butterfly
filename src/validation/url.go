package validation

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"chimbori.dev/butterfly/db"
)

// ValidateUrl validates a URL provided by the user, and returns a formatted URL as a string.
func ValidateUrl(ctx context.Context, q *db.Queries, userUrl string) (validatedUrl string, hostname string, err error) {
	if userUrl == "" {
		return "", "", errors.New("missing url")
	}

	if !strings.HasPrefix(userUrl, "https://") && !strings.HasPrefix(userUrl, "http://") && !strings.Contains(userUrl, "://") {
		userUrl = "https://" + userUrl
	}

	u, err := url.Parse(userUrl)
	if err != nil {
		return "", "", errors.New("invalid url")
	}

	authorized, err := IsAuthorized(ctx, q, u)
	if err != nil {
		return "", u.Hostname(), err
	}

	if !authorized {
		return "", u.Hostname(), errors.New("domain " + u.Hostname() + " not authorized")
	}

	return u.String(), u.Hostname(), nil
}

// IsAuthorized returns true if the given URL’s domain is in the list of authorized domains.
// As a side effect, if the domain is not authorized and doesn’t exist in the database,
// it will be added (default blocked) for future triage.
func IsAuthorized(ctx context.Context, q *db.Queries, u *url.URL) (bool, error) {
	hostname := u.Hostname()
	authorized, err := q.IsAuthorized(ctx, hostname)
	if err != nil {
		return false, err
	}

	// If not authorized, add it to the database for future triage.
	if !authorized {
		err = q.InsertUnauthorizedDomain(ctx, hostname)
		if err != nil {
			// Log the error but don’t fail the authorization check.
			// Caller can decide how to handle this.
		}
	}

	return authorized, nil
}
