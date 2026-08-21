"""룰 회귀 픽스처 — finguard.python.credential-file-write (#31).

런타임에 발급받은 자격증명을 로컬 파일에 떨구면서 권한을 좁히지 않는 형태.
합성 코드이며 실제 자격증명 값은 들어 있지 않다.
"""

import json
import os
import pickle
import stat
from pathlib import Path


# 증거 형태 — 한국투자증권 open-trading-api kis_auth.save_token 과 동일 구조
def save_token(my_token, valid_date):
    with open(token_tmp, "w", encoding="utf-8") as f:
        # EXPECT: finguard.python.credential-file-write
        f.write(f"token: {my_token}\n")
        # EXPECT: finguard.python.credential-file-write
        f.write(f"valid-date: {valid_date}\n")


# json.dump 싱크 — 대상 경로 변수명에 자격증명 어휘가 있다
def dump_credentials(cfg):
    with open(access_token_path, "w") as f:
        # EXPECT: finguard.python.credential-file-write
        json.dump(cfg, f)


# Path.write_text 싱크
def persist_secret(value):
    # EXPECT: finguard.python.credential-file-write
    secret_path.write_text(value)


# pickle 싱크
def dump_session(obj):
    with open(credential_file, "wb") as f:
        # EXPECT: finguard.python.credential-file-write
        pickle.dump(obj, f)


# configparser 싱크 — 쓰기 인자가 파일 핸들이다
def dump_ini(cp):
    with open(api_key_file, "w") as f:
        # EXPECT: finguard.python.credential-file-write
        cp.write(f)


# chmod 는 하지만 0o644 라 소유자 전용이 아니다 — 억제 대상이 아니다
def save_token_loose_mode(my_token):
    with open(token_tmp, "w") as f:
        # EXPECT: finguard.python.credential-file-write
        f.write(f"access_token: {my_token}")
    os.chmod(token_tmp, 0o644)


# 무관한 대상에 대한 chmod 한 줄이 있어도 억제되면 안 된다
def save_token_unrelated_chmod(my_token):
    os.chmod(LOG_DIR, 0o777)
    with open(token_tmp, "w") as f:
        # EXPECT: finguard.python.credential-file-write
        f.write(f"access_token: {my_token}")
