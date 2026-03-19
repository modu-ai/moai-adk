package lessons

import (
	"fmt"
	"strings"
	"testing"
)

// mockStore는 테스트용 인메모리 LessonStore 구현체다.
type mockStore struct {
	lessons map[string]*Lesson
	saveErr error
}

func newMockStore() *mockStore {
	return &mockStore{lessons: make(map[string]*Lesson)}
}

func (m *mockStore) Save(lesson *Lesson) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.lessons[lesson.ID] = lesson
	return nil
}

func (m *mockStore) List(filter LessonFilter) ([]*Lesson, error) {
	var result []*Lesson
	for _, l := range m.lessons {
		result = append(result, l)
	}
	return result, nil
}

func (m *mockStore) Get(id string) (*Lesson, error) {
	l, ok := m.lessons[id]
	if !ok {
		return nil, fmt.Errorf("not found: %s", id)
	}
	return l, nil
}

func (m *mockStore) Archive(id string) error {
	l, ok := m.lessons[id]
	if !ok {
		return fmt.Errorf("not found: %s", id)
	}
	l.Active = false
	return nil
}

func (m *mockStore) Count() (int, error) {
	count := 0
	for _, l := range m.lessons {
		if l.Active {
			count++
		}
	}
	return count, nil
}

// --- ExtractFromQualityFailure 테스트 ---

func TestExtractor_ExtractFromQualityFailure_Basic(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	ext := NewExtractor(store)

	err := ext.ExtractFromQualityFailure("SPEC-001", "coverage", "coverage below 85%", []string{"unit-test"})
	if err != nil {
		t.Fatalf("ExtractFromQualityFailure 실패: %v", err)
	}

	if len(store.lessons) != 1 {
		t.Fatalf("레슨이 1개 저장되어야 함: got %d", len(store.lessons))
	}

	var saved *Lesson
	for _, l := range store.lessons {
		saved = l
	}

	if saved.Type != "quality_failure" {
		t.Errorf("Type = %q, want %q", saved.Type, "quality_failure")
	}
	if saved.Source != "SPEC-001" {
		t.Errorf("Source = %q, want %q", saved.Source, "SPEC-001")
	}
	if saved.Severity != "high" {
		t.Errorf("Severity = %q, want %q (quality_failure는 항상 high)", saved.Severity, "high")
	}
	if !saved.Active {
		t.Error("Active = false, want true")
	}
	if saved.HitCount != 1 {
		t.Errorf("HitCount = %d, want 1", saved.HitCount)
	}
}

func TestExtractor_ExtractFromQualityFailure_TagsAssignment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		failureType string
		inputTags   []string
		wantTags    []string // 반드시 포함되어야 할 태그들
	}{
		{
			name:        "태그가 없을 때 quality와 failureType만 추가",
			failureType: "lint",
			inputTags:   nil,
			wantTags:    []string{"quality", "lint"},
		},
		{
			name:        "기존 태그에 quality와 failureType 추가",
			failureType: "coverage",
			inputTags:   []string{"python", "backend"},
			wantTags:    []string{"python", "backend", "quality", "coverage"},
		},
		{
			name:        "security 실패 타입",
			failureType: "security",
			inputTags:   []string{"api"},
			wantTags:    []string{"api", "quality", "security"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := newMockStore()
			ext := NewExtractor(store)

			if err := ext.ExtractFromQualityFailure("SPEC-T", tc.failureType, "desc", tc.inputTags); err != nil {
				t.Fatalf("ExtractFromQualityFailure 실패: %v", err)
			}

			var saved *Lesson
			for _, l := range store.lessons {
				saved = l
			}

			tagSet := make(map[string]bool)
			for _, tag := range saved.Tags {
				tagSet[tag] = true
			}

			for _, want := range tc.wantTags {
				if !tagSet[want] {
					t.Errorf("태그 %q가 없음. 실제 태그: %v", want, saved.Tags)
				}
			}
		})
	}
}

func TestExtractor_ExtractFromQualityFailure_PatternContainsDescription(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	ext := NewExtractor(store)

	desc := "missing test for AuthService"
	if err := ext.ExtractFromQualityFailure("SPEC-002", "coverage", desc, nil); err != nil {
		t.Fatal(err)
	}

	for _, l := range store.lessons {
		if !strings.Contains(l.Pattern, desc) {
			t.Errorf("Pattern에 description이 포함되어야 함: got %q", l.Pattern)
		}
		if !strings.Contains(l.Pattern, "coverage") {
			t.Errorf("Pattern에 failureType이 포함되어야 함: got %q", l.Pattern)
		}
	}
}

// --- ExtractFromSPECCompletion 테스트 ---

