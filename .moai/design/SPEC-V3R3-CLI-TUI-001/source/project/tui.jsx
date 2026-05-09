/* MoAI-ADK TUI components — lipgloss/bubbles/huh visual language mapped to CSS */
/* eslint-disable */

const ThemeCtx = React.createContext("light");
const useTheme = () => React.useContext(ThemeCtx);

// ───────── Tokens (모두의AI deep teal core) ─────────
const TOK = {
  light: {
    chrome:        "#e8e6e0",
    chromeBorder:  "#bdbab2",
    titleBar:      "linear-gradient(180deg,#efece5 0%,#e1ddd3 100%)",
    bg:            "#fbfaf6",            // ivory terminal bg
    panel:         "#f3f3f3",
    fg:            "#0e1513",
    body:          "#1f2826",
    dim:           "#5b625f",
    faint:         "#8c918d",
    rule:          "#dcd9d2",
    ruleSoft:      "#ebe8e1",
    accent:        "#144a46",
    accentDeep:    "#0a2825",
    accentSoft:    "rgba(20,74,70,0.10)",
    accentSofter:  "rgba(20,74,70,0.05)",
    success:       "#0e7a6c",
    successSoft:   "rgba(14,122,108,0.12)",
    warning:       "#a86412",
    warningSoft:   "rgba(168,100,18,0.13)",
    danger:        "#b1432f",
    dangerSoft:    "rgba(177,67,47,0.12)",
    info:          "#1f6f72",
    infoSoft:      "rgba(31,111,114,0.12)",
    cursor:        "#144a46",
    selection:     "rgba(20,74,70,0.18)",
    promptArrow:   "#144a46",
    promptPath:    "#1f6f72",
    shadow:        "0 24px 48px -22px rgba(9,17,15,0.22), 0 1px 0 rgba(255,255,255,0.6) inset",
  },
  dark: {
    chrome:        "#0c1413",
    chromeBorder:  "#1c2624",
    titleBar:      "linear-gradient(180deg,#131b19 0%,#0a1110 100%)",
    bg:            "#0a110f",
    panel:         "#0f1816",
    fg:            "#eef2ef",
    body:          "#d8dedb",
    dim:           "#9aa3a0",
    faint:         "#6b7370",
    rule:          "#1c2624",
    ruleSoft:      "#152019",
    accent:        "#3eb3a4",
    accentDeep:    "#22938a",
    accentSoft:    "rgba(62,179,164,0.16)",
    accentSofter:  "rgba(62,179,164,0.07)",
    success:       "#3fcfa6",
    successSoft:   "rgba(63,207,166,0.14)",
    warning:       "#e3a14a",
    warningSoft:   "rgba(227,161,74,0.14)",
    danger:        "#ed7d6b",
    dangerSoft:    "rgba(237,125,107,0.15)",
    info:          "#5cc7c9",
    infoSoft:      "rgba(92,199,201,0.14)",
    cursor:        "#3eb3a4",
    selection:     "rgba(62,179,164,0.25)",
    promptArrow:   "#3eb3a4",
    promptPath:    "#5cc7c9",
    shadow:        "0 30px 60px -22px rgba(0,0,0,0.65), 0 1px 0 rgba(255,255,255,0.03) inset",
  },
};
const useTok = () => TOK[useTheme()];

// ───────── Type helpers ─────────
const FONT_SANS = '"Pretendard", system-ui, -apple-system, sans-serif';
const FONT_MONO = '"JetBrains Mono", ui-monospace, "SF Mono", Menlo, monospace';

// Inline color span
const Tx = ({ c, b, mono, children, style }) => {
  const t = useTok();
  return <span style={{
    color: c ? (t[c] || c) : "inherit",
    fontWeight: b ? 700 : "inherit",
    fontFamily: mono ? FONT_MONO : "inherit",
    ...style,
  }}>{children}</span>;
};

