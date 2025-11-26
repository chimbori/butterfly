package dashboard

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/lmittmann/tint"
	"go.chimbori.app/butterfly/core"
	"go.chimbori.app/butterfly/db"
)

// GET /dashboard/domains - List all domains
var getDomainsHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	queries := db.New(db.Pool)
	domains, err := queries.ListDomains(ctx)
	if err != nil {
		slog.Error("failed to list domains",
			"error", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	DomainsTempl(domains).Render(ctx, w)
})

// PUT /dashboard/domains - Add a new domain, or update existing one if present.
var putDomainHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	queries := db.New(db.Pool)

	err := req.ParseForm()
	if err != nil {
		slog.Error("failed to parse form",
			"error", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	domain := strings.TrimSpace(req.FormValue("domain"))
	if domain == "" {
		err := fmt.Errorf("empty domain")
		slog.Error(err.Error(),
			"error", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	includeSubdomains := req.FormValue("include_subdomains") == "on"
	authorizeAction := strings.ToLower(req.FormValue("authorized"))

	_, err = queries.UpsertDomain(ctx, db.UpsertDomainParams{
		Domain:            domain,
		IncludeSubdomains: &includeSubdomains,
		Authorized:        isAuthorized(authorizeAction),
	})
	if err != nil {
		slog.Error("failed to update domain",
			"error", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated list
	domains, err := queries.ListDomains(ctx)
	if err != nil {
		slog.Error("failed to list domains",
			"error", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	DomainsListTempl(domains).Render(ctx, w)
})

// DELETE /dashboard/domains?domain=example.com - Delete a domain
var deleteDomainHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	queries := db.New(db.Pool)

	domain := req.URL.Query().Get("domain")
	err := queries.DeleteDomain(ctx, domain)
	if err != nil {
		slog.Error("failed to delete domain",
			"error", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated list
	domains, err := queries.ListDomains(ctx)
	if err != nil {
		slog.Error("failed to list domains",
			"error", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL.String(),
			"status", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	DomainsListTempl(domains).Render(ctx, w)
})

func isAuthorized(authorizeAction string) *bool {
	switch strings.TrimSpace(authorizeAction) {
	case "":
		return nil
	case "allow":
		return core.Ptr(true)
	case "block":
		return core.Ptr(false)
	}
	return nil
}
