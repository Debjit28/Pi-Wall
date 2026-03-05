import uvicorn

if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host="127.0.0.1",   # localhost only — not exposed externally
        port=5000,
        log_level="info",
        workers=1,           # Pi Zero 2 W has limited RAM; 1 worker is right
    )
```

---

### `firewall/requirements.txt`
```
fastapi==0.111.0
uvicorn[standard]==0.29.0
pydantic==2.7.1