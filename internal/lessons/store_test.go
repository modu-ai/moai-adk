package lessons

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// newActiveLesson은 테스트용 활성 Lesson을 생성하는 헬퍼 함수다.
func newActiveLesson(id, lessonType string) *Lesson {
	return &Lesson{
		ID:       id,
		Type:     lessonType,
		Source:   "SPEC-001",
		Pattern:  "test pattern",
		Severity: "medium",
		Tags:     []string{"test", "golang"},
		Active:   true,
		HitCount: 0,
	}
}

// writeLessonsFile은 JSONL 형식으로 레슨을 파일에 직접 기록하는 헬퍼다.
func writeLessonsFile(t *testing.T, path string, lessons []*Lesson) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("디렉토리 생성 실패: %v", err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("파일 생성 실패: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, l := range lessons {
		if err := enc.Encode(l); err != nil {
			t.Fatalf("레슨 인코딩 실패: %v", err)
		}
	}
}

// --- FileStore.Save 테스트 ---

func TestFileStore_Save_NewLesson(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	lesson := newActiveLesson("lesson-1", "quality_failure")
	if err := store.Save(lesson); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	// 저장 후 파일이 존재해야 한다
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("레슨 파일이 생성되지 않음: %v", err)
	}

	// 저장된 레슨을 Get으로 조회할 수 있어야 한다
	got, err := store.Get("lesson-1")
	if err != nil {
		t.Fatalf("Get 실패: %v", err)
	}
	if got.ID != "lesson-1" {
		t.Errorf("ID 불일치: got %q, want %q", got.ID, "lesson-1")
	}
}

func TestFileStore_Save_UpdateExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	// 최초 저장
	lesson := newActiveLesson("lesson-dup", "spec_completion")
	if err := store.Save(lesson); err != nil {
		t.Fatalf("최초 Save 실패: %v", err)
	}

	// 동일 ID로 패턴 변경 후 재저장
	lesson.Pattern = "updated pattern"
	lesson.HitCount = 5
	if err := store.Save(lesson); err != nil {
		t.Fatalf("업데이트 Save 실패: %v", err)
	}

	// 파일에 레슨이 1개만 있어야 한다 (중복 없음)
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count 실패: %v", err)
	}
	if count != 1 {
		t.Errorf("중복 저장됨: got count %d, want 1", count)
	}

	// 업데이트된 값이 반영되어야 한다
	got, err := store.Get("lesson-dup")
	if err != nil {
		t.Fatalf("Get 실패: %v", err)
	}
	if got.Pattern != "updated pattern" {
		t.Errorf("Pattern 미반영: got %q, want %q", got.Pattern, "updated pattern")
	}
	if got.HitCount != 5 {
		t.Errorf("HitCount 미반영: got %d, want 5", got.HitCount)
	}
}

func TestFileStore_Save_AutoArchiveWhenOverMax(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	maxActive := 3
	store := NewFileStore(path, "", maxActive)

	// maxActive+1개 저장 → 가장 오래된 것이 아카이브되어야 한다
	for i := 0; i < maxActive+1; i++ {
		l := newActiveLesson(fmt.Sprintf("lesson-%d", i), "quality_failure")
		// UpdatedAt을 명시적으로 설정해 순서 보장
		l.UpdatedAt = time.Now().Add(time.Duration(i) * time.Second)
		if err := store.Save(l); err != nil {
			t.Fatalf("Save(%d) 실패: %v", i, err)
		}
	}

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count 실패: %v", err)
	}
	if count > maxActive {
		t.Errorf("maxActive 초과: got active count %d, want <= %d", count, maxActive)
	}
}

func TestFileStore_Save_SetsCreatedAt(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	// CreatedAt이 zero인 경우 자동 설정되어야 한다
	before := time.Now()
	lesson := newActiveLesson("lesson-ts", "quality_failure")
	lesson.CreatedAt = time.Time{} // zero로 초기화
	if err := store.Save(lesson); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}
	after := time.Now()

	got, err := store.Get("lesson-ts")
	if err != nil {
		t.Fatalf("Get 실패: %v", err)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt이 설정되지 않음")
	}
	if got.CreatedAt.Before(before) || got.CreatedAt.After(after) {
		t.Errorf("CreatedAt 범위 벗어남: %v", got.CreatedAt)
	}
}

