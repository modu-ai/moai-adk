/* eslint-disable */
// Screen renderers for moai-adk TUI

// ─────────────────────────────────────────────────────────
// 1. install.sh  (macOS / Linux)
// ─────────────────────────────────────────────────────────
function ScreenInstallSh() {
  const t = useTok();
  return <Term title="bash — install.sh" width={760} badge="zsh · 132×38">
    <Prompt host="yuna@air" path="~" cmd={<>curl -fsSL <Tx c="info">https://moai-adk.dev/install.sh</Tx> | sh</>} />

    <div style={{ marginTop: 10, marginBottom: 14 }}>
      <ThickBox accent padding="14px 18px">
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 12 }}>
          <div>
            <div style={{ fontFamily: FONT_SANS, fontSize: 16, fontWeight: 800, letterSpacing: "-0.03em", color: t.accent }}>
              MoAI-ADK Installer
            </div>
            <div style={{ fontFamily: FONT_SANS, fontSize: 12.5, color: t.dim, marginTop: 2, letterSpacing: "-0.02em" }}>
              모두의AI · Agentic Development Kit
            </div>
          </div>
          <Pill kind="primary" solid>v3.2.4</Pill>
        </div>
      </ThickBox>
    </div>

    <Section title="환경 감지" sub="Layer 1 · 환경 변수 우선" />
    <CheckLine status="ok"   label="플랫폼"     value="darwin · arm64" hint="macOS 14.5" />
    <CheckLine status="ok"   label="셸"         value="zsh · 5.9" />
    <CheckLine status="ok"   label="Git"        value="2.43.0" />
    <CheckLine status="ok"   label="Claude Code" value="1.0.18" hint="권장 버전" />

    <Section title="다운로드" right="moai-adk_3.2.4_darwin_arm64.tar.gz · 19.4 MB" />
    <div style={{ marginTop: 4 }}>
      <Progress value={100} label="아카이브" color={t.success} />
    </div>
    <div style={{ marginTop: 6 }}>
      <Progress value={100} label="체크섬"   color={t.success} />
    </div>
    <div style={{ marginTop: 8, fontFamily: FONT_MONO, fontSize: 12, color: t.dim }}>
      <Tx c="success" b>✓</Tx> sha256 <Tx c="fg">3e4b7c2…a91f</Tx> 검증 완료
    </div>

    <div style={{ marginTop: 16 }}>
      <Box title="설치 완료" accent padding="12px 16px">
        <KV k="바이너리" v="/usr/local/bin/moai" />
        <KV k="버전"     v="3.2.4" />
        <KV k="셸 보강"  v="zsh · fish (자동완성 등록)" />
        <div style={{ marginTop: 10, paddingTop: 10, borderTop: `1px dashed ${t.ruleSoft}`, fontFamily: FONT_SANS, fontSize: 12.5, color: t.dim, letterSpacing: "-0.02em" }}>
          다음 단계 <Tx c="accent" b mono>moai init my-app</Tx>
        </div>
      </Box>
    </div>

    <div style={{ marginTop: 12 }}><Prompt host="yuna@air" path="~" /></div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 2. install.ps1  (Windows PowerShell)
