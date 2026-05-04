// MoAI-ADK v3 Console — Drawers + Modals
const { useState: useSt, useEffect: useEf } = React;

function SpecDrawer({ spec, open, onClose }) {
  const { EARS_SAMPLE } = window.MoaiData;
  const [tab, setTab] = useSt("ears");
  if (!spec) return null;
  return (
    <Drawer open={open} onClose={onClose}
      title={<span className="mono" style={{color:"var(--color-primary)"}}>{spec.id}</span>}
      badge={statusChip(spec.status)}
      footer={<>
        <button className="btn ghost" onClick={onClose}>닫기</button>
        <button className="btn outline"><Icon name="git-branch" size={14}/>브랜치 열기</button>
        <button className="btn primary"><Icon name="play" size={14}/>재실행</button>
      </>}>
      <div style={{marginBottom:18}}>
        <div style={{fontSize:18, fontWeight:800, letterSpacing:"-0.035em", marginBottom:8}}>{spec.title}</div>
        <div style={{display:"flex", gap:6, flexWrap:"wrap"}}>
          {priorityChip(spec.priority)}
          <span className="chip outline">domain={spec.domain}</span>
          <span className="chip outline">phase={spec.phase}</span>
          <span className="chip outline">owner={spec.owner}</span>
        </div>
      </div>

      <div className="grid-3" style={{marginBottom:18, gap:10}}>
        <div className="card compact"><div style={{fontSize:11, color:"var(--fg-3)", fontWeight:600}}>EARS</div><div style={{fontSize:22, fontWeight:800, fontVariantNumeric:"tabular-nums"}}>{spec.ears}</div></div>
        <div className="card compact"><div style={{fontSize:11, color:"var(--fg-3)", fontWeight:600}}>커버리지</div><div style={{fontSize:22, fontWeight:800}}>{spec.coverage}%</div></div>
        <div className="card compact"><div style={{fontSize:11, color:"var(--fg-3)", fontWeight:600}}>@MX 태그</div><div style={{fontSize:22, fontWeight:800, color: spec.mxTags > 0 ? "var(--color-warning)" : "var(--fg-1)"}}>{spec.mxTags}</div></div>
      </div>

      <div className="tabs" style={{marginBottom:14}}>
        <button className="tab" aria-selected={tab==="ears"} onClick={()=>setTab("ears")}>EARS 요구사항</button>
        <button className="tab" aria-selected={tab==="contract"} onClick={()=>setTab("contract")}>Contract</button>
        <button className="tab" aria-selected={tab==="mx"} onClick={()=>setTab("mx")}>@MX</button>
      </div>

      {tab === "ears" && <div>{EARS_SAMPLE.slice(0, spec.ears).map(r => (
        <div key={r.id} className="ears-row">
          <div className="id-row">
            <span className="mono" style={{fontWeight:700, fontSize:11.5, color:"var(--color-primary)"}}>{r.id}</span>
            <span className="chip outline" style={{fontSize:10}}>{r.type}</span>
            {r.status === "pass" && <span className="chip success" style={{marginLeft:"auto"}}>pass</span>}
            {r.status === "running" && <span className="chip primary" style={{marginLeft:"auto"}}>running</span>}
            {r.status === "fail" && <span className="chip danger" style={{marginLeft:"auto"}}>fail</span>}
            {r.status === "pending" && <span className="chip outline" style={{marginLeft:"auto"}}>pending</span>}
          </div>
          <div className="req">{r.text.pre} <span className="kw">{r.text.kw}</span> {r.text.post}</div>
        </div>
      ))}</div>}

      {tab === "contract" && <pre className="code-block">{`criteria:
  REQ-001..003: pass
  REQ-004: running
  REQ-005: pending
  REQ-006: fail   ← evaluator iter 3
iteration: 3
evaluator_memory: per_iteration
checkpoint: .moai/state/checkpoint-${spec.id.toLowerCase()}-iter3.yaml`}</pre>}

      {tab === "mx" && (spec.mxTags > 0 ? (
        <div>{Array.from({length: spec.mxTags}).map((_, i) => (
          <div key={i} className="config-row">
            <span className="chip warning">@MX:WARN</span>
            <div style={{flex:1, fontSize:12.5, fontFamily:"var(--font-mono)"}}>internal/permission/bubble.go:L{42+i*8}</div>
            <span style={{fontSize:11.5, color:"var(--fg-3)"}}>{i===0?"timeout boundary":i===1?"audit log":"egress allowlist"}</span>
          </div>
        ))}</div>
      ) : <Empty title="@MX 태그 없음" sub="이 SPEC의 모든 관찰값이 정상 처리되었습니다." mascot={false}/>)}
    </Drawer>
  );
}

