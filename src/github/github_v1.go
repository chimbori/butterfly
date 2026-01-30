package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"chimbori.dev/butterfly/conf"
	"chimbori.dev/butterfly/core"
	"github.com/lmittmann/tint"
)

var Cache *core.DiskCache

// setCORSHeaders configures permissive CORS headers so this endpoint can be called from any origin.
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
}

func Init(mux *http.ServeMux) {
	Cache = core.NewDiskCache(
		filepath.Join(conf.Config.DataDir, "cache", "github"),
		core.WithTTL(time.Hour), // 1 hour
	)
	mux.HandleFunc("GET /github/v1/{user}/{repo}/{type}", handleGithubV1)
	mux.HandleFunc("OPTIONS /github/v1/{user}/{repo}/{type}", handleGithubV1)
}

func handleGithubV1(w http.ResponseWriter, req *http.Request) {
	setCORSHeaders(w)
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	user := req.PathValue("user")
	repo := req.PathValue("repo")
	reqType := req.PathValue("type")

	if user == "" || repo == "" || reqType == "" {
		err := fmt.Errorf("missing parameters: user, repo, type")
		slog.Error("Invalid request", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("repos/%s/%s", user, repo)
	var data []byte
	var err error

	if Cache != nil {
		data, err = Cache.Find(key)
		if err != nil {
			slog.Error("Error checking GitHub cache", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", req.URL)
		}
	}

	if data == nil {
		gitHubApiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo)
		resp, err := http.Get(gitHubApiUrl)
		if err != nil {
			slog.Error("Error fetching from GitHub", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", req.URL)
			http.Error(w, "Error fetching from GitHub", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("GitHub API error: %d", resp.StatusCode)
			slog.Error("GitHub API error", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", req.URL,
				"status", resp.Status)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("Error reading GitHub response body", tint.Err(err),
				"method", req.Method,
				"path", req.URL.Path,
				"url", req.URL)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if Cache != nil {
			if err := Cache.Write(key, data); err != nil {
				slog.Error("Error writing to GitHub cache", tint.Err(err),
					"method", req.Method,
					"path", req.URL.Path,
					"url", req.URL)
			}
		}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		slog.Error("Error unmarshaling GitHub JSON", tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL)
		http.Error(w, "invalid JSON from GitHub", http.StatusBadGateway)
		return
	}

	field := reqType
	switch reqType {
	case "name":
		field = "name"
	case "description":
		field = "description"
	case "stars":
		field = "stargazers_count"
	case "forks":
		field = "forks_count"
	case "issues":
		field = "open_issues_count"
	case "watchers":
		field = "subscribers_count"
	}

	val, ok := result[field]
	if !ok {
		err := fmt.Errorf("Field '%s' not found", reqType)
		slog.Error(err.Error(), tint.Err(err),
			"method", req.Method,
			"path", req.URL.Path,
			"url", req.URL)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "%v", val)
}
