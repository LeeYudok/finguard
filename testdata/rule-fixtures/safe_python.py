"""finguard.python 룰 8종 오탐 방지 픽스처 (#20).

아래 블록은 각 룰의 안전한 변형을 모아둔 것으로, python.yaml 룰 8종은 물론
rules/ 전체 룰셋 어디에도 걸리면 안 된다.
"""

import ast
import hashlib
import os
import ssl
import subprocess

import httpx
import requests
import yaml


# 1) finguard.python.hardcoded-secret — 환경변수 조회, 빈 문자열, 플레이스홀더는 시크릿이 아니다.
PAYMENT_API_KEY = os.getenv("PAYMENT_API_KEY")
DB_PASSWORD = ""
SECRET_PLACEHOLDER = "changeme"
# 단독 `key` 는 공통 어휘에서 제외한다 (#36) — 맵 키·헤더명 오탐이 폭증한다.
key = "Authorization-Bearer-Header"
sort_key = "settlement_date_desc"
cache_key = "user:1234:profile:v2"


# 2) finguard.python.weak-hash — 비보안 용도(체크섬)로 명시한 MD5 는 대상이 아니다.
def checksum(data: bytes) -> str:
    return hashlib.md5(data, usedforsecurity=False).hexdigest()


# 3) finguard.python.http-url — localhost 와 XML 네임스페이스 URI 는 대상이 아니다 (이슈 #3 계열).
LOCAL_HEALTHCHECK_URL = "http://localhost:8000/health"
SOAP_NAMESPACE = "http://schemas.xmlsoap.org/soap/envelope/"


# 4) finguard.python.sql-format — 바인딩 파라미터를 쓰면 대상이 아니다.
def find_user_safe(cursor, user_id):
    cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))


# 5) finguard.python.subprocess-shell — 리스트 인자 + shell=False 는 대상이 아니다.
def list_dir_safe(path):
    subprocess.run(["ls", path], shell=False)


# 6) finguard.python.eval-exec — 리터럴 상수만 전달하거나 ast.literal_eval 을 쓰면 대상이 아니다.
def compute_safe(user_expr):
    return ast.literal_eval(user_expr)


CONST_RESULT = eval("1 + 1")


# 7) finguard.python.tls-verify-disabled — verify=True 명시 또는 기본값은 대상이 아니다.
def fetch_verified(url):
    return requests.get(url, verify=True)


def fetch_default(url):
    return requests.get(url)


# 7b) finguard.python.tls-verify-disabled — 기본 컨텍스트로 되돌리는 대입은 안전하다 (#26).
def restore_tls_default():
    ssl._create_default_https_context = ssl.create_default_context


# 7c) finguard.python.tls-verify-disabled — 검증을 켜는 설정은 대상이 아니다.
def build_strict_context():
    ctx = ssl.create_default_context()
    ctx.verify_mode = ssl.CERT_REQUIRED
    return ctx


def fetch_httpx_verified(url):
    return httpx.Client(verify=True).get(url)


# 9) finguard.python.cleartext-websocket — wss 와 로컬 개발 서버는 대상이 아니다 (#30).
REALTIME_FEED_URL = "wss://ops.example-broker.co.kr:21000"
LOCAL_FEED_URL = "ws://localhost:8765"
LOOPBACK_FEED_URL = "ws://127.0.0.1:8765"


# 8) finguard.python.yaml-unsafe-load — safe_load 또는 SafeLoader 는 대상이 아니다.
def load_config_safe(stream):
    return yaml.safe_load(stream)


def load_config_explicit_safe(stream):
    return yaml.load(stream, Loader=yaml.SafeLoader)