// --- FileStore.List 테스트 ---

func TestFileStore_List_FilterByType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	if err := store.Save(newActiveLesson("l1", "quality_failure")); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(newActiveLesson("l2", "spec_completion")); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(newActiveLesson("l3", "quality_failure")); err != nil {
		t.Fatal(err)
	}

	results, err := store.List(LessonFilter{Type: "quality_failure"})
	if err != nil {
		t.Fatalf("List 실패: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("타입 필터 결과 수 불일치: got %d, want 2", len(results))
	}
	for _, r := range results {
		if r.Type != "quality_failure" {
			t.Errorf("타입 필터 오류: got %q", r.Type)
		}
	}
}

func TestFileStore_List_FilterByTags(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	l1 := newActiveLesson("t1", "quality_failure")
	l1.Tags = []string{"security", "injection"}
	l2 := newActiveLesson("t2", "quality_failure")
	l2.Tags = []string{"performance"}
	l3 := newActiveLesson("t3", "quality_failure")
	l3.Tags = []string{"security", "xss"}

	for _, l := range []*Lesson{l1, l2, l3} {
		if err := store.Save(l); err != nil {
			t.Fatal(err)
		}
	}

	// "security" 태그가 있는 레슨만 조회
	results, err := store.List(LessonFilter{Tags: []string{"security"}})
	if err != nil {
		t.Fatalf("List 실패: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("태그 필터 결과 수 불일치: got %d, want 2", len(results))
	}
}

func TestFileStore_List_FilterByActiveStatus(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	active := newActiveLesson("active-1", "quality_failure")
	inactive := newActiveLesson("inactive-1", "quality_failure")
	inactive.Active = false

	if err := store.Save(active); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(inactive); err != nil {
		t.Fatal(err)
	}

	trueVal := true
	falseVal := false

	tests := []struct {
		name       string
		filter     LessonFilter
		wantCount  int
	}{
		{"active만 조회", LessonFilter{Active: &trueVal}, 1},
		{"inactive만 조회", LessonFilter{Active: &falseVal}, 1},
		{"전체 조회(nil)", LessonFilter{}, 2},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			results, err := store.List(tc.filter)
			if err != nil {
				t.Fatalf("List 실패: %v", err)
			}
			if len(results) != tc.wantCount {
				t.Errorf("%s: got %d, want %d", tc.name, len(results), tc.wantCount)
			}
		})
	}
}

func TestFileStore_List_Limit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	for i := 0; i < 5; i++ {
		l := newActiveLesson(fmt.Sprintf("lim-%d", i), "quality_failure")
		l.HitCount = i
		if err := store.Save(l); err != nil {
			t.Fatal(err)
		}
	}

	results, err := store.List(LessonFilter{Limit: 3})
	if err != nil {
		t.Fatalf("List 실패: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Limit 결과 수 불일치: got %d, want 3", len(results))
	}
}