function AgentDrawer({ agent, open, onClose }) {
  if (!agent) return null;
  return (
    <Drawer open={open} onClose={onClose}
      title={<span className="mono">{agent.id}</span>}
      badge={<Status kind={agent.status==="live"?"live":agent.status==="warn"?"warn":"idle"}>{agent.status.toUpperCase()}</Status>}
      footer={<>
        <button className="btn ghost" onClick={onClose}>닫기</button>
        <button className="btn outline"><Icon name="file-text" size={14}/>frontmatter</button>
        <button className="btn primary"><Icon name="play" size={14}/>호출</button>
      </>}>
      <div style={{display:"flex", gap:14, alignItems:"center", marginBottom:18}}>
        <div className={`agent-icon ${agent.kind}`} style={{width:56, height:56, fontSize:14, borderRadius:12}}>
          {agent.kind === "manager" ? "MGR" : agent.kind === "expert" ? "EXP" : agent.kind === "builder" ? "BLD" : agent.kind === "evaluator" ? "EVL" : "RES"}
        </div>
        <div>
          <div style={{fontSize:18, fontWeight:800, letterSpacing:"-0.03em"}}>{agent.role}</div>
          <div style={{fontSize:12.5, color:"var(--fg-2)", marginTop:3}}>kind: {agent.kind} · last run {agent.lastRun}</div>
        </div>
      </div>
      <div className="card compact" style={{marginBottom:14}}>
        <div style={{fontSize:13, lineHeight:1.6}}>{agent.desc}</div>
      </div>

      <h4 style={{margin:"16px 0 10px", fontSize:12, fontWeight:700, color:"var(--fg-3)", textTransform:"uppercase", letterSpacing:"0.04em"}}>Frontmatter</h4>
      <pre className="code-block">{`name: ${agent.id}
kind: ${agent.kind}
effort: ${agent.effort}
isolation: ${agent.isolation}
permissionMode: ${agent.kind === "evaluator" || agent.kind === "researcher" ? "plan" : "acceptEdits"}
tools:
  - Read, Glob, Grep
${agent.kind !== "evaluator" ? "  - Write, Edit\n" : ""}  - Bash (allowlist)
memory: project`}</pre>

      <h4 style={{margin:"16px 0 10px", fontSize:12, fontWeight:700, color:"var(--fg-3)", textTransform:"uppercase", letterSpacing:"0.04em"}}>최근 호출 내역</h4>
      <div>
        {[
          ["2분 전", "SPEC-V3-CLI-001 / iter 3", "ok"],
          ["18분 전", "SPEC-V3-AGT-002 / iter 1", "ok"],
          ["1시간 전", "SPEC-V3-CLI-001 / iter 2", "warn"],
          ["3시간 전", "SPEC-V3-HOOKS-002", "ok"],
        ].map(([t, ctx, s], i) => (
          <div key={i} className="config-row">
            <div className="mono" style={{fontSize:11.5, color:"var(--fg-3)", width:80}}>{t}</div>
            <div style={{flex:1, fontSize:12.5}}>{ctx}</div>
            <span className={`chip ${s==="ok"?"success":"warning"}`} style={{fontSize:10}}>{s}</span>
          </div>
        ))}
      </div>
    </Drawer>
  );
}

