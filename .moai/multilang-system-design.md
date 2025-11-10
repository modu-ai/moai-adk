# MoAI-ADK 다국어 동적 치환 시스템 설계

## 📋 설계 개요

MoAI-ADK 프로젝트의 다국어 동적 치환 시스템 및 전문가 위임 아키텍처 포괄 설계

**작성일**: 2025-11-11  
**버전**: 1.0.0  
**설계자**: Alfred SuperAgent

---

## 🔍 현황 분석

### ✅ 이미 구현된 기능

#### 1. 다국어 동적 치환 시스템 (완벽하게 구현됨)
- **TemplateEngine**: Jinja2 기반 템플릿 엔진 (`src/moai_adk/core/template_engine.py`)
- **변수 치환**: `{{CONVERSATION_LANGUAGE_NAME}}` placeholder 전역 지원
- **언어 맵핑**: `migration.py`에 언어 코드→이름 자동 변환
- **지원 언어**: 
  ```python
  language_names = {
      "en": "English",
      "ko": "한국어", 
      "ja": "日本語",
      "zh": "中文",
      "es": "Español"
  }
  ```
- **Fallback 처리**: 알려지지 않은 언어는 "English"로 처리

#### 2. 현재 아키텍처 구조
- **Commands → Agents → Skills** 3계층 아키텍처
- **project-manager**가 `/alfred:0-project`의 모든 모드 처리
- **TemplateProcessor**가 동적 템플릿 치환 담당
- **Language First** 원칙 이미 구현됨

### 🔧 개선이 필요한 부분

#### 1. 전문가 위임 시스템 재설계
현재 project-manager가 모든 것을 처리하므로 전문 분리 필요

#### 2. 언어 시스템 확장
현재 5개 언어에서 더 많은 언어로 확장 필요

---

## 🏗️ 개선 설계: 전문가 위임 아키텍처

### 1. 전문가 에이전트 역할 재정의

#### 🎨 Frontend Expert (frontend-expert)
**담당 모드**: `settings`
**전문 분야**: 
- 사용자 인터페이스 및 상호작용 설계
- AskUserQuestion 통한 언어 선택 UX
- 설정 관리의 직관성 최적화
- 실시간 검증 및 피드백 시스템

**주요 기능**:
```python
# settings mode 핵심 로직
def handle_settings_mode():
    # 1. 언어 컨텍스트 확인
    current_language = get_conversation_language_name(config)
    
    # 2. 사용자 상호작용 (AskUserQuestion)
    language_choice = ask_user_language_selection(current_language)
    
    # 3. 실시간 검증 및 적용
    if language_choice != current_language:
        update_language_config(language_choice)
        trigger_template_substitution()
```

#### 📝 Doc-Syncer (doc-syncer)  
**담당 모드**: `update`
**전문 분야**:
- 문서 동기화 및 템플릿 최적화
- 패키지 템플릿 변경 감지
- 스마트 머징 및 충돌 해결
- 백업 및 롤백 시스템

**주요 기능**:
```python
# update mode 핵심 로직  
def handle_update_mode():
    # 1. 템플릿 변경 감지
    template_changes = detect_template_modifications()
    
    # 2. 스마트 머징
    merge_result = perform_smart_merging(template_changes)
    
    # 3. 다국어 변수 치환
    substitute_multilingual_variables(merge_result)
    
    # 4. 검증 및 롤백 준비
    create_backup_and_validate(merge_result)
```

#### 🚀 Project Manager (project-manager)
**담당 모드**: `init` 
**전문 분야**:
- 프로젝트 초기화 및 설정 생성
- Language-first 초기화 워크플로우
- 프로젝트 타입 감지 및 최적화
- 도메인별 설정 자동화

**주요 기능**:
```python
# init mode 핵심 로직
def handle_init_mode():
    # 1. Language-first 접근
    language_config = init_language_selection()
    
    # 2. 프로젝트 인터뷰
    project_profile = conduct_project_interview(language_config)
    
    # 3. 설정 생성
    config_result = generate_project_config(project_profile)
    
    # 4. 문서 자동 생성
    create_project_documentation(config_result)
```

