package dashboard

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"butterfly.chimbori.dev/conf"
	"butterfly.chimbori.dev/core"
	"github.com/justinas/alice"
	"github.com/lmittmann/tint"
	"golang.org/x/crypto/bcrypt"
)

var sessionStore = core.NewInMemorySessionStore(24 * time.Hour)

const sessionCookieName = "butterfly_session"

func Init(mux *http.ServeMux) {
	if *conf.Config.LinkPreviews.Cache.Enabled {
		ThumbnailCache = core.NewDiskCache(
			filepath.Join(conf.Config.DataDir, "cache", "link-previews-thumbnails"),
			core.WithTTL(conf.Config.LinkPreviews.Cache.TTL),
			core.WithMaxSize(conf.Config.LinkPreviews.Cache.MaxSizeBytes),
		)
	}

	chain := alice.New(authHandler)

	mux.Handle("GET /dashboard", chain.ThenFunc(homeHandler))

	mux.Handle("GET /dashboard/link-previews", chain.ThenFunc(linkPreviewsPageHandler))
	mux.Handle("GET /dashboard/link-previews/list", chain.ThenFunc(linkPreviewsListHandler))
	mux.Handle("GET /dashboard/link-previews/image", chain.ThenFunc(serveLinkPreviewHandler))
	mux.Handle("GET /dashboard/link-previews/stats", chain.ThenFunc(linkPreviewsStatsHandler))
	mux.Handle("GET /dashboard/link-previews/user-agents", chain.ThenFunc(linkPreviewsUserAgentsHandler))
	mux.Handle("DELETE /dashboard/link-previews/url", chain.ThenFunc(deleteLinkPreviewHandler))

	mux.Handle("GET /dashboard/qr-codes", chain.ThenFunc(listQrCodesHandler))
	mux.Handle("DELETE /dashboard/qr-codes/url", chain.ThenFunc(deleteQrCodeHandler))

	mux.Handle("GET /dashboard/domains", chain.ThenFunc(domainsPageHandler))
	mux.Handle("PUT /dashboard/domains/domain", chain.ThenFunc(putDomainHandler))
	mux.Handle("DELETE /dashboard/domains/domain", chain.ThenFunc(deleteDomainHandler))

	mux.Handle("GET /dashboard/logs", chain.ThenFunc(logsHandler))
	mux.Handle("GET /dashboard/logs/data", chain.ThenFunc(logsDataHandler))
}

// GET /dashboard
func homeHandler(w http.ResponseWriter, req *http.Request) {
	HomeTempl(conf.AppName).Render(req.Context(), w)
}

// Checks whether the user is authorized, and either returns an error, or executes the passed [http.Handler].
func authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if cookie, err := req.Cookie(sessionCookieName); err == nil {
			if sessionStore.IsValid(cookie.Value) {
				next.ServeHTTP(w, req)
				return
			}
		}

		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok || reqUsername != conf.Config.Dashboard.Username {
			slog.Warn("no credentials provided", tint.Err(fmt.Errorf("no credentials (from: %s)", core.ReadUserIP(req))),
				"method", req.Method,
				"path", req.URL.Path,
				"status", http.StatusUnauthorized)
			w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, conf.AppName))
			w.WriteHeader(http.StatusUnauthorized)
			ContentTempl("Unauthorized", ErrorTempl("Please provide valid credentials to access this section.")).Render(req.Context(), w)
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(conf.Config.Dashboard.Password), []byte(reqPassword))
		if err != nil {
			slog.Error("invalid credentials provided", tint.Err(fmt.Errorf("invalid credentials (from: %s)", core.ReadUserIP(req))),
				"method", req.Method,
				"path", req.URL.Path,
				"status", http.StatusUnauthorized)
			w.WriteHeader(http.StatusUnauthorized)
			ContentTempl("Unauthorized", ErrorTempl("Please provide valid credentials to access this section.")).Render(req.Context(), w)
			return
		}

		sessionID, err := sessionStore.Create()
		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   int((24 * time.Hour).Seconds()),
			})
		} else {
			slog.Error("failed to create session", tint.Err(err))
		}

		next.ServeHTTP(w, req)
	})
}

func CleanupExpiredSessions() {
	sessionStore.CleanupExpired()
}
