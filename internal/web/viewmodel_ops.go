// 개요 · 칸반 · SPEC · 모니터 뷰모델.
// internal/spec · session · goal · verify · kanban 을 읽기만 한다 — 쓰기 없음.
//
// 이 파일의 규율: 모르는 것을 아는 것처럼 쓰지 않는다.
//   - 세션 활성은 PID 생존을 확인한 것만 StateLive 로 올린다.
//   - 단계 상태는 하트비트 추정이며 StageEstimated=true 로 표시된다.
//   - 역할/모델/effort/컨텍스트는 기록이 없으면 빈 값으로 남기고 화면이 "—"로 그린다.
//
// 로더가 상태 파일을 직접 읽는 이유: session.QueryActiveWork 등 일부 패키지
// 진입점은 프로세스 전역 기본 경로에 묶여 있어 콘솔이 설정한 ProjectRoot 를
// 존중하지 않는다. 여기서는 프로젝트 루트를 명시적으로 받아 그 아래만 읽는다.
package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/spec"
	"github.com/modu-ai/moai-adk/internal/verify"
)

const (
	StateLive  = "live"  // PID 생존 확인
	StateStale = "stale" // 기록은 있으나 하트비트가 낡음 / PID 미확인
	StateIdle  = "idle"  // 세션 자체가 없음 = 체인 결함

	StageDone    = "done"
	StageActive  = "active"
	StageWait    = "wait"
	StageBlocked = "blocked"

	staleAfter = 10 * time.Minute // 이보다 오래된 하트비트는 활성으로 단정하지 않는다

	maxOverviewRows = 8
	maxVerifyRows   = 10
)

// ChainRoles 는 kanban-dispatch.md 의 역할 순서를 그대로 따른다.
var ChainRoles = []string{"lead", "plan", "run", "sync"}

// KanbanRecord 는 디스크에 있는 칸반 세션 기록이다.
type KanbanRecord = kanban.Record

// StatVM 의 Note 는 영어 baseline("4 in-progress")이고 NoteKey 가 실제 표시
// 언어를 바꾼다. 개수가 문장 안에 박힌 부제는 평평한 키 하나로 담을 수 없어
// NoteParams 가 {0} 자리에 숫자를 공급한다 — 키는 개수와 무관하게 하나다.
type StatVM struct {
	Label, Value, Note string
	NoteKey            string
	NoteParams         string
}

type AttentionVM struct {
	Icon      string // alert | clock
	Source    string
	Text      string
	// Role 는 이 행이 칸반 미기동 역할 알림일 때 그 역할 이름이다. 비어 있으면
	// Text 를 그대로 렌더하고, 차 있으면 Text 대신 Role + i18n 키 조각으로
	// 렌더한다 — 같은 안내가 체인 띠와 여기 두 경로로 나오는데 번역 키를
	// 공유해야 한쪽만 한국어가 되는 일이 없다.
	Role      string
	Badge     string
	BadgeKind string // danger | outline
	Href      string
}

type SessionVM struct {
	ID        string
	SpecID    string
	Backend   string // claude | glm | "" (미기록)
	State     string
	Heartbeat string
	Cwd       string
}

type RoleVM struct {
	Role           string
	Session        string
	Backend        string
	Model          string // 미기록 — 3단계에서 채워진다
	Effort         string // 미기록 — 3단계
	ContextPct     int    // -1 = 기록 없음
	State          string
	Stage          string
	StageEstimated bool
	Heartbeat      string
}

type ChainVM struct {
	Present  bool
	CardID   string
	IdleRole string // 가장 앞선 미기동 역할 — 체인이 멈춘 지점
	Roles    []RoleVM
}

type SpecRowVM struct {
	ID, Title, Status, Tier, Era, Updated, Drift, Session string
}

type PipeColumnVM struct {
	ID     string // plan | run | sync | done
	Status string // draft | in-progress | implemented | completed
	Cards  []SpecRowVM
}

type OverviewVM struct {
	Stats      []StatVM
	Chain      ChainVM
	InProgress []SpecRowVM
	Attention  []AttentionVM
	Sessions   []SessionVM
}

