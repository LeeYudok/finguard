"""finguard.python 룰 8종 정탐 픽스처 (#20).

각 블록은 대응하는 룰 하나씩만 발화해야 한다. semgrep 이 정적으로만 읽는
예제 코드라서 import 대상 모듈이 실제 설치돼 있지 않아도 무방하다.
"""

import hashlib
import subprocess
import requests
import yaml


# 1) finguard.python.hardcoded-secret — API 키가 코드에 직접 박혀 있다.
PAYMENT_API_KEY = "sk-test-9f8e7d6c5b4a3210"


# 2) finguard.python.weak-hash — MD5 로 비밀번호 해시를 생성한다.
def hash_password(password: str) -> str:
    return hashlib.md5(password.encode()).hexdigest()


# 3) finguard.python.http-url — 평문 HTTP 로 내부 결제 API 를 호출한다.
PAYMENT_ENDPOINT = "http://payment-gateway.corp.local/api/transfer"


# 4) finguard.python.sql-format — f-string 으로 조립한 SQL 을 그대로 실행한다.
def find_user(cursor, user_id):
    cursor.execute(f"SELECT * FROM users WHERE id = {user_id}")


# 5) finguard.python.subprocess-shell — 조립된 명령을 shell=True 로 실행한다.
def list_dir(user_supplied_path):
    cmd = "ls " + user_supplied_path
    subprocess.run(cmd, shell=True)


# 6) finguard.python.eval-exec — 동적으로 조립된 문자열을 eval 에 전달한다.
def compute(user_expr):
    return eval(user_expr)


# 7) finguard.python.tls-verify-disabled — TLS 인증서 검증을 끈다.
def fetch(url):
    return requests.get(url, verify=False)


# 8) finguard.python.yaml-unsafe-load — 안전하지 않은 로더로 역직렬화한다.
def load_config(stream):
    return yaml.load(stream, Loader=yaml.Loader)