#### 🔍 Implementation Planner (implementation-planner)
**담당 모드**: `auto-detect`
**전문 분야**: 
- 프로젝트 상태 분석 및 진단
- 설정 최적화 기회 탐지
- 마이그레이션 경로 제안
- 성능 분석 및 개선 제안

**주요 기능**:
```python
# auto-detect mode 핵심 로직
def handle_auto_detect_mode():
    # 1. 프로젝트 상태 분석
    project_status = analyze_project_configuration()
    
    # 2. 최적화 기회 탐지
    optimization_opportunities = detect_improvement_areas()
    
    # 3. 수정 옵션 제안
    recommendations = generate_recommendations(project_status)
    
    # 4. 사용자 선택에 따른 라우팅
    route_to_specialist(recommendations)
```

---

## 🌐 언어 시스템 확장 설계

### 1. 확장된 언어 맵핑 테이블

```python
# 확장된 언어 맵핑 시스템
EXTENDED_LANGUAGE_MAP = {
    # 현재 지원 언어
    "en": {"name": "English", "native": "English", "rtl": False},
    "ko": {"name": "Korean", "native": "한국어", "rtl": False},
    "ja": {"name": "Japanese", "native": "日本語", "rtl": False}, 
    "zh": {"name": "Chinese", "native": "中文", "rtl": False},
    "es": {"name": "Spanish", "native": "Español", "rtl": False},
    
    # 추가 지원 언어
    "fr": {"name": "French", "native": "Français", "rtl": False},
    "de": {"name": "German", "native": "Deutsch", "rtl": False},
    "it": {"name": "Italian", "native": "Italiano", "rtl": False},
    "pt": {"name": "Portuguese", "native": "Português", "rtl": False},
    "ru": {"name": "Russian", "native": "Русский", "rtl": False},
    "ar": {"name": "Arabic", "native": "العربية", "rtl": True},
    "hi": {"name": "Hindi", "native": "हिन्दी", "rtl": False},
    "th": {"name": "Thai", "native": "ไทย", "rtl": False},
    "vi": {"name": "Vietnamese", "native": "Tiếng Việt", "rtl": False},
    "nl": {"name": "Dutch", "native": "Nederlands", "rtl": False},
    "pl": {"name": "Polish", "native": "Polski", "rtl": False},
    "tr": {"name": "Turkish", "native": "Türkçe", "rtl": False},
    "sv": {"name": "Swedish", "native": "Svenska", "rtl": False},
    "da": {"name": "Danish", "native": "Dansk", "rtl": False},
    "no": {"name": "Norwegian", "native": "Norsk", "rtl": False},
    "fi": {"name": "Finnish", "native": "Suomi", "rtl": False},
    "cs": {"name": "Czech", "native": "Čeština", "rtl": False},
    "hu": {"name": "Hungarian", "native": "Magyar", "rtl": False},
    "ro": {"name": "Romanian", "native": "Română", "rtl": False},
    "bg": {"name": "Bulgarian", "native": "Български", "rtl": False},
    "hr": {"name": "Croatian", "native": "Hrvatski", "rtl": False},
    "sk": {"name": "Slovak", "native": "Slovenčina", "rtl": False},
    "sl": {"name": "Slovenian", "native": "Slovenščina", "rtl": False},
    "et": {"name": "Estonian", "native": "Eesti", "rtl": False},
    "lv": {"name": "Latvian", "native": "Latviešu", "rtl": False},
    "lt": {"name": "Lithuanian", "native": "Lietuvių", "rtl": False},
    "el": {"name": "Greek", "native": "Ελληνικά", "rtl": False},
    "he": {"name": "Hebrew", "native": "עברית", "rtl": True},
    "fa": {"name": "Persian", "native": "فارسی", "rtl": True},
    "ur": {"name": "Urdu", "native": "اردو", "rtl": True},
    "bn": {"name": "Bengali", "native": "বাংলা", "rtl": False},
    "ta": {"name": "Tamil", "native": "தமிழ்", "rtl": False},
    "te": {"name": "Telugu", "native": "తెలుగు", "rtl": False},
    "ml": {"name": "Malayalam", "native": "മലയാളം", "rtl": False},
    "kn": {"name": "Kannada", "native": "ಕನ್ನಡ", "rtl": False},
    "gu": {"name": "Gujarati", "native": "ગુજરાતી", "rtl": False},
    "pa": {"name": "Punjabi", "native": "ਪੰਜਾਬੀ", "rtl": False},
    "mr": {"name": "Marathi", "native": "मराठी", "rtl": False},
    "ne": {"name": "Nepali", "native": "नेपाली", "rtl": False},
    "si": {"name": "Sinhala", "native": "සිංහල", "rtl": False},
    "my": {"name": "Myanmar", "native": "မြန်မာ", "rtl": False},
    "km": {"name": "Khmer", "native": "ខ្មែរ", "rtl": False},
    "lo": {"name": "Lao", "native": "ລາວ", "rtl": False},
    "ka": {"name": "Georgian", "native": "ქართული", "rtl": False},
    "am": {"name": "Amharic", "native": "አማርኛ", "rtl": False},
    "sw": {"name": "Swahili", "native": "Kiswahili", "rtl": False},
    "zu": {"name": "Zulu", "native": "isiZulu", "rtl": False},
    "af": {"name": "Afrikaans", "native": "Afrikaans", "rtl": False},
    "is": {"name": "Icelandic", "native": "Íslenska", "rtl": False},
    "mt": {"name": "Maltese", "native": "Malti", "rtl": False},
    "cy": {"name": "Welsh", "native": "Cymraeg", "rtl": False},
    "ga": {"name": "Irish", "native": "Gaeilge", "rtl": False},
    "gd": {"name": "Scottish Gaelic", "native": "Gàidhlig", "rtl": False},
    "eu": {"name": "Basque", "native": "Euskara", "rtl": False},
    "ca": {"name": "Catalan", "native": "Català", "rtl": False},
    "gl": {"name": "Galician", "native": "Galego", "rtl": False},
    "oc": {"name": "Occitan", "native": "Occitan", "rtl": False},
}
```

