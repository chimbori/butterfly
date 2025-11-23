package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "time/tzdata"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmittmann/tint"
	"go.chimbori.app/butterfly/conf"
	"go.chimbori.app/butterfly/core"
	"go.chimbori.app/butterfly/dashboard"
	"go.chimbori.app/butterfly/db"
	"go.chimbori.app/butterfly/embedfs"
	"go.chimbori.app/butterfly/slogdb"
)

func main() {
	tintHandler := tint.NewHandler(os.Stderr, &tint.Options{TimeFormat: "2006-01-02 15:04:05.000"})
	slog.SetDefault(slog.New(tintHandler))
	slog.Info(conf.AppName, "build-timestamp", conf.BuildTimestamp)

	configYmlFlag := flag.String("config", "butterfly.yml", "path to butterfly.yml")
	flag.Parse()

	// Read config before any routine maintenance is performed.
	var err error
	if conf.Config, err = conf.ReadConfig(*configYmlFlag); err != nil {
		slog.Error("Failed to parse config", tint.Err(err))
		flag.PrintDefaults()
		os.Exit(1)
	}

	// If debug mode was turned on in the config file, print logs at DEBUG or above.
	if conf.Config.Debug {
		slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: "2006-01-02 15:04:05.000",
		})))
	}

	// Run migrations using [database/sql] before connecting to the DB using [pgxpool.Pool].
	if err := core.RunMigrations(conf.Config.Database.Url, db.EmbedMigrations); err != nil {
		slog.Error("Error running critical migrations", tint.Err(err))
		os.Exit(1)
	}
	db.Pool, err = pgxpool.New(context.Background(), conf.Config.Database.Url)
	if err != nil {
		slog.Error("Unable to connect to database", tint.Err(err))
		os.Exit(1)
	}
	slog.Info("Connected to database successfully")

	// Now that the database is connected, wrap the console handler with the DB handler
	// so that all error-level logs are also written to the database.
	slog.SetDefault(slog.New(slogdb.NewDBHandler(tintHandler, db.Pool)))
	slog.Info("Database error logging enabled")

	initCache()

	// Set up a graceful cleanup for when the process is terminated.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signalCh
		fmt.Println()
		slog.Info("Shutdown successfully!")
		os.Exit(0)
	}()

	// Set up the Web server and start serving.
	mux := http.NewServeMux()
	core.SetupHealthz(mux)
	core.ServeWebManifest(mux, conf.AppName, conf.Config.Web.PublicUrl, "#FFD92E")
	embedfs.ServeStaticFS(mux)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, req *http.Request) {
		IndexTempl().Render(req.Context(), w)
	})
	mux.HandleFunc("GET /link-preview/v1", handleLinkPreview)
	dashboard.SetupHandlers(mux)

	addr := net.JoinHostPort(conf.Config.Web.Host, strconv.Itoa(conf.Config.Web.Port))
	slog.Info("Listening", "url", "http://"+addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