// ─────────────────────────────────────────────────────────
function ScreenInstallPs1() {
  const t = useTok();
  return <Term title="PowerShell — install.ps1" width={760} badge="pwsh · 7.4">
    <div style={{ fontFamily: FONT_MONO, fontSize: 13, marginBottom: 8 }}>
      <Tx c="info" b>PS</Tx> <Tx c="promptPath" b>C:\Users\yuna</Tx><Tx c="dim">{">"}</Tx>{" "}
      irm <Tx c="info">https://moai-adk.dev/install.ps1</Tx> | iex
    </div>

    <ThickBox accent padding="12px 16px">
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
        <div style={{ fontFamily: FONT_SANS, fontWeight: 800, fontSize: 14, color: t.accent, letterSpacing: "-0.03em" }}>
          MoAI-ADK Installer · Windows
        </div>
        <Pill kind="primary" solid>v3.2.4</Pill>
      </div>
    </ThickBox>

    <Section title="아키텍처 · 환경" />
    <CheckLine status="ok" label="플랫폼"   value="windows · amd64" />
    <CheckLine status="ok" label="PowerShell" value="7.4.5" />
    <CheckLine status="ok" label="실행정책" value="RemoteSigned" hint="현재 사용자 범위" />

    <Section title="다운로드" right="moai-adk_3.2.4_windows_amd64.zip · 19.7 MB" />
    <div style={{ marginTop: 4 }}><Progress value={100} label="아카이브" color={t.success} /></div>
    <div style={{ marginTop: 6 }}><Progress value={100} label="체크섬"   color={t.success} /></div>

    <div style={{ marginTop: 14 }}>
      <Box title="설치 완료" accent padding="12px 16px">
        <KV k="바이너리"   v="$env:LOCALAPPDATA\Programs\moai\moai.exe" />
        <KV k="버전"       v="3.2.4" />
        <KV k="PATH"       v="사용자 범위로 등록됨" />
        <div style={{ marginTop: 10, paddingTop: 10, borderTop: `1px dashed ${t.ruleSoft}`, fontFamily: FONT_SANS, fontSize: 12.5, color: t.dim, letterSpacing: "-0.02em" }}>
          새 PowerShell 세션에서 <Tx c="accent" b mono>moai init my-app</Tx>
        </div>
      </Box>
    </div>

    <div style={{ marginTop: 12, fontFamily: FONT_MONO, fontSize: 13 }}>
      <Tx c="info" b>PS</Tx> <Tx c="promptPath" b>C:\Users\yuna</Tx><Tx c="dim">{">"}</Tx> <Cursor />
    </div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 3. install.bat  (cmd 폴백)
// ─────────────────────────────────────────────────────────
function ScreenInstallBat() {
  const t = useTok();
  return <Term title="cmd.exe — install.bat" width={760} badge="cmd · 폴백">
    <div style={{ fontFamily: FONT_MONO, fontSize: 13, marginBottom: 8 }}>
      <Tx c="promptPath" b>C:\Users\yuna{">"}</Tx> install.bat
    </div>

    <ThickBox accent padding="12px 16px" color={t.warning}>
      <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
        <StatusIcon status="warn" />
        <div style={{ fontFamily: FONT_SANS }}>
          <div style={{ fontWeight: 800, fontSize: 13.5, color: t.warning, letterSpacing: "-0.025em" }}>오프라인 / 제한 환경 모드</div>
          <div style={{ fontSize: 12, color: t.dim, marginTop: 1 }}>PowerShell 차단 감지 · cmd.exe 폴백 사용</div>
        </div>
      </div>
    </ThickBox>

    <Section title="설치 단계" />
    <CheckLine status="ok"   label="바이너리 위치"     value=".\bin\moai.exe" />
    <CheckLine status="ok"   label="checksum.txt 검증" value="sha256 일치" />
    <CheckLine status="ok"   label="복사"              value="3 파일 · 18.4 MB" hint="%LOCALAPPDATA%\Programs\moai" />
    <CheckLine status="ok"   label="PATH 등록"         value="setx 명령" />

    <div style={{ marginTop: 14 }}>
      <Box title="사용 가능한 명령" padding="10px 14px">
        <KV k="moai init my-app" v="새 프로젝트 부트스트랩" kw={170} />
        <KV k="moai doctor"      v="환경 진단"           kw={170} />
        <KV k="moai version"     v="버전 확인"           kw={170} />
      </Box>
    </div>

    <div style={{ marginTop: 12, fontFamily: FONT_MONO, fontSize: 13 }}>
      <Tx c="promptPath" b>C:\Users\yuna{">"}</Tx> <Cursor />
    </div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 4. moai (배너 + 메인)
// ─────────────────────────────────────────────────────────
function ScreenBanner() {
  const t = useTok();
  return <Term title="moai — zsh" width={760} footer={<>
    <span><Tx c="dim">베타 테스터 빌드</Tx> · <Tx c="accent" b mono>v3.2.4</Tx></span>
    <span><Tx c="dim">텔레메트리 옵트인</Tx></span>
  </>}>
    <Prompt path="~/work" cmd="moai" />

    <div style={{
      marginTop: 8, marginBottom: 14, padding: "18px 20px",
      border: `1px solid ${t.rule}`, borderRadius: 10,
      background: `linear-gradient(180deg, ${t.accentSofter} 0%, transparent 100%)`,
    }}>
      <div style={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between", gap: 16 }}>
        <div>
          <pre style={{
            fontFamily: FONT_MONO, fontWeight: 700, fontSize: 13, lineHeight: 1.25,
            color: t.accent, margin: 0, letterSpacing: "0.02em",
          }}>{` __  __  __  ___    _
|  \\/  |/_/ /   |  (_)
| |\\/| |  | / /| | / /
| |  | |  |/ ___ |/ /
|_|  |_| /_/ |_/_/`}</pre>
          <div style={{ fontFamily: FONT_SANS, fontWeight: 800, fontSize: 16, color: t.fg, marginTop: 10, letterSpacing: "-0.03em" }}>
            모두의 Agentic Development Kit
          </div>
          <div style={{ fontFamily: FONT_SANS, fontSize: 12.5, color: t.dim, marginTop: 2, letterSpacing: "-0.02em" }}>
            SPEC → Plan → Implement → Sync · CX 7원칙 거버넌스
          </div>
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 6, alignItems: "flex-end" }}>
          <Pill kind="primary" solid>v3.2.4</Pill>
          <Pill kind="ok">go 1.23</Pill>
          <Pill kind="info">claude 1.0.18</Pill>
        </div>
      </div>
    </div>

    <Section title="자주 쓰는 명령" right="moai help · ?" />
    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "4px 24px" }}>
      <KV k="moai init"      v="새 프로젝트 부트스트랩" kw={120} />
      <KV k="moai doctor"    v="환경 · 설정 진단"      kw={120} />
      <KV k="moai status"    v="프로젝트 현황 요약"     kw={120} />
      <KV k="moai cc"        v="Claude Code 실행"     kw={120} />
      <KV k="moai loop"      v="자율 개발 루프 시작"    kw={120} />
      <KV k="moai sync"      v="문서 · 태그 동기화"     kw={120} />
    </div>

    <HelpBar items={[
      ["?", "도움말"],
      ["q", "종료"],
      ["v", "버전"],
      ["moai <cmd> --help", "상세 도움말"],
    ]} />
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 5. moai help
// ─────────────────────────────────────────────────────────
function ScreenHelp() {
  const t = useTok();
  const groups = [
    { title: "프로젝트", items: [
      ["init", "새 프로젝트 부트스트랩 (위저드)"],
      ["doctor", "환경 · 의존성 · 설정 진단"],
      ["status", "현재 프로젝트 요약"],
      ["update", "MoAI-ADK 자체 업데이트"],
      ["version", "버전 정보"],
    ]},
    { title: "런처", items: [
      ["cc", "Claude Code 실행 (브릿지 포함)"],
      ["statusline", "tmux/vim용 상태줄 출력"],
    ]},
    { title: "자율 개발", items: [
      ["loop", "Spec→Plan→Impl→Sync 루프"],
      ["spec", "스펙 카드 관리"],
      ["worktree", "워크트리 격리 작업"],
    ]},
    { title: "거버넌스", items: [
      ["constitution", "CX 7원칙 검사"],
      ["mx", "MX 앵커 · 그래프"],
      ["telemetry", "사용 통계 (옵트인)"],
    ]},
  ];
  return <Term title="moai help — zsh" width={760} scrollHint footer={<>
    <span><Tx c="dim">help · </Tx><Tx c="fg" b mono>moai &lt;명령&gt; --help</Tx></span>
    <span><Tx c="dim">v3.2.4</Tx></span>
  </>}>
    <Prompt path="~/work/my-app" cmd="moai help" />

    <div style={{ marginTop: 6, fontFamily: FONT_SANS, fontSize: 12.5, color: t.dim, letterSpacing: "-0.02em" }}>
      <Tx c="fg" b>moai-adk</Tx> · 모두의 Agentic Development Kit · 모듈식 명령 모음
    </div>

    {groups.map((g, i) => <React.Fragment key={i}>
      <Section title={g.title} />
      {g.items.map(([cmd, desc], j) => <div key={j} style={{
        display: "flex", alignItems: "baseline", gap: 12, padding: "3px 0",
        fontFamily: FONT_SANS, letterSpacing: "-0.02em",
      }}>
        <span style={{ minWidth: 130, fontFamily: FONT_MONO, fontSize: 12.5, color: t.accent, fontWeight: 700 }}>moai {cmd}</span>
        <span style={{ color: t.body, fontSize: 13 }}>{desc}</span>
      </div>)}
    </React.Fragment>)}

    <HelpBar items={[
      ["↑↓", "스크롤"], ["/", "검색"], ["q", "닫기"], ["enter", "선택"],
    ]} />
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 6. moai init  (huh wizard, step 3 / 6)
// ─────────────────────────────────────────────────────────
function ScreenInit() {
  const t = useTok();
  return <Term title="moai init — huh form" width={760} footer={<>
    <span><Tx c="accent" b>◆</Tx> 선택 · <Tx c="dim">↑↓ 이동 · enter 다음 · esc 취소</Tx></span>
    <span><Tx c="dim">3/6</Tx></span>
  </>}>
    <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 12 }}>
      <div>
        <div style={{ fontFamily: FONT_SANS, fontWeight: 800, fontSize: 14, color: t.accent, letterSpacing: "-0.03em" }}>
          새 프로젝트 부트스트랩
        </div>
        <div style={{ fontFamily: FONT_SANS, fontSize: 12, color: t.dim, marginTop: 1 }}>my-app · ~/work/my-app</div>
      </div>
      <Stepper current={3} total={6} />
    </div>

    <Box title="언어 · 런타임" accent padding="10px 8px" sub="프로젝트의 주 언어를 선택하세요. 이후 단계에서 추가할 수 있어요.">
      <RadioRow                  label="Go"          hint="1.23+" sub="bubbletea · cobra 친화. moai-adk 자체와 동일한 스택" />
      <RadioRow selected         label="TypeScript"  hint="Node 20 / Bun 1" sub="React · Next · CLI 모두 지원" />
      <RadioRow                  label="Python"      hint="3.11+" sub="poetry · uv 자동 감지" />
      <RadioRow                  label="Rust"        hint="stable" sub="cargo workspace 인식" />
      <RadioRow                  label="다국어 모노레포" sub="언어를 나중에 워크스페이스별로 지정" />
    </Box>

    <div style={{ marginTop: 10, padding: "10px 14px", borderRadius: 8, border: `1px dashed ${t.rule}`, fontFamily: FONT_SANS, fontSize: 12.5, color: t.dim, letterSpacing: "-0.02em" }}>
      <Tx c="info" b>i</Tx> 선택한 언어로 <Tx c="fg" b mono>.moai/config.yaml</Tx>의 <Tx mono>language</Tx> 필드와 SPEC 템플릿이 결정됩니다.
    </div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 7. moai doctor
