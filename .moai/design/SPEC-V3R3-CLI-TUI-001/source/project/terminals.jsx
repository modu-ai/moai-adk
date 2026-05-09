/* moai-adk CLI 터미널 디자인 — ASCII 텍스트 + 모두의AI 색상 */
/* eslint-disable */

// ───────────────────────────────────────────────
// Theme tokens (passed via React context)
// ───────────────────────────────────────────────
const ThemeCtx = React.createContext("light");

function useTheme() { return React.useContext(ThemeCtx); }

// Color tokens for terminal — derived from 모두의AI guide
// Light terminal: ivory bg, deep teal foreground accent, ink body text
// Dark terminal: ink bg, teal-cyan accent, off-white body
const TERMINAL_TOKENS = {
  light: {
    chrome: "#e8e6e1",           // window chrome
    chromeBorder: "#bcbcbc",
    bg: "#fafaf7",               // terminal background (ivory, not pure white)
    panel: "#f3f3f3",
    fg: "#09110f",               // ink body
    dim: "#6e6e6e",
    faint: "#959595",
    accent: "#144a46",           // primary teal
    accentDeep: "#0a2825",
    accentSoft: "rgba(20,74,70,0.10)",
    success: "#1c7c70",
    warning: "#c47b2a",
    danger: "#c44a3a",
    info: "#2a8a8c",
    selection: "rgba(20,74,70,0.18)",
    rule: "#d4d4d4",
    cursor: "#144a46",
    titleBar: "linear-gradient(180deg, #efece6 0%, #e2dfd8 100%)",
    promptUser: "#144a46",
    promptHost: "#4c4c4c",
    promptPath: "#2a8a8c",
  },
  dark: {
    chrome: "#0e1513",
    chromeBorder: "#1a1f1d",
    bg: "#09110f",
    panel: "#0e1513",
    fg: "#eef0ee",
    dim: "#9aa3a0",
    faint: "#6b7370",
    accent: "#3aa89c",           // brightened for AA on dark
    accentDeep: "#22938a",
    accentSoft: "rgba(58,168,156,0.14)",
    success: "#3ec9a8",
    warning: "#e3a14a",
    danger: "#e87766",
    info: "#5cc7c9",
    selection: "rgba(58,168,156,0.22)",
    rule: "#1a2522",
    cursor: "#3aa89c",
    titleBar: "linear-gradient(180deg, #131a18 0%, #0a1110 100%)",
    promptUser: "#3aa89c",
    promptHost: "#9aa3a0",
    promptPath: "#5cc7c9",
  },
};

// ───────────────────────────────────────────────
// Terminal frame chrome
// ───────────────────────────────────────────────
function Term({ title = "moai — zsh", cols = 92, children, height, footer, scheme }) {
  const theme = useTheme();
  const t = TERMINAL_TOKENS[scheme || theme];
  const ch = 7.6; // approx px per mono char at 13px
  const width = cols * ch + 32; // padding

  return (
    <div className="term" style={{
      width: width,
      background: t.bg,
      border: `1px solid ${t.chromeBorder}`,
      borderRadius: 10,
      overflow: "hidden",
      fontFamily: '"JetBrains Mono", ui-monospace, monospace',
      fontSize: 13,
      lineHeight: 1.55,
      color: t.fg,
      boxShadow: scheme === "dark" || theme === "dark"
        ? "0 24px 60px -20px rgba(0,0,0,0.55), 0 2px 0 rgba(255,255,255,0.02) inset"
        : "0 18px 40px -18px rgba(9,17,15,0.18), 0 1px 0 rgba(255,255,255,0.6) inset",
    }}>
      {/* Title bar */}
      <div style={{
        height: 30,
        background: t.titleBar,
        borderBottom: `1px solid ${t.chromeBorder}`,
        display: "flex",
        alignItems: "center",
        padding: "0 12px",
        gap: 8,
      }}>
        <div style={{ display: "flex", gap: 6 }}>
          <span style={{ width: 11, height: 11, borderRadius: "50%", background: scheme === "dark" || theme === "dark" ? "#4a3030" : "#e06c5a", border: `0.5px solid rgba(0,0,0,0.15)` }} />
          <span style={{ width: 11, height: 11, borderRadius: "50%", background: scheme === "dark" || theme === "dark" ? "#4a4530" : "#e0b85a", border: `0.5px solid rgba(0,0,0,0.15)` }} />
          <span style={{ width: 11, height: 11, borderRadius: "50%", background: scheme === "dark" || theme === "dark" ? "#2e4a3a" : "#5ec275", border: `0.5px solid rgba(0,0,0,0.15)` }} />
        </div>
        <div style={{
          flex: 1, textAlign: "center", fontSize: 11.5,
          fontFamily: '"Pretendard", system-ui, sans-serif',
          color: t.dim, fontWeight: 500, letterSpacing: "-0.01em",
        }}>{title}</div>
        <div style={{ width: 50 }} />
      </div>
      {/* Body */}
      <div style={{
        padding: "14px 18px 18px",
        minHeight: height,
        whiteSpace: "pre",
        fontVariantLigatures: "none",
      }}>
        {children}
      </div>
      {footer ? (
        <div style={{
          borderTop: `1px solid ${t.rule}`,
          padding: "6px 18px",
          fontSize: 11,
          color: t.dim,
          fontFamily: '"Pretendard", system-ui, sans-serif',
          letterSpacing: "-0.01em",
          background: t.panel,
          display: "flex", justifyContent: "space-between",
        }}>{footer}</div>
      ) : null}
    </div>
  );
}

