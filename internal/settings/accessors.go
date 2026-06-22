package settings

// 이 파일은 스키마 데이터(schema.go)에 대한 공개 접근자를 제공한다. 두 표면
// (internal/cli TUI, internal/web 콘솔)은 이 접근자를 통해 필드를 균일하게
// 순회한다.

// AllFields는 6개 섹션의 34개 정규 필드를 렌더 순서대로 반환한다.
func AllFields() []FieldDef {
	return allFields()
}

// SectionFields는 주어진 섹션에 속한 필드를 렌더 순서대로 반환한다.
func SectionFields(section SectionID) []FieldDef {
	var out []FieldDef
	for _, f := range allFields() {
		if f.Section == section {
			out = append(out, f)
		}
	}
	return out
}

// Field는 논리 필드명으로 FieldDef 를 조회한다. 두 번째 반환값은 존재 여부다.
func Field(name string) (FieldDef, bool) {
	for _, f := range allFields() {
		if f.Name == name {
			return f, true
		}
	}
	return FieldDef{}, false
}

// FieldNames는 34개 정규 필드의 논리명 집합을 렌더 순서대로 반환한다.
// AC-WC10-010 의 set-equality 비교에 사용된다.
func FieldNames() []string {
	fields := allFields()
	names := make([]string, 0, len(fields))
	for _, f := range fields {
		names = append(names, f.Name)
	}
	return names
}

// StatuslineSegmentKeys는 정규 15개 세그먼트 키를 렌더 순서대로 반환한다.
// 웹 statusline 섹션 재추가(REQ-WC10-009)와 TUI MultiSelect 가 동일한 15개
// 세그먼트를 단일 원천에서 파생하도록 한다.
func StatuslineSegmentKeys() []string {
	return statuslineSegmentKeys()
}

// SelectOptions는 select 필드의 정규 옵션 값 슬라이스를 반환한다(옵션이 없는
// 필드는 빈 슬라이스). 웹 view-model 의 옵션 목록 조립에 사용된다.
func (f FieldDef) SelectOptions() []string {
	out := make([]string, 0, len(f.Options))
	for _, o := range f.Options {
		out = append(out, o.Value)
	}
	return out
}

// ModelOptionValues는 정규 모델 옵션 값을 반환한다(웹 view-model 헬퍼).
func ModelOptionValues() []string {
	return fieldOptionValues("model")
}

// EffortOptionValues는 정규 effort level 옵션 값을 반환한다(웹 view-model 헬퍼).
func EffortOptionValues() []string {
	return fieldOptionValues("effort_level")
}

// LanguageOptionValues는 정규 언어 옵션 값(en/ko/ja/zh)을 반환한다.
func LanguageOptionValues() []string {
	return fieldOptionValues("conversation_lang")
}

// DevelopmentModeOptionValues는 정규 development mode 옵션 값을 반환한다.
func DevelopmentModeOptionValues() []string {
	return fieldOptionValues("development_mode")
}

// ConventionOptionValues는 정규 git convention 옵션 값을 반환한다.
func ConventionOptionValues() []string {
	return fieldOptionValues("git_convention")
}

// fieldOptionValues는 주어진 필드의 옵션 값 슬라이스를 반환한다(필드 부재 시 nil).
func fieldOptionValues(name string) []string {
	if f, ok := Field(name); ok {
		return f.SelectOptions()
	}
	return nil
}

// EmptyLabelFor는 주어진 필드명의 정규 빈 옵션 라벨을 반환한다(필드 부재 또는
// 빈 라벨이 없는 필드는 빈 문자열). 두 표면이 동일한 라벨을 렌더하도록 단일화한다
// (REQ-WC10-013, AC-WC10-014).
func EmptyLabelFor(name string) string {
	if f, ok := Field(name); ok {
		return f.EmptyLabel
	}
	return ""
}