// ───────── lipgloss Border() — CSS rounded border box ─────────
function Box({
  title, titleAlign = "left", borderColor, accent = false, padding = "12px 16px",
  width, children, badge, footer, sub,
}) {
  const t = useTok();
  const bc = borderColor || (accent ? t.accent : t.rule);
  return (
    <div style={{
      border: `1px solid ${bc}`,
      borderRadius: 8,
      width: width,
      background: accent ? t.accentSofter : "transparent",
      position: "relative",
      marginTop: title ? 10 : 0,
    }}>
      {title ? (
        <div style={{
          position: "absolute", top: -10,
          left: titleAlign === "left" ? 14 : "50%",
          transform: titleAlign === "center" ? "translateX(-50%)" : undefined,
          background: t.bg, padding: "0 8px",
          fontFamily: FONT_SANS, fontWeight: 700, fontSize: 12,
          letterSpacing: "-0.02em", color: accent ? t.accent : t.dim,
          display: "flex", alignItems: "center", gap: 8,
        }}>
          <span>{title}</span>
          {badge ? <span>{badge}</span> : null}
        </div>
      ) : null}
      <div style={{ padding }}>
        {sub ? <div style={{
          fontFamily: FONT_SANS, fontSize: 12, color: t.dim,
          letterSpacing: "-0.02em", marginBottom: 10,
        }}>{sub}</div> : null}
        {children}
      </div>
      {footer ? (
        <div style={{
          borderTop: `1px solid ${t.ruleSoft}`,
          padding: "8px 16px", fontSize: 11.5,
          fontFamily: FONT_SANS, color: t.dim, letterSpacing: "-0.015em",
        }}>{footer}</div>
      ) : null}
    </div>
  );
}

// ───────── lipgloss "thick border" — accent-bordered card ─────────
function ThickBox({ title, children, accent, color, padding = "14px 18px" }) {
  const t = useTok();
  const c = color || t[accent ? "accent" : "rule"];
  return (
    <div style={{
      border: `2px solid ${c}`,
      borderRadius: 10,
      background: accent ? t.accentSofter : "transparent",
      padding: padding,
      position: "relative",
    }}>
      {title ? <div style={{
        fontFamily: FONT_SANS, fontWeight: 800, fontSize: 13,
        letterSpacing: "-0.025em", color: c, marginBottom: 8,
      }}>{title}</div> : null}
      {children}
    </div>
  );
}

// ───────── Tag / Pill ─────────
function Pill({ kind = "info", solid = false, children, mono = false }) {
  const t = useTok();
  const map = {
    info:    [t.info, t.infoSoft],
    ok:      [t.success, t.successSoft],
    warn:    [t.warning, t.warningSoft],
    err:     [t.danger, t.dangerSoft],
    primary: [t.accent, t.accentSoft],
    neutral: [t.dim, t.ruleSoft],
  };
  const [fg, bg] = map[kind] || map.info;
  return <span style={{
    background: solid ? fg : bg,
    color: solid ? (kind === "neutral" ? t.bg : "#fff") : fg,
    fontWeight: 700, fontSize: 11.5, letterSpacing: "-0.005em",
    padding: "2px 8px", borderRadius: 999,
    fontFamily: mono ? FONT_MONO : FONT_SANS,
    display: "inline-flex", alignItems: "center", gap: 4, lineHeight: 1.4,
    whiteSpace: "nowrap",
  }}>{children}</span>;
}

// Status icon
function StatusIcon({ status }) {
  const t = useTok();
  const map = {
    ok:    { c: t.success, ch: "✓" },
    warn:  { c: t.warning, ch: "!" },
    err:   { c: t.danger, ch: "✗" },
    info:  { c: t.info, ch: "·" },
    pending: { c: t.faint, ch: "○" },
    active: { c: t.accent, ch: "●" },
    arrow: { c: t.accent, ch: "→" },
  };
  const m = map[status] || map.info;
  return <span style={{
    display: "inline-flex", width: 16, height: 16, justifyContent: "center",
    alignItems: "center", color: m.c, fontWeight: 800, fontFamily: FONT_MONO,
    fontSize: 12,
  }}>{m.ch}</span>;
}

