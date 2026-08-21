# 대조군 픽스처 — 어떤 룰에도 걸리면 안 된다 (#35).
# 오리진 화이트리스트를 쓰는 정상 구현. safe_ 접두라 EXPECT 마커를 가질 수 없다.
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from flask import Flask
from flask_cors import CORS

app = FastAPI()

ALLOWED_ORIGINS = [
    "https://ops.example.com",
    "https://desk.example.com",
]

# 자격증명을 허용하지만 오리진은 명시적 화이트리스트다
app.add_middleware(
    CORSMiddleware,
    allow_origins=ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["GET", "POST"],
    allow_headers=["Authorization"],
)

# 리터럴 화이트리스트도 마찬가지로 정상이다
app.add_middleware(
    CORSMiddleware,
    allow_origins=["https://desk.example.com"],
    allow_credentials=True,
)

flask_app = Flask(__name__)
CORS(flask_app, origins=["https://ops.example.com"], supports_credentials=True)

# 메서드만 와일드카드인 경우는 오리진 통제와 무관하다
app.add_middleware(
    CORSMiddleware,
    allow_origins=["https://ops.example.com"],
    allow_methods=["*"],
    allow_headers=["*"],
)