### 2. 향상된 언어 처리 함수

```python
def get_enhanced_language_config(language_code: str) -> dict:
    """향상된 언어 설정 반환 (RTL 지원, native 이름 등)"""
    return EXTENDED_LANGUAGE_MAP.get(language_code, {
        "name": "English", 
        "native": "English", 
        "rtl": False
    })

def substitute_multilingual_variables(content: str, config: dict) -> str:
    """다국어 변수 동적 치환 (확장 버전)"""
    language_code = config.get("conversation_language", "en")
    lang_config = get_enhanced_language_config(language_code)
    
    substitutions = {
        "{{CONVERSATION_LANGUAGE_NAME}}": lang_config["name"],
        "{{CONVERSATION_LANGUAGE_NATIVE}}": lang_config["native"], 
        "{{CONVERSATION_LANGUAGE_CODE}}": language_code,
        "{{IS_RTL_LANGUAGE}}": "true" if lang_config["rtl"] else "false",
        "{{TEXT_DIRECTION}}": "rtl" if lang_config["rtl"] else "ltr",
    }
    
    for placeholder, value in substitutions.items():
        content = content.replace(placeholder, value)
    
    return content
```

---

## 🔄 개선된 /alfred:0-project 워크플로우

### Phase 1: 명령어 라우팅 및 분석

```python
# 향상된 명령어 라우팅 시스템
def route_alfred_command(subcommand: str) -> str:
    """전문가 에이전트로 명령어 라우팅"""
    
    routing_map = {
        "setting": "frontend-expert",
        "update": "doc-syncer", 
        "": "auto-detect"  # 기본값
    }
    
    # config.json 존재 여부 확인
    if not subcommand:
        if config_exists():
            return "auto-detect"  # implementation-planner
        else:
            return "init"  # project-manager
    
    return routing_map.get(subcommand, "auto-detect")
```

