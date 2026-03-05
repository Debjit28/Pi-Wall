package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// ProxyHandler routes HTTP and HTTPS CONNECT requests through the firewall.
type ProxyHandler struct {
	firewall  *FirewallClient
	logger    *Logger
	transport *http.Transport
}

func NewProxyHandler(fw *FirewallClient, lg *Logger) *ProxyHandler {
	return &ProxyHandler{
		firewall: fw,
		logger:   lg,
		transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}

	meta := RequestMeta{
		ClientIP:  clientIP,
		Method:    r.Method,
		Host:      host,
		Path:      r.URL.Path,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	decision, err := p.firewall.Check(meta)
	if err != nil {
		p.logger.Log(meta, "error", 502, fmt.Sprintf("firewall unreachable: %v", err))
		http.Error(w, "Firewall unavailable", http.StatusBadGateway)
		return
	}

	if decision.Action == "block" {
		p.logger.Log(meta, "block", 403, decision.Reason)
		http.Error(w, fmt.Sprintf("Blocked: %s", decision.Reason), http.StatusForbidden)
		return
	}

	if r.Method == http.MethodConnect {
		p.handleCONNECT(w, r, meta)
	} else {
		p.handleHTTP(w, r, meta)
	}
}

// handleHTTP forwards plain HTTP requests.
func (p *ProxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request, meta RequestMeta) {
	outReq := r.Clone(r.Context())
	outReq.RequestURI = ""
	outReq.Header.Del("Proxy-Connection")

	resp, err := p.transport.RoundTrip(outReq)
	if err != nil {
		p.logger.Log(meta, "allow", 502, fmt.Sprintf("upstream error: %v", err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	p.logger.Log(meta, "allow", resp.StatusCode, "")
}

// handleCONNECT tunnels HTTPS without SSL interception.
func (p *ProxyHandler) handleCONNECT(w http.ResponseWriter, r *http.Request, meta RequestMeta) {
	destConn, err := net.DialTimeout("tcp", r.Host, 15*time.Second)
	if err != nil {
		p.logger.Log(meta, "allow", 502, fmt.Sprintf("dial error: %v", err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer destConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		p.logger.Log(meta, "allow", 500, "hijacking not supported")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		p.logger.Log(meta, "allow", 500, fmt.Sprintf("hijack error: %v", err))
		return
	}
	defer clientConn.Close()

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	p.logger.Log(meta, "allow", 200, "CONNECT tunnel")

	// Bidirectional pipe — no SSL interception, raw bytes only
	done := make(chan struct{}, 2)
	pipe := func(dst, src net.Conn) {
		io.Copy(dst, src)
		done <- struct{}{}
	}
	go pipe(destConn, clientConn)
	go pipe(clientConn, destConn)
	<-done
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, v := range values {
			dst.Add(key, v)
		}
	}
}
