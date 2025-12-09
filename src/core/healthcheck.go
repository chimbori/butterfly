package core

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

func SetupHealthCheck(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, req *http.Request) {
		slog.Info("/healthcheck: ok", "from", ReadUserIP(req))
		w.Write([]byte("ok"))
	})
}

func VerifyHealthCheck(host string, port int) int {
	url := "http://" + net.JoinHostPort(host, strconv.Itoa(port)) + "/healthcheck"
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("failed: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("failed: %s\n", resp.Status)
		return 1
	}

	fmt.Println("ok")
	return 0
}