function CreateSpecModal({ open, onClose }) {
  const [domain, setDomain] = useSt("CLI");
  const [title, setTitle] = useSt("");
  const [priority, setPriority] = useSt("medium");
  return (
    <Modal open={open} onClose={onClose}
      title="새 SPEC 작성"
      sub="EARS 형식으로 작성됩니다. 작성 후 manager-spec이 검증."
      footer={<>
        <button className="btn ghost" onClick={onClose}>취소</button>
        <button className="btn outline"><Icon name="zap" size={14}/>AI 초안</button>
        <button className="btn primary"><Icon name="check" size={14}/>SPEC 생성 →</button>
      </>}>
      <div className="field">
        <label>제목</label>
        <input className="input" placeholder="예: Permission bubble model with provenance" value={title} onChange={e=>setTitle(e.target.value)} />
      </div>
      <div className="grid-2" style={{gap:12}}>
        <div className="field">
          <label>도메인</label>
          <select className="select" value={domain} onChange={e=>setDomain(e.target.value)}>
            {["CLI","AGT","HOOKS","MEM","SCH","SKL","CLN","OUT","PLG"].map(d => <option key={d}>{d}</option>)}
          </select>
          <div className="hint">SPEC ID: SPEC-V3-{domain}-{String(Math.floor(Math.random()*900)+100)}</div>
        </div>
        <div className="field">
          <label>우선순위</label>
          <div className="seg" style={{width:"fit-content"}}>
            {["low","medium","high","critical"].map(p => (
              <button key={p} aria-pressed={priority===p} onClick={()=>setPriority(p)} type="button">{p}</button>
            ))}
          </div>
        </div>
      </div>
      <div className="field">
        <label>EARS 요구사항 초안 (선택)</label>
        <textarea className="textarea" rows={5} defaultValue={`# Ubiquitous
시스템은 [항상] [조건]을(를) 보장해야 한다.

# Event-driven
[이벤트 발생] WHEN, 시스템은 [응답]해야 한다.`} />
      </div>
      <div className="field">
        <label>오너 에이전트</label>
        <select className="select" defaultValue="manager-spec">
          <option>manager-spec</option><option>manager-cycle</option><option>manager-strategy</option>
        </select>
      </div>
    </Modal>
  );
}

function CmdPalette({ open, onClose, onNav }) {
  const [q, setQ] = useSt("");
  const [active, setActive] = useSt(0);
  const items = [
    { group:"Nav", label:"오버뷰로 이동", target:"dashboard", kbd:"G O", icon:"layout-dashboard" },
    { group:"Nav", label:"SPEC 목록", target:"specs", kbd:"G S", icon:"file-code-2" },
    { group:"Nav", label:"파이프라인", target:"pipeline", kbd:"G P", icon:"git-branch" },
    { group:"Nav", label:"에이전트", target:"agents", kbd:"G A", icon:"users" },
    { group:"Nav", label:"후크 & 하니스", target:"hooks", kbd:"G H", icon:"webhook" },
    { group:"Nav", label:"안전 & 평가", target:"safety", kbd:"G T", icon:"shield-check" },
    { group:"Nav", label:"스킬 & 룰", target:"library", kbd:"G L", icon:"library" },
    { group:"Action", label:"새 SPEC 생성…", target:"createSpec", kbd:"⌘ N", icon:"plus" },
    { group:"Action", label:"테마 토글", target:"toggleTheme", kbd:"⌘ J", icon:"sun-moon" },
  ];
  const filtered = items.filter(it => !q || (it.label + it.group).toLowerCase().includes(q.toLowerCase()));
  useEf(() => { if (open) setQ(""), setActive(0); }, [open]);
  useEf(() => {
    if (!open) return;
    const onKey = (e) => {
      if (e.key === "ArrowDown") { e.preventDefault(); setActive(a => Math.min(a+1, filtered.length-1)); }
      else if (e.key === "ArrowUp") { e.preventDefault(); setActive(a => Math.max(a-1, 0)); }
      else if (e.key === "Enter") {
        const it = filtered[active]; if (it) { onNav(it.target); onClose(); }
      } else if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, filtered, active]);
  return (
    <>
      <div className={`drawer-scrim ${open ? "open" : ""}`} onClick={onClose}/>
      <div className={`cmd-pal ${open ? "open" : ""}`}>
        <input autoFocus={open} placeholder="명령어 검색…" value={q} onChange={e=>{setQ(e.target.value); setActive(0);}}/>
        <div className="cmd-list">
          {filtered.map((it, i) => (
            <div key={it.target} className={`cmd-row ${i===active?"active":""}`}
              onMouseEnter={()=>setActive(i)}
              onClick={()=>{onNav(it.target); onClose();}}>
              <Icon name={it.icon} size={14} />
              <span>{it.label}</span>
              <span className="group">{it.group}</span>
              <span className="kbd">{it.kbd}</span>
            </div>
          ))}
          {filtered.length === 0 && <div style={{padding:24, textAlign:"center", color:"var(--fg-3)", fontSize:13}}>일치하는 명령이 없습니다.</div>}
        </div>
      </div>
    </>
  );
}

Object.assign(window, { SpecDrawer, AgentDrawer, CreateSpecModal, CmdPalette });
