// MoAI-ADK v3 Console — App shell + tweaks + routing
const { useState: uS, useEffect: uE, useRef: uR } = React;

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "theme": "light",
  "density": "compact",
  "layout": "sidebar",
  "live": true,
  "mascot": true
}/*EDITMODE-END*/;

const NAV = [
  { group: "워크플로우", items: [
    { id:"dashboard", label:"오버뷰", icon:"layout-dashboard" },
    { id:"specs",     label:"SPEC", icon:"file-code-2", badge:"12" },
    { id:"pipeline",  label:"파이프라인", icon:"git-branch", live:true },
  ]},
  { group: "운영", items: [
    { id:"agents",   label:"에이전트", icon:"users", badge:"16" },
    { id:"hooks",    label:"후크 & 하니스", icon:"webhook" },
    { id:"safety",   label:"안전 & 평가", icon:"shield-check" },
    { id:"library",  label:"스킬 & 룰", icon:"library" },
  ]},
];

function App() {
  const [tweaks, setTweak] = useTweaks(TWEAK_DEFAULTS);
  const [route, setRoute] = uS("dashboard");
  const [collapsed, setCollapsed] = uS(false);
  const [palOpen, setPalOpen] = uS(false);
  const [createOpen, setCreateOpen] = uS(false);
  const [openSpec, setOpenSpec] = uS(null);
  const [openAgent, setOpenAgent] = uS(null);
  // expose tweaks so live log knows whether to stream
  uE(() => { window.MoaiTweaks = tweaks; }, [tweaks]);

  uE(() => {
    document.documentElement.setAttribute("data-theme", tweaks.theme);
    document.documentElement.setAttribute("data-density", tweaks.density);
  }, [tweaks.theme, tweaks.density]);

  uE(() => {
    const onNav = (e) => setRoute(e.detail);
    const onCreate = () => setCreateOpen(true);
    window.addEventListener("nav", onNav);
    window.addEventListener("openCreateSpec", onCreate);
    return () => { window.removeEventListener("nav", onNav); window.removeEventListener("openCreateSpec", onCreate); };
  }, []);

  uE(() => {
    const onKey = (e) => {
      if ((e.metaKey||e.ctrlKey) && e.key === "k") { e.preventDefault(); setPalOpen(o=>!o); }
      else if ((e.metaKey||e.ctrlKey) && e.key === "j") { e.preventDefault(); setTweak("theme", tweaks.theme==="light"?"dark":"light"); }
      else if ((e.metaKey||e.ctrlKey) && e.key.toLowerCase() === "n") { e.preventDefault(); setCreateOpen(true); }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [tweaks.theme]);

  const handlePal = (target) => {
    if (target === "createSpec") setCreateOpen(true);
    else if (target === "toggleTheme") setTweak("theme", tweaks.theme === "light" ? "dark" : "light");
    else setRoute(target);
  };

  const crumbLabel = NAV.flatMap(g => g.items).find(i => i.id === route)?.label || "오버뷰";

  return (
    <div className="app" data-collapsed={collapsed} data-layout={tweaks.layout}>
      {tweaks.layout === "sidebar" && (
        <aside className="sidebar">
          <div className="sidebar-header">
            <div className="sidebar-logo">M</div>
            {!collapsed && <div className="sidebar-brand">
              <span className="name">MoAI-ADK</span>
              <span className="tag">v3 · console</span>
            </div>}
            <button className="icon-btn" onClick={()=>setCollapsed(c=>!c)} style={{marginLeft:"auto"}} aria-label="접기">
              <Icon name={collapsed?"panel-left-open":"panel-left-close"} size={16}/>
            </button>
          </div>
          <nav className="nav">
            {NAV.map(g => (
              <div key={g.group} className="nav-section">
                {!collapsed && <div className="nav-label">{g.group}</div>}
                {g.items.map(it => (
                  <button key={it.id} className="nav-item" aria-current={route===it.id?"page":undefined} onClick={()=>setRoute(it.id)}>
                    <Icon name={it.icon} size={16}/>
                    {!collapsed && <span>{it.label}</span>}
                    {!collapsed && it.live && <span className="dot" />}
                    {!collapsed && it.badge && <span className="badge">{it.badge}</span>}
                  </button>
                ))}
              </div>
            ))}
          </nav>
          <div className="sidebar-footer">
            <div className="avatar">개</div>
            {!collapsed && <div className="user-info">
              <div className="name">개발자</div>
              <div className="role">harness=standard</div>
            </div>}
            {!collapsed && <button className="icon-btn" aria-label="설정"><Icon name="settings" size={14}/></button>}
          </div>
        </aside>
      )}

      <header className="topbar">
        {tweaks.layout === "topbar" && <div style={{display:"flex", alignItems:"center", gap:10, marginRight:16}}>
          <div className="sidebar-logo" style={{width:30, height:30}}>M</div>
          <div className="sidebar-brand"><span className="name">MoAI-ADK</span><span className="tag">v3</span></div>
        </div>}

        <div className="crumbs">
          <span>MoAI-ADK</span>
          <span className="sep">/</span>
          <span className="current">{crumbLabel}</span>
        </div>

        {tweaks.layout === "topbar" && (
          <nav style={{display:"flex", gap:4, marginLeft:16}}>
            {NAV.flatMap(g=>g.items).map(it => (
              <button key={it.id} className="nav-item" aria-current={route===it.id?"page":undefined} style={{padding:"6px 12px", fontSize:13}} onClick={()=>setRoute(it.id)}>
                <Icon name={it.icon} size={14}/>{it.label}
              </button>
            ))}
          </nav>
        )}

        <button className="cmd-trigger" onClick={()=>setPalOpen(true)}>
          <Icon name="search" size={13}/>
          <span>명령어 또는 SPEC 검색…</span>
          <kbd>⌘K</kbd>
        </button>

        <button className="icon-btn" onClick={()=>setTweak("theme", tweaks.theme==="light"?"dark":"light")} aria-label="테마">
          <Icon name={tweaks.theme==="light"?"moon":"sun"} size={16}/>
        </button>
        <button className="icon-btn" aria-label="알림"><Icon name="bell" size={16}/><span className="pip"/></button>
        <button className="icon-btn" aria-label="문서"><Icon name="book-open" size={16}/></button>
      </header>

      <main className="main" key={route}>
        {route === "dashboard" && <DashboardScreen onOpenSpec={setOpenSpec} onOpenAgent={setOpenAgent} />}
        {route === "specs"     && <SpecsScreen onOpenSpec={setOpenSpec} />}
        {route === "pipeline"  && <PipelineScreen />}
        {route === "agents"    && <AgentsScreen onOpen={setOpenAgent} />}
        {route === "hooks"     && <HooksScreen />}
        {route === "safety"    && <SafetyScreen />}
        {route === "library"   && <LibraryScreen />}
      </main>

      <SpecDrawer spec={openSpec} open={!!openSpec} onClose={()=>setOpenSpec(null)} />
      <AgentDrawer agent={openAgent} open={!!openAgent} onClose={()=>setOpenAgent(null)} />
      <CreateSpecModal open={createOpen} onClose={()=>setCreateOpen(false)} />
      <CmdPalette open={palOpen} onClose={()=>setPalOpen(false)} onNav={handlePal} />

      <TweaksPanel title="Tweaks">
        <TweakSection title="외형">
          <TweakRadio label="테마" value={tweaks.theme} onChange={v=>setTweak("theme", v)} options={[{value:"light", label:"Light"},{value:"dark", label:"Dark"}]} />
          <TweakRadio label="밀도" value={tweaks.density} onChange={v=>setTweak("density", v)} options={[{value:"cozy", label:"Cozy"},{value:"compact", label:"Compact"},{value:"dense", label:"Dense"}]} />
          <TweakRadio label="레이아웃" value={tweaks.layout} onChange={v=>setTweak("layout", v)} options={[{value:"sidebar", label:"Sidebar"},{value:"topbar", label:"Topbar"}]} />
        </TweakSection>
        <TweakSection title="동작">
          <TweakToggle label="실시간 로그 스트리밍" value={tweaks.live} onChange={v=>setTweak("live", v)} />
          <TweakToggle label="마스코트 표시" value={tweaks.mascot} onChange={v=>setTweak("mascot", v)} />
        </TweakSection>
        <TweakSection title="단축키">
          <div style={{fontSize:12, color:"var(--fg-2)", lineHeight:1.8}}>
            <div><kbd style={{fontSize:10.5, padding:"1px 6px", border:"1px solid var(--border-1)", borderRadius:4}}>⌘K</kbd> 명령 팔레트</div>
            <div><kbd style={{fontSize:10.5, padding:"1px 6px", border:"1px solid var(--border-1)", borderRadius:4}}>⌘J</kbd> 테마 토글</div>
            <div><kbd style={{fontSize:10.5, padding:"1px 6px", border:"1px solid var(--border-1)", borderRadius:4}}>⌘N</kbd> 새 SPEC</div>
          </div>
        </TweakSection>
      </TweaksPanel>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById("root")).render(<App />);
