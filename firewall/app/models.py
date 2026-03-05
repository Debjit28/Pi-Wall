from pydantic import BaseModel
from typing import Optional

class CheckRequest(BaseModel):
    client_ip: str
    method: str
    host: str
    path: str
    timestamp: str

class CheckResponse(BaseModel):
    action: str                   # "allow" | "block"
    reason: Optional[str] = None