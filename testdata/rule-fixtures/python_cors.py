# 룰 회귀 픽스처 — finguard.python.cors-wildcard-credentials / cors-wildcard-origin (#35)
#
# 기대값은 검출이 기대되는 줄 바로 위의 EXPECT 마커에 적는다.
# semgrep 은 여러 줄에 걸친 호출을 호출 시작 줄로 보고하므로 마커는 호출 첫 줄 위에 둔다.
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from flask import Flask
from flask_cors import CORS

app = FastAPI()

# 와일드카드 + 자격증명 — 요청 Origin 반향으로 실제 악용 가능
# EXPECT: finguard.python.cors-wildcard-credentials
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 키워드 인자 순서를 뒤집어도 같은 취약점이다
# EXPECT: finguard.python.cors-wildcard-credentials
app.add_middleware(
    CORSMiddleware,
    allow_credentials=True,
    allow_methods=["*"],
    allow_origins=["*"],
)

flask_app = Flask(__name__)
# EXPECT: finguard.python.cors-wildcard-credentials
CORS(flask_app, origins="*", supports_credentials=True)

# 자격증명 없는 와일드카드 — 위험도가 달라 WARNING 으로 분리된다
# EXPECT: finguard.python.cors-wildcard-origin
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["GET"],
)

# EXPECT: finguard.python.cors-wildcard-origin
CORS(flask_app, origins=["*"])