// Spinner (bubbles/spinner Dot style)
function Spinner({ label }) {
  const t = useTok();
  return <span style={{ color: t.accent, display: "inline-flex", alignItems: "center", gap: 6 }}>
    <span style={{
      display: "inline-block", width: 12, height: 12, borderRadius: "50%",
      border: `1.5px solid ${t.accent}`, borderTopColor: "transparent",
      animation: "spin 0.7s linear infinite",
    }} />
    {label ? <span style={{ fontFamily: FONT_SANS, fontSize: 12.5, color: t.fg, letterSpacing: "-0.02em" }}>{label}</span> : null}
  </span>;
}

// Progress bar (bubbles/progress with gradient)
function Progress({ value, total = 100, width = 320, label, percent = true, color }) {
  const t = useTok();
  const pct = Math.min(100, Math.round((value / total) * 100));
  const c = color || t.accent;
  return <div style={{ display: "flex", alignItems: "center", gap: 10, fontFamily: FONT_SANS }}>
    {label ? <span style={{ fontSize: 12, color: t.dim, letterSpacing: "-0.02em", minWidth: 70 }}>{label}</span> : null}
    <div style={{
      flex: 1, maxWidth: width, height: 6, borderRadius: 999,
      background: t.ruleSoft, overflow: "hidden",
    }}>
      <div style={{
        width: `${pct}%`, height: "100%",
        background: `linear-gradient(90deg, ${c} 0%, ${t.accentDeep} 100%)`,
        transition: "width .25s ease",
      }} />
    </div>
    {percent ? <span style={{
      fontFamily: FONT_MONO, fontSize: 11.5, fontWeight: 700, color: t.fg, minWidth: 36, textAlign: "right",
    }}>{pct}%</span> : null}
  </div>;
}

// Stepper dots (huh form progress)
function Stepper({ current, total }) {
  const t = useTok();
  return <div style={{ display: "inline-flex", gap: 6, alignItems: "center" }}>
    {Array.from({ length: total }).map((_, i) => (
      <span key={i} style={{
        width: i < current ? 18 : 6, height: 6, borderRadius: 3,
        background: i < current ? t.accent : (i === current ? t.accentDeep : t.rule),
        transition: "all .2s",
      }} />
    ))}
    <span style={{
      marginLeft: 8, fontFamily: FONT_MONO, fontSize: 11.5, color: t.dim,
    }}>{current}/{total}</span>
  </div>;
}

// huh radio option row
function RadioRow({ selected, label, hint, sub }) {
  const t = useTok();
  return <div style={{
    display: "flex", alignItems: "flex-start", gap: 10,
    padding: "8px 12px", borderRadius: 6,
    background: selected ? t.accentSoft : "transparent",
    borderLeft: `2px solid ${selected ? t.accent : "transparent"}`,
  }}>
    <span style={{
      color: selected ? t.accent : t.faint, fontFamily: FONT_MONO,
      fontWeight: 700, fontSize: 13, marginTop: 1,
    }}>{selected ? "◆" : "◇"}</span>
    <div style={{ flex: 1 }}>
      <div style={{
        fontFamily: FONT_SANS, fontWeight: selected ? 700 : 500, fontSize: 13.5,
        color: selected ? t.accent : t.fg, letterSpacing: "-0.025em",
      }}>{label} {hint ? <span style={{ color: t.dim, fontWeight: 400, fontSize: 12 }}>· {hint}</span> : null}</div>
      {sub ? <div style={{
        fontFamily: FONT_SANS, fontSize: 12, color: t.dim, marginTop: 2,
        letterSpacing: "-0.02em",
      }}>{sub}</div> : null}
    </div>
  </div>;
}