type KanbanVM struct {
	CardID   string
	IdleRole string
	Roles    []RoleVM
	Columns  []PipeColumnVM
	Total    int
}

type FilterVM struct {
	Label  string
	Count  int
	Href   string
	Active bool
}

type SortVM struct {
	Label  string
	Href   string
	Active bool
}

type FindingVM struct{ Severity, Message, File string }

type SpecDetailVM struct {
	ID, Title, Status, Tier, Era, SHA, Path, VerifyNote string
	Docs                                                []string
	Findings                                            []FindingVM
	Verify                                              []bool
}

type SpecListVM struct {
	Query      string
	SelectedID string
	Filters    []FilterVM
	Sorts      []SortVM
	Rows       []SpecRowVM
	Detail     *SpecDetailVM

	// CloseDebt 는 구현은 끝났지만(status: implemented) lifecycle 이 completed 로
	// 닫히지 않은 SPEC 이다. 필터와 무관하게 카탈로그 전체에서 모은다 — 지금 보고
	// 있는 필터 결과가 아니라 "아직 닫지 않은 것 전부"가 알고 싶은 값이기 때문이다.
	CloseDebt []SpecRowVM

	// MustFix 는 조치 명령을 가진 MUST-FIX drift 다. 명령은 화면이 실행하지 않는다
	// — 복사만 하고 실행은 사람이 자기 터미널에서 한다.
	MustFix []MustFixVM
}

// MustFixVM 은 조치 명령을 동반한 MUST-FIX drift 하나다.
type MustFixVM struct {
	SpecID      string
	FindingType string
	Remediation string
}

type GoalVM struct {
	Session, Condition, Verdict string
	Turns, TurnPct              int
	Stalled                     bool
}

type VerifyVM struct {
	Key, When string
	OK        bool
	History   []bool
}

type EpicVM struct {
	Prefix, Progress string
	Pct              int
}

type MonitorVM struct {
	Sessions   []SessionVM
	Goals      []GoalVM
	Verify     []VerifyVM
	Epics      []EpicVM
	VerifyKeys int
	Cwd        string
}

// ─── 상태 판정 ───────────────────────────────────────────────────────────────

// sessionState 는 하트비트와 PID 생존으로만 판정한다.
// 레지스트리에 항목이 있다는 사실만으로 활성이라고 쓰지 않는다 — 레지스트리에는
// 종료된 프로세스의 항목이 남는다.
func sessionState(lastHeartbeat time.Time, pid int, now time.Time) string {
	if pid > 0 && processAlive(pid) && now.Sub(lastHeartbeat) <= staleAfter {
		return StateLive
	}
	return StateStale
}

func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscallZero) == nil
}

// roleOf 는 칸반 기록에서 역할을 읽는다. 기록이 없으면 빈 문자열을 돌려주고,
// 화면은 그 칸을 미기동으로 정직하게 그린다 — 그럴듯한 값을 채우지 않는다.
func roleOf(r KanbanRecord) string { return strings.ToLower(strings.TrimSpace(r.Role)) }

// buildChain 은 칸반 세션 기록과 활성 세션을 역할 5칸에 배치한다.
func buildChain(records []KanbanRecord, sessions map[string]SessionVM, cardID string) ChainVM {
	byRole := map[string]KanbanRecord{}
	for _, r := range records {
		if role := roleOf(r); role != "" {
			byRole[role] = r
		}
	}
	out := ChainVM{Present: len(records) > 0, CardID: cardID}
	for _, role := range ChainRoles {
		rec, ok := byRole[role]
		if !ok {
			out.Roles = append(out.Roles, RoleVM{Role: role, State: StateIdle, Stage: StageBlocked, ContextPct: -1})
			if out.IdleRole == "" {
				out.IdleRole = role
			}
			continue
		}
		s := sessions[rec.SessionID]
		if s.State == "" {
			s.State = StateStale
		}
		stage, estimated := estimateStage(s)
		out.Roles = append(out.Roles, RoleVM{
			Role:           role,
			Session:        rec.SessionID,
			Backend:        rec.Backend,
			Model:          "", // 3단계: Record 에 모델 스냅샷이 추가되면 채운다
			Effort:         "", // 3단계
			ContextPct:     -1, // 3단계: context-usage/<session-id>.json 분리 후
			State:          s.State,
			Stage:          stage,
			StageEstimated: estimated,
			Heartbeat:      s.Heartbeat,
		})
	}
	return out
}