func TestExtractor_ExtractFromSPECCompletion_SeverityThresholds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		iterations   int
		wantSeverity string
	}{
		{"3 이하 → low", 3, "low"},
		{"4 → medium (3 초과)", 4, "medium"},
		{"5 → medium", 5, "medium"},
		{"6 → high (5 초과)", 6, "high"},
		{"10 → high", 10, "high"},
		{"1 → low", 1, "low"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := newMockStore()
			ext := NewExtractor(store)

			if err := ext.ExtractFromSPECCompletion("SPEC-SV", tc.iterations, nil); err != nil {
				t.Fatalf("ExtractFromSPECCompletion 실패: %v", err)
			}

			for _, l := range store.lessons {
				if l.Severity != tc.wantSeverity {
					t.Errorf("iterations=%d: Severity = %q, want %q",
						tc.iterations, l.Severity, tc.wantSeverity)
				}
			}
		})
	}
}

func TestExtractor_ExtractFromSPECCompletion_TagsIncludeSpecCompletion(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	ext := NewExtractor(store)

	if err := ext.ExtractFromSPECCompletion("SPEC-003", 2, []string{"golang", "api"}); err != nil {
		t.Fatal(err)
	}

	for _, l := range store.lessons {
		found := false
		for _, tag := range l.Tags {
			if tag == "spec_completion" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("spec_completion 태그가 없음: %v", l.Tags)
		}
	}
}

func TestExtractor_ExtractFromSPECCompletion_PatternContainsIterations(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	ext := NewExtractor(store)

	if err := ext.ExtractFromSPECCompletion("SPEC-004", 7, nil); err != nil {
		t.Fatal(err)
	}

	for _, l := range store.lessons {
		if !strings.Contains(l.Pattern, "7") {
			t.Errorf("Pattern에 iteration 수가 포함되어야 함: got %q", l.Pattern)
		}
		if !strings.Contains(l.Pattern, "SPEC-004") {
			t.Errorf("Pattern에 specID가 포함되어야 함: got %q", l.Pattern)
		}
	}
}

// --- ExtractFromCritic 테스트 ---

func TestExtractor_ExtractFromCritic_IDUniqueness(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	ext := NewExtractor(store)

	highValueQs := []string{"Is auth secure?", "Is data validated?", "Is rate limited?"}
	dismissedQs := []string{"Should we add logging?", "Any comments?"}

	if err := ext.ExtractFromCritic("SPEC-005", highValueQs, dismissedQs); err != nil {
		t.Fatalf("ExtractFromCritic 실패: %v", err)
	}

	expectedTotal := len(highValueQs) + len(dismissedQs)
	if len(store.lessons) != expectedTotal {
		t.Errorf("총 레슨 수 불일치: got %d, want %d", len(store.lessons), expectedTotal)
	}

	// 모든 ID가 고유해야 한다 (루프 인덱스로 충돌 방지 확인)
	ids := make(map[string]bool)
	for id := range store.lessons {
		if ids[id] {
			t.Errorf("중복 ID 발견: %q", id)
		}
		ids[id] = true
	}
}

func TestExtractor_ExtractFromCritic_HighValueSeverity(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	ext := NewExtractor(store)

	if err := ext.ExtractFromCritic("SPEC-006", []string{"Is auth secure?"}, nil); err != nil {
		t.Fatal(err)
	}

	for _, l := range store.lessons {
		if l.Severity != "high" {
			t.Errorf("high_value 질문의 Severity = %q, want %q", l.Severity, "high")
		}
		tagSet := make(map[string]bool)
		for _, tag := range l.Tags {
			tagSet[tag] = true
		}
		if !tagSet["high_value"] {
			t.Errorf("high_value 태그가 없음: %v", l.Tags)
		}
		if !tagSet["critic"] {
			t.Errorf("critic 태그가 없음: %v", l.Tags)
		}
	}
}

func TestExtractor_ExtractFromCritic_DismissedSeverity(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	ext := NewExtractor(store)

	if err := ext.ExtractFromCritic("SPEC-007", nil, []string{"Should we add logging?"}); err != nil {
		t.Fatal(err)
	}

	for _, l := range store.lessons {
		if l.Severity != "low" {
			t.Errorf("dismissed 질문의 Severity = %q, want %q", l.Severity, "low")
		}
		tagSet := make(map[string]bool)
		for _, tag := range l.Tags {
			tagSet[tag] = true
		}
		if !tagSet["low_value"] {
			t.Errorf("low_value 태그가 없음: %v", l.Tags)
		}
	}
}

func TestExtractor_ExtractFromCritic_MultipleQuestionsAllSaved(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	ext := NewExtractor(store)

	highValueQs := []string{
		"Is authentication properly implemented?",
		"Are edge cases handled?",
	}
	dismissedQs := []string{
		"Should we rename the variable?",
		"Could we add more comments?",
		"Is this naming convention correct?",
	}

	if err := ext.ExtractFromCritic("SPEC-008", highValueQs, dismissedQs); err != nil {
		t.Fatalf("ExtractFromCritic 실패: %v", err)
	}

	// 고가치 + 기각 질문 합산만큼 저장되어야 한다
	total := len(highValueQs) + len(dismissedQs)
	if len(store.lessons) != total {
		t.Errorf("저장된 레슨 수: got %d, want %d", len(store.lessons), total)
	}
}

