package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RequestMeta is the payload sent to the Python firewall.
type RequestMeta struct {
	ClientIP  string `json:"client_ip"`
	Method    string `json:"method"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	Timestamp string `json:"timestamp"`
}

// FirewallDecision is the response from the Python firewall.
type FirewallDecision struct {
	Action string `json:"action"` // "allow" | "block"
	Reason string `json:"reason"`
}

// FirewallClient sends request metadata to the Python firewall REST API.
type FirewallClient struct {
	endpoint string
	client   *http.Client
}

func NewFirewallClient(endpoint string) *FirewallClient {
	return &FirewallClient{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (fc *FirewallClient) Check(meta RequestMeta) (*FirewallDecision, error) {
	body, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	resp, err := fc.client.Post(fc.endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	var decision FirewallDecision
	if err := json.NewDecoder(resp.Body).Decode(&decision); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &decision, nil
}
