import json
import logging
from pathlib import Path
from .models import CheckRequest, CheckResponse

logger = logging.getLogger(__name__)


class RuleEngine:
    """Loads blacklist from rules.json and evaluates requests."""

    def __init__(self, rules_path: str = "rules.json"):
        self.rules_path = Path(rules_path)
        self.blacklist: set[str] = set()

    def load(self) -> None:
        """(Re)load rules from disk. Safe to call at runtime for hot reload."""
        try:
            with self.rules_path.open() as f:
                data = json.load(f)
            # Normalise: strip leading dots, lowercase
            self.blacklist = {
                d.strip().lstrip(".").lower()
                for d in data.get("blacklist", [])
            }
            logger.info("Loaded %d blacklist entries", len(self.blacklist))
        except FileNotFoundError:
            logger.warning("rules.json not found — empty blacklist")
            self.blacklist = set()
        except json.JSONDecodeError as exc:
            logger.error("Invalid rules.json: %s", exc)

    def _is_blocked(self, host: str) -> bool:
        """Return True if host OR any parent domain is blacklisted.
        e.g. "ads.example.com" is blocked if "example.com" is in the list.
        """
        host = host.split(":")[0].lower()  # strip port
        parts = host.split(".")
        for i in range(len(parts)):
            if ".".join(parts[i:]) in self.blacklist:
                return True
        return False

    def evaluate(self, req: CheckRequest) -> CheckResponse:
        if self._is_blocked(req.host):
            logger.info("BLOCK  %s %s %s", req.client_ip, req.method, req.host)
            return CheckResponse(action="block", reason="blacklisted domain")
        logger.info("ALLOW  %s %s %s", req.client_ip, req.method, req.host)
        return CheckResponse(action="allow")