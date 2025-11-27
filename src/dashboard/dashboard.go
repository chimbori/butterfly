package dashboard

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/justinas/alice"
	"github.com/lmittmann/tint"
	"go.chimbori.app/butterfly/conf"
	"go.chimbori.app/butterfly/core"
	"golang.org/x/crypto/bcrypt"
)

func SetupHandlers(mux *http.ServeMux) {
	chain := alice.New(authHandler)

	mux.Handle("GET /dashboard", chain.Then(homeHandler))

	mux.Handle("GET /dashboard/domains", chain.Then(getDomainsHandler))
	mux.Handle("PUT /dashboard/domains", chain.Then(putDomainHandler))
	mux.Handle("DELETE /dashboard/domains", chain.Then(deleteDomainHandler))

	mux.Handle("GET /dashboard/logs", chain.Then(logsHandler))
	mux.Handle("GET /dashboard/logs/data", chain.Then(logsDataHandler))
}

// GET /dashboard
var homeHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	HomeTempl(conf.AppName).Render(req.Context(), w)
})

// Checks whether the user is authorized, and either returns an error, or executes the passed [http.Handler].
func authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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

		next.ServeHTTP(w, req)
	})
}
