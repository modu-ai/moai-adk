# SPEC-SKILL-REFACTOR-001 구현 계획서

> Claude Code Skills 표준화 작업 계획

---

## 📋 작업 개요

**목표**: 50개 Skills를 Anthropic 공식 표준에 맞게 표준화

**범위**:
1. 파일명 표준화 (skill.md → SKILL.md)
2. 중복 템플릿 삭제
3. YAML 필드 정리
4. allowed-tools 필드 추가

**예상 작업 시간**: 25분 (병렬 처리 시)

---

## 🎯 작업 그룹

### 그룹 1: 중복 템플릿 삭제 (1분)

**작업**: moai-cc-*-template 5개 디렉토리 삭제

**스크립트**:
```bash
# .claude/skills/ 에서 삭제
rm -rf /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-cc-agent-template
rm -rf /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-cc-command-template
rm -rf /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-cc-skill-template
rm -rf /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-cc-plugin-template
rm -rf /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-cc-settings-template

# src/moai_adk/templates/.claude/skills/ 에서 삭제 (이미 완료)
# (이미 3개는 삭제되었고, plugin, settings 템플릿만 확인 필요)
```

**검증**:
```bash
ls .claude/skills/ | grep moai-cc- | wc -l  # 0이어야 함
```

---

### 그룹 2: 파일명 표준화 (5분)

**작업**: skill.md → SKILL.md (50개)

**스크립트**:
```bash
# .claude/skills/ 디렉토리
cd /Users/goos/MoAI/MoAI-ADK/.claude/skills
for dir in moai-*/; do
  if [ -f "$dir/skill.md" ]; then
    echo "Renaming: $dir/skill.md → SKILL.md"
    mv "$dir/skill.md" "$dir/SKILL.md"
  fi
done

# src/moai_adk/templates/.claude/skills/ 디렉토리
cd /Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills
for dir in moai-*/; do
  if [ -f "$dir/skill.md" ]; then
    echo "Renaming: $dir/skill.md → SKILL.md"
    mv "$dir/skill.md" "$dir/SKILL.md"
  fi
done
```

**검증**:
```bash
# skill.md 파일이 없어야 함
find .claude/skills/ -name "skill.md" | wc -l  # 0
find src/moai_adk/templates/.claude/skills/ -name "skill.md" | wc -l  # 0

# SKILL.md 파일이 51개여야 함
find .claude/skills/ -name "SKILL.md" | wc -l  # 51
```

---

### 그룹 3: YAML 필드 정리 (10분)

**작업**: version, author, license, tags, model 필드 제거

**Python 스크립트** (`scripts/clean_yaml_fields.py`):
```python
#!/usr/bin/env python3
"""
YAML 필드 정리 스크립트
- version, author, license, tags, model 필드 제거
- name, description, allowed-tools만 유지
"""

import sys
from pathlib import Path
from ruamel.yaml import YAML

yaml = YAML()
yaml.preserve_quotes = True
yaml.default_flow_style = False

def clean_yaml_fields(skill_file: Path):
    """YAML frontmatter에서 불필요한 필드 제거"""
    content = skill_file.read_text()
    
    if not content.startswith('---'):
        print(f"No YAML frontmatter: {skill_file}")
        return
    
    # YAML frontmatter 추출
    parts = content.split('---', 2)
    if len(parts) < 3:
        print(f"Invalid YAML structure: {skill_file}")
        return
    
    yaml_str = parts[1]
    body = parts[2]
    
    # YAML 파싱
    data = yaml.load(yaml_str)
    
    # 보존할 필드
    preserved = {}
    if 'name' in data:
        preserved['name'] = data['name']
    if 'description' in data:
        preserved['description'] = data['description']
    if 'allowed-tools' in data:
        preserved['allowed-tools'] = data['allowed-tools']
    
    # 파일 재작성
    new_yaml = yaml.dump_to_string(preserved)
    new_content = f"---\n{new_yaml}---{body}"
    skill_file.write_text(new_content)
    print(f"Cleaned: {skill_file}")

def main():
    # .claude/skills/
    skills_dir = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")
    for skill_file in skills_dir.glob("*/SKILL.md"):
        clean_yaml_fields(skill_file)
    
    # src/moai_adk/templates/.claude/skills/
    templates_dir = Path("/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills")
    for skill_file in templates_dir.glob("*/SKILL.md"):
        clean_yaml_fields(skill_file)

if __name__ == "__main__":
    main()
```

**실행**:
```bash
python scripts/clean_yaml_fields.py
```

**검증**:
```bash
# version 필드가 없어야 함
rg "^version:" .claude/skills/*/SKILL.md | wc -l  # 0
# model 필드가 없어야 함
rg "^model:" .claude/skills/*/SKILL.md | wc -l  # 0
```

---

### 그룹 4: allowed-tools 필드 추가 (20분)

**작업**: 25개 스킬에 allowed-tools 추가

**스킬 유형별 도구 권한**:

| 스킬 유형 | allowed-tools | 예시 |
|---------|---------------|------|
| **Alfred 에이전트** | Read, Write, Edit, Bash, TodoWrite | debugger-pro, code-reviewer |
| **Lang 스킬** | Read, Bash | python, typescript, rust |
| **Domain 스킬** | Read, Bash | backend, frontend, cli-tool |

**Python 스크립트** (`scripts/add_allowed_tools.py`):
```python
#!/usr/bin/env python3
"""
allowed-tools 필드 추가 스크립트
"""

import sys
from pathlib import Path
from ruamel.yaml import YAML

yaml = YAML()
yaml.preserve_quotes = True
yaml.default_flow_style = False

# 스킬 유형별 도구 매핑
ALFRED_TOOLS = ["Read", "Write", "Edit", "Bash", "TodoWrite"]
LANG_TOOLS = ["Read", "Bash"]
DOMAIN_TOOLS = ["Read", "Bash"]

def add_allowed_tools(skill_file: Path):
    """allowed-tools 필드 추가"""
    content = skill_file.read_text()
    
    if not content.startswith('---'):
        print(f"No YAML frontmatter: {skill_file}")
        return
    
    # YAML frontmatter 추출
    parts = content.split('---', 2)
    if len(parts) < 3:
        print(f"Invalid YAML structure: {skill_file}")
        return
    
    yaml_str = parts[1]
    body = parts[2]
    
    # YAML 파싱
    data = yaml.load(yaml_str)
    
    # 이미 allowed-tools가 있으면 건너뛰기
    if 'allowed-tools' in data:
        print(f"Skip (already has allowed-tools): {skill_file}")
        return
    
    # 스킬 이름으로 유형 결정
    name = data.get('name', '')
    
    if 'alfred' in name:
        tools = ALFRED_TOOLS
    elif 'lang' in name:
        tools = LANG_TOOLS
    elif 'domain' in name:
        tools = DOMAIN_TOOLS
    else:
        # 기본값 (참조 전용)
        tools = ["Read"]
    
    data['allowed-tools'] = tools
    
    # 파일 재작성
    new_yaml = yaml.dump_to_string(data)
    new_content = f"---\n{new_yaml}---{body}"
    skill_file.write_text(new_content)
    print(f"Added tools to: {skill_file} ({len(tools)} tools)")

def main():
    # .claude/skills/
    skills_dir = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")
    for skill_file in skills_dir.glob("*/SKILL.md"):
        add_allowed_tools(skill_file)
    
    # src/moai_adk/templates/.claude/skills/
    templates_dir = Path("/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills")
    for skill_file in templates_dir.glob("*/SKILL.md"):
        add_allowed_tools(skill_file)

if __name__ == "__main__":
    main()
```

**실행**:
```bash
python scripts/add_allowed_tools.py
```

**검증**:
```bash
# allowed-tools 필드가 51개여야 함
rg "^allowed-tools:" .claude/skills/*/SKILL.md | wc -l  # 51
```

---

## 🔄 실행 순서

### 순차 실행 (안전)
```bash
# 1. 중복 템플릿 삭제
./scripts/delete_duplicate_templates.sh

# 2. 파일명 변경
./scripts/rename_skill_files.sh

# 3. YAML 필드 정리
python scripts/clean_yaml_fields.py

# 4. allowed-tools 추가
python scripts/add_allowed_tools.py
```

### 병렬 실행 (빠름)
```bash
# 1. 중복 템플릿 삭제 (단독)
./scripts/delete_duplicate_templates.sh

# 2. 나머지 병렬 실행
./scripts/rename_skill_files.sh &
python scripts/clean_yaml_fields.py &
python scripts/add_allowed_tools.py &
wait
```

---

## ✅ 최종 검증

### 체크리스트

- [ ] skill.md 파일 0개 (모두 SKILL.md로 변경)
- [ ] SKILL.md 파일 51개 (전체 스킬)
- [ ] moai-cc-*-template 디렉토리 0개 (모두 삭제)
- [ ] version, author, license, tags, model 필드 0개 (모두 제거)
- [ ] allowed-tools 필드 51개 (모든 스킬에 추가)