// Checkbox row
function CheckRow({ checked, label, hint }) {
  const t = useTok();
  return <div style={{ display: "flex", alignItems: "center", gap: 10, padding: "4px 0" }}>
    <span style={{
      width: 18, height: 18, borderRadius: 4,
      border: `1.5px solid ${checked ? t.accent : t.rule}`,
      background: checked ? t.accent : "transparent",
      color: "#fff", fontFamily: FONT_MONO, fontWeight: 800, fontSize: 11,
      display: "inline-flex", alignItems: "center", justifyContent: "center",
    }}>{checked ? "✓" : ""}</span>
    <span style={{ fontFamily: FONT_SANS, fontSize: 13, color: t.fg, letterSpacing: "-0.02em" }}>
      {label} {hint ? <Tx c="dim" style={{ fontSize: 12 }}>· {hint}</Tx> : null}
    </span>
  </div>;
}

// Key-value table row (lipgloss `renderKeyValueLines`)
function KV({ k, v, kw = 110, vColor }) {
  const t = useTok();
  return <div style={{
    display: "flex", alignItems: "baseline", gap: 12, padding: "3px 0",
    fontFamily: FONT_SANS, letterSpacing: "-0.02em", fontSize: 13,
  }}>
    <span style={{ color: t.dim, minWidth: kw, fontSize: 12.5 }}>{k}</span>
    <span style={{ color: vColor || t.fg, fontFamily: FONT_MONO, fontSize: 12.5 }}>{v}</span>
  </div>;
}

// Status check row (doctor)
function CheckLine({ status, label, value, hint }) {
  const t = useTok();
  return <div style={{
    display: "flex", alignItems: "baseline", gap: 10, padding: "3px 0",
    fontFamily: FONT_SANS, letterSpacing: "-0.02em",
  }}>
    <StatusIcon status={status} />
    <span style={{ color: t.fg, fontSize: 13, minWidth: 220 }}>{label}</span>
    <span style={{ color: t.dim, fontSize: 12.5, fontFamily: FONT_MONO }}>{value}</span>
    {hint ? <span style={{ color: t.faint, fontSize: 12, marginLeft: 6 }}>· {hint}</span> : null}
  </div>;
}

// Section heading (subtle accent line + title)
function Section({ title, right, sub }) {
  const t = useTok();
  return <div style={{ marginTop: 14, marginBottom: 8 }}>
    <div style={{ display: "flex", alignItems: "baseline", justifyContent: "space-between" }}>
      <div style={{ display: "flex", alignItems: "baseline", gap: 10 }}>
        <span style={{
          width: 4, height: 14, background: t.accent, borderRadius: 2,
          display: "inline-block", transform: "translateY(2px)",
        }} />
        <span style={{
          fontFamily: FONT_SANS, fontWeight: 800, fontSize: 13,
          letterSpacing: "-0.03em", color: t.fg,
        }}>{title}</span>
        {sub ? <span style={{ fontSize: 12, color: t.dim, letterSpacing: "-0.02em" }}>{sub}</span> : null}
      </div>
      {right ? <span style={{ fontSize: 12, color: t.dim, fontFamily: FONT_MONO }}>{right}</span> : null}
    </div>
  </div>;
}

// Prompt line ($ moai cmd)
function Prompt({ path = "~/work/my-app", branch, dirty, cmd, host = "yuna@air" }) {
  const t = useTok();
  return <div style={{ fontFamily: FONT_MONO, fontSize: 13, lineHeight: 1.6, marginBottom: 6 }}>
    <Tx c="promptArrow" b>❯</Tx>{" "}
    <Tx c="promptPath" b>{path}</Tx>
    {branch ? <>
      {" "}<Tx c="dim">on</Tx>{" "}
      <Tx c={dirty ? "warning" : "success"} b> {branch}</Tx>
      {dirty ? <Tx c="warning"> ✗</Tx> : null}
    </> : null}
    {cmd ? <>{"  "}<Tx c="fg">{cmd}</Tx></> : <>{"  "}<Cursor /></>}
  </div>;
}