// estimateStage — 하트비트 추정. 전이 기록이 생기면 estimated=false 로 바뀐다.
func estimateStage(s SessionVM) (string, bool) {
	switch s.State {
	case StateLive:
		return StageActive, true
	case StateStale:
		return StageWait, true
	default:
		return StageBlocked, false // 세션 없음은 추정이 아니라 사실이다
	}
}

// pipelineColumns — 뷰 B. SPEC status 가 정직하게 말해주는 4단계만 그린다.
// backlog 는 대응 status 가 없어 만들지 않는다.
func pipelineColumns(specs []SpecRowVM) []PipeColumnVM {
	cols := []PipeColumnVM{
		{ID: "plan", Status: "draft"},
		{ID: "run", Status: "in-progress"},
		{ID: "sync", Status: "implemented"},
		{ID: "done", Status: "completed"},
	}
	idx := map[string]int{"draft": 0, "planned": 0, "in-progress": 1, "implemented": 2, "completed": 3}
	for _, s := range specs {
		i, ok := idx[s.Status]
		if !ok {
			continue // superseded / archived / rejected 는 보드 밖 — 필터로만 본다
		}
		cols[i].Cards = append(cols[i].Cards, s)
	}
	for i := range cols {
		col := cols[i]
		sort.SliceStable(col.Cards, func(a, b int) bool { return col.Cards[a].Updated > col.Cards[b].Updated })
	}
	return cols
}

func humanSince(t time.Time, now time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return itoa(int(d.Minutes())) + "m"
	case d < 24*time.Hour:
		return itoa(int(d.Hours())) + "h"
	default:
		return itoa(int(d.Hours()/24)) + "d"
	}
}

// ─── 로더 (전부 읽기 전용) ───────────────────────────────────────────────────

// loadSpecRows 는 SPEC 목록과 SPEC별 드리프트 findings 를 함께 만든다.
// board.go 와 같은 원천(spec.ListDocs + spec.Audit)을 쓰고 git 경로는 건드리지 않는다.
// closeDebtShown 은 종료 부채 패널이 한 번에 보여 주는 최대 줄 수다. 카탈로그가
// 큰 저장소에서는 이 목록이 수백 줄까지 자라, 패널을 접지 않으면 그 아래 표가
// 화면 밖으로 밀려난다. 잘린 사실과 전체 수는 패널이 함께 적는다.
const closeDebtShown = 10

// closeDebtRows 는 status: implemented 인 SPEC 을 최근 갱신 순으로 모은다.
// 구현이 끝났다는 표시는 있는데 lifecycle 을 닫은 표시는 없는 상태다.
func closeDebtRows(rows []SpecRowVM) []SpecRowVM {
	var out []SpecRowVM
	for _, r := range rows {
		if r.Status == "implemented" {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Updated > out[j].Updated })
	return out
}

// boundedRows 는 패널에 실제로 그릴 앞부분만 잘라 준다.
func boundedRows(rows []SpecRowVM) []SpecRowVM {
	if len(rows) > closeDebtShown {
		return rows[:closeDebtShown]
	}
	return rows
}

// mustFixFindings 는 MUST-FIX drift 를 SPEC 행 순서대로 펼친다. loadSpecRows 가
// 이미 채운 findings 맵을 재사용한다 — 감사 스캔을 두 번 돌리지 않는다.
//
// FindingVM 은 감사 결과를 담을 때 Message 에 Remediation 을, File 에
// FindingType 을 실어 둔다 (loadSpecRows 참조). 이름과 내용이 어긋나 있어
// 여기서 제자리 이름으로 되돌려 담는다.
func mustFixFindings(rows []SpecRowVM, findings map[string][]FindingVM) []MustFixVM {
	var out []MustFixVM
	for _, r := range rows {
		for _, f := range findings[r.ID] {
			if f.Severity != "MUST-FIX" {
				continue
			}
			out = append(out, MustFixVM{
				SpecID:      r.ID,
				FindingType: f.File,
				Remediation: f.Message,
			})
		}
	}
	return out
}