### 검증 명령어

```bash
# 종합 검증 스크립트
cat > scripts/verify_standardization.sh << 'EOF'
#!/bin/bash

echo "=== Skills 표준화 검증 ==="

# 1. 파일명 검증
skill_md_count=$(find .claude/skills/ -name "skill.md" 2>/dev/null | wc -l | tr -d ' ')
SKILL_md_count=$(find .claude/skills/ -name "SKILL.md" 2>/dev/null | wc -l | tr -d ' ')

echo "1. 파일명 표준화:"
echo "   - skill.md (비표준): $skill_md_count (0이어야 함)"
echo "   - SKILL.md (표준): $SKILL_md_count (51이어야 함)"

# 2. 중복 템플릿 검증
duplicate_count=$(ls .claude/skills/ 2>/dev/null | grep -c "moai-cc-.*-template" || echo 0)

echo "2. 중복 템플릿:"
echo "   - moai-cc-*-template: $duplicate_count (0이어야 함)"

# 3. YAML 필드 검증
version_count=$(rg "^version:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')
model_count=$(rg "^model:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')
allowed_tools_count=$(rg "^allowed-tools:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')

echo "3. YAML 필드:"
echo "   - version 필드: $version_count (0이어야 함)"
echo "   - model 필드: $model_count (0이어야 함)"
echo "   - allowed-tools 필드: $allowed_tools_count (51이어야 함)"

# 종합 판정
if [ "$skill_md_count" -eq 0 ] && \
   [ "$SKILL_md_count" -eq 51 ] && \
   [ "$duplicate_count" -eq 0 ] && \
   [ "$version_count" -eq 0 ] && \
   [ "$model_count" -eq 0 ] && \
   [ "$allowed_tools_count" -eq 51 ]; then
    echo ""
    echo "✅ 모든 검증 통과!"
    exit 0
else
    echo ""
    echo "❌ 검증 실패. 위 항목을 확인하세요."
    exit 1
fi
EOF

chmod +x scripts/verify_standardization.sh
./scripts/verify_standardization.sh
```

---

## 📦 산출물

### 파일 변경 목록

**삭제**:
- .claude/skills/moai-cc-agent-template/
- .claude/skills/moai-cc-command-template/
- .claude/skills/moai-cc-skill-template/
- .claude/skills/moai-cc-plugin-template/
- .claude/skills/moai-cc-settings-template/

**파일명 변경** (100개):
- .claude/skills/*/skill.md → SKILL.md (50개)
- src/moai_adk/templates/.claude/skills/*/skill.md → SKILL.md (50개)

**파일 수정** (100개):
- YAML 필드 정리 (50개 × 2 = 100개)
- allowed-tools 추가 (25개 × 2 = 50개)

### Git 커밋 계획

```bash
# Commit 1: 중복 템플릿 삭제
git add .claude/skills/
git commit -m "🗑️ DELETE: 중복 CC 템플릿 5개 삭제

- moai-cc-agent-template
- moai-cc-command-template
- moai-cc-skill-template
- moai-cc-plugin-template
- moai-cc-settings-template

→ moai-claude-code로 통합 완료"

# Commit 2: 파일명 표준화
git add .claude/skills/ src/moai_adk/templates/
git commit -m "♻️ REFACTOR: Skills 파일명 표준화 (skill.md → SKILL.md)

- Anthropic 공식 표준 준수 (대문자 SKILL.md)
- 영향받는 스킬: 50개
- .claude/skills/ 및 templates/ 동기화"

# Commit 3: YAML 필드 정리
git add .claude/skills/ src/moai_adk/templates/
git commit -m "♻️ REFACTOR: YAML frontmatter 공식 표준 준수

- 제거: version, author, license, tags, model (174개 필드)
- 유지: name, description, allowed-tools
- Skills는 model 필드 불필요 (Agent 전용)"

# Commit 4: allowed-tools 추가
git add .claude/skills/ src/moai_adk/templates/
git commit -m "✨ FEATURE: allowed-tools 필드 추가 (명시적 권한 관리)

- Alfred 에이전트: Read, Write, Edit, Bash, TodoWrite
- Lang 스킬: Read, Bash
- Domain 스킬: Read, Bash
- 영향받는 스킬: 25개 (누락 스킬)"
```

---

**최종 업데이트**: 2025-10-19
**작성자**: @Goos