function Cursor() {
  const t = useTok();
  return <span style={{
    display: "inline-block", width: 8, height: 14,
    background: t.cursor, verticalAlign: "-2px",
    animation: "blink 1.05s steps(1) infinite",
  }} />;
}

// Window chrome (terminal frame)
function Term({ title = "moai — zsh", width = 760, children, footer, badge, scrollHint = false }) {
  const theme = useTheme();
  const t = TOK[theme];
  return <div style={{
    width, background: t.bg, color: t.fg,
    border: `1px solid ${t.chromeBorder}`, borderRadius: 12,
    overflow: "hidden", boxShadow: t.shadow,
    fontFamily: FONT_SANS,
  }}>
    <div style={{
      display: "flex", alignItems: "center", height: 32, padding: "0 12px",
      background: t.titleBar, borderBottom: `1px solid ${t.chromeBorder}`,
    }}>
      <div style={{ display: "flex", gap: 7 }}>
        {[
          theme === "dark" ? "#3a2826" : "#e0584a",
          theme === "dark" ? "#3a3624" : "#e0a73c",
          theme === "dark" ? "#243a2c" : "#54b25c",
        ].map((c, i) => <span key={i} style={{
          width: 12, height: 12, borderRadius: "50%", background: c,
          border: `0.5px solid rgba(0,0,0,0.18)`,
        }} />)}
      </div>
      <div style={{
        flex: 1, textAlign: "center", fontFamily: FONT_SANS, fontSize: 12,
        color: t.dim, fontWeight: 600, letterSpacing: "-0.015em",
      }}>{title}</div>
      <div style={{ minWidth: 60, textAlign: "right" }}>
        {badge ? <span style={{
          fontSize: 11, color: t.dim, fontFamily: FONT_MONO, opacity: 0.8,
        }}>{badge}</span> : null}
      </div>
    </div>
    <div style={{ padding: "16px 18px 18px", color: t.body }}>
      {children}
    </div>
    {footer ? <div style={{
      borderTop: `1px solid ${t.rule}`, background: t.panel,
      padding: "8px 18px", fontSize: 11.5, color: t.dim,
      fontFamily: FONT_SANS, letterSpacing: "-0.015em",
      display: "flex", justifyContent: "space-between", alignItems: "center",
    }}>{footer}</div> : null}
    {scrollHint ? <div style={{
      borderTop: `1px dashed ${t.ruleSoft}`, padding: "4px 18px",
      fontSize: 10.5, color: t.faint, fontFamily: FONT_SANS, letterSpacing: 0,
    }}>↓ 더 보기 (j/k 또는 ↑↓) · q 종료 · /검색</div> : null}
  </div>;
}

// Help footer (bubbles/help style)
function HelpBar({ items }) {
  const t = useTok();
  return <div style={{
    display: "flex", gap: 16, flexWrap: "wrap", marginTop: 12,
    padding: "8px 12px", borderTop: `1px solid ${t.ruleSoft}`,
    fontFamily: FONT_SANS, fontSize: 11.5, color: t.dim,
    letterSpacing: "-0.015em",
  }}>
    {items.map(([k, v], i) => <span key={i}>
      <kbd style={{
        background: t.ruleSoft, color: t.fg, padding: "1px 6px",
        borderRadius: 3, fontFamily: FONT_MONO, fontSize: 11, fontWeight: 600,
        border: `1px solid ${t.rule}`, marginRight: 4,
      }}>{k}</kbd>
      {v}
    </span>)}
  </div>;
}

Object.assign(window, {
  ThemeCtx, useTheme, useTok, TOK,
  FONT_SANS, FONT_MONO,
  Tx, Box, ThickBox, Pill, StatusIcon, Spinner, Progress, Stepper,
  RadioRow, CheckRow, KV, CheckLine, Section, Prompt, Cursor, Term, HelpBar,
});