func loadSpecRows(root string) ([]SpecRowVM, map[string][]FindingVM, error) {
	records, err := spec.ListDocs(root)
	if err != nil {
		return nil, nil, err
	}
	findings := map[string][]FindingVM{}
	if res, err := spec.Audit(spec.AuditOptions{BaseDir: root}); err == nil && res != nil {
		for _, f := range res.DriftFindings {
			findings[f.SpecID] = append(findings[f.SpecID], FindingVM{
				Severity: f.Severity,
				Message:  f.Remediation,
				File:     f.FindingType,
			})
		}
	}

	rows := make([]SpecRowVM, 0, len(records))
	for _, rec := range records {
		id := rec.Frontmatter.ID
		if id == "" {
			id = filepath.Base(filepath.Dir(rec.Path))
		}
		row := SpecRowVM{
			ID:      id,
			Title:   rec.Frontmatter.Title,
			Status:  rec.Frontmatter.Status,
			Tier:    rec.Frontmatter.Tier,
			Era:     rec.Frontmatter.Era,
			Updated: rec.Frontmatter.Updated,
		}
		if fs := findings[id]; len(fs) > 0 {
			row.Drift = fs[0].Severity
		}
		rows = append(rows, row)
	}
	return rows, findings, nil
}

// loadSessions 는 활성 세션 레지스트리를 프로젝트 루트 아래에서 직접 읽는다.
func loadSessions(root string, now time.Time) ([]SessionVM, map[string]SessionVM) {
	path := filepath.Join(root, ".moai", "state", "active-sessions.json")
	data, err := os.ReadFile(path) // #nosec G304 — 프로젝트 루트 하위 고정 경로
	if err != nil {
		return nil, map[string]SessionVM{}
	}
	var entries []session.Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, map[string]SessionVM{}
	}

	out := make([]SessionVM, 0, len(entries))
	byID := make(map[string]SessionVM, len(entries))
	for _, e := range entries {
		vm := SessionVM{
			ID:        shortID(e.SessionID),
			SpecID:    e.SpecID,
			State:     sessionState(e.LastHeartbeat, e.PID, now),
			Heartbeat: humanSince(e.LastHeartbeat, now),
			Cwd:       e.CWD,
		}
		out = append(out, vm)
		byID[e.SessionID] = vm
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].ID < out[b].ID })
	return out, byID
}

// loadKanbanRecords 는 .moai/state/kanban/*.json 을 읽는다.
func loadKanbanRecords(root string) []KanbanRecord {
	dir := filepath.Join(root, ".moai", "state", "kanban")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []KanbanRecord
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name())) // #nosec G304
		if err != nil {
			continue
		}
		var rec KanbanRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			continue
		}
		out = append(out, rec)
	}
	return out
}

// loadGoals 는 무장된 목표를 세션별로 읽는다. goal.LoadGoal 이 프로젝트 루트를
// 받으므로 그대로 위임하고, 세션 목록만 디렉터리에서 뽑는다.
func loadGoals(root string) []GoalVM {
	dir := filepath.Join(root, ".moai", "state", "goal")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []GoalVM
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".verdict.json") {
			continue
		}
		sessionID := strings.TrimSuffix(name, ".json")
		g, err := goal.LoadGoal(root, sessionID)
		if err != nil || g == nil {
			continue
		}
		ceiling := g.Ceiling.MaxTurns
		pct := 0
		if ceiling > 0 {
			pct = clampPct(g.TurnsUsed * 100 / ceiling)
		}
		out = append(out, GoalVM{
			Session:   shortID(g.SessionID),
			Condition: g.Goal,
			Verdict:   string(g.Status),
			Turns:     g.TurnsUsed,
			TurnPct:   pct,
			Stalled:   ceiling > 0 && g.TurnsUsed >= ceiling,
		})
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Session < out[b].Session })
	return out
}

