"""
Integrated Documentation Sync

SPEC-006: 자동 문서 동기화, API 문서 생성, 일관성 검증, 다국어 지원

@DESIGN:PLUGIN-ARCH-001 - 플러그인 기반 문서 생성 엔진
@REQ:DOC-CORE-001 - 핵심 문서 동기화 요구사항
"""

__version__ = "0.1.0"
__author__ = "MoAI-ADK"

# Constitution 5원칙 준수
# Article I: Simplicity - 3개 핵심 모듈 (sync, generate, validate)
# Article II: Architecture - 플러그인 기반 확장 가능한 설계
# Article III: Testing - TDD 기반 개발 준비
# Article IV: Observability - 구조화 로깅 준비
# Article V: Versioning - 시맨틱 버전 관리