package web

// 이 파일은 M5-b D3 — 4개 스키마 위젯(schemaTextRow / schemaNumberRow /
// schemaToggleRow / schemaSelectRow) 이 방출하는 f.<name>.title / f.<name>.desc
// data-i18n 훅의 baseline 텍스트를 생성한다.
//
// M2b까지 이 필드들은 기술 식별자 key chip 으로만 렌더되었고 title/desc 가
// 없었다 (fieldsets.templ 기존 주석 참조). D3 는 이를 기존 langSelect /
// optSelect / toggle / numberField 위젯과 동일한 field__title + field__key +
// field__desc 3-요소 패턴으로 끌어올린다 — title/desc 는 data-i18n 으로 4-locale
// 렌더되고, key chip (yaml 키) 은 여전히 data-i18n 없이 영문 코드 토큰으로
// 렌더된다 (TestDataI18nKeysNoCodeChip 계약 보존).
//
// baseline 텍스트는 영문 humanized 형태다. 실제 4-locale 번역은 i18n.js 사전이
// 담당한다 (applyI18n 이 data-i18n 키의 사전 값으로 교체).

import (
	"strings"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// schemaFieldTitle 은 FieldDef 의 마지막 path segment 를 humanized 한 영문
// baseline title 이다 (예: "token_threshold" → "Token threshold").
// role_profiles.<role>.<field> 패턴은 role 이름을 앞에 배치한다.
func schemaFieldTitle(f settings.FieldDef) string {
	return humanizeLastSegment(f.Name)
}

// schemaFieldDesc 은 FieldDef 의 타입과 path 에서 파생한 간단한 영문 baseline
// 설명이다. 사전값이 우선한다 (data-i18n).
func schemaFieldDesc(f settings.FieldDef) string {
	return fieldDescByType(f)
}

// humanizeLastSegment 는 dot-path 의 마지막 segment 를 사람이 읽기 쉬운 형태로
// 변환한다 (예: "auto_clear.token_threshold" → "Token threshold").
// role_profiles 패턴은 "<Role> <field>" 형태로 확장한다.
func humanizeLastSegment(name string) string {
	parts := strings.Split(name, ".")
	last := parts[len(parts)-1]
	// role_profiles.<role>.<field> 패턴 감지.
	for i, p := range parts {
		if p == "role_profiles" && i+1 < len(parts) {
			role := parts[i+1]
			return humanizeIdent(role) + " " + humanizeIdent(last)
		}
	}
	return humanizeIdent(last)
}

// humanizeIdent 는 snake_case 식별자를 "First word rest" 형태로 humanize 한다.
// 첫 단어만 대문자, 나머지는 소문자 그대로 (예: "token_threshold" → "Token
// threshold", "pre_push" → "Pre push").
func humanizeIdent(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	if s == "" {
		return s
	}
	// 첫 글자만 대문자로.
	return strings.ToUpper(s[:1]) + s[1:]
}

// fieldDescByType 은 필드 타입과 이름에서 간단한 설명을 파생한다.
func fieldDescByType(f settings.FieldDef) string {
	last := lastSegment(f.Name)
	switch f.Type {
	case settings.TypeBool:
		return "Toggle this setting on or off."
	case settings.TypeInt, settings.TypeFloat:
		return "Numeric value for " + humanizeIdent(last) + "."
	case settings.TypeSelect:
		return "Select the " + humanizeIdent(last) + "."
	default:
		return "Text value for " + humanizeIdent(last) + "."
	}
}

// lastSegment 는 dot-path 의 마지막 segment 를 반환한다.
func lastSegment(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

// ─── SPEC-WEB-CONSOLE-014 M2/M4: read-only / raw view note-key 해석 ──────────
//
// ReadOnlyField.NoteKey / RawBlockRef.NoteKey 가 비어 있으면 제네릭 라벨 키를,
// 아니면 정직한 설명 라벨 키를 반환한다. baseline 텍스트는 applyI18n 실행 전
// fallback 이며 4-locale 번역은 i18n.js 사전이 담당한다.

// roNoteKey 는 read-only 설명 라벨의 data-i18n 키를 반환한다.
func roNoteKey(noteKey string) string {
	if noteKey == "" {
		return "ro.note"
	}
	return noteKey
}

// roNoteBaseline 은 read-only 설명 라벨의 영문 baseline 텍스트를 반환한다.
func roNoteBaseline(noteKey string) string {
	switch noteKey {
	case "ro.note.governance":
		return "read-only — governance FROZEN (auto_apply stays false)"
	case "ro.note.dead_config":
		return "read-only — path fixed by the runtime (informational)"
	default:
		return "read-only (runtime-managed)"
	}
}

// rawNoteKey 는 raw view 요약 라벨의 data-i18n 키를 반환한다.
func rawNoteKey(noteKey string) string {
	if noteKey == "" {
		return "raw.note"
	}
	return noteKey
}

// rawNoteBaseline 은 raw view 요약 라벨의 영문 baseline 텍스트를 반환한다.
func rawNoteBaseline(noteKey string) string {
	switch noteKey {
	case "raw.note.informational":
		return "informational — displayed value is not wired to runtime enforcement"
	default:
		return "structured block (read-only)"
	}
}