func TestFileStore_List_GlobalAndLocalMerge(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localPath := filepath.Join(dir, "local", "lessons.jsonl")
	globalPath := filepath.Join(dir, "global", "lessons.jsonl")

	// 로컬과 글로벌 파일을 사전 준비
	localLesson := newActiveLesson("local-1", "quality_failure")
	globalLesson := newActiveLesson("global-1", "spec_completion")
	writeLessonsFile(t, localPath, []*Lesson{localLesson})
	writeLessonsFile(t, globalPath, []*Lesson{globalLesson})

	store := NewFileStore(localPath, globalPath, 50)
	results, err := store.List(LessonFilter{})
	if err != nil {
		t.Fatalf("List 실패: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("로컬+글로벌 병합 결과 수 불일치: got %d, want 2", len(results))
	}
}

func TestFileStore_List_SortByHitCountDesc(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	for i, hits := range []int{1, 5, 3} {
		l := newActiveLesson(fmt.Sprintf("sort-%d", i), "quality_failure")
		l.HitCount = hits
		if err := store.Save(l); err != nil {
			t.Fatal(err)
		}
	}

	results, err := store.List(LessonFilter{})
	if err != nil {
		t.Fatalf("List 실패: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("결과 부족: got %d", len(results))
	}
	// 첫 번째 결과의 HitCount가 가장 높아야 한다
	for i := 1; i < len(results); i++ {
		if results[i-1].HitCount < results[i].HitCount {
			t.Errorf("정렬 오류: results[%d].HitCount=%d < results[%d].HitCount=%d",
				i-1, results[i-1].HitCount, i, results[i].HitCount)
		}
	}
}

// --- FileStore.Get 테스트 ---

func TestFileStore_Get_Found(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	lesson := newActiveLesson("get-1", "quality_failure")
	if err := store.Save(lesson); err != nil {
		t.Fatal(err)
	}

	got, err := store.Get("get-1")
	if err != nil {
		t.Fatalf("Get 실패: %v", err)
	}
	if got.ID != "get-1" {
		t.Errorf("ID 불일치: got %q, want %q", got.ID, "get-1")
	}
}

func TestFileStore_Get_NotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	_, err := store.Get("nonexistent-id")
	if err == nil {
		t.Error("존재하지 않는 ID에 대해 오류가 반환되어야 함")
	}
}

// --- FileStore.Archive 테스트 ---

func TestFileStore_Archive_MarksInactive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	lesson := newActiveLesson("arch-1", "quality_failure")
	if err := store.Save(lesson); err != nil {
		t.Fatal(err)
	}

	if err := store.Archive("arch-1"); err != nil {
		t.Fatalf("Archive 실패: %v", err)
	}

	got, err := store.Get("arch-1")
	if err != nil {
		t.Fatalf("Archive 후 Get 실패: %v", err)
	}
	if got.Active {
		t.Error("Archive 후 Active가 false여야 함")
	}
}

func TestFileStore_Archive_NotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	err := store.Archive("ghost-id")
	if err == nil {
		t.Error("존재하지 않는 ID에 대해 오류가 반환되어야 함")
	}
}

// --- 원자적 쓰기 테스트 ---

func TestFileStore_AtomicWrite_TempFilePattern(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	lesson := newActiveLesson("atomic-1", "quality_failure")
	if err := store.Save(lesson); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	// 원자적 쓰기 후 임시 파일(.tmp)이 남아 있으면 안 된다
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("임시 파일(.tmp)이 정리되지 않음")
	}

	// 최종 파일은 존재해야 한다
	if _, err := os.Stat(path); err != nil {
		t.Errorf("최종 파일이 존재해야 함: %v", err)
	}
}

// --- 동시성 테스트 ---

func TestFileStore_ConcurrentSave(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 100)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			l := newActiveLesson(fmt.Sprintf("concurrent-%d", i), "quality_failure")
			if err := store.Save(l); err != nil {
				// 동시성 환경에서 오류가 없어야 한다 (레이스 컨디션 방지 확인)
				t.Errorf("goroutine %d Save 실패: %v", i, err)
			}
		}()
	}

	wg.Wait()

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count 실패: %v", err)
	}
	if count != goroutines {
		t.Errorf("동시 저장 결과 수 불일치: got %d, want %d", count, goroutines)
	}
}

// --- matchesFilter 테스트 ---

