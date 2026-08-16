#!/usr/bin/env python3
"""이슈 #10: kisa_item 을 구판 SER-NNN 표기에서 2026판 웹·모바일·HTS 실제 ID(1~66)로 재매핑.

구판 SER 번호 20종이 2026판의 서로 다른 실재 항목 번호와 충돌(교차 오염 포함)하므로
순차 치환이 아니라 dict 로 원문 전체 문자열을 한 번에 치환한다.
대응표 검증 원천: docs/평가기준_웹모바일HTS_66항목.md
"""
import pathlib
import sys

# 구판 표기(항목명 구표기 포함) → 2026판 (실제 ID, 원문 항목명)
REMAP = {
    "웹/모바일/HTS-SER-001 (SQL Injection)": (19, "SQL Injection"),
    "웹/모바일/HTS-SER-008 (단말기 내 중요정보 저장 방지)": (26, "단말기 내 중요정보 저장 방지"),
    "웹/모바일/HTS-SER-010 (파일 다운로드)": (28, "파일 다운로드"),
    "웹/모바일/HTS-SER-012 (유추 가능한 세션ID)": (30, "유추 가능한 세션ID"),
    "웹/모바일/HTS-SER-014 (운영체제 명령실행)": (32, "운영체제 명령실행"),
    "웹/모바일/HTS-SER-015 (XML 외부객체 공격 (XXE))": (33, "XML 외부객체 공격 (XXE)"),
    "웹/모바일/HTS-SER-016 (리다이렉트 기능을 이용한 피싱 공격)": (34, "리다이렉트 기능을 이용한 피싱 공격"),
    "웹/모바일/HTS-SER-017 (LDAP Injection)": (35, "LDAP Injection"),
    "웹/모바일/HTS-SER-022 (버퍼오버플로우)": (40, "버퍼오버플로우 (Buffer Overflow Attack)"),
    "웹/모바일/HTS-SER-023 (포맷스트링)": (41, "포맷스트링 (Format String Attack)"),
    "웹/모바일/HTS-SER-025 (앱 소스코드 운영정보 노출 방지)": (43, "앱 소스코드 내 운영정보 노출 방지"),
    "웹/모바일/HTS-SER-030 (서버 인증서 무결성 검증)": (47, "서버 인증서 무결성 검증"),
    "웹/모바일/HTS-SER-034 (취약한 HTTPS 프로토콜 이용 제한)": (51, "취약한 HTTPS 프로토콜 이용 제한"),
    "웹/모바일/HTS-SER-035 (취약한 HTTPS 암호 알고리즘 이용)": (52, "취약한 HTTPS 암호 알고리즘 이용"),
    "웹/모바일/HTS-SER-041 (크로스사이트 스크립팅(XSS))": (58, "크로스 사이트 스크립팅 (XSS)"),
    "웹/모바일/HTS-SER-042 (디버그 로그 내 중요정보 노출 방지)": (59, "디버그 로그 내 중요정보 노출 방지"),
    "웹/모바일/HTS-SER-046 (서버 사이드 요청 위조(SSRF))": (61, "서버 사이드 요청 위조 (SSRF)"),
    "웹/모바일/HTS-SER-051 (통신구간 암호화 적용)": (65, "통신구간 암호화 적용"),
    "웹/모바일/HTS-SER-052 (서버 사이드 템플릿 인젝션(SSTI))": (66, "서버 사이드 템플릿 인젝션(SSTI)"),
}

NEW_FMT = "전자금융기반시설 평가기준(2026) 웹·모바일·HTS {id} ({name})"


def main() -> int:
    path = pathlib.Path(__file__).resolve().parent.parent / "mapping" / "rules.yaml"
    text = path.read_text(encoding="utf-8")

    total = 0
    for old, (item_id, name) in REMAP.items():
        new = NEW_FMT.format(id=item_id, name=name)
        count = text.count(old)
        if count == 0:
            print(f"경고: 매치 0건 — {old}", file=sys.stderr)
            return 1
        text = text.replace(old, new)
        total += count
        print(f"{count}건: {old.split(' ')[0]} → {item_id} ({name})")

    if "SER-" in text:
        print("오류: 치환 후에도 SER- 표기 잔존", file=sys.stderr)
        return 1

    path.write_text(text, encoding="utf-8")
    print(f"완료: 총 {total}건 치환")
    return 0


if __name__ == "__main__":
    sys.exit(main())