// ───────────────────────────────────────────────
// Inline color helpers for terminal text
// ───────────────────────────────────────────────
function useTok() {
  const theme = useTheme();
  return TERMINAL_TOKENS[theme];
}

const C = ({ c, b, children, style }) => {
  const t = useTok();
  return <span style={{ color: t[c] || c, fontWeight: b ? 700 : "inherit", ...style }}>{children}</span>;
};

// Prompt line:  ➜ ~/work/my-app  ⌥
function Prompt({ path = "~/work/my-app", branch = "main", cmd, dirty = false }) {
  const t = useTok();
  return (
    <div>
      <span style={{ color: t.accent, fontWeight: 700 }}>➜</span>{" "}
      <span style={{ color: t.promptPath, fontWeight: 600 }}>{path}</span>
      {branch ? <>
        {" "}
        <span style={{ color: t.dim }}>git:(</span>
        <span style={{ color: dirty ? t.warning : t.success, fontWeight: 600 }}>{branch}</span>
        <span style={{ color: t.dim }}>)</span>
        {dirty ? <span style={{ color: t.warning }}>{" ✗"}</span> : null}
      </> : null}
      {"  "}
      {cmd ? <span style={{ color: t.fg }}>{cmd}</span> : <span style={{ color: t.cursor, animation: "blink 1.05s steps(1) infinite" }}>▌</span>}
    </div>
  );
}

// Tag pill rendered in terminal text
function Tag({ kind = "info", children }) {
  const t = useTok();
  const map = {
    info: { bg: "rgba(42,138,140,0.16)", fg: t.info },
    ok:   { bg: "rgba(28,124,112,0.16)", fg: t.success },
    warn: { bg: "rgba(196,123,42,0.16)", fg: t.warning },
    err:  { bg: "rgba(196,74,58,0.16)", fg: t.danger },
    primary: { bg: t.accentSoft, fg: t.accent },
  };
  const m = map[kind] || map.info;
  return <span style={{
    background: m.bg, color: m.fg, fontWeight: 600,
    padding: "1px 7px", borderRadius: 4,
    fontSize: "0.92em", letterSpacing: 0,
  }}>{children}</span>;
}

// Section divider
function Rule({ ch = "─", n = 88, label }) {
  const t = useTok();
  if (label) {
    const padding = Math.max(0, n - label.length - 4);
    const left = Math.floor(padding / 2);
    const right = padding - left;
    return <div style={{ color: t.rule }}>
      {ch.repeat(left)} <span style={{ color: t.accent, fontWeight: 700 }}>{label}</span> {ch.repeat(right)}
    </div>;
  }
  return <div style={{ color: t.rule }}>{ch.repeat(n)}</div>;
}

// ───────────────────────────────────────────────
// Banner ASCII (from internal/cli/banner.go) — re-styled in teal
// ───────────────────────────────────────────────
const MOAI_BANNER = String.raw`███╗   ███╗          █████╗ ██╗       █████╗ ██████╗ ██╗  ██╗
████╗ ████║ ██████╗ ██╔══██╗██║      ██╔══██╗██╔══██╗██║ ██╔╝
██╔████╔██║██║   ██║███████║██║█████╗███████║██║  ██║█████╔╝
██║╚██╔╝██║██║   ██║██╔══██║██║╚════╝██╔══██║██║  ██║██╔═██╗
██║ ╚═╝ ██║╚██████╔╝██║  ██║██║      ██║  ██║██████╔╝██║  ██╗
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝      ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝`;

function Banner({ version = "v3.2.4" }) {
  const t = useTok();
  return (
    <>
      <div style={{ color: t.accent, fontWeight: 700, lineHeight: 1.1, fontSize: 12 }}>{MOAI_BANNER}</div>
      <div style={{ height: 6 }} />
      <div style={{ color: t.dim, paddingLeft: 2 }}>
        Modu-AI's Agentic Development Kit <span style={{ color: t.accent }}>w/</span> SuperAgent <span style={{ color: t.accent, fontWeight: 700 }}>MoAI</span>
      </div>
      <div style={{ color: t.faint, paddingLeft: 2 }}>
        Version: <span style={{ color: t.fg }}>{version}</span>   ·   <span style={{ color: t.accent }}>한국어 / English</span>   ·   <a style={{ color: t.info }}>github.com/modu-ai/moai-adk</a>
      </div>
    </>
  );
}

// ───────────────────────────────────────────────
// Box drawing helpers
// ───────────────────────────────────────────────
function Box({ width = 88, title, children, accent = false }) {
  const t = useTok();
  const top = "╭" + "─".repeat(width - 2) + "╮";
  const bot = "╰" + "─".repeat(width - 2) + "╯";
  const titleLine = title ? "├ " + title + " " + "─".repeat(Math.max(0, width - title.length - 5)) + "┤" : null;
  const color = accent ? t.accent : t.rule;
  return (
    <div>
      <div style={{ color }}>{top}</div>
      {title ? <div style={{ color }}>
        <span>├ </span>
        <span style={{ color: t.accent, fontWeight: 700 }}>{title}</span>
        <span>{" " + "─".repeat(Math.max(0, width - title.length - 5)) + "┤"}</span>
      </div> : null}
      {children}
      <div style={{ color }}>{bot}</div>
    </div>
  );
}

// Boxed line: "│ ...content... │"
function BL({ children, width = 88, color }) {
  const t = useTok();
  return <div>
    <span style={{ color: color || t.rule }}>│ </span>
    {children}
    {/* note: content widths are eyeballed; trailing │ is decorative */}
  </div>;
}

Object.assign(window, { Term, Prompt, C, Tag, Rule, Banner, Box, BL, ThemeCtx, useTok, TERMINAL_TOKENS });