func TestExtractor_ExtractFromCritic_EmptyLists(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	ext := NewExtractor(store)

	// 빈 리스트로 호출해도 오류 없이 동작해야 한다
	if err := ext.ExtractFromCritic("SPEC-009", nil, nil); err != nil {
		t.Fatalf("빈 리스트 ExtractFromCritic 실패: %v", err)
	}
	if len(store.lessons) != 0 {
		t.Errorf("빈 리스트면 레슨이 없어야 함: got %d", len(store.lessons))
	}
}

func TestExtractor_ExtractFromCritic_IDContainsIndexNotCollide(t *testing.T) {
	t.Parallel()

	// 동일 specID + 같은 타임스탬프 기준으로 고가치와 기각 질문의 인덱스가 각각 0부터 시작하더라도
	// prefix(hv vs lv)로 구분되어 ID 충돌이 없어야 한다.
	store := newMockStore()
	ext := NewExtractor(store)

	if err := ext.ExtractFromCritic("SPEC-010",
		[]string{"q1"},
		[]string{"q2"},
	); err != nil {
		t.Fatal(err)
	}

	ids := make([]string, 0, 2)
	for id := range store.lessons {
		ids = append(ids, id)
	}
	if len(ids) != 2 {
		t.Fatalf("ID 충돌로 인해 레슨 1개만 저장됨: %v", ids)
	}

	// hv(고가치)와 lv(기각) prefix가 모두 존재해야 한다
	hvFound, lvFound := false, false
	for _, id := range ids {
		if strings.Contains(id, "-hv-") {
			hvFound = true
		}
		if strings.Contains(id, "-lv-") {
			lvFound = true
		}
	}
	if !hvFound {
		t.Errorf("고가치 ID(-hv-)가 없음: %v", ids)
	}
	if !lvFound {
		t.Errorf("기각 ID(-lv-)가 없음: %v", ids)
	}
}

// --- Injector 테스트 ---

func TestInjector_GetRelevantLessons(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := fmt.Sprintf("%s/lessons.jsonl", dir)
	fileStore := NewFileStore(path, "", 50)

	// 태그가 다른 레슨들을 저장
	l1 := newActiveLesson("inj-1", "quality_failure")
	l1.Tags = []string{"security"}
	l1.HitCount = 10
	l2 := newActiveLesson("inj-2", "quality_failure")
	l2.Tags = []string{"performance"}

	if err := fileStore.Save(l1); err != nil {
		t.Fatal(err)
	}
	if err := fileStore.Save(l2); err != nil {
		t.Fatal(err)
	}

	injector := NewInjector(fileStore, 5)
	results, err := injector.GetRelevantLessons([]string{"security"})
	if err != nil {
		t.Fatalf("GetRelevantLessons 실패: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("결과 수 불일치: got %d, want 1", len(results))
	}
	if results[0].ID != "inj-1" {
		t.Errorf("잘못된 레슨 반환: got %q, want %q", results[0].ID, "inj-1")
	}
}

func TestInjector_FormatForInjection_EmptyLessons(t *testing.T) {
	t.Parallel()

	injector := NewInjector(newMockStore(), 5)
	result := injector.FormatForInjection(nil)
	if result != "" {
		t.Errorf("빈 레슨 포맷: got %q, want %q", result, "")
	}
}

func TestInjector_FormatForInjection_ContainsLessonInfo(t *testing.T) {
	t.Parallel()

	injector := NewInjector(newMockStore(), 5)
	lessons := []*Lesson{
		{
			ID:       "fmt-1",
			Severity: "high",
			Pattern:  "security vulnerability found",
			Source:   "SPEC-100",
			HitCount: 3,
		},
	}

	result := injector.FormatForInjection(lessons)
	if !strings.Contains(result, "high") {
		t.Errorf("포맷 결과에 severity가 없음: %q", result)
	}
	if !strings.Contains(result, "security vulnerability found") {
		t.Errorf("포맷 결과에 pattern이 없음: %q", result)
	}
	if !strings.Contains(result, "SPEC-100") {
		t.Errorf("포맷 결과에 source가 없음: %q", result)
	}
}

func TestInjector_DefaultLimit(t *testing.T) {
	t.Parallel()

	// limit <= 0이면 기본값 5로 설정되어야 한다
	injector := NewInjector(newMockStore(), 0)
	if injector.limit != 5 {
		t.Errorf("기본 limit: got %d, want 5", injector.limit)
	}
}
