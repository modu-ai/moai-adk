package settings

import (
	"reflect"
	"testing"
)

// TestFieldLookup은 Field 가 존재하는 필드를 찾고 부재 필드에 false 를 반환함을 검증한다.
func TestFieldLookup(t *testing.T) {
	f, ok := Field("model")
	if !ok {
		t.Fatal("Field(model) not found")
	}
	if f.Section != SectionLaunch {
		t.Errorf("model field section = %q, want %q", f.Section, SectionLaunch)
	}
	if _, ok := Field("__nonexistent__"); ok {
		t.Error("Field(__nonexistent__) reported found")
	}
}

// TestSectionFieldsOrder는 SectionFields 가 섹션별 필드를 반환함을 검증한다.
func TestSectionFieldsOrder(t *testing.T) {
	statusline := SectionFields(SectionStatusline)
	if len(statusline) != 16 {
		t.Fatalf("Statusline section field count = %d, want 16", len(statusline))
	}
	if statusline[0].Name != "statusline_theme" {
		t.Errorf("first statusline field = %q, want statusline_theme", statusline[0].Name)
	}
}

// TestStatuslineSegmentKeys는 15개 정규 세그먼트 키를 반환함을 검증한다.
func TestStatuslineSegmentKeys(t *testing.T) {
	keys := StatuslineSegmentKeys()
	if len(keys) != 15 {
		t.Fatalf("StatuslineSegmentKeys count = %d, want 15", len(keys))
	}
	// SegmentRepo(16번째)는 제외되어야 한다.
	for _, k := range keys {
		if k == "repo" {
			t.Error("StatuslineSegmentKeys must NOT include 'repo' (16th, outside 15-key schema)")
		}
	}
}

// TestSelectOptions는 select 필드의 옵션 값 슬라이스를 반환함을 검증한다.
func TestSelectOptions(t *testing.T) {
	f, _ := Field("model")
	vals := f.SelectOptions()
	want := []string{"opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan"}
	if !reflect.DeepEqual(vals, want) {
		t.Errorf("model SelectOptions = %v, want %v", vals, want)
	}

	// text 필드는 빈 슬라이스를 반환한다.
	textField, _ := Field("user_name")
	if len(textField.SelectOptions()) != 0 {
		t.Errorf("text field SelectOptions should be empty, got %v", textField.SelectOptions())
	}
}

// TestOptionValueHelpers는 view-model 헬퍼들이 정규 옵션 값을 반환함을 검증한다.
func TestOptionValueHelpers(t *testing.T) {
	cases := []struct {
		name string
		got  []string
		want []string
	}{
		{"model", ModelOptionValues(), []string{"opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan"}},
		{"effort", EffortOptionValues(), []string{"low", "medium", "high", "xhigh", "max"}},
		{"language", LanguageOptionValues(), []string{"en", "ko", "ja", "zh"}},
		{"development_mode", DevelopmentModeOptionValues(), []string{"ddd", "tdd"}},
		{"convention", ConventionOptionValues(), []string{"auto", "conventional-commits", "angular", "karma"}},
	}
	for _, c := range cases {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("%s option values = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// TestFieldOptionValuesAbsent는 옵션이 없는 필드에 nil 을 반환함을 검증한다.
func TestFieldOptionValuesAbsent(t *testing.T) {
	if vals := fieldOptionValues("__nonexistent__"); vals != nil {
		t.Errorf("fieldOptionValues(absent) = %v, want nil", vals)
	}
	if vals := fieldOptionValues("user_name"); len(vals) != 0 {
		t.Errorf("fieldOptionValues(text field) = %v, want empty", vals)
	}
}

// TestEmptyLabelFor는 정규 빈 옵션 라벨을 단일 원천에서 반환함을 검증한다(REQ-WC10-013).
func TestEmptyLabelFor(t *testing.T) {
	cases := map[string]string{
		"model":             emptyLabelProjectDefault,
		"effort_level":      emptyLabelRuntimeDefault,
		"conversation_lang": emptyLabelUnset,
		"git_convention":    emptyLabelProjectDefault,
		"development_mode":  emptyLabelProjectDefault,
	}
	for name, want := range cases {
		if got := EmptyLabelFor(name); got != want {
			t.Errorf("EmptyLabelFor(%q) = %q, want %q", name, got, want)
		}
	}
	if got := EmptyLabelFor("__nonexistent__"); got != "" {
		t.Errorf("EmptyLabelFor(absent) = %q, want empty", got)
	}
}

// TestNormalizePermissionMode는 acceptEdits → "" 정규화를 검증한다(REQ-WC10-014).
func TestNormalizePermissionMode(t *testing.T) {
	if got := normalizePermissionMode("acceptEdits"); got != "" {
		t.Errorf("normalizePermissionMode(acceptEdits) = %q, want empty", got)
	}
	if got := normalizePermissionMode("plan"); got != "plan" {
		t.Errorf("normalizePermissionMode(plan) = %q, want plan", got)
	}
	if got := normalizePermissionMode(""); got != "" {
		t.Errorf("normalizePermissionMode(empty) = %q, want empty", got)
	}

	// permission_mode 필드의 Persist.Normalize 가 이 함수를 가리킴을 확인.
	f, _ := Field("permission_mode")
	if f.Persist.Normalize == nil {
		t.Fatal("permission_mode field has nil Persist.Normalize")
	}
	if got := f.Persist.Normalize("acceptEdits"); got != "" {
		t.Errorf("permission_mode Persist.Normalize(acceptEdits) = %q, want empty", got)
	}
}