### Phase 2: 전문가 에이전트 실행

```python
# 전문가 에이전트 실행 프레임워크
def execute_specialist_agent(agent_type: str, mode: str, context: dict):
    """지정된 전문가 에이전트 실행"""
    
    specialist_handlers = {
        "frontend-expert": FrontendExpertHandler(),
        "doc-syncer": DocSyncerHandler(), 
        "project-manager": ProjectManagerHandler(),
        "implementation-planner": ImplementationPlannerHandler()
    }
    
    handler = specialist_handlers.get(agent_type)
    if handler:
        return handler.handle(mode, context)
    else:
        raise ValueError(f"Unknown specialist agent: {agent_type}")
```

---

## 🎯 구현 계획

### Phase 1: 전문가 에이전트 핸들러 구현 (1-2주)
1. **FrontendExpertHandler** 구현
   - settings mode 전문 처리
   - AskUserQuestion 통합 UI/UX
   - 실시간 언어 전환 시스템

2. **DocSyncerHandler** 구현  
   - update mode 전문 처리
   - 스마트 템플릿 머징
   - 백업 및 롤백 시스템

3. **ProjectManagerHandler** 개선
   - init mode 전문화
   - Language-first 워크플로우 최적화
   - 도메인별 설정 자동화

4. **ImplementationPlannerHandler** 구현
   - auto-detect mode 전문 처리
   - 상태 분석 및 진단 시스템
   - 최적화 추천 엔진

### Phase 2: 언어 시스템 확장 (1주)
1. **확장된 언어 맵핑 테이블** 통합
2. **향상된 다국어 치환 함수** 구현
3. **RTL 언어 지원** 추가
4. **지역화된 템플릿** 시스템

### Phase 3: 통합 및 테스트 (1주)
1. **통합 테스트** 수행
2. **다국어 시나리오** 검증  
3. **성능 최적화**
4. **문서 업데이트**

---

## 📊 기대 효과

### ✅ 성능 향상
- **전문 분리**: 각 모드 전문 처리로 성능 30% 향상 예상
- **병렬 처리**: 독립적인 전문가 에이전트로 병렬 실행 가능
- **캐싱 최적화**: 언어 설정 및 템플릿 캐싱으로 속도 개선

### 🌍 국제화 강화  
- **50+ 언어 지원**: 전 세계 주요 언어全覆盖
- **RTL 언어 지원**: 아랍어, 히브리어 등 RTL 언어 완벽 지원
- **지역화 템플릿**: 언어별 최적화된 템플릿 제공

### 👥 사용자 경험 개선
- **직관적 설정**: frontend-expert의 UX 전문 설계
- **실시간 피드백**: 즉각적인 설정 변경 및 적용
- **스마트 추천**: implementation-planner의 지능적 제안

### 🔧 유지보수성 향상
- **명확한 역할 분리**: 각 전문가의 책임 범위 명확
- **확장성 용이**: 새로운 전문가 및 기능 추가 쉬움
- **테스트 용이성**: 독립적인 전문가 단위 테스트 가능

---

## 🚀 결론

MoAI-ADK에는 이미 훌륭한 다국어 동적 치환 시스템이 구현되어 있습니다. 본 설계는 이 기반 위에서 **전문가 위임 시스템**을 재설계하여 **효율성**, **확장성**, **사용자 경험**을 획기적으로 개선하는 것을 목표로 합니다.

핵심은 **"각자의 전문 분야에 집중하는 전문가 에이전트들"**을 통해 **더 빠르고, 더 정확하며, 더 사용자 친화적인** MoAI-ADK 시스템을 구축하는 것입니다.

---

**다음 단계**: Phase 1 전문가 에이전트 핸들러 구현 시작
