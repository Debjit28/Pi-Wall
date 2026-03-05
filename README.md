# PiWall 🔥

# minor-project-sem-6

> A lightweight proxy-based firewall running on a Raspberry Pi Zero 2 W.  
> Built with Go (forward proxy) + Python/FastAPI (rule engine).

---

## What It Does

PiWall sits between your devices and the internet, inspecting every HTTP/HTTPS request before it leaves your network. Blocked domains get a `403 Forbidden`. Everything else passes through transparently — with no SSL interception.

```
Your Device (laptop/phone)
        │
        ▼
 Pi Zero 2 W — port 8080
  ┌─────────────────────┐
  │   Go Forward Proxy  │──► POST /check ──► Python Firewall
  │                     │◄── allow/block ◄──  (rules.json)
  └─────────────────────┘
        │ (if allowed)
        ▼
    Internet
```

---

## Project Structure

```
piwall/
├── proxy/               # Go forward proxy (HTTP + HTTPS CONNECT)
│   ├── main.go          # Entry point, CLI flags
│   ├── handler.go       # HTTP and CONNECT tunnel handling
│   ├── firewall.go      # Firewall API client + request/response types
│   ├── logger.go        # Structured JSON logger
│   └── go.mod
│
├── firewall/            # Python rule engine (FastAPI)
│   ├── run.py           # Entry point
│   ├── requirements.txt
│   ├── rules.json       # Blacklist configuration
│   └── app/
│       ├── main.py      # FastAPI routes (/check, /health, /reload)
│       ├── models.py    # Pydantic request/response models
│       └── rules.py     # RuleEngine — loads blacklist, evaluates requests
```

---

## Tech Stack

| Component      | Technology                          |
|----------------|-------------------------------------|
| Forward Proxy  | Go — stdlib only, zero dependencies |
| Rule Engine    | Python 3.13, FastAPI, uvicorn       |
| Communication  | HTTP REST (localhost:5000)          |
| Logging        | Structured JSON (newline-delimited) |
| Target Hardware| Raspberry Pi Zero 2 W (512MB RAM)  |

---

## Prerequisites

**On the Raspberry Pi:**
```bash
sudo apt update
sudo apt install golang python3 python3-venv -y
```

---

## Setup & Installation

### 1. Clone the repository
```bash
git clone https://github.com/yourname/piwall.git
cd piwall
```

### 2. Set up Python firewall
```bash
cd firewall
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 3. Build Go proxy
```bash
cd ../proxy
go build -o proxy .
```

---

## Running PiWall

You need two terminals on the Pi.

### Terminal 1 — Start the firewall
```bash
cd firewall
source venv/bin/activate
python run.py
```

Expected output:
```
INFO:     Started server process [XXXX]
[firewall] ready - 7 blacklisted domains
INFO:     Uvicorn running on http://127.0.0.1:5000
```

### Terminal 2 — Start the proxy
```bash
cd proxy
./proxy -addr 0.0.0.0:8080 -log proxy.log
```

Expected output:
```
Starting proxy on 0.0.0.0:8080
```

---

## Configuring Your Devices

Find the Pi's IP address:
```bash
hostname -I
```

Then set the proxy on your device (all must be on the same WiFi):

| Platform | Where to set it |
|----------|----------------|
| Windows  | Settings → Network → Proxy → Manual → `<pi-ip>:8080` |
| macOS    | System Settings → Network → Proxies → Web Proxy → `<pi-ip>:8080` |
| Android  | WiFi → Long press → Modify → Advanced → Proxy: Manual → `<pi-ip>:8080` |
| Linux    | `export http_proxy=http://<pi-ip>:8080` |

---

## Firewall Rules (rules.json)

Edit `firewall/rules.json` to control what gets blocked:

```json
{
  "blacklist": [
    "doubleclick.net",
    "googlesyndication.com",
    "adservice.google.com",
    "facebook.com",
    "malware-site.com"
  ]
}
```

Subdomain matching is automatic — blocking `facebook.com` also blocks `www.facebook.com`, `static.facebook.com`, etc.

### Hot-reload rules without restarting
```bash
curl -X POST http://localhost:5000/reload
```

---

## API Reference (Firewall)

| Method | Endpoint  | Description                        |
|--------|-----------|------------------------------------|
| POST   | `/check`  | Evaluate a request (used by proxy) |
| GET    | `/health` | Health check + blacklist count     |
| POST   | `/reload` | Reload rules.json from disk        |

### Request payload (`/check`)
```json
{
  "client_ip": "192.168.1.5",
  "method": "CONNECT",
  "host": "example.com:443",
  "path": "",
  "timestamp": "2026-03-05T10:00:00Z"
}
```

### Response
```json
{ "action": "allow" }
{ "action": "block", "reason": "blacklisted domain" }
```

---

## Log Format

Logs are written to `proxy/proxy.log` as newline-delimited JSON:

```json
{"client_ip":"192.168.1.5","method":"CONNECT","host":"example.com:443","path":"","timestamp":"2026-03-05T10:00:00Z","decision":"allow","status":200}
{"client_ip":"192.168.1.5","method":"GET","host":"doubleclick.net","path":"/ads","timestamp":"2026-03-05T10:00:01Z","decision":"block","status":403,"reason":"blacklisted domain"}
```

Watch live:
```bash
tail -f proxy/proxy.log
```

Pretty-print:
```bash
tail -f proxy/proxy.log | python3 -m json.tool
```

---

## Testing

```bash
# Allowed request
curl -x http://<pi-ip>:8080 http://example.com

# Blocked request — expect 403
curl -x http://<pi-ip>:8080 http://doubleclick.net

# HTTPS tunnel — expect 200 Connection Established
curl -x http://<pi-ip>:8080 https://example.com

# Firewall health check
curl http://localhost:5000/health
```

---

## CLI Flags (proxy)

```
-addr     string   Listen address         (default ":8080")
-firewall string   Firewall API URL       (default "http://localhost:5000/check")
-log      string   Log file path          (default "proxy.log")
```

---

## Memory Footprint

Designed to be lightweight for the Pi Zero 2 W (512MB RAM):

| Process         | Approx. RSS |
|-----------------|-------------|
| Go proxy        | ~8 MB       |
| Python firewall | ~55 MB      |
| **Total**       | **~63 MB**  |

---

## Limitations

- HTTPS traffic is tunneled via CONNECT — hostname is visible but content is not inspected (no SSL interception by design).
- Firewall runs as a single uvicorn worker — sufficient for home/lab network use.
- Rules are domain-based only; no IP, port, or regex matching.

---

## Author

**Debjit** — Minor Project, Semester 6

---

## License

MIT
