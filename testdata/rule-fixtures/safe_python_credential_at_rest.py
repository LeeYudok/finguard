"""대조군 — 어떤 룰에도 걸리면 안 된다 (#31).

토큰을 저장하되 소유자 전용 권한을 실제로 부여하는 정상 구현들. chmod 를 with 블록
밖에서 호출하는 실무 관용구(파일 핸들이 필요 없는 오퍼레이션)도 함께 둔다.
"""

import os
import stat
from pathlib import Path


# with 블록을 빠져나온 뒤 소유자 전용 권한 부여 — 실무에서 가장 흔한 안전 형태
def save_token(my_token):
    with open(token_tmp, "w", encoding="utf-8") as f:
        f.write(f"token: {my_token}\n")
    os.chmod(token_tmp, 0o600)


# stat 상수 조합
def save_token_stat(my_token):
    with open(token_tmp, "w", encoding="utf-8") as f:
        f.write(f"token: {my_token}\n")
    os.chmod(token_tmp, stat.S_IRUSR | stat.S_IWUSR)


# Path.chmod
def save_token_pathlib(my_token):
    p = Path(token_tmp)
    p.write_text(f"token: {my_token}")
    p.chmod(0o600)


# umask 로 그룹·타인 권한을 미리 차단
def save_token_umask(my_token):
    os.umask(0o077)
    with open(token_tmp, "w", encoding="utf-8") as f:
        f.write(f"token: {my_token}\n")


# 자격증명이 아닌 일반 파일 쓰기는 대상이 아니다
def write_report(rows, report_path):
    with open(report_path, "w", encoding="utf-8") as f:
        for row in rows:
            f.write(row + "\n")
