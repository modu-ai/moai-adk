// MoAI-ADK v3 Console — UI primitives shared across screens
const { useState, useEffect, useRef, useMemo, useCallback } = React;

// ── Lucide icon helper (uses lucide global from CDN)
function Icon({ name, size = 16, strokeWidth = 1.75, style }) {
  const ref = useRef(null);
  useEffect(() => {
    if (window.lucide && ref.current) {
      ref.current.setAttribute("data-lucide", name);
      ref.current.innerHTML = "";
      window.lucide.createIcons({ icons: window.lucide.icons, attrs: {}, nameAttr: "data-lucide" });
    }
  }, [name]);
  return <i ref={ref} data-lucide={name} style={{ width: size, height: size, display: "inline-flex", strokeWidth, ...style }} />;
}

// ── Sparkline (mini SVG line chart)
function Sparkline({ data, color, height = 32, fill = true }) {
  const w = 120, h = height;
  const max = Math.max(...data), min = Math.min(...data);
  const range = max - min || 1;
  const stepX = w / (data.length - 1);
  const pts = data.map((v, i) => [i * stepX, h - ((v - min) / range) * (h - 4) - 2]);
  const path = pts.map((p, i) => (i ? "L" : "M") + p[0].toFixed(1) + " " + p[1].toFixed(1)).join(" ");
  const fillPath = path + ` L ${w} ${h} L 0 ${h} Z`;
  const id = useMemo(() => "sg" + Math.random().toString(36).slice(2, 8), []);
  return (
    <svg viewBox={`0 0 ${w} ${h}`} preserveAspectRatio="none" style={{ width: "100%", height: h, display: "block" }}>
      <defs>
        <linearGradient id={id} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity="0.28" />
          <stop offset="100%" stopColor={color} stopOpacity="0" />
        </linearGradient>
      </defs>
      {fill && <path d={fillPath} fill={`url(#${id})`} />}
      <path d={path} fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

// ── Status pill helpers
function Status({ kind = "idle", children }) {
  return <span className={`status-pill ${kind}`}>{children}</span>;
}

// ── Stat card
function Stat({ label, value, delta, deltaKind = "pos", spark, sparkColor = "var(--color-primary)" }) {
  return (
    <div className="card stat">
      <div className="label">{label}</div>
      <div className="value">{value}</div>
      {delta && <div className={`delta ${deltaKind === "pos" ? "" : deltaKind === "neg" ? "neg" : "flat"}`}>
        <Icon name={deltaKind === "pos" ? "trending-up" : deltaKind === "neg" ? "trending-down" : "minus"} size={12} />
        {delta}
      </div>}
      {spark && <div className="stat-spark"><Sparkline data={spark} color={sparkColor} /></div>}
    </div>
  );
}

// ── Drawer
function Drawer({ open, onClose, title, badge, children, footer }) {
  return (
    <>
      <div className={`drawer-scrim ${open ? "open" : ""}`} onClick={onClose} />
      <aside className={`drawer ${open ? "open" : ""}`} role="dialog" aria-modal="true">
        <header className="drawer-head">
          <h2>{title}</h2>
          {badge}
          <button className="icon-btn" onClick={onClose} style={{ marginLeft: "auto" }} aria-label="닫기">
            <Icon name="x" size={18} />
          </button>
        </header>
        <div className="drawer-body">{children}</div>
        {footer && <footer className="drawer-foot">{footer}</footer>}
      </aside>
    </>
  );
}

// ── Modal
function Modal({ open, onClose, title, sub, children, footer, width = 600 }) {
  return (
    <div className={`modal-scrim ${open ? "open" : ""}`} onClick={(e) => { if (e.target === e.currentTarget) onClose && onClose(); }}>
      <div className="modal" style={{ width }}>
        <header className="modal-head">
          <h2>{title}</h2>
          {sub && <div className="sub">{sub}</div>}
        </header>
        <div className="modal-body">{children}</div>
        {footer && <footer className="modal-foot">{footer}</footer>}
      </div>
    </div>
  );
}

// ── Empty state
function Empty({ title, sub, action, mascot = true }) {
  return (
    <div className="empty card">
      {mascot && <img src="assets/moai-logo-3.png" alt="" className="mascot" />}
      <h3>{title}</h3>
      <p>{sub}</p>
      {action}
    </div>
  );
}

// ── Status to chip mapping
function statusChip(s) {
  const map = {
    "in-progress": ["info",    "진행 중"],
    "review":      ["warning", "검토"],
    "done":        ["success", "완료"],
    "blocked":     ["danger",  "블록"],
    "todo":        ["outline", "대기"],
  };
  const [k, l] = map[s] || ["outline", s];
  return <span className={`chip ${k} dot`}>{l}</span>;
}

function priorityChip(p) {
  const map = { critical: ["danger", "Critical"], high: ["warning", "High"], medium: ["info", "Medium"], low: ["outline", "Low"] };
  const [k, l] = map[p] || ["outline", p];
  return <span className={`chip ${k}`}>{l}</span>;
}

Object.assign(window, { Icon, Sparkline, Status, Stat, Drawer, Modal, Empty, statusChip, priorityChip });
