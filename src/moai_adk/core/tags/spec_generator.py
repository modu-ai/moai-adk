#!/usr/bin/env python3
# @CODE:SPEC-GENERATOR-001 | @SPEC:TAG-SPEC-GENERATION-001 | @DOC:SPEC-AUTO-GEN-001
"""SPEC 템플릿 자동 생성기

코드 파일을 분석하여 EARS 포맷의 SPEC 템플릿을 자동으로 생성합니다.
도메인 추론, 신뢰도 계산, 편집 가이드 제공 포함.

주요 기능:
- AST 기반 Python/JavaScript/Go 코드 분석
- 파일 경로에서 도메인 자동 추론
- EARS 포맷 SPEC 템플릿 생성
- 신뢰도 점수 계산 (0-1)
- 편집 가이드 생성 (TODO 체크리스트)
"""

import ast
import re
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Set


@dataclass
class CodeAnalysis:
    """코드 분석 결과

    Attributes:
        functions: 함수 목록 및 정보
        classes: 클래스 목록 및 정보
        imports: import 정보
        docstring: 모듈 docstring
        domain_keywords: 도메인 관련 키워드
        has_clear_structure: 명확한 구조 여부
    """
    functions: Dict[str, Dict[str, Any]] = field(default_factory=dict)
    classes: Dict[str, Dict[str, Any]] = field(default_factory=dict)
    imports: Dict[str, List[str]] = field(default_factory=dict)
    docstring: Optional[str] = None
    domain_keywords: Set[str] = field(default_factory=set)
    has_clear_structure: bool = False