// ─────────────────────────────────────────────────────────
function ScreenDoctor() {
  const t = useTok();
  return <Term title="moai doctor — zsh" width={760} footer={<>
    <span><Tx c="success" b>✓ 8</Tx> 통과 · <Tx c="warning" b>! 1</Tx> 주의 · <Tx c="danger" b>✗ 0</Tx> 실패</span>
    <span><Tx c="dim">14:32:08 · 1.2s</Tx></span>
  </>}>
    <Prompt path="~/work/my-app" cmd="moai doctor" />

    <Section title="시스템" sub="필수 런타임 · 도구 체인" />
    <CheckLine status="ok" label="Go"          value="1.23.4"  hint="권장 1.23+" />
    <CheckLine status="ok" label="Git"         value="2.43.0"  hint="LFS 사용 가능" />
    <CheckLine status="ok" label="Claude Code" value="1.0.18"  hint="브릿지 활성" />
    <CheckLine status="ok" label="GitHub CLI"  value="2.62.0"  hint="인증 ✓ yuna-afamily" />

    <Section title="MoAI-ADK" />
    <CheckLine status="ok"   label=".moai/config.yaml" value="schema v3" />
    <CheckLine status="ok"   label="훅 · 슬래시 명령"   value="9개 · 14개" />
    <CheckLine status="ok"   label="MX 앵커 인덱스"   value="412개"  hint="0.3s 캐시" />
    <CheckLine status="warn" label="Glamour 캐시"    value="갱신 필요" hint="moai sync --themes 권장" />
    <CheckLine status="ok"   label="텔레메트리"      value="옵트인 · 익명" />

    <div style={{ marginTop: 14 }}>
      <Box title="요약" accent padding="10px 14px">
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
          <Pill kind="ok">8 통과</Pill>
          <Pill kind="warn">1 주의</Pill>
          <Pill kind="err">0 실패</Pill>
          <Pill kind="neutral">14:32:08</Pill>
        </div>
        <div style={{ marginTop: 10, fontFamily: FONT_SANS, fontSize: 12.5, color: t.dim, letterSpacing: "-0.02em" }}>
          다음 단계 <Tx c="accent" b mono>moai sync --themes</Tx> 로 Glamour 캐시를 갱신하세요.
        </div>
      </Box>
    </div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 8. moai status
// ─────────────────────────────────────────────────────────
function ScreenStatus() {
  const t = useTok();
  return <Term title="moai status — zsh" width={760}>
    <Prompt path="~/work/my-app" branch="feat/SPEC-AUTH-007" dirty cmd="moai status" />

    <Section title="프로젝트" right="my-app · v0.4.2-rc1" />
    <KV k="언어"   v="TypeScript · pnpm 9.4" kw={110} />
    <KV k="브랜치" v="feat/SPEC-AUTH-007" kw={110} vColor={t.warning} />
    <KV k="커밋"   v="8 ahead · 0 behind origin/main" kw={110} />
    <KV k="워크트리" v="2개 활성" kw={110} />

    <Section title="SPEC 카드" right="총 14개" />
    <div style={{ display: "flex", gap: 8, flexWrap: "wrap", padding: "4px 0" }}>
      <Pill kind="ok">완료 9</Pill>
      <Pill kind="primary">진행 3</Pill>
      <Pill kind="warn">초안 1</Pill>
      <Pill kind="neutral">백로그 1</Pill>
    </div>

    <Section title="진행 중" />
    <div style={{ display: "flex", flexDirection: "column", gap: 8, marginTop: 4 }}>
      {[
        { id: "SPEC-AUTH-007", title: "OAuth 토큰 회전 정책", phase: "구현", pct: 62 },
        { id: "SPEC-DOC-014",  title: "온보딩 가이드 동기화", phase: "동기화", pct: 88 },
        { id: "SPEC-API-022",  title: "캐시 헤더 정규화",     phase: "계획",   pct: 18 },
      ].map((s, i) => <div key={i} style={{
        display: "grid", gridTemplateColumns: "auto 1fr auto auto", gap: 12, alignItems: "center",
        padding: "8px 12px", border: `1px solid ${t.rule}`, borderRadius: 8,
      }}>
        <Pill kind="primary" mono>{s.id}</Pill>
        <span style={{ fontFamily: FONT_SANS, fontSize: 13, color: t.fg, letterSpacing: "-0.025em" }}>{s.title}</span>
        <Pill kind={s.pct > 80 ? "ok" : (s.pct < 30 ? "warn" : "info")}>{s.phase}</Pill>
        <div style={{ width: 140 }}><Progress value={s.pct} width={140} /></div>
      </div>)}
    </div>

    <HelpBar items={[
      ["enter", "카드 열기"], ["n", "새 SPEC"], ["s", "동기화"], ["q", "닫기"],
    ]} />
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 9. moai version
// ─────────────────────────────────────────────────────────
function ScreenVersion() {
  const t = useTok();
  return <Term title="moai version — zsh" width={760}>
    <Prompt path="~/work/my-app" cmd="moai version" />

    <div style={{ marginTop: 10 }}>
      <ThickBox accent padding="16px 20px">
        <div style={{ display: "grid", gridTemplateColumns: "auto 1fr auto", gap: 18, alignItems: "center" }}>
          <pre style={{ fontFamily: FONT_MONO, fontSize: 11, color: t.accent, margin: 0, lineHeight: 1.2 }}>{`__  __  __
|  \\/  |/_/
| |\\/| |
|_|  |_|`}</pre>
          <div>
            <div style={{ fontFamily: FONT_SANS, fontWeight: 800, fontSize: 18, letterSpacing: "-0.035em", color: t.fg }}>MoAI-ADK</div>
            <div style={{ fontFamily: FONT_MONO, fontSize: 13, color: t.accent, fontWeight: 700, marginTop: 2 }}>v3.2.4 · 2026-04-28</div>
            <div style={{ fontFamily: FONT_SANS, fontSize: 12, color: t.dim, marginTop: 4, letterSpacing: "-0.02em" }}>모두의AI · 베타 테스터 빌드</div>
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 6, alignItems: "flex-end" }}>
            <Pill kind="ok">최신</Pill>
            <Pill kind="info">go 1.23</Pill>
          </div>
        </div>
      </ThickBox>
    </div>

    <Section title="구성 요소" />
    <KV k="commit"      v="3e4b7c2a91f8 · main" kw={140} />
    <KV k="빌드 시각"    v="2026-04-28 09:14 KST" kw={140} />
    <KV k="플랫폼"      v="darwin / arm64" kw={140} />
    <KV k="bubbletea"   v="v1.3.10" kw={140} />
    <KV k="lipgloss"    v="v1.1.1" kw={140} />
    <KV k="huh · glamour" v="v0.8.0 · v0.10.0" kw={140} />
    <KV k="claude code"   v="v1.0.18" kw={140} />
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 10. moai update
// ─────────────────────────────────────────────────────────
function ScreenUpdate() {
  const t = useTok();
  return <Term title="moai update — zsh" width={760}>
    <Prompt path="~/work/my-app" cmd="moai update" />

    <div style={{ marginTop: 8, marginBottom: 12 }}>
      <Box title="업데이트 가능" accent padding="12px 16px" badge={<Pill kind="primary" solid>v3.3.0</Pill>}>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
          <div>
            <div style={{ fontFamily: FONT_SANS, fontSize: 13, color: t.fg, letterSpacing: "-0.025em" }}>
              <Tx c="dim">현재</Tx> <Tx mono b>v3.2.4</Tx>{" → "}<Tx c="accent" b mono>v3.3.0</Tx>
            </div>
            <div style={{ fontFamily: FONT_SANS, fontSize: 12, color: t.dim, marginTop: 2, letterSpacing: "-0.02em" }}>
              호환성 · 마이너 · MX 앵커 그래프 v2 · huh 폼 테마 추가
            </div>
          </div>
          <Pill kind="info">62일 만에 갱신</Pill>
        </div>
      </Box>
    </div>

    <Section title="다운로드 진행" />
    <div style={{ marginTop: 4, display: "flex", flexDirection: "column", gap: 6 }}>
      <Progress value={100} label="아카이브" color={t.success} />
      <Progress value={100} label="체크섬" color={t.success} />
      <Progress value={64}  label="설치" color={t.accent} />
    </div>

    <div style={{ marginTop: 12, fontFamily: FONT_SANS, fontSize: 12.5, color: t.dim, letterSpacing: "-0.02em" }}>
      <Spinner label="설정 마이그레이션 중 · .moai/config.yaml v3 → v3.1" />
    </div>

    <div style={{ marginTop: 14 }}>
      <ThickBox accent padding="10px 14px" color={t.success}>
        <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
          <StatusIcon status="ok" />
          <span style={{ fontFamily: FONT_SANS, fontSize: 13, color: t.success, fontWeight: 700, letterSpacing: "-0.025em" }}>
            안전 모드 사용 중
          </span>
          <Tx c="dim" style={{ fontSize: 12 }}>· 백업 <Tx mono>~/.moai/backups/3.2.4-2604.tar.gz</Tx></Tx>
        </div>
      </ThickBox>
    </div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 11. moai cc  (Claude Code 부트스트랩)
// ─────────────────────────────────────────────────────────
function ScreenCC() {
  const t = useTok();
  return <Term title="moai cc — zsh" width={760}>
    <Prompt path="~/work/my-app" cmd="moai cc" />

    <Section title="Claude Code 브릿지" sub="MoAI 컨텍스트 주입 → Claude Code 인계" />
    <CheckLine status="ok"   label="구성 파일 검증"     value=".moai/config.yaml" />
    <CheckLine status="ok"   label="훅 · 슬래시 명령"   value="9 · 14 등록" />
    <CheckLine status="ok"   label="MX 앵커 그래프"    value="412 노드 · 0.28s" />
    <CheckLine status="ok"   label="MCP 서버"          value="filesystem · git · github" />
    <CheckLine status="active" label="Claude 1.0.18 부트스트랩" value="ws://127.0.0.1:5179" />

    <div style={{ marginTop: 12 }}>
      <Spinner label="컨텍스트 패킹 중 · 8.4 MB / 12.0 MB" />
    </div>
    <div style={{ marginTop: 8 }}>
      <Progress value={70} label="패킹" />
    </div>

    <div style={{ marginTop: 14 }}>
      <Box title="실행 정보" padding="10px 14px">
        <KV k="세션 ID"   v="cc-3e4b7c2a91f" kw={120} />
        <KV k="모델"      v="claude-sonnet-4.5 · 200k" kw={120} />
        <KV k="스코프"    v="my-app · feat/SPEC-AUTH-007" kw={120} />
        <KV k="안전 모드" v="ON · auto-edit OFF" kw={120} vColor={t.success} />
      </Box>
    </div>

    <HelpBar items={[
      ["ctrl+c", "취소"], ["ctrl+r", "재시작"], ["?", "단축키"],
    ]} />
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 12. moai loop start
// ─────────────────────────────────────────────────────────
function ScreenLoop() {
  const t = useTok();
  const phases = [
    { name: "Spec",   status: "ok",     desc: "SPEC-AUTH-007 카드 동기화" },
    { name: "Plan",   status: "ok",     desc: "8개 작업 · 3개 위험" },
    { name: "Implement", status: "active", desc: "5/8 작업 완료 · 테스트 통과" },
    { name: "Sync",   status: "pending", desc: "문서 · 태그 동기화 대기" },
  ];
  return <Term title="moai loop start — bubbletea" width={760} footer={<>
    <span><Tx c="accent" b>●</Tx> 자율 모드 · <Tx c="dim">14:08:22부터 1시간 12분 경과</Tx></span>
    <span><Tx c="dim">p 일시정지 · q 종료</Tx></span>
  </>}>
    <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 12 }}>
      <div>
        <div style={{ fontFamily: FONT_SANS, fontWeight: 800, fontSize: 14, color: t.accent, letterSpacing: "-0.03em" }}>
          자율 개발 루프
        </div>
        <div style={{ fontFamily: FONT_SANS, fontSize: 12, color: t.dim, marginTop: 2, letterSpacing: "-0.02em" }}>
          SPEC-AUTH-007 · OAuth 토큰 회전 정책
        </div>
      </div>
      <Pill kind="primary" solid>실행 중</Pill>
    </div>

    {/* phase strip */}
    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr 1fr", gap: 10 }}>
      {phases.map((p, i) => <div key={i} style={{
        padding: "10px 12px", borderRadius: 8,
        border: `1px solid ${p.status === "active" ? t.accent : t.rule}`,
        background: p.status === "active" ? t.accentSofter : "transparent",
      }}>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
          <span style={{ fontFamily: FONT_SANS, fontSize: 12, color: t.dim, letterSpacing: "-0.02em" }}>{i + 1}</span>
          <StatusIcon status={p.status} />
        </div>
        <div style={{ fontFamily: FONT_SANS, fontWeight: 800, fontSize: 13, color: t.fg, marginTop: 2, letterSpacing: "-0.025em" }}>{p.name}</div>
        <div style={{ fontFamily: FONT_SANS, fontSize: 11.5, color: t.dim, marginTop: 2, letterSpacing: "-0.02em" }}>{p.desc}</div>
      </div>)}
    </div>

    <Section title="현재 작업" right="5/8" />
    <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
      {[
        { s: "ok",     n: "토큰 모델 · refresh 필드 추가" },
        { s: "ok",     n: "회전 정책 인터페이스 정의" },
        { s: "ok",     n: "메모리 어댑터 구현" },
        { s: "ok",     n: "Redis 어댑터 구현" },
        { s: "ok",     n: "단위 테스트 (78개)" },
        { s: "active", n: "통합 테스트 작성 중 · 12/18" },
        { s: "pending", n: "문서 갱신" },
        { s: "pending", n: "MX 앵커 동기화" },
      ].map((x, i) => <div key={i} style={{ display: "flex", gap: 10, alignItems: "center" }}>
        <StatusIcon status={x.s} />
        <span style={{ fontFamily: FONT_SANS, fontSize: 13, color: x.s === "pending" ? t.dim : t.fg, letterSpacing: "-0.02em" }}>{x.n}</span>
      </div>)}
    </div>

    <div style={{ marginTop: 12 }}>
      <Progress value={62} label="전체 진행" />
    </div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 13. moai spec view
// ─────────────────────────────────────────────────────────
function ScreenSpec() {
  const t = useTok();
  return <Term title="moai spec view SPEC-AUTH-007 — glamour" width={760}>
    <Prompt path="~/work/my-app" cmd="moai spec view SPEC-AUTH-007" />

    <div style={{ marginTop: 8, padding: "14px 18px", border: `1px solid ${t.rule}`, borderRadius: 10 }}>
      <div style={{ display: "flex", gap: 10, alignItems: "baseline", justifyContent: "space-between", marginBottom: 6 }}>
        <div style={{ display: "flex", gap: 8, alignItems: "baseline" }}>
          <Pill kind="primary" mono>SPEC-AUTH-007</Pill>
          <span style={{ fontFamily: FONT_SANS, fontWeight: 800, fontSize: 16, color: t.fg, letterSpacing: "-0.03em" }}>
            OAuth 토큰 회전 정책
          </span>
        </div>
        <Pill kind="info">구현 단계</Pill>
      </div>

      {/* glamour-style metadata */}
      <div style={{ display: "flex", gap: 16, fontFamily: FONT_SANS, fontSize: 12, color: t.dim, letterSpacing: "-0.02em", marginBottom: 12, flexWrap: "wrap" }}>
        <span><Tx c="dim">소유</Tx> <Tx c="fg" b>@yuna</Tx></span>
        <span><Tx c="dim">생성</Tx> <Tx mono>2026-04-12</Tx></span>
        <span><Tx c="dim">의존</Tx> <Tx c="fg" mono>SPEC-AUTH-002 · SPEC-API-014</Tx></span>
        <span><Tx c="dim">CX</Tx> <Tx c="success" b>7/7</Tx></span>
      </div>

      {/* glamour rendered markdown */}
      <div style={{ fontFamily: FONT_SANS, fontSize: 13.5, color: t.body, lineHeight: 1.65, letterSpacing: "-0.02em" }}>
        <div style={{ display: "flex", gap: 8, alignItems: "baseline", marginTop: 4 }}>
          <span style={{ color: t.accent, fontWeight: 800 }}>##</span>
          <span style={{ color: t.fg, fontWeight: 800, fontSize: 14 }}>배경</span>
        </div>
        <p style={{ margin: "4px 0 12px", color: t.body }}>
          기존 발급 토큰의 수명이 30일로 길어 침해 시 위험이 큽니다. 회전 정책을 도입해 단명 액세스 토큰과 장기 리프레시 토큰을 분리합니다.
        </p>

        <div style={{ display: "flex", gap: 8, alignItems: "baseline" }}>
          <span style={{ color: t.accent, fontWeight: 800 }}>##</span>
          <span style={{ color: t.fg, fontWeight: 800, fontSize: 14 }}>요구사항</span>
        </div>
        <ul style={{ margin: "4px 0 12px", paddingLeft: 18, color: t.body }}>
          <li>액세스 토큰 수명 <Tx c="info" mono b>15분</Tx></li>
          <li>리프레시 토큰 수명 <Tx c="info" mono b>14일</Tx>, 1회 사용 후 폐기</li>
          <li>도용 감지 시 동일 패밀리 전체 무효화</li>
        </ul>

        <div style={{ display: "flex", gap: 8, alignItems: "baseline" }}>
          <span style={{ color: t.accent, fontWeight: 800 }}>##</span>
          <span style={{ color: t.fg, fontWeight: 800, fontSize: 14 }}>인수 기준</span>
        </div>
        <div style={{ marginTop: 6, display: "flex", flexDirection: "column", gap: 4 }}>
          <CheckRow checked label="78개 단위 테스트 통과" />
          <CheckRow checked label="Redis · 메모리 어댑터 동등성 검증" />
          <CheckRow label="통합 테스트 18개 통과" hint="진행 중 12/18" />
          <CheckRow label="MX 앵커 · 문서 동기화" />
        </div>
      </div>
    </div>

    <HelpBar items={[
      ["e", "편집"], ["c", "댓글"], ["enter", "Plan 보기"], ["q", "닫기"],
    ]} />
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 14. moai worktree list
// ─────────────────────────────────────────────────────────
function ScreenWorktree() {
  const t = useTok();
  const rows = [
    { active: true,  branch: "feat/SPEC-AUTH-007", path: "~/work/my-app",                      ahead: 8, behind: 0, age: "2일",   status: "더티" },
    { active: false, branch: "feat/SPEC-DOC-014",  path: "~/work/.worktrees/spec-doc-014",     ahead: 3, behind: 1, age: "5시간", status: "클린" },
    { active: false, branch: "fix/SPEC-API-022",   path: "~/work/.worktrees/spec-api-022",     ahead: 1, behind: 2, age: "31분",  status: "더티" },
  ];
  return <Term title="moai worktree list — bubbles/table" width={760}>
    <Prompt path="~/work/my-app" cmd="moai worktree list" />

    <div style={{ marginTop: 10, border: `1px solid ${t.rule}`, borderRadius: 10, overflow: "hidden" }}>
      {/* header */}
      <div style={{
        display: "grid", gridTemplateColumns: "auto 1fr 1.4fr auto auto auto",
        gap: 14, padding: "10px 14px",
        background: t.panel, borderBottom: `1px solid ${t.rule}`,
        fontFamily: FONT_SANS, fontSize: 11.5, color: t.dim, fontWeight: 700,
        letterSpacing: "0.05em", textTransform: "uppercase",
      }}>
        <span style={{ width: 14 }}> </span>
        <span>브랜치</span>
        <span>경로</span>
        <span>ahead/behind</span>
        <span>나이</span>
        <span>상태</span>
      </div>
      {rows.map((r, i) => <div key={i} style={{
        display: "grid", gridTemplateColumns: "auto 1fr 1.4fr auto auto auto",
        gap: 14, padding: "9px 14px", alignItems: "center",
        background: r.active ? t.accentSofter : "transparent",
        borderTop: i > 0 ? `1px solid ${t.ruleSoft}` : "none",
        fontFamily: FONT_SANS, fontSize: 12.5, letterSpacing: "-0.02em",
      }}>
        <span style={{ color: r.active ? t.accent : t.faint, fontFamily: FONT_MONO, fontWeight: 800 }}>{r.active ? "●" : "○"}</span>
        <span style={{ fontFamily: FONT_MONO, color: r.active ? t.accent : t.fg, fontWeight: r.active ? 700 : 500, fontSize: 12.5 }}>{r.branch}</span>
        <span style={{ fontFamily: FONT_MONO, color: t.dim, fontSize: 12 }}>{r.path}</span>
        <span style={{ fontFamily: FONT_MONO, fontSize: 12 }}>
          <Tx c="success" b>↑{r.ahead}</Tx> <Tx c={r.behind ? "warning" : "dim"} b>↓{r.behind}</Tx>
        </span>
        <span style={{ color: t.dim, fontSize: 12 }}>{r.age}</span>
        <Pill kind={r.status === "클린" ? "ok" : "warn"}>{r.status}</Pill>
      </div>)}
    </div>

    <HelpBar items={[
      ["enter", "이동"], ["a", "추가"], ["d", "제거"], ["s", "동기화"], ["q", "닫기"],
    ]} />
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 15. moai constitution check  (CX 7원칙)
// ─────────────────────────────────────────────────────────
function ScreenConstitution() {
  const t = useTok();
  const principles = [
    { n: 1, title: "사양 우선",       score: 7, of: 7, note: "모든 변경은 SPEC 카드에서 출발" },
    { n: 2, title: "역추적 가능성",    score: 7, of: 7, note: "MX 앵커 412개 · 누락 0" },
    { n: 3, title: "테스트 주도",      score: 7, of: 7, note: "커버리지 91.4% · 회귀 0" },
    { n: 4, title: "안전 자율성",      score: 6, of: 7, note: "auto-edit 1건 검토 필요" },
    { n: 5, title: "거버넌스 투명성",  score: 7, of: 7, note: "변경 로그 · CR 동기화" },
    { n: 6, title: "현지화 · 접근성",  score: 7, of: 7, note: "ko · en 동기화" },
    { n: 7, title: "지속 가능성",      score: 7, of: 7, note: "에너지 · 의존성 예산 준수" },
  ];
  const total = principles.reduce((a, b) => a + b.score, 0);
  return <Term title="moai constitution check — bubbles/list" width={760}>
    <Prompt path="~/work/my-app" cmd="moai constitution check" />

    <div style={{ marginTop: 8, marginBottom: 12 }}>
      <ThickBox accent padding="14px 18px">
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 12 }}>
          <div>
            <div style={{ fontFamily: FONT_SANS, fontWeight: 800, fontSize: 14, color: t.accent, letterSpacing: "-0.03em" }}>CX 7원칙 점검</div>
            <div style={{ fontFamily: FONT_SANS, fontSize: 12, color: t.dim, marginTop: 2, letterSpacing: "-0.02em" }}>
              <Tx mono>moai-adk</Tx>는 모두의AI 거버넌스 7원칙을 따릅니다.
            </div>
          </div>
          <div style={{ textAlign: "right" }}>
            <div style={{ fontFamily: FONT_MONO, fontSize: 22, fontWeight: 800, color: t.accent }}>{total}/49</div>
            <div style={{ fontFamily: FONT_SANS, fontSize: 11, color: t.dim }}>총합</div>
          </div>
        </div>
      </ThickBox>
    </div>

    <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
      {principles.map((p, i) => <div key={i} style={{
        display: "grid", gridTemplateColumns: "auto 1fr auto auto", gap: 12, alignItems: "center",
        padding: "9px 12px", border: `1px solid ${t.rule}`, borderRadius: 8,
      }}>
        <span style={{ fontFamily: FONT_MONO, fontSize: 12, color: t.dim, fontWeight: 700, width: 24 }}>0{p.n}</span>
        <div>
          <div style={{ fontFamily: FONT_SANS, fontWeight: 700, fontSize: 13, color: t.fg, letterSpacing: "-0.025em" }}>{p.title}</div>
          <div style={{ fontFamily: FONT_SANS, fontSize: 12, color: t.dim, marginTop: 1, letterSpacing: "-0.02em" }}>{p.note}</div>
        </div>
        <div style={{ width: 100 }}>
          <Progress value={(p.score / p.of) * 100} percent={false} width={100} />
        </div>
        <Pill kind={p.score === p.of ? "ok" : "warn"} mono>{p.score}/{p.of}</Pill>
      </div>)}
    </div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 16. moai mx graph
// ─────────────────────────────────────────────────────────
function ScreenMx() {
  const t = useTok();
  // a small ascii-ish graph using boxes + lines
  return <Term title="moai mx graph SPEC-AUTH-007 — bubbletea" width={760}>
    <Prompt path="~/work/my-app" cmd="moai mx graph SPEC-AUTH-007" />

    <Section title="MX 앵커 그래프" sub="SPEC → Plan → Code → Test → Doc 추적" />

    <div style={{
      marginTop: 8, padding: "20px 24px",
      border: `1px solid ${t.rule}`, borderRadius: 10,
      background: t.panel,
    }}>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr 1fr", gap: 16, alignItems: "center" }}>
        {[
          { label: "SPEC",   id: "AUTH-007",    n: 1, kind: "primary" },
          { label: "Plan",   id: "8 작업",       n: 8, kind: "info" },
          { label: "Code",   id: "12 파일",      n: 12, kind: "info" },
          { label: "Test",   id: "78 + 12 IT",  n: 90, kind: "ok" },
        ].map((node, i) => <React.Fragment key={i}>
          <div style={{
            border: `1.5px solid ${t[node.kind === "primary" ? "accent" : (node.kind === "ok" ? "success" : "info")]}`,
            borderRadius: 8, padding: "10px 12px",
            background: t.bg, position: "relative",
          }}>
            <div style={{ fontFamily: FONT_SANS, fontSize: 11, color: t.dim, fontWeight: 700, letterSpacing: "0.06em", textTransform: "uppercase" }}>{node.label}</div>
            <div style={{ fontFamily: FONT_MONO, fontSize: 13, fontWeight: 700, color: t.fg, marginTop: 4 }}>{node.id}</div>
            <div style={{ fontFamily: FONT_MONO, fontSize: 11, color: t.accent, marginTop: 2 }}>{node.n}개 앵커</div>
          </div>
        </React.Fragment>)}
      </div>
      {/* arrows row */}
      <div style={{
        display: "grid", gridTemplateColumns: "1fr 1fr 1fr 1fr", gap: 16,
        marginTop: -8, marginBottom: -8, padding: "12px 0",
      }}>
        {["→","→","→",""].map((a, i) => <div key={i} style={{ textAlign: "right", paddingRight: 8, fontFamily: FONT_MONO, color: t.faint, fontSize: 13 }}>{a}</div>)}
      </div>

      <div style={{ marginTop: 6, display: "flex", justifyContent: "center" }}>
        <div style={{
          border: `1.5px solid ${t.success}`, borderRadius: 8, padding: "10px 14px",
          background: t.bg, minWidth: 200, textAlign: "center",
        }}>
          <div style={{ fontFamily: FONT_SANS, fontSize: 11, color: t.dim, fontWeight: 700, letterSpacing: "0.06em", textTransform: "uppercase" }}>Doc</div>
          <div style={{ fontFamily: FONT_MONO, fontSize: 13, fontWeight: 700, color: t.fg, marginTop: 4 }}>4개 가이드</div>
          <div style={{ fontFamily: FONT_MONO, fontSize: 11, color: t.success, marginTop: 2 }}>모두 동기화 ✓</div>
        </div>
      </div>
    </div>

    <div style={{ marginTop: 14, display: "flex", gap: 8, flexWrap: "wrap" }}>
      <Pill kind="ok">앵커 누락 0</Pill>
      <Pill kind="info">총 412개</Pill>
      <Pill kind="primary">갱신 0.28s</Pill>
      <Pill kind="neutral">캐시 적중 96%</Pill>
    </div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 17. moai telemetry report
// ─────────────────────────────────────────────────────────
function ScreenTelemetry() {
  const t = useTok();
  // sparkline
  const data = [3, 7, 4, 9, 6, 11, 8, 14, 12, 18, 15, 21];
  const max = Math.max(...data);
  return <Term title="moai telemetry report — bubbles/table" width={760}>
    <Prompt path="~/work/my-app" cmd="moai telemetry report --week" />

    <div style={{ marginTop: 8, marginBottom: 12, padding: "12px 16px", border: `1px solid ${t.rule}`, borderRadius: 10 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 10 }}>
        <div>
          <div style={{ fontFamily: FONT_SANS, fontSize: 12, color: t.dim, letterSpacing: "-0.02em" }}>지난 7일 명령 호출</div>
          <div style={{ fontFamily: FONT_MONO, fontSize: 22, fontWeight: 800, color: t.fg, marginTop: 2 }}>148<Tx c="dim" style={{ fontSize: 14, fontWeight: 500 }}> 회</Tx></div>
        </div>
        <div style={{ display: "flex", alignItems: "flex-end", gap: 3, height: 40 }}>
          {data.map((v, i) => <div key={i} style={{
            width: 8, height: `${(v / max) * 100}%`,
            background: i === data.length - 1 ? t.accent : t.accentSoft,
            borderRadius: 2,
          }} />)}
        </div>
        <div>
          <Pill kind="ok">+38% w/w</Pill>
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 12, paddingTop: 10, borderTop: `1px solid ${t.ruleSoft}` }}>
        {[
          { k: "loop start",  v: "42", d: "+18%" },
          { k: "spec view",   v: "31", d: "+9%" },
          { k: "doctor",      v: "22", d: "−4%" },
        ].map((m, i) => <div key={i}>
          <div style={{ fontFamily: FONT_SANS, fontSize: 12, color: t.dim, letterSpacing: "-0.02em" }}>{m.k}</div>
          <div style={{ fontFamily: FONT_MONO, fontSize: 18, fontWeight: 800, color: t.fg, marginTop: 2 }}>{m.v}</div>
          <div style={{ fontFamily: FONT_MONO, fontSize: 11, color: m.d.startsWith("−") ? t.danger : t.success, marginTop: 1 }}>{m.d}</div>
        </div>)}
      </div>
    </div>

    <Section title="익명 통계 · 옵트인" sub="기기 식별자 · 경로 · 코드 내용은 수집되지 않습니다" />
    <CheckLine status="ok"   label="옵트인 상태"     value="ON · 익명 ID rk-2604" />
    <CheckLine status="ok"   label="전송 빈도"      value="일 1회 · 최대 8KB" />
    <CheckLine status="info" label="전송된 카테고리" value="명령 · 에러 코드 · 지속 시간" />
    <CheckLine status="info" label="다음 전송"       value="04-29 03:00 KST" />

    <HelpBar items={[
      ["o", "옵트아웃"], ["d", "데이터 보기"], ["q", "닫기"],
    ]} />
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 18. statusline (tmux/vim 임베드)
// ─────────────────────────────────────────────────────────
function ScreenStatusline() {
  const t = useTok();
  // simulate a tmux right-status
  const segs = [
    { bg: t.accent,    fg: "#fff",          text: "MoAI" },
    { bg: t.panel,     fg: t.fg,            text: "feat/SPEC-AUTH-007" },
    { bg: t.warningSoft, fg: t.warning,     text: "✗ 3" },
    { bg: t.successSoft, fg: t.success,     text: "↑8" },
    { bg: t.infoSoft,    fg: t.info,        text: "loop · 5/8" },
    { bg: t.accentSofter, fg: t.accent,     text: "claude 1.0.18" },
  ];
  return <Term title="moai statusline — tmux 미리보기" width={760}>
    <Prompt path="~/work/my-app" cmd="moai statusline --preview tmux" />

    <Section title="tmux 우측 상태줄" />
    <div style={{
      display: "flex", marginTop: 8, padding: 0,
      border: `1px solid ${t.rule}`, borderRadius: 8, overflow: "hidden",
      fontFamily: FONT_MONO, fontSize: 12,
    }}>
      {segs.map((s, i) => <div key={i} style={{
        background: s.bg, color: s.fg, padding: "6px 12px",
        fontWeight: 700, letterSpacing: "-0.005em",
        borderRight: i < segs.length - 1 ? `1px solid ${t.bg}` : "none",
      }}>{s.text}</div>)}
    </div>

    <Section title="vim · airline 통합" />
    <div style={{
      marginTop: 6, padding: "6px 0", borderRadius: 8,
      background: t.panel, border: `1px solid ${t.rule}`,
      display: "flex", overflow: "hidden",
    }}>
      <div style={{ background: t.accent, color: "#fff", padding: "6px 14px", fontFamily: FONT_MONO, fontWeight: 800, fontSize: 12 }}>NORMAL</div>
      <div style={{ background: t.accentSoft, color: t.accent, padding: "6px 12px", fontFamily: FONT_MONO, fontWeight: 700, fontSize: 12 }}>moai · AUTH-007</div>
      <div style={{ flex: 1, padding: "6px 12px", color: t.dim, fontFamily: FONT_MONO, fontSize: 12 }}>internal/auth/rotate.go</div>
      <div style={{ background: t.warningSoft, color: t.warning, padding: "6px 12px", fontFamily: FONT_MONO, fontWeight: 700, fontSize: 12 }}>3 ⚠</div>
      <div style={{ background: t.accent, color: "#fff", padding: "6px 12px", fontFamily: FONT_MONO, fontWeight: 800, fontSize: 12 }}>62%</div>
    </div>

    <div style={{ marginTop: 14 }}>
      <Box title="설정" padding="10px 14px">
        <KV k="셸"   v="zsh prompt: %1~ $(moai statusline)" kw={70} />
        <KV k="tmux" v="set -g status-right '#(moai statusline -t tmux)'" kw={70} />
        <KV k="vim"  v="airline_section_b = 'moai#statusline()'" kw={70} />
      </Box>
    </div>
  </Term>;
}

// ─────────────────────────────────────────────────────────
// 19. error  (실패 케이스)
// ─────────────────────────────────────────────────────────
function ScreenError() {
  const t = useTok();
  return <Term title="moai loop start — 에러" width={760}>
    <Prompt path="~/work/my-app" branch="feat/SPEC-AUTH-007" dirty cmd="moai loop start" />

    <div style={{ marginTop: 10 }}>
      <ThickBox color={t.danger} padding="14px 18px">
        <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 8 }}>
          <StatusIcon status="err" />
          <div style={{ fontFamily: FONT_SANS, fontWeight: 800, fontSize: 14, color: t.danger, letterSpacing: "-0.03em" }}>
            거버넌스 위반 · 루프 시작 거부
          </div>
        </div>
        <div style={{ fontFamily: FONT_SANS, fontSize: 13, color: t.body, letterSpacing: "-0.02em", lineHeight: 1.55 }}>
          <Tx mono c="danger" b>SPEC-AUTH-007</Tx>의 인수 기준 중 일부가 누락되어 있어요. CX 1원칙(사양 우선)에 따라 자율 루프를 시작할 수 없습니다.
        </div>
      </ThickBox>
    </div>

    <Section title="누락 항목" />
    <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
      <CheckLine status="err"  label="인수 기준 #3 · 통합 테스트" value="기대 결과 미정의" />
      <CheckLine status="err"  label="MX 앵커 · 문서"          value="3개 미연결" />
      <CheckLine status="warn" label="플랜 검토자"               value="배정되지 않음" hint="권장 1인 이상" />
    </div>

    <div style={{ marginTop: 14 }}>
      <Box title="제안된 다음 단계" accent padding="10px 14px">
        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          {[
            ["moai spec edit SPEC-AUTH-007", "인수 기준 보강"],
            ["moai mx link SPEC-AUTH-007 docs/auth/rotate.md", "앵커 연결"],
            ["moai plan review --assign @팀원",                "검토자 지정"],
          ].map(([cmd, desc], i) => <div key={i} style={{ display: "flex", gap: 12, alignItems: "baseline" }}>
            <Tx c="accent" b mono style={{ fontSize: 12.5 }}>{cmd}</Tx>
            <Tx c="dim" style={{ fontSize: 12.5 }}>· {desc}</Tx>
          </div>)}
        </div>
      </Box>
    </div>

    <HelpBar items={[
      ["e", "스펙 편집"], ["i", "무시 (위험)"], ["q", "닫기"],
    ]} />
  </Term>;
}

Object.assign(window, {
  ScreenInstallSh, ScreenInstallPs1, ScreenInstallBat,
  ScreenBanner, ScreenHelp, ScreenInit, ScreenDoctor, ScreenStatus, ScreenVersion, ScreenUpdate,
  ScreenCC, ScreenLoop, ScreenSpec, ScreenWorktree,
  ScreenConstitution, ScreenMx, ScreenTelemetry,
  ScreenStatusline, ScreenError,
});