// loadVerify 는 검증 스냅샷을 읽어 최근 결과와 짧은 추이를 만든다.
func loadVerify(root string, limit int) ([]VerifyVM, int) {
	dir := filepath.Join(root, ".moai", "state", "verify", "snapshots")
	entries, err := os.ReadDir(dir)
	if err != nil {
		// 구버전 배치(스냅샷 하위 디렉터리 없이 평면)도 읽어본다.
		dir = filepath.Join(root, ".moai", "state", "verify")
		entries, err = os.ReadDir(dir)
		if err != nil {
			return nil, 0
		}
	}

	type keyed struct {
		vm VerifyVM
		at time.Time
	}
	var all []keyed
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		key := strings.TrimSuffix(e.Name(), ".json")
		snap, err := verify.Load(root, key)
		if err != nil || snap == nil {
			continue
		}
		vm := VerifyVM{Key: key, When: humanSince(snap.RecordedAt, time.Now()), OK: true}
		for _, c := range snap.Checks {
			ok := c.ExitCode == 0
			vm.History = append(vm.History, ok)
			if !ok {
				vm.OK = false
			}
		}
		if n := len(vm.History); n > 8 {
			vm.History = vm.History[n-8:]
		}
		all = append(all, keyed{vm: vm, at: snap.RecordedAt})
	}
	total := len(all)
	sort.SliceStable(all, func(a, b int) bool { return all[a].at.After(all[b].at) })
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	out := make([]VerifyVM, 0, len(all))
	for _, k := range all {
		out = append(out, k.vm)
	}
	return out, total
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// ─── 화면 조립 ───────────────────────────────────────────────────────────────

func (a *app) buildOverview(now time.Time) (OverviewVM, error) {
	root := a.cfg.ProjectRoot
	rows, findings, err := loadSpecRows(root)
	if err != nil {
		return OverviewVM{}, err
	}
	sessions, byID := loadSessions(root, now)
	records := loadKanbanRecords(root)

	var inProgress []SpecRowVM
	mustFix := 0
	for _, r := range rows {
		if r.Status == "in-progress" {
			inProgress = append(inProgress, r)
		}
	}
	for _, fs := range findings {
		for _, f := range fs {
			if strings.HasPrefix(f.Severity, "MUST") {
				mustFix++
			}
		}
	}
	live := 0
	for _, s := range sessions {
		if s.State == StateLive {
			live++
		}
	}
	verifyRows, verifyKeys := loadVerify(root, 1)
	lastVerify := "—"
	if len(verifyRows) > 0 {
		if verifyRows[0].OK {
			lastVerify = "pass"
		} else {
			lastVerify = "fail"
		}
	}

	vm := OverviewVM{
		Stats: []StatVM{
			{Label: "SPEC", Value: itoa(len(rows)), Note: itoa(len(inProgress)) + " in-progress", NoteKey: "statNote.in-progress", NoteParams: itoa(len(inProgress))},
			{Label: "drift", Value: itoa(mustFix), Note: "MUST-FIX", NoteKey: "statNote.must-fix"},
			{Label: "session", Value: itoa(live) + "/" + itoa(len(sessions)), Note: "PID confirmed / registry", NoteKey: "statNote.pid-confirmed-registry"},
			{Label: "verify", Value: lastVerify, Note: itoa(verifyKeys) + " keys", NoteKey: "statNote.keys", NoteParams: itoa(verifyKeys)},
		},
		Chain:    buildChain(records, byID, chainCardID(records)),
		Sessions: sessions,
	}
	if len(inProgress) > maxOverviewRows {
		inProgress = inProgress[:maxOverviewRows]
	}
	vm.InProgress = inProgress
	vm.Attention = buildAttention(rows, findings, vm.Chain)
	return vm, nil
}

// buildAttention 은 사람이 손대야 하는 것만 모은다 — MUST-FIX 드리프트와
// 미기동 역할. 정상 상태를 나열하지 않는다.
func buildAttention(rows []SpecRowVM, findings map[string][]FindingVM, chain ChainVM) []AttentionVM {
	var out []AttentionVM
	if chain.Present && chain.IdleRole != "" {
		out = append(out, AttentionVM{
			Icon:      "alert",
			Source:    "kanban",
			Text:      chain.IdleRole + " session not started — the chain stops here",
			Role:      chain.IdleRole,
			Badge:     "idle",
			BadgeKind: "danger",
			Href:      "/kanban",
		})
	}
	for _, r := range rows {
		for _, f := range findings[r.ID] {
			if !strings.HasPrefix(f.Severity, "MUST") {
				continue
			}
			out = append(out, AttentionVM{
				Icon:      "alert",
				Source:    r.ID,
				Text:      f.Message,
				Badge:     f.Severity,
				BadgeKind: "danger",
				Href:      "/specs?id=" + r.ID,
			})
			if len(out) >= maxOverviewRows {
				return out
			}
		}
	}
	return out
}

