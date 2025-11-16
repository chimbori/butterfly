package core

import (
	"net"
	"net/http"
)

// GetOutboundIP returns the preferred outbound IP address of this machine.
// It establishes a UDP connection to 8.8.8.8 to determine the local IP.
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP
}

// ReadUserIP extracts the client IP address from an HTTP request.
// It checks X-Real-Ip, X-Forwarded-For headers, and falls back to RemoteAddr.
func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
