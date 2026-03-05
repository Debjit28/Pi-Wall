package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8080", "Proxy listen address")
	firewallURL := flag.String("firewall", "http://localhost:5000/check", "Firewall API endpoint")
	logFile := flag.String("log", "proxy.log", "Log file path")
	flag.Parse()

	logger, err := NewLogger(*logFile)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logger.Close()

	firewall := NewFirewallClient(*firewallURL)
	handler := NewProxyHandler(firewall, logger)

	log.Printf("Starting proxy on %s", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("proxy error: %v", err)
	}
}