class SpecGenerator:
    """SPEC 템플릿 자동 생성기

    코드 파일을 분석하여 SPEC-first 원칙에 맞는 SPEC 템플릿을 자동으로 생성합니다.

    Usage:
        generator = SpecGenerator()
        result = generator.generate_spec_template(
            code_file=Path("src/auth/login.py"),
            domain="AUTH"
        )
        print(result["spec_path"])
        print(result["content"])
        print(f"신뢰도: {result['confidence']:.0%}")
    """

    # 도메인 추론용 키워드 맵
    DOMAIN_KEYWORDS = {
        "AUTH": {"authenticate", "login", "logout", "token", "password", "user"},
        "PAYMENT": {"payment", "pay", "billing", "transaction", "charge", "amount"},
        "USER": {"user", "profile", "account", "registration", "signup"},
        "API": {"endpoint", "request", "response", "handler", "route"},
        "DATA": {"database", "query", "record", "data", "model"},
        "FILE": {"upload", "download", "file", "storage", "bucket"},
        "EMAIL": {"email", "mail", "message", "notification"},
        "LOG": {"log", "logging", "trace", "debug", "audit"},
    }

    def __init__(self):
        """초기화"""
        self.creation_timestamp = datetime.now().isoformat()

    def generate_spec_template(
        self,
        code_file: Path,
        domain: Optional[str] = None,
    ) -> Dict[str, Any]:
        """SPEC 템플릿 생성 (주요 메서드)

        Args:
            code_file: 코드 파일 경로
            domain: 도메인 (미지정 시 자동 추론)

        Returns:
            생성 결과 딕셔너리:
            - spec_path: SPEC 파일 경로
            - content: SPEC 템플릿 내용
            - domain: 추론된 도메인
            - confidence: 신뢰도 (0-1)
            - suggestions: 편집 가이드
        """
        result = {
            "success": False,
            "spec_path": None,
            "content": None,
            "domain": None,
            "confidence": 0.0,
            "suggestions": [],
            "editing_guide": []
        }

        try:
            code_file = Path(code_file)
            if not code_file.exists():
                result["error"] = f"파일을 찾을 수 없습니다: {code_file}"
                return result

            # 코드 분석
            analysis = self._analyze_code_file(code_file)

            # 도메인 추론
            if not domain:
                domain = self._infer_domain(code_file, analysis)

            # SPEC 경로 생성
            spec_path = Path(f".moai/specs/SPEC-{domain}/spec.md")

            # EARS 포맷 템플릿 생성
            content = self._create_ears_template(code_file, domain, analysis)

            # 신뢰도 계산
            confidence = self._calculate_confidence(code_file, analysis, domain)

            # 편집 가이드 생성
            editing_guide = self._generate_editing_guide(analysis, confidence, domain)

            result.update({
                "success": True,
                "spec_path": str(spec_path),
                "content": content,
                "domain": domain,
                "confidence": confidence,
                "editing_guide": editing_guide
            })

        except Exception as e:
            result["error"] = str(e)

        return result

    def _analyze_code_file(self, code_file: Path) -> CodeAnalysis:
        """코드 파일 분석

        파일 유형에 따라 AST 또는 정규식 분석.

        Args:
            code_file: 코드 파일 경로

        Returns:
            CodeAnalysis 객체
        """
        analysis = CodeAnalysis()
        suffix = code_file.suffix.lower()

        try:
            content = code_file.read_text(encoding="utf-8", errors="ignore")

            if suffix == ".py":
                self._analyze_python(content, analysis)
            elif suffix in {".js", ".jsx", ".ts", ".tsx"}:
                self._analyze_javascript(content, analysis)
            elif suffix == ".go":
                self._analyze_go(content, analysis)

        except Exception:
            pass

        return analysis

    def _analyze_python(self, content: str, analysis: CodeAnalysis) -> None:
        """Python 코드 분석 (AST 기반)"""
        try:
            tree = ast.parse(content)

            # 모듈 docstring
            if ast.get_docstring(tree):
                analysis.docstring = ast.get_docstring(tree)

            # 함수 추출
            for node in ast.walk(tree):
                if isinstance(node, ast.FunctionDef):
                    analysis.functions[node.name] = {
                        "docstring": ast.get_docstring(node),
                        "params": [arg.arg for arg in node.args.args],
                        "lineno": node.lineno
                    }

                elif isinstance(node, ast.ClassDef):
                    methods = []
                    for item in node.body:
                        if isinstance(item, ast.FunctionDef):
                            methods.append(item.name)

                    analysis.classes[node.name] = {
                        "docstring": ast.get_docstring(node),
                        "methods": methods,
                        "lineno": node.lineno
                    }

            # 구조 평가
            analysis.has_clear_structure = bool(analysis.functions or analysis.classes)

        except SyntaxError:
            pass

    def _analyze_javascript(self, content: str, analysis: CodeAnalysis) -> None:
        """JavaScript 코드 분석 (정규식 기반)"""
        # 함수 추출
        func_pattern = r"(?:async\s+)?function\s+(\w+)\s*\([^)]*\)|const\s+(\w+)\s*=\s*(?:async\s+)?\([^)]*\)\s*=>"
        for match in re.finditer(func_pattern, content):
            func_name = match.group(1) or match.group(2)
            analysis.functions[func_name] = {"lineno": content[:match.start()].count("\n")}

        # JSDoc 추출
        jsdoc_pattern = r"/\*\*\s*([\s\S]*?)\*/"
        for match in re.finditer(jsdoc_pattern, content):
            if match.group(1).strip():
                analysis.docstring = match.group(1).strip()

        analysis.has_clear_structure = bool(analysis.functions)

    def _analyze_go(self, content: str, analysis: CodeAnalysis) -> None:
        """Go 코드 분석 (정규식 기반)"""
        # 함수 추출
        func_pattern = r"func\s+(?:\([^)]*\)\s+)?(\w+)\s*\("
        for match in re.finditer(func_pattern, content):
            analysis.functions[match.group(1)] = {"lineno": content[:match.start()].count("\n")}

        # 주석 추출
        comment_pattern = r"//\s*(.+)"
        for match in re.finditer(comment_pattern, content):
            text = match.group(1).strip()
            if text and not analysis.docstring:
                analysis.docstring = text

        analysis.has_clear_structure = bool(analysis.functions)

    def _infer_domain(self, code_file: Path, analysis: CodeAnalysis) -> str:
        """도메인 추론

        우선 순위:
        1. 파일명/경로에서 추론
        2. 코드의 클래스명에서 추론
        3. 함수명에서 추론
        4. docstring에서 추론

        Args:
            code_file: 코드 파일 경로
            analysis: 코드 분석 결과

        Returns:
            추론된 도메인 (예: AUTH, PAYMENT)
        """
        # 1. 파일 경로에서 추론
        path_str = str(code_file).upper()
        for domain, keywords in self.DOMAIN_KEYWORDS.items():
            if any(kw.upper() in path_str for kw in keywords):
                analysis.domain_keywords.add(domain)
                return domain

        # 2. 클래스명에서 추론
        for class_name in analysis.classes.keys():
            for domain, keywords in self.DOMAIN_KEYWORDS.items():
                if any(kw.lower() in class_name.lower() for kw in keywords):
                    analysis.domain_keywords.add(domain)
                    return domain

        # 3. 함수명에서 추론
        for func_name in analysis.functions.keys():
            for domain, keywords in self.DOMAIN_KEYWORDS.items():
                if any(kw.lower() in func_name.lower() for kw in keywords):
                    analysis.domain_keywords.add(domain)
                    return domain

        # 4. docstring에서 추론
        if analysis.docstring:
            for domain, keywords in self.DOMAIN_KEYWORDS.items():
                if any(kw.lower() in analysis.docstring.lower() for kw in keywords):
                    analysis.domain_keywords.add(domain)
                    return domain

        # 기본값
        return "COMMON"

    def _create_ears_template(
        self,
        code_file: Path,
        domain: str,
        analysis: CodeAnalysis
    ) -> str:
        """EARS 포맷 SPEC 템플릿 생성

        Args:
            code_file: 코드 파일 경로
            domain: 도메인
            analysis: 코드 분석 결과

        Returns:
            EARS 포맷 SPEC 템플릿 내용
        """
        spec_id = f"SPEC-{domain}-001"
        created_at = datetime.now().strftime("%Y-%m-%d")

        template = f"""# {spec_id} | @SPEC:{domain}

**프로젝트**: {code_file.parent.parent.parent.name}
**파일**: `{code_file}`
**생성일**: {created_at}
**상태**: draft

---

## HISTORY

- {created_at}: 자동 생성된 SPEC 템플릿

---

## Overview

> **수정 필요**: 이 섹션을 작성하여 {domain} 기능의 전체 목표와 범위를 설명하세요.

{self._get_function_documentation(analysis)}

---

## Requirements

### Ubiquitous Requirements (항상 만족해야 함)

```
THE SYSTEM SHALL [요구사항 설명]
```

### State-Driven Requirements (특정 상태에서)

```
WHEN [조건]
THE SYSTEM SHALL [동작]
```

### Event-Driven Requirements (이벤트 발생 시)

```
IF [이벤트]
THE SYSTEM SHALL [동작]
```

### Optional Requirements (선택 사항)

```
THE SYSTEM MAY [선택 기능]
```

### Unwanted Behaviors (방지해야 함)

```
THE SYSTEM SHALL NOT [금지 동작]
```

---

## Environment

> **수정 필요**: 이 SPEC이 동작하기 위한 환경 조건을 작성하세요.

- 필요한 외부 서비스: [예: Database, API Gateway]
- 필요한 라이브러리: [예: cryptography v41.0.0+]
- 환경 변수: [예: DATABASE_URL, SECRET_KEY]

---

## Assumptions

> **수정 필요**: 이 SPEC이 성립하기 위한 가정들을 작성하세요.

1. [가정 1]
2. [가정 2]

---

## Test Cases

### 정상 케이스 (Happy Path)

```
GIVEN [초기 상태]
WHEN [사용자 행동]
THEN [예상 결과]
```

### 에러 케이스 (Error Handling)

```
GIVEN [초기 상태]
WHEN [에러 발생 시나리오]
THEN [에러 처리 결과]
```

### 엣지 케이스 (Edge Cases)

```
GIVEN [경계 상황]
WHEN [경계 행동]
THEN [경계 결과]
```

---

## Implementation Notes

### Code References

- 관련 코드: `{code_file}`

### Design Decisions

> **수정 필요**: 설계 결정 사항을 기록하세요.

### Dependencies

> **수정 필요**: 외부 의존성을 나열하세요.

### Performance Considerations

> **수정 필요**: 성능 관련 고려사항을 작성하세요.

---

## Related Specifications

> **수정 필요**: 이와 관련된 다른 SPEC들을 참조하세요.

- SPEC-PARENT: 상위 SPEC (있으면)
- SPEC-RELATED: 관련 SPEC (있으면)

---

## TODO Checklist

완성하기 전에 이 체크리스트를 검토하세요:

- [ ] Overview 섹션 작성 완료
- [ ] 최소 3개 이상의 요구사항 정의
- [ ] Environment 섹션 상세 작성
- [ ] Assumptions 섹션 작성
- [ ] Test Cases 3가지 이상 정의
- [ ] Code References 확인
- [ ] Related Specifications 검토
- [ ] 팀 리뷰 완료

---

**작성자**: @user
**최종 검수**: Pending
"""
        return template

    def _get_function_documentation(self, analysis: CodeAnalysis) -> str:
        """코드에서 추출한 함수/클래스 정보를 문서화"""
        doc_parts = []

        if analysis.classes:
            doc_parts.append("### 주요 클래스\n")
            for class_name, info in analysis.classes.items():
                doc_parts.append(f"- **{class_name}**")
                if info.get("docstring"):
                    doc_parts.append(f"  - {info['docstring']}")
                if info.get("methods"):
                    doc_parts.append(f"  - 메서드: {', '.join(info['methods'])}")
            doc_parts.append("")

        if analysis.functions:
            doc_parts.append("### 주요 함수\n")
            for func_name, info in analysis.functions.items():
                doc_parts.append(f"- **{func_name}**")
                if info.get("docstring"):
                    doc_parts.append(f"  - {info['docstring']}")
                if info.get("params"):
                    doc_parts.append(f"  - 파라미터: {', '.join(info['params'])}")
            doc_parts.append("")

        return "\n".join(doc_parts) if doc_parts else ""

    def _calculate_confidence(
        self,
        code_file: Path,
        analysis: CodeAnalysis,
        domain: str
    ) -> float:
        """신뢰도 계산 (0-1)

        팩터:
        - 명확한 코드 구조 (30%)
        - 도메인 추론 명확성 (40%)
        - docstring 존재 여부 (30%)

        Args:
            code_file: 코드 파일 경로
            analysis: 코드 분석 결과
            domain: 추론된 도메인

        Returns:
            신뢰도 점수 (0-1)
        """
        score = 0.0

        # 명확한 구조 (최대 0.3)
        if analysis.has_clear_structure:
            score += 0.3

        # 도메인 추론 명확성 (최대 0.4)
        if domain in analysis.domain_keywords:
            score += 0.4
        elif domain != "COMMON":
            score += 0.2

        # docstring (최대 0.3)
        if analysis.docstring:
            score += 0.2
        if analysis.functions and any(f.get("docstring") for f in analysis.functions.values()):
            score += 0.1

        return min(score, 1.0)

    def _generate_editing_guide(
        self,
        analysis: CodeAnalysis,
        confidence: float,
        domain: str
    ) -> List[str]:
        """편집 가이드 생성 (TODO 항목)

        신뢰도가 낮을수록 더 많은 가이드 제시.

        Args:
            analysis: 코드 분석 결과
            confidence: 신뢰도
            domain: 도메인

        Returns:
            편집 가이드 항목 리스트
        """
        guide = [
            "[ ] 개요(Overview) 섹션 작성",
            "[ ] 요구사항(Requirements) 최소 3개 정의",
            "[ ] 환경(Environment) 섹션 상세 작성",
            "[ ] 가정(Assumptions) 항목 정의",
            "[ ] 테스트 케이스(정상/에러/엣지) 작성",
        ]

        # 낮은 신뢰도 → 더 많은 가이드
        if confidence < 0.5:
            guide.extend([
                "[ ] 도메인 '{domain}' 확인 (자동 추론됨)",
                "[ ] 관련 함수/클래스 요구사항과 연결",
                "[ ] 외부 API/라이브러리 의존성 나열",
                "[ ] 성능 고려사항 검토",
            ])

        if not analysis.docstring:
            guide.append("[ ] 모듈 docstring 추가 (코드 가독성 향상)")

        if not any(f.get("docstring") for f in analysis.functions.values()):
            guide.append("[ ] 함수 docstring 추가 (자동 SPEC 생성 품질 향상)")

        return guide
