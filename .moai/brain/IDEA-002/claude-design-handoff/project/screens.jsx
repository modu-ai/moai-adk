// MoAI-ADK v3 — Screen modules
const { useState: useS, useEffect: useE, useMemo: useM, useRef: useR } = React;

// ───────── DASHBOARD ─────────
function DashboardScreen({ onOpenSpec, onOpenAgent }) {
  const { SPECS, AGENTS, SPARKS, LOG_SEED, TRUST } = window.MoaiData;
  const liveSpecs = SPECS.filter(s => s.status === "in-progress");
  const blocked = SPECS.filter(s => s.status === "blocked");

  return (
    <div className="fade-up">
      <div className="page-head">
        <div>
          <h1>오버뷰</h1>
          <div className="sub">SPEC 12개 · 에이전트 16개 · 후크 12개 · TRUST 5 게이트 활성</div>
        </div>
        <div style={{ display:"flex", gap:10, alignItems:"center" }}>
          <Status kind="live">harness=standard</Status>
          <button className="btn outline"><Icon name="refresh-cw" size={14} />동기화</button>
          <button className="btn primary" onClick={() => window.dispatchEvent(new CustomEvent("openCreateSpec"))}><Icon name="plus" size={14} />새 SPEC</button>
        </div>
      </div>

      <div className="stat-grid">
        <Stat label="활성 RUN" value="4" delta="+2 (24h)" spark={SPARKS.runs} />
        <Stat label="평균 커버리지" value="89%" delta="+1.2 pp" spark={SPARKS.coverage} sparkColor="var(--color-success)" />
        <Stat label="실패율" value="0.3%" delta="-2.1 pp" deltaKind="pos" spark={SPARKS.failures} sparkColor="var(--color-danger)" />
        <Stat label="누적 토큰 비용" value="$3.4" delta="+0.2" deltaKind="flat" spark={SPARKS.cost} sparkColor="var(--color-warning)" />
      </div>

      <div className="grid-2">
        <div className="card">
          <div className="card-head">
            <h3>활성 SPEC ({liveSpecs.length})</h3>
            <button className="btn ghost sm" onClick={() => window.dispatchEvent(new CustomEvent("nav", { detail: "specs" }))}>모두 보기 <Icon name="arrow-right" size={12} /></button>
          </div>
          {liveSpecs.slice(0,5).map(s => (
            <div key={s.id} className="config-row" style={{cursor:"pointer"}} onClick={()=>onOpenSpec(s)}>
              <div style={{flex:1, minWidth:0}}>
                <div style={{display:"flex", gap:8, alignItems:"center", marginBottom:4}}>
                  <span className="mono" style={{fontSize:11.5, color:"var(--color-primary)", fontWeight:600}}>{s.id}</span>
                  {priorityChip(s.priority)}
                </div>
                <div style={{fontSize:13.5, fontWeight:600, letterSpacing:"-0.02em", overflow:"hidden", textOverflow:"ellipsis", whiteSpace:"nowrap"}}>{s.title}</div>
              </div>
              <div style={{minWidth:120, textAlign:"right"}}>
                <div style={{fontSize:11.5, color:"var(--fg-3)", marginBottom:4, fontFamily:"var(--font-mono)"}}>{s.coverage}%</div>
                <div className={`progress ${s.coverage > 80 ? "success" : s.coverage > 40 ? "" : "warning"}`} style={{width:120}}>
                  <div style={{width:`${s.coverage}%`}} />
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="card">
          <div className="card-head">
            <h3>TRUST 5 게이트</h3>
            <span className="meta">최근 평가 · 12초 전</span>
          </div>
          {TRUST.map((t, i) => (
            <div key={i} className="trust-row" style={{borderBottom: i === TRUST.length-1 ? "none" : ""}}>
              <div className={`trust-letter ${t.status}`}>{t.letter}</div>
              <div style={{flex:1, minWidth:0}}>
                <div style={{display:"flex", gap:8, alignItems:"center"}}>
                  <span style={{fontWeight:700, fontSize:13.5}}>{t.name}</span>
                  <span className="chip" style={{marginLeft:"auto"}}>{t.value}</span>
                </div>
                <div style={{fontSize:12, color:"var(--fg-2)", marginTop:3, lineHeight:1.5}}>{t.note}</div>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="grid-2" style={{marginTop:14}}>
        <LiveLogCard />
        <div className="card">
          <div className="card-head">
            <h3>활성 에이전트</h3>
            <span className="meta">{AGENTS.filter(a => a.status === "live").length} live · {AGENTS.filter(a => a.status === "idle").length} idle</span>
          </div>
          <div style={{display:"flex", flexDirection:"column", gap:8}}>
            {AGENTS.filter(a => a.status === "live" || a.status === "warn").map(a => (
              <div key={a.id} className="config-row" style={{padding:"10px 0", cursor:"pointer"}} onClick={()=>onOpenAgent(a)}>
                <div className={`agent-icon ${a.kind}`} style={{width:32, height:32, fontSize:11}}>
                  {a.kind === "manager" ? "MGR" : a.kind === "expert" ? "EXP" : a.kind === "builder" ? "BLD" : a.kind === "evaluator" ? "EVL" : "RES"}
                </div>
                <div style={{flex:1, minWidth:0}}>
                  <div style={{fontWeight:600, fontSize:13, fontFamily:"var(--font-mono)"}}>{a.id}</div>
                  <div style={{fontSize:11.5, color:"var(--fg-3)"}}>{a.role}</div>
                </div>
                <Status kind={a.status === "live" ? "live" : a.status === "warn" ? "warn" : "idle"}>
                  {a.status === "live" ? "RUNNING" : a.status === "warn" ? "WARN" : "IDLE"}
                </Status>
              </div>
            ))}
          </div>
          {blocked.length > 0 && (
            <div style={{marginTop:14, padding:"10px 12px", background:"rgba(196,74,58,0.08)", borderRadius:8, fontSize:12.5, display:"flex", gap:8, alignItems:"center"}}>
              <Icon name="alert-triangle" size={14} style={{color:"var(--color-danger)"}} />
              {blocked.length}개 SPEC이 BlockerReport 대기 중
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ───────── LIVE LOG CARD (streaming) ─────────
function LiveLogCard({ height = 320 }) {
  const { LOG_SEED } = window.MoaiData;
  const [lines, setLines] = useS(LOG_SEED);
  const [paused, setPaused] = useS(false);
  const liveRef = useR();

  useE(() => {
    if (paused || !window.MoaiTweaks?.live) return;
    const samples = [
      ["info", "expert-backend", "rewriting tests/permission_test.go (8 cases)"],
      ["dbg",  "harness",        "model=opus-4.7 effort=high temp=0.0"],
      ["ok",   "trust",          "T  tests green   149/149 ✓"],
      ["info", "manager-cycle",  "phase RED → GREEN — 2 tests passing"],
      ["warn", "evaluator-active","REQ-006 — coverage drop on error path"],
      ["dbg",  "checkpoint",     ".moai/state/iter4.yaml flushed (3.6 KB)"],
      ["ok",   "expert-refactoring","✓ ast-grep migration applied — 14 sites"],
      ["info", "context",        "loaded SPEC + 8 EARS — 2,196 tokens"],
      ["err",  "sandbox",        "egress denied: api.example.com (not in allowlist)"],
    ];
    const t = setInterval(() => {
      const now = new Date();
      const ts = now.toTimeString().slice(0,8);
      const [lvl, src, msg] = samples[Math.floor(Math.random() * samples.length)];
      setLines(prev => [...prev.slice(-40), [ts, lvl, src, msg]]);
    }, 1400);
    return () => clearInterval(t);
  }, [paused]);

  useE(() => { if (liveRef.current) liveRef.current.scrollTop = liveRef.current.scrollHeight; }, [lines]);

  return (
    <div className="card">
      <div className="card-head">
        <h3>스트리밍 로그</h3>
        <div style={{display:"flex", gap:8}}>
          <button className="btn ghost sm" onClick={()=>setPaused(p=>!p)}>
            <Icon name={paused ? "play" : "pause"} size={12} />{paused ? "재개" : "일시정지"}
          </button>
          <button className="btn ghost sm" onClick={()=>setLines([])}>
            <Icon name="trash-2" size={12} />지우기
          </button>
        </div>
      </div>
      <div className="log" ref={liveRef} style={{height}}>
        {lines.map(([ts, lvl, src, msg], i) => (
          <div className="ln" key={i}>
            <span className="ts">{ts}</span>
            <span className={`lvl ${lvl}`}>{lvl.toUpperCase()}</span>
            <span className="msg"><span style={{color:"#a8aeac"}}>[{src}]</span> {msg}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

// ───────── SPECS SCREEN ─────────
function SpecsScreen({ onOpenSpec }) {
  const { SPECS } = window.MoaiData;
  const [filter, setFilter] = useS("all");
  const [q, setQ] = useS("");
  const [sortKey, setSortKey] = useS("updated");

  const filtered = SPECS.filter(s => {
    if (filter !== "all" && s.status !== filter) return false;
    if (q && !`${s.id} ${s.title}`.toLowerCase().includes(q.toLowerCase())) return false;
    return true;
  });

  return (
    <div className="fade-up">
      <div className="page-head">
        <div>
          <h1>SPEC</h1>
          <div className="sub">EARS 형식의 헌법 계약 — 모든 phase의 단일 진실 출처(SSOT)</div>
        </div>
        <div style={{display:"flex", gap:10}}>
          <button className="btn outline"><Icon name="download" size={14} />내보내기</button>
          <button className="btn primary" onClick={() => window.dispatchEvent(new CustomEvent("openCreateSpec"))}><Icon name="plus" size={14} />새 SPEC</button>
        </div>
      </div>

      <div className="tabs">
        {[
          ["all", "전체", SPECS.length],
          ["in-progress", "진행 중", SPECS.filter(s => s.status === "in-progress").length],
          ["review", "검토", SPECS.filter(s => s.status === "review").length],
          ["blocked", "블록", SPECS.filter(s => s.status === "blocked").length],
          ["done", "완료", SPECS.filter(s => s.status === "done").length],
          ["todo", "대기", SPECS.filter(s => s.status === "todo").length],
        ].map(([k, l, n]) => (
          <button key={k} className="tab" aria-selected={filter === k} onClick={() => setFilter(k)}>
            {l}<span className="count">{n}</span>
          </button>
        ))}
      </div>

      <div className="filter-bar">
        <div className="search-input">
          <Icon name="search" size={14} style={{color:"var(--fg-3)"}} />
          <input placeholder="SPEC ID 또는 제목 검색…" value={q} onChange={e=>setQ(e.target.value)} />
        </div>
        <div className="seg">
          <button aria-pressed={sortKey==="updated"} onClick={()=>setSortKey("updated")}>최신</button>
          <button aria-pressed={sortKey==="priority"} onClick={()=>setSortKey("priority")}>우선순위</button>
          <button aria-pressed={sortKey==="coverage"} onClick={()=>setSortKey("coverage")}>커버리지</button>
        </div>
        <div style={{marginLeft:"auto", color:"var(--fg-3)", fontSize:12, fontFamily:"var(--font-mono)"}}>{filtered.length} / {SPECS.length}</div>
      </div>

      <div className="card" style={{padding:0, overflow:"hidden"}}>
        <table className="table">
          <thead>
            <tr>
              <th style={{width:200}}><span className="sort">ID <Icon name="chevron-down" size={11} /></span></th>
              <th>제목</th>
              <th style={{width:120}}>상태</th>
              <th style={{width:90}}>우선</th>
              <th style={{width:110}}>오너</th>
              <th style={{width:140}}>커버리지</th>
              <th style={{width:60}}>EARS</th>
              <th style={{width:100}}>업데이트</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map(s => (
              <tr key={s.id} onClick={()=>onOpenSpec(s)}>
                <td className="mono" style={{color:"var(--color-primary)", fontWeight:600}}>{s.id}</td>
                <td>
                  <div style={{fontWeight:600, letterSpacing:"-0.02em"}}>{s.title}</div>
                  <div style={{fontSize:11.5, color:"var(--fg-3)", marginTop:2}}>
                    <span className="mono">domain={s.domain}</span> · phase={s.phase}
                    {s.mxTags > 0 && <span className="chip warning" style={{marginLeft:8, fontSize:10, padding:"1px 6px"}}>@MX × {s.mxTags}</span>}
                  </div>
                </td>
                <td>{statusChip(s.status)}</td>
                <td>{priorityChip(s.priority)}</td>
                <td className="mono" style={{fontSize:11.5, color:"var(--fg-2)"}}>{s.owner}</td>
                <td>
                  <div style={{display:"flex", gap:8, alignItems:"center"}}>
                    <div className={`progress ${s.coverage > 80 ? "success" : s.coverage > 40 ? "" : "warning"}`} style={{flex:1, maxWidth:80}}>
                      <div style={{width:`${s.coverage}%`}} />
                    </div>
                    <span className="mono" style={{fontSize:11.5}}>{s.coverage}%</span>
                  </div>
                </td>
                <td className="mono">{s.ears}</td>
                <td className="mono" style={{fontSize:11.5, color:"var(--fg-3)"}}>{s.updated}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {filtered.length === 0 && <Empty title="검색 결과가 없습니다" sub="다른 키워드를 시도해 보세요." action={<button className="btn outline sm" onClick={()=>{setQ(""); setFilter("all");}}>필터 초기화</button>} />}
      </div>
    </div>
  );
}

// ───────── PIPELINE / RUN ─────────
function PipelineScreen() {
  const [activeTask, setActiveTask] = useS("CLI-001-T3");
  const [tab, setTab] = useS("dag");

  const tasks = {
    plan: [
      { id: "CLI-001-T1", title: "EARS 8개 작성 + 수용기준", agent: "manager-spec", state: "done", deps: [] },
      { id: "CLI-001-T2", title: "DAG 합성 + reads/writes 선언", agent: "manager-strategy", state: "done", deps: ["CLI-001-T1"] },
      { id: "CLI-001-T2b", title: "회의주의 감사 (plan-auditor)", agent: "plan-auditor", state: "done", deps: ["CLI-001-T2"] },
    ],
    run: [
      { id: "CLI-001-T3", title: "permission/bubble.go 구현", agent: "expert-backend", state: "running", deps: ["CLI-001-T2"] },
      { id: "CLI-001-T4", title: "내장 권한 정책 마이그레이션", agent: "expert-refactoring", state: "running", deps: ["CLI-001-T2"] },
      { id: "CLI-001-T5", title: "샌드박스 wrapper 통합", agent: "expert-security", state: "queued", deps: ["CLI-001-T3"] },
    ],
    sync: [
      { id: "CLI-001-T6", title: "TRUST 5 게이트 통과", agent: "manager-quality", state: "queued", deps: ["CLI-001-T3","CLI-001-T4","CLI-001-T5"] },
      { id: "CLI-001-T7", title: "GAN 평가 (fresh context)", agent: "evaluator-active", state: "queued", deps: ["CLI-001-T6"] },
      { id: "CLI-001-T8", title: "@MX 동기화 + git commit", agent: "manager-git", state: "queued", deps: ["CLI-001-T7"] },
    ],
  };

  const stateChip = (s) => ({
    done:    <span className="chip success dot">완료</span>,
    running: <span className="chip primary dot">실행 중</span>,
    queued:  <span className="chip outline">대기</span>,
    failed:  <span className="chip danger dot">실패</span>,
  }[s]);

  return (
    <div className="fade-up">
      <div className="page-head">
        <div>
          <h1>파이프라인</h1>
          <div className="sub"><span className="mono" style={{color:"var(--color-primary)"}}>SPEC-V3-CLI-001</span> · Permission bubble model · iteration 3/∞</div>
        </div>
        <div style={{display:"flex", gap:10, alignItems:"center"}}>
          <Status kind="live">RUNNING</Status>
          <button className="btn outline"><Icon name="square" size={14} />중단</button>
          <button className="btn outline"><Icon name="refresh-cw" size={14} />재시작</button>
        </div>
      </div>

      <div className="tabs">
        <button className="tab" aria-selected={tab==="dag"} onClick={()=>setTab("dag")}>의존성 DAG</button>
        <button className="tab" aria-selected={tab==="timeline"} onClick={()=>setTab("timeline")}>타임라인</button>
        <button className="tab" aria-selected={tab==="logs"} onClick={()=>setTab("logs")}>실시간 로그</button>
        <button className="tab" aria-selected={tab==="contract"} onClick={()=>setTab("contract")}>Sprint Contract</button>
      </div>

      {tab === "dag" && (
        <>
          <div className="pipeline">
            {[["plan","Plan",tasks.plan],["run","Run",tasks.run],["sync","Sync",tasks.sync]].map(([k, label, ts]) => (
              <div key={k} className="pipe-col">
                <h4>
                  <Icon name={k==="plan"?"git-branch":k==="run"?"play":"check-circle-2"} size={12} />
                  {label}
                  <span className="chip" style={{marginLeft:"auto", fontSize:10}}>{ts.length}</span>
                </h4>
                {ts.map(t => (
                  <div key={t.id} className={`pipe-task ${activeTask===t.id?"active":""}`} onClick={()=>setActiveTask(t.id)}>
                    <div className="id">{t.id}</div>
                    <div className="title">{t.title}</div>
                    <div className="meta">
                      {stateChip(t.state)}
                      <span className="chip outline" style={{fontSize:10}}>@{t.agent}</span>
                    </div>
                  </div>
                ))}
              </div>
            ))}
          </div>
          <div className="card" style={{marginTop:14}}>
            <div className="card-head">
              <h3>Sprint Contract — durable state</h3>
              <span className="meta">.moai/state/sprint-cli-001.yaml · 3.4 KB</span>
            </div>
            <pre className="code-block">
<span className="kw">criteria:</span>
  <span className="str">REQ-001</span>: <span className="kw">pass</span>     <span className="com"># provenance in response</span>
  <span className="str">REQ-002</span>: <span className="kw">pass</span>     <span className="com"># bubble on risk≥medium</span>
  <span className="str">REQ-003</span>: <span className="kw">pass</span>     <span className="com"># audit log on bypass</span>
  <span className="str">REQ-004</span>: <span className="kw">running</span>  <span className="com"># 5s timeout → BlockerReport</span>
  <span className="str">REQ-005</span>: <span className="kw">pending</span>  <span className="com"># allowlist skip</span>
  <span className="str">REQ-006</span>: <span className="kw">fail</span>     <span className="com"># worktree write boundary ← evaluator (iter 3)</span>
<span className="kw">iteration:</span> 3
<span className="kw">evaluator_memory:</span> per_iteration   <span className="com"># P-Z01 fix applied</span>
<span className="kw">checkpoint:</span> .moai/state/checkpoint-cli-001-iter3.yaml
            </pre>
          </div>
        </>
      )}

      {tab === "timeline" && (
        <div className="card">
          <div className="card-head"><h3>실행 타임라인</h3><span className="meta">10:42:01 → now</span></div>
          <div className="timeline">
            {[
              ["done", "manager-spec — EARS 8개 작성", "10:38:14 · 142 tokens · ✓"],
              ["done", "manager-strategy — DAG 합성", "10:39:02 · 8 tasks 의존성 매핑 ✓"],
              ["done", "plan-auditor — 회의주의 감사", "10:39:48 · 0 누락된 의존성 ✓"],
              ["running", "expert-backend — bubble.go 구현", "10:42:04 · iteration 3 진행 중"],
              ["running", "expert-refactoring — 권한 정책 ast-grep 마이그레이션", "10:42:09 · 14/22 sites done"],
              ["failed", "evaluator-active iter 3 — REQ-006 worktree boundary", "10:42:21 · BlockerReport 생성"],
              ["done", "checkpoint flush", "10:42:23 · 3.4 KB ✓"],
            ].map(([state, title, meta], i) => (
              <div key={i} className={`tl-item ${state}`}>
                <div className="title">{title}</div>
                <div className="meta">{meta}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {tab === "logs" && <LiveLogCard height={520} />}

      {tab === "contract" && (
        <div className="card">
          <div className="card-head">
            <h3>Sprint Contract 진화</h3>
            <span className="meta">평가자는 매 iteration마다 fresh context</span>
          </div>
          {[1,2,3].map(it => (
            <div key={it} className="config-row" style={{padding:"14px 0"}}>
              <div style={{minWidth:100, fontFamily:"var(--font-mono)", fontSize:12, color:"var(--fg-3)"}}>iteration {it}</div>
              <div style={{flex:1}}>
                <div style={{display:"flex", gap:6, flexWrap:"wrap", marginBottom:6}}>
                  {["REQ-001","REQ-002","REQ-003"].map(r => <span key={r} className="chip success" style={{fontSize:10}}>{r}</span>)}
                  {it >= 2 && ["REQ-004","REQ-005"].map(r => <span key={r} className={`chip ${it===3?"warning":"info"}`} style={{fontSize:10}}>{r}</span>)}
                  {it === 3 && <span className="chip danger" style={{fontSize:10}}>REQ-006 ← new failure</span>}
                </div>
                <div style={{fontSize:12, color:"var(--fg-2)"}}>
                  {it === 1 && "초기 통과: 3/8. evaluator는 모든 사례를 처음 봄."}
                  {it === 2 && "+REQ-004, REQ-005 추가 통과. evaluator memory cleared."}
                  {it === 3 && "REQ-006 boundary leak 발견. previous judgment 메모리 없이 fresh."}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ───────── AGENTS SCREEN ─────────
function AgentsScreen({ onOpen }) {
  const { AGENTS } = window.MoaiData;
  const [kind, setKind] = useS("all");
  const [q, setQ] = useS("");
  const filtered = AGENTS.filter(a => (kind==="all"||a.kind===kind) && (!q || a.id.includes(q.toLowerCase())));

  return (
    <div className="fade-up">
      <div className="page-head">
        <div>
          <h1>에이전트</h1>
          <div className="sub">매니저 7 · 전문가 5 · 빌더 1 · 평가자 2 · 리서처 1 — v2.x 22개 → v3 16개로 통합</div>
        </div>
        <div style={{display:"flex", gap:10}}>
          <button className="btn outline"><Icon name="file-text" size={14} />agent-authoring.md</button>
          <button className="btn primary"><Icon name="plus" size={14} />새 에이전트</button>
        </div>
      </div>

      <div className="filter-bar">
        <div className="search-input"><Icon name="search" size={14} style={{color:"var(--fg-3)"}}/><input placeholder="에이전트 검색…" value={q} onChange={e=>setQ(e.target.value)}/></div>
        <div className="seg">
          {["all","manager","expert","builder","evaluator","researcher"].map(k => (
            <button key={k} aria-pressed={kind===k} onClick={()=>setKind(k)}>{k==="all"?"전체":k}</button>
          ))}
        </div>
      </div>

      <div style={{display:"grid", gridTemplateColumns:"repeat(auto-fill, minmax(280px, 1fr))", gap:14}}>
        {filtered.map(a => (
          <div key={a.id} className="agent-card" onClick={()=>onOpen(a)}>
            <div className="head">
              <div className={`agent-icon ${a.kind}`}>
                {a.kind === "manager" ? "MGR" : a.kind === "expert" ? "EXP" : a.kind === "builder" ? "BLD" : a.kind === "evaluator" ? "EVL" : "RES"}
              </div>
              <div style={{flex:1, minWidth:0}}>
                <div className="name">{a.id}</div>
                <div className="role">{a.role}</div>
              </div>
              <Status kind={a.status === "live" ? "live" : a.status === "warn" ? "warn" : "idle"}>
                {a.status === "live" ? "LIVE" : a.status === "warn" ? "WARN" : "IDLE"}
              </Status>
            </div>
            <div className="desc">{a.desc}</div>
            <div className="footer">
              <span className="chip outline" style={{fontSize:10}}>effort={a.effort}</span>
              <span className="chip outline" style={{fontSize:10}}>iso={a.isolation}</span>
              {a.touches > 0 && <span className="chip primary" style={{fontSize:10}}>{a.touches} files</span>}
              <span style={{marginLeft:"auto", fontSize:10.5, color:"var(--fg-3)", fontFamily:"var(--font-mono)"}}>{a.lastRun}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ───────── HOOKS SCREEN ─────────
function HooksScreen() {
  const { HOOKS } = window.MoaiData;
  return (
    <div className="fade-up">
      <div className="page-head">
        <div>
          <h1>후크 & 하니스</h1>
          <div className="sub">27개 핸들러 — 활성 6 · stub 5 · orphan 1. JSON 응답 프로토콜 마이그레이션 진행 중.</div>
        </div>
        <div style={{display:"flex", gap:10}}>
          <button className="btn outline"><Icon name="settings-2" size={14} />settings.json</button>
          <button className="btn primary"><Icon name="zap" size={14} />JSON 마이그레이션</button>
        </div>
      </div>

      <div className="stat-grid" style={{gridTemplateColumns:"repeat(4, 1fr)"}}>
        <Stat label="활성 핸들러" value="6 / 12" delta="JSON 4 / exit 2" deltaKind="flat" />
        <Stat label="훅 호출 (24h)" value="7,056" delta="+12%" />
        <Stat label="평균 지연" value="14ms" delta="-2ms" />
        <Stat label="실패율" value="0.02%" delta="-0.01pp" />
      </div>

      <div className="grid-2">
        <div className="card" style={{padding:0, overflow:"hidden"}}>
          <div className="card-head" style={{padding:"16px 20px", marginBottom:0, borderBottom:"1px solid var(--border-1)"}}>
            <h3>핸들러 매트릭스</h3>
            <span className="meta">.claude/settings.json</span>
          </div>
          <div>
            {HOOKS.map(h => (
              <div key={h.event} className="hook-row">
                <div>
                  {h.state === "active" ? <Icon name="check-circle-2" size={16} style={{color:"var(--color-success)"}}/> :
                   h.state === "stub" ? <Icon name="circle-dashed" size={16} style={{color:"var(--fg-3)"}}/> :
                                        <Icon name="alert-triangle" size={16} style={{color:"var(--color-warning)"}}/>}
                </div>
                <div>
                  <div className="name">{h.event}</div>
                  <div className="desc">{h.desc}</div>
                </div>
                <div>
                  {h.protocol === "json" ? <span className="chip primary">JSON</span> :
                   h.protocol === "exit" ? <span className="chip outline">exit code</span> :
                                            <span className="chip" style={{fontSize:10}}>—</span>}
                </div>
                <div className="mono" style={{fontSize:11.5, color:"var(--fg-3)"}}>
                  {h.calls > 0 ? `${h.calls.toLocaleString()} · ${h.latency}` : "—"}
                </div>
                <div style={{textAlign:"right"}}>
                  {h.state === "active" ? <Status kind="live">ON</Status> :
                   h.state === "orphan" ? <Status kind="error">ORPHAN</Status> :
                                          <Status kind="idle">STUB</Status>}
                </div>
              </div>
            ))}
          </div>
        </div>

        <div style={{display:"flex", flexDirection:"column", gap:14}}>
          <div className="card">
            <div className="card-head">
              <h3>하니스 라우팅</h3>
              <span className="chip primary">standard</span>
            </div>
            <div style={{display:"flex", flexDirection:"column", gap:10}}>
              {[
                ["minimal", "단순 변경", "린트, 포맷, 단일파일 수정", 0.4],
                ["standard", "표준 SPEC", "다중파일 + 테스트 + 리뷰", 1.0],
                ["thorough", "복잡 SPEC", "+ adversarial, +논문, +실험", 2.4],
              ].map(([level, label, desc, mult]) => (
                <div key={level} className="config-row" style={{padding:"10px 0"}}>
                  <div style={{flex:1}}>
                    <div style={{display:"flex", alignItems:"center", gap:8, marginBottom:3}}>
                      <span className="mono" style={{fontWeight:700, fontSize:12.5, color:level==="standard"?"var(--color-primary)":"var(--fg-1)"}}>{level}</span>
                      <span style={{fontSize:13, fontWeight:600}}>{label}</span>
                      {level==="standard" && <span className="chip success" style={{fontSize:10}}>active</span>}
                    </div>
                    <div style={{fontSize:11.5, color:"var(--fg-3)"}}>{desc}</div>
                  </div>
                  <div className="mono" style={{fontSize:12, color:"var(--fg-2)"}}>×{mult}</div>
                </div>
              ))}
            </div>
          </div>

          <div className="card">
            <div className="card-head"><h3>JSON 응답 예시</h3><span className="meta">PostToolUse</span></div>
            <pre className="code-block">{`{
  "additionalContext": "@MX:WARN at L42",
  "permissionDecision": "ask",
  "updatedInput": null,
  "systemMessage": "egress denied",
  "continue": true
}`}</pre>
          </div>
        </div>
      </div>
    </div>
  );
}

// ───────── SAFETY SCREEN (TRUST + sandbox + permissions) ─────────
function SafetyScreen() {
  const { TRUST, PERMISSIONS } = window.MoaiData;
  const [sandboxOn, setSandboxOn] = useS(true);
  const [bubbleOn, setBubbleOn] = useS(true);
  const [egressOn, setEgressOn] = useS(true);

  return (
    <div className="fade-up">
      <div className="page-head">
        <div>
          <h1>안전 & 평가</h1>
          <div className="sub">TRUST 5 게이트 · 샌드박스 · 권한 버블 · OWASP Top 10 Agentic Apps 정합성</div>
        </div>
        <Status kind="live">SANDBOXED</Status>
      </div>

      <div className="grid-2">
        <div className="card">
          <div className="card-head">
            <h3>TRUST 5 — 최근 평가</h3>
            <span className="meta">SPEC-V3-CLI-001 · 12초 전</span>
          </div>
          {TRUST.map((t, i) => (
            <div key={i} className="trust-row" style={{borderBottom: i === TRUST.length-1 ? "none" : ""}}>
              <div className={`trust-letter ${t.status}`}>{t.letter}</div>
              <div style={{flex:1, minWidth:0}}>
                <div style={{display:"flex", gap:8, alignItems:"center"}}>
                  <span style={{fontWeight:700, fontSize:13.5}}>{t.name}</span>
                  <span className="chip" style={{marginLeft:"auto"}}>{t.value}</span>
                </div>
                <div style={{fontSize:12, color:"var(--fg-2)", marginTop:3}}>{t.note}</div>
              </div>
            </div>
          ))}
        </div>

        <div className="card">
          <div className="card-head"><h3>샌드박스 설정</h3><span className="meta">security.yaml</span></div>
          <div className="config-row">
            <div style={{flex:1}}>
              <div className="label">에이전트 실행 격리</div>
              <div className="desc">Bubblewrap (Linux) / Seatbelt (macOS) / Docker (CI)</div>
            </div>
            <div className="switch" role="switch" aria-checked={sandboxOn} onClick={()=>setSandboxOn(v=>!v)} />
          </div>
          <div className="config-row">
            <div style={{flex:1}}>
              <div className="label">위험 도구 → bubble 모드</div>
              <div className="desc">risk≥medium 도구 호출 시 부모 터미널 승인</div>
            </div>
            <div className="switch" role="switch" aria-checked={bubbleOn} onClick={()=>setBubbleOn(v=>!v)} />
          </div>
          <div className="config-row">
            <div style={{flex:1}}>
              <div className="label">네트워크 egress 차단</div>
              <div className="desc">기본 deny — allowlist만 허용</div>
            </div>
            <div className="switch" role="switch" aria-checked={egressOn} onClick={()=>setEgressOn(v=>!v)} />
          </div>
          <div className="config-row" style={{borderBottom:"none"}}>
            <div style={{flex:1}}>
              <div className="label">파일쓰기 스코프</div>
              <div className="desc">에이전트별 worktree 외부 쓰기 거부</div>
            </div>
            <span className="chip success">ENFORCED</span>
          </div>
        </div>
      </div>

      <div className="card" style={{marginTop:14, padding:0, overflow:"hidden"}}>
        <div className="card-head" style={{padding:"16px 20px", borderBottom:"1px solid var(--border-1)", marginBottom:0}}>
          <h3>권한 스택 — provenance per value</h3>
          <span className="meta">policy &gt; user &gt; project &gt; local &gt; plugin &gt; skill &gt; session &gt; builtin</span>
        </div>
        <table className="table">
          <thead>
            <tr><th>도구</th><th>스코프</th><th>출처</th><th>위험</th><th>모드</th><th></th></tr>
          </thead>
          <tbody>
            {PERMISSIONS.map((p, i) => (
              <tr key={i}>
                <td className="mono" style={{color:"var(--color-primary)", fontWeight:600}}>{p.tool}</td>
                <td style={{fontSize:12.5, color:"var(--fg-2)"}}>{p.scope}</td>
                <td><span className="chip outline" style={{fontSize:10.5}}>{p.source}</span></td>
                <td>{
                  p.risk === "critical" ? <span className="chip danger">critical</span> :
                  p.risk === "high"     ? <span className="chip warning">high</span> :
                  p.risk === "medium"   ? <span className="chip info">medium</span> :
                                          <span className="chip success">low</span>
                }</td>
                <td className="mono" style={{fontSize:12}}>{p.mode}</td>
                <td className="right">
                  {p.mode === "deny" ? <span className="chip danger" style={{fontSize:10}}>BLOCKED</span> :
                   p.mode === "bubble" ? <span className="chip warning" style={{fontSize:10}}>ASK</span> :
                                          <span className="chip success" style={{fontSize:10}}>ALLOW</span>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ───────── SKILLS / RULES SCREEN ─────────
function LibraryScreen() {
  const { SKILLS, RULES } = window.MoaiData;
  const [tab, setTab] = useS("skills");
  const [g, setG] = useS("all");
  const groups = ["all","foundation","workflow","cmd","domain","design","tool","platform"];

  return (
    <div className="fade-up">
      <div className="page-head">
        <div>
          <h1>스킬 & 룰 라이브러리</h1>
          <div className="sub">v3 표면 축소 — 스킬 48 → {SKILLS.length} (목표 24) · 룰 통합 진행 중</div>
        </div>
        <div style={{display:"flex", gap:10}}>
          <button className="btn outline"><Icon name="file-search" size={14} />정합성 검사</button>
          <button className="btn primary"><Icon name="package-plus" size={14} />새 스킬</button>
        </div>
      </div>

      <div className="tabs">
        <button className="tab" aria-selected={tab==="skills"} onClick={()=>setTab("skills")}>스킬 <span className="count">{SKILLS.length}</span></button>
        <button className="tab" aria-selected={tab==="rules"} onClick={()=>setTab("rules")}>룰 <span className="count">{RULES.length}</span></button>
      </div>

      {tab === "skills" && (
        <>
          <div className="filter-bar">
            <div className="seg">
              {groups.map(k => <button key={k} aria-pressed={g===k} onClick={()=>setG(k)}>{k}</button>)}
            </div>
            <div style={{marginLeft:"auto", fontSize:12, color:"var(--fg-3)", fontFamily:"var(--font-mono)"}}>
              progressive disclosure: ref=3000, specialist=5000, orchestrator=8000+
            </div>
          </div>
          <div style={{display:"grid", gridTemplateColumns:"repeat(auto-fill, minmax(290px, 1fr))", gap:12}}>
            {SKILLS.filter(s => g==="all"||s.group===g).map(s => (
              <div key={s.name} className="skill-tile">
                <div className="name">{s.name}</div>
                <div className="desc">{s.desc}</div>
                <div className="meta">
                  <span className="chip outline" style={{fontSize:10}}>L{s.level}</span>
                  <span className="chip outline" style={{fontSize:10}}>{s.size}</span>
                  <span className="chip outline" style={{fontSize:10}}>refs: {s.refs}</span>
                  <span className="chip primary" style={{fontSize:10, marginLeft:"auto"}}>{s.group}</span>
                </div>
              </div>
            ))}
          </div>
        </>
      )}

      {tab === "rules" && (
        <div className="card" style={{padding:0, overflow:"hidden"}}>
          <table className="table">
            <thead><tr><th>이름</th><th>경로</th><th style={{width:80}}>구역</th><th style={{width:80}} className="right">LOC</th><th></th></tr></thead>
            <tbody>
              {RULES.map(r => (
                <tr key={r.name}>
                  <td>
                    <div className="mono" style={{fontWeight:600, color:"var(--color-primary)"}}>{r.name}</div>
                    <div style={{fontSize:11.5, color:"var(--fg-3)", marginTop:3}}>{r.desc}</div>
                  </td>
                  <td className="mono" style={{fontSize:11.5, color:"var(--fg-2)"}}>{r.path}</td>
                  <td>{r.kind === "FROZEN" ? <span className="chip danger">FROZEN</span> : <span className="chip info">EVOLVABLE</span>}</td>
                  <td className="mono right">{r.lines}</td>
                  <td className="right"><Icon name="external-link" size={14} style={{color:"var(--fg-3)"}}/></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

Object.assign(window, { DashboardScreen, SpecsScreen, PipelineScreen, AgentsScreen, HooksScreen, SafetyScreen, LibraryScreen, LiveLogCard });