// chainCardID 는 기록된 SPEC 이 있으면 그것을 카드로 본다. 없으면 빈 문자열 —
// plan 단계부터 시작한 체인은 아직 카드 식별자가 없다.
func chainCardID(records []KanbanRecord) string {
	for _, r := range records {
		if r.SpecID != "" {
			return r.SpecID
		}
	}
	return ""
}

func (a *app) buildKanban(now time.Time) (KanbanVM, error) {
	root := a.cfg.ProjectRoot
	rows, _, err := loadSpecRows(root)
	if err != nil {
		return KanbanVM{}, err
	}
	_, byID := loadSessions(root, now)
	records := loadKanbanRecords(root)
	chain := buildChain(records, byID, chainCardID(records))

	return KanbanVM{
		CardID:   chain.CardID,
		IdleRole: chain.IdleRole,
		Roles:    chain.Roles,
		Columns:  pipelineColumns(rows),
		Total:    len(rows),
	}, nil
}

func (a *app) buildMonitor(now time.Time) (MonitorVM, error) {
	root := a.cfg.ProjectRoot
	sessions, _ := loadSessions(root, now)
	verifyRows, verifyKeys := loadVerify(root, maxVerifyRows)
	return MonitorVM{
		Sessions:   sessions,
		Goals:      loadGoals(root),
		Verify:     verifyRows,
		VerifyKeys: verifyKeys,
		Cwd:        root,
	}, nil
}

// buildSpecList 는 검색어·상태 필터·선택 항목을 반영한 SPEC 목록을 만든다.
func (a *app) buildSpecList(query, status, selected string) (SpecListVM, error) {
	rows, findings, err := loadSpecRows(a.cfg.ProjectRoot)
	if err != nil {
		return SpecListVM{}, err
	}

	counts := map[string]int{}
	for _, r := range rows {
		counts[r.Status]++
	}

	filtered := make([]SpecRowVM, 0, len(rows))
	q := strings.ToLower(strings.TrimSpace(query))
	for _, r := range rows {
		if status != "" && r.Status != status {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(r.ID), q) && !strings.Contains(strings.ToLower(r.Title), q) {
			continue
		}
		filtered = append(filtered, r)
	}
	sort.SliceStable(filtered, func(a, b int) bool { return filtered[a].Updated > filtered[b].Updated })

	vm := SpecListVM{
		Query:      query,
		SelectedID: selected,
		Rows:       filtered,
		Filters:    buildFilters(counts, status, query, len(rows)),
		CloseDebt:  closeDebtRows(rows),
		MustFix:    mustFixFindings(rows, findings),
	}
	if selected != "" {
		for _, r := range rows {
			if r.ID != selected {
				continue
			}
			d := SpecDetailVM{
				ID: r.ID, Title: r.Title, Status: r.Status, Tier: r.Tier, Era: r.Era,
				Path:     filepath.Join(".moai", "specs", r.ID),
				Docs:     []string{"spec.md", "plan.md", "acceptance.md", "progress.md"},
				Findings: findings[r.ID],
			}
			vm.Detail = &d
			break
		}
	}
	return vm, nil
}

func buildFilters(counts map[string]int, active, query string, total int) []FilterVM {
	href := func(status string) string {
		v := "/specs?"
		if query != "" {
			v += "q=" + query + "&"
		}
		if status == "" {
			return strings.TrimSuffix(v, "&")
		}
		return v + "status=" + status
	}
	out := []FilterVM{{Label: "all", Count: total, Href: href(""), Active: active == ""}}
	for _, s := range []string{"draft", "in-progress", "implemented", "completed"} {
		out = append(out, FilterVM{Label: s, Count: counts[s], Href: href(s), Active: active == s})
	}
	return out
}
