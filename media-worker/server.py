import hmac
import io
import json
import os
import threading
import zipfile
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from models import load_models
from moderation import review
from transcription import transcribe

KEY = os.environ["MEDIA_WORKER_KEY"]
if len(KEY) < 32:
    raise RuntimeError("worker key too short")
MODELS = load_models(os.environ.get("MODEL_DIRECTORY", "/models"))
SLOTS = threading.BoundedSemaphore(2)

def source_archive():
    buffer = io.BytesIO()
    with zipfile.ZipFile(buffer, "w", zipfile.ZIP_DEFLATED) as archive:
        for path in Path(__file__).parent.iterdir():
            if path.is_file() and (path.suffix in [".py", ".txt", ".md"] or path.name in ["LICENSE", "Dockerfile"]):
                archive.write(path, path.name)
    return buffer.getvalue()

class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_args):
        return None

    def reply(self, status, value):
        data = json.dumps(value).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)
        return status

    def do_GET(self):
        if self.path == "/source":
            data = source_archive()
            self.send_response(200)
            self.send_header("Content-Type", "application/zip")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)
            return 200
        if self.path == "/health":
            return self.reply(200, {"ready": True})
        return self.reply(404, {"error": "not_found"})

    def do_POST(self):
        if not hmac.compare_digest(self.headers.get("Authorization", ""), "Bearer " + KEY):
            return self.reply(401, {"error": "unauthorized"})
        if self.path not in ["/moderate", "/transcribe"]:
            return self.reply(404, {"error": "not_found"})
        try:
            length = int(self.headers.get("Content-Length", "0"))
        except ValueError:
            return self.reply(400, {"error": "invalid_length"})
        if not 0 < length <= 5*1024*1024:
            return self.reply(413, {"error": "payload_size"})
        if not SLOTS.acquire(blocking=False):
            return self.reply(503, {"error": "busy"})
        try:
            self.connection.settimeout(10)
            raw = self.rfile.read(length)
            if len(raw) != length:
                return self.reply(400, {"error": "incomplete_payload"})
            result = review(MODELS[0], raw) if self.path == "/moderate" else transcribe(MODELS[1], raw)
            return self.reply(200, result)
        except ValueError:
            return self.reply(400, {"error": "invalid_media"})
        except Exception:
            return self.reply(503, {"error": "review_unavailable"})
        finally:
            SLOTS.release()

if __name__ == "__main__":
    ThreadingHTTPServer(("0.0.0.0", 8090), Handler).serve_forever()