func TestMatchesFilter(t *testing.T) {
	t.Parallel()

	trueVal := true
	falseVal := false

	tests := []struct {
		name   string
		lesson *Lesson
		filter LessonFilter
		want   bool
	}{
		{
			name: "타입 일치",
			lesson: &Lesson{Type: "quality_failure", Active: true, Tags: []string{}, HitCount: 0},
			filter: LessonFilter{Type: "quality_failure"},
			want:   true,
		},
		{
			name: "타입 불일치",
			lesson: &Lesson{Type: "spec_completion", Active: true, Tags: []string{}, HitCount: 0},
			filter: LessonFilter{Type: "quality_failure"},
			want:   false,
		},
		{
			name: "Active 필터 일치",
			lesson: &Lesson{Type: "", Active: true, Tags: []string{}, HitCount: 0},
			filter: LessonFilter{Active: &trueVal},
			want:   true,
		},
		{
			name: "Active 필터 불일치",
			lesson: &Lesson{Type: "", Active: false, Tags: []string{}, HitCount: 0},
			filter: LessonFilter{Active: &trueVal},
			want:   false,
		},
		{
			name: "Inactive 필터 일치",
			lesson: &Lesson{Type: "", Active: false, Tags: []string{}, HitCount: 0},
			filter: LessonFilter{Active: &falseVal},
			want:   true,
		},
		{
			name: "MinHits 만족",
			lesson: &Lesson{Type: "", Active: true, Tags: []string{}, HitCount: 5},
			filter: LessonFilter{MinHits: 3},
			want:   true,
		},
		{
			name: "MinHits 미만",
			lesson: &Lesson{Type: "", Active: true, Tags: []string{}, HitCount: 2},
			filter: LessonFilter{MinHits: 3},
			want:   false,
		},
		{
			name: "태그 일치",
			lesson: &Lesson{Type: "", Active: true, Tags: []string{"security", "injection"}, HitCount: 0},
			filter: LessonFilter{Tags: []string{"injection"}},
			want:   true,
		},
		{
			name: "태그 불일치",
			lesson: &Lesson{Type: "", Active: true, Tags: []string{"performance"}, HitCount: 0},
			filter: LessonFilter{Tags: []string{"security"}},
			want:   false,
		},
		{
			name: "빈 필터는 모두 통과",
			lesson: &Lesson{Type: "quality_failure", Active: false, Tags: []string{"x"}, HitCount: 0},
			filter: LessonFilter{},
			want:   true,
		},
	}

	store := &FileStore{}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := store.matchesFilter(tc.lesson, tc.filter)
			if got != tc.want {
				t.Errorf("matchesFilter() = %v, want %v", got, tc.want)
			}
		})
	}
}

// --- archiveOldest 테스트 ---

func TestArchiveOldest(t *testing.T) {
	t.Parallel()

	now := time.Now()
	lessons := []*Lesson{
		{ID: "old-1", Active: true, UpdatedAt: now.Add(-3 * time.Hour)},
		{ID: "old-2", Active: true, UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "new-1", Active: true, UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "inactive-1", Active: false, UpdatedAt: now.Add(-4 * time.Hour)},
	}

	store := &FileStore{}
	// 활성 레슨 중 가장 오래된 1개를 아카이브
	store.archiveOldest(lessons, 1)

	// "old-1"이 아카이브되어야 한다 (UpdatedAt이 가장 오래됨)
	for _, l := range lessons {
		if l.ID == "old-1" && l.Active {
			t.Error("old-1이 아카이브되지 않음")
		}
		if l.ID == "old-2" && !l.Active {
			t.Error("old-2가 잘못 아카이브됨")
		}
		if l.ID == "new-1" && !l.Active {
			t.Error("new-1이 잘못 아카이브됨")
		}
	}
}

// --- Count 테스트 ---

func TestFileStore_Count_EmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count 실패: %v", err)
	}
	if count != 0 {
		t.Errorf("빈 스토어 Count: got %d, want 0", count)
	}
}

func TestFileStore_Count_OnlyActive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lessons.jsonl")
	store := NewFileStore(path, "", 50)

	active := newActiveLesson("count-active", "quality_failure")
	inactive := newActiveLesson("count-inactive", "quality_failure")
	inactive.Active = false

	if err := store.Save(active); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(inactive); err != nil {
		t.Fatal(err)
	}

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count 실패: %v", err)
	}
	if count != 1 {
		t.Errorf("Count는 활성 레슨만 집계해야 함: got %d, want 1", count)
	}
}
