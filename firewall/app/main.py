from contextlib import asynccontextmanager
from fastapi import FastAPI
from .models import CheckRequest, CheckResponse
from .rules import RuleEngine

rule_engine: RuleEngine | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global rule_engine
    rule_engine = RuleEngine("rules.json")
    rule_engine.load()
    print(f"[firewall] ready — {len(rule_engine.blacklist)} blacklisted domains")
    yield


app = FastAPI(title="PiWall Firewall", lifespan=lifespan)


@app.post("/check", response_model=CheckResponse)
def check(req: CheckRequest) -> CheckResponse:
    return rule_engine.evaluate(req)


@app.get("/health")
def health():
    return {"status": "ok", "blacklist_count": len(rule_engine.blacklist)}


@app.post("/reload")
def reload():
    """Hot-reload rules.json without restarting the service."""
    rule_engine.load()
    return {"status": "reloaded", "blackli
