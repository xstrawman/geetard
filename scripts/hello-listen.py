#!/usr/bin/env python3
"""LAN doorbell for geetard Windows hello.ps1."""
from http.server import BaseHTTPRequestHandler, HTTPServer
from datetime import datetime
import sys

LOG = "/tmp/geetard-hello.log"


class H(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        pass

    def do_POST(self):
        n = int(self.headers.get("Content-Length") or 0)
        body = self.rfile.read(n).decode("utf-8", "replace")
        peer = self.client_address[0]
        line = f"\n===== {datetime.now().isoformat()} from {peer} {self.path}\n{body}\n"
        sys.stdout.write(line)
        sys.stdout.flush()
        with open(LOG, "a", encoding="utf-8") as f:
            f.write(line)
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"ok\n")

    def do_GET(self):
        self.do_POST()


if __name__ == "__main__":
    HTTPServer(("0.0.0.0", 9876), H).serve_forever()
