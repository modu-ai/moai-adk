#!/usr/bin/env node
// Deterministic source lint for hand-authored SVG infographics.
//
// Runs on the Node 18+ standard library alone: no package install, no network
// access, no browser. It reads the SVG as text, so it reports a stable
// file:line:column for every finding and never depends on a rendering engine.
//
// Usage:
//   node check-svg.mjs <file.svg> [options]
//
// Options:
//   --json          emit machine-readable JSON instead of text
//   --strict        treat warnings as failures
//   --pad <n>       inner padding assumed for text-fit checks (default 8)
//   --help          print this usage block
//
// Exit codes:
//   0  no errors (warnings may be present unless --strict was passed)
//   1  at least one error, or a warning under --strict
//   2  usage error, or the file could not be read
//
// Diagnostic tiers:
//   error    deterministic and structural; always fix before rendering
//   warning  heuristic (character-advance estimation); confirm against the PNG
//
// The error tier covers document structure (SVG001-SVG050) and the
// accessible-SVG contract (SVG060-SVG064: role, aria-labelledby, <title> first,
// <desc> present, ids prefixed per diagram). Fixtures exercising the contract in
// both directions live in fixtures/a11y-present.svg and fixtures/a11y-missing.svg.

import { readFileSync } from 'node:fs';

const ERROR = 'error';
const WARNING = 'warning';

// ---------------------------------------------------------------------------
// Argument handling
// ---------------------------------------------------------------------------

function parseArgs(argv) {
  const opts = { file: null, json: false, strict: false, pad: 8, help: false };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--help' || a === '-h') opts.help = true;
    else if (a === '--json') opts.json = true;
    else if (a === '--strict') opts.strict = true;
    else if (a === '--pad') opts.pad = Number(argv[++i]);
    else if (a.startsWith('-')) return { error: `unknown option: ${a}` };
    else if (opts.file === null) opts.file = a;
    else return { error: `unexpected extra argument: ${a}` };
  }
  if (!opts.help && opts.file === null) return { error: 'missing <file.svg>' };
  if (!Number.isFinite(opts.pad) || opts.pad < 0) return { error: '--pad expects a non-negative number' };
  return opts;
}

const USAGE = [
  'Usage: node check-svg.mjs <file.svg> [--json] [--strict] [--pad <n>]',
  '',
  'Lints an SVG source file for structural errors and heuristic layout warnings.',
  'Exit 0 = no errors, 1 = errors (or warnings under --strict), 2 = usage/read failure.',
].join('\n');

// ---------------------------------------------------------------------------
// Tokenizer: walks the raw source, skipping comments, CDATA, and declarations
// ---------------------------------------------------------------------------

const ATTR_PATTERN = /([A-Za-z_:][\w.:-]*)\s*=\s*("([^"]*)"|'([^']*)')/g;

function parseAttributes(raw) {
  const attrs = {};
  ATTR_PATTERN.lastIndex = 0;
  let m;
  while ((m = ATTR_PATTERN.exec(raw)) !== null) {
    attrs[m[1]] = m[3] !== undefined ? m[3] : m[4];
  }
  return attrs;
}

function tokenize(src) {
  const tokens = [];
  const n = src.length;
  let i = 0;

  while (i < n) {
    const lt = src.indexOf('<', i);
    if (lt === -1) {
      if (i < n) tokens.push({ kind: 'text', value: src.slice(i), offset: i });
      break;
    }
    if (lt > i) tokens.push({ kind: 'text', value: src.slice(i, lt), offset: i });

    if (src.startsWith('<!--', lt)) {
      const end = src.indexOf('-->', lt + 4);
      i = end === -1 ? n : end + 3;
      continue;
    }
    if (src.startsWith('<![CDATA[', lt)) {
      const end = src.indexOf(']]>', lt + 9);
      i = end === -1 ? n : end + 3;
      continue;
    }
    if (src.startsWith('<?', lt) || src.startsWith('<!', lt)) {
      const end = src.indexOf('>', lt);
      i = end === -1 ? n : end + 1;
      continue;
    }

    // Regular tag: scan to the closing angle bracket, respecting quoted values.
    let j = lt + 1;
    let quote = null;
    while (j < n) {
      const c = src[j];
      if (quote !== null) {
        if (c === quote) quote = null;
      } else if (c === '"' || c === "'") {
        quote = c;
      } else if (c === '>') {
        break;
      }
      j++;
    }
    if (j >= n) {
      tokens.push({ kind: 'unterminated', offset: lt });
      break;
    }

    let raw = src.slice(lt + 1, j);
    let kind = 'open';
    if (raw.startsWith('/')) {
      kind = 'close';
      raw = raw.slice(1);
    } else if (raw.endsWith('/')) {
      kind = 'self';
      raw = raw.slice(0, -1);
    }
    const nameMatch = /^\s*([A-Za-z_][\w.:-]*)/.exec(raw);
    tokens.push({
      kind,
      name: nameMatch ? nameMatch[1] : '',
      attrs: kind === 'close' ? {} : parseAttributes(raw),
      offset: lt,
    });
    i = j + 1;
  }
  return tokens;
}

// ---------------------------------------------------------------------------
// Light element tree
// ---------------------------------------------------------------------------

function buildTree(src) {
  const tokens = tokenize(src);
  const root = { name: '#root', attrs: {}, children: [], parent: null, text: '', offset: 0 };
  const stack = [root];
  const structural = [];

  for (const t of tokens) {
    const top = stack[stack.length - 1];

    if (t.kind === 'text') {
      top.text += t.value;
      continue;
    }
    if (t.kind === 'unterminated') {
      structural.push({ offset: t.offset, message: 'tag is not terminated by ">"' });
      continue;
    }
    if (t.kind === 'close') {
      const depth = stack.findLastIndex((el) => el.name === t.name);
      if (depth <= 0) {
        structural.push({ offset: t.offset, message: `closing tag </${t.name}> has no matching opening tag` });
        continue;
      }
      for (let k = stack.length - 1; k > depth; k--) {
        structural.push({
          offset: stack[k].offset,
          message: `<${stack[k].name}> is not closed before </${t.name}>`,
        });
      }
      stack.length = depth;
      continue;
    }

    const el = {
      name: t.name,
      attrs: t.attrs,
      children: [],
      parent: top,
      text: '',
      offset: t.offset,
    };
    top.children.push(el);
    if (t.kind === 'open') stack.push(el);
  }

  for (let k = stack.length - 1; k > 0; k--) {
    structural.push({ offset: stack[k].offset, message: `<${stack[k].name}> is never closed` });
  }
  return { root, structural };
}

function walk(el, visit) {
  for (const child of el.children) {
    visit(child);
    walk(child, visit);
  }
}

function lineIndex(src) {
  const starts = [0];
  for (let i = 0; i < src.length; i++) {
    if (src[i] === '\n') starts.push(i + 1);
  }
  return starts;
}

function positionOf(starts, offset) {
  let lo = 0;
  let hi = starts.length - 1;
  while (lo < hi) {
    const mid = (lo + hi + 1) >> 1;
    if (starts[mid] <= offset) lo = mid;
    else hi = mid - 1;
  }
  return { line: lo + 1, column: offset - starts[lo] + 1 };
}

// ---------------------------------------------------------------------------
// Text measurement (the same model the skill body documents)
// ---------------------------------------------------------------------------

const NARROW = new Set([...'iljtIfr.,:;\'`|!()[]{}-']);

function isFullWidth(cp) {
  return (
    (cp >= 0x1100 && cp <= 0x115f) ||
    (cp >= 0x2e80 && cp <= 0x303e) ||
    (cp >= 0x3041 && cp <= 0x33ff) ||
    (cp >= 0x3400 && cp <= 0x4dbf) ||
    (cp >= 0x4e00 && cp <= 0x9fff) ||
    (cp >= 0xa000 && cp <= 0xa4cf) ||
    (cp >= 0xac00 && cp <= 0xd7a3) ||
    (cp >= 0xf900 && cp <= 0xfaff) ||
    (cp >= 0xfe30 && cp <= 0xfe6f) ||
    (cp >= 0xff00 && cp <= 0xff60) ||
    (cp >= 0xffe0 && cp <= 0xffe6) ||
    (cp >= 0x20000 && cp <= 0x2fffd) ||
    (cp >= 0x30000 && cp <= 0x3fffd)
  );
}

// Advance width in em units. Full-width scripts occupy a whole em, which is why
// a CJK line holds roughly 60% of the characters a Latin line holds.
function advanceEm(ch) {
  const cp = ch.codePointAt(0);
  if (isFullWidth(cp)) return 1.0;
  if (ch === ' ') return 0.3;
  if (NARROW.has(ch)) return 0.3;
  if (ch >= 'A' && ch <= 'Z') return 0.68;
  return 0.55;
}

function estimateWidth(text, fontSize) {
  let em = 0;
  for (const ch of text) em += advanceEm(ch);
  return em * fontSize;
}

function num(value) {
  if (value === undefined || value === null) return null;
  const cleaned = String(value).trim().replace(/(px|pt|%)$/, '');
  const parsed = Number(cleaned);
  return Number.isFinite(parsed) ? parsed : null;
}

function inheritedFontSize(el) {
  for (let cur = el; cur !== null; cur = cur.parent) {
    const direct = num(cur.attrs['font-size']);
    if (direct !== null) return direct;
    const style = cur.attrs.style;
    if (style) {
      const m = /font-size\s*:\s*([0-9.]+)/.exec(style);
      if (m) return Number(m[1]);
    }
  }
  return 16;
}

function textContent(el) {
  let out = el.text;
  for (const child of el.children) out += textContent(child);
  return out.replace(/\s+/g, ' ').trim();
}

function hasTransform(el) {
  for (let cur = el; cur !== null; cur = cur.parent) {
    if (cur.attrs.transform) return true;
  }
  return false;
}

// ---------------------------------------------------------------------------
// Connector geometry (the C2 / C4 / C6 rules of references/authoring.md 2.5)
// ---------------------------------------------------------------------------

// The documented connector idiom: an arrowhead stands markerLen off the box it
// points at, elbows are rounded at r = 8, crossings hop at rHop = 5.
const MARKER_LEN = 10;
// C2's band: a label mask clears its own stroke by 6 to 10 units inclusive.
const C2_MIN_CLEARANCE = 6;
const C2_MAX_CLEARANCE = 10;
// A mask associates to its own connector only within the band's upper bound
// plus a tolerance of 6, C2's own lower bound. Beyond that it labels nothing.
const MASK_WINDOW = C2_MAX_CLEARANCE + C2_MIN_CLEARANCE;
// C4's fan floor, selected by the length of the edge the arrivals land on.
const C4_LONG_EDGE = 120;
const C4_FLOOR_LONG = 12;
const C4_FLOOR_SHORT = 8;

// Definitions, not painted geometry: document order says nothing about their
// paint order either, so they are wrong inputs to every check below.
const NON_RENDERED = new Set(['defs', 'marker', 'symbol', 'clipPath', 'mask', 'pattern']);

const EDGE_NAMES = ['top', 'right', 'bottom', 'left'];

function inNonRendered(el) {
  for (let cur = el; cur !== null; cur = cur.parent) {
    if (NON_RENDERED.has(cur.name)) return true;
  }
  return false;
}

const PATH_TOKEN = /([MmLlHhVvQqAaZz])|([+-]?(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?)|([\s,]+)|([^])/g;

// Splits a d attribute into command letters and numbers, or returns null on the
// first character the reader does not handle (C, S, T, malformed numbers).
function tokenizePath(d) {
  const out = [];
  PATH_TOKEN.lastIndex = 0;
  let m;
  while ((m = PATH_TOKEN.exec(d)) !== null) {
    if (m[1] !== undefined) out.push({ cmd: m[1] });
    else if (m[2] !== undefined) out.push({ n: Number(m[2]) });
    else if (m[3] !== undefined) continue;
    else return null;
  }
  return out;
}

function quadraticAt(p0, c, p1, t) {
  const u = 1 - t;
  return [
    u * u * p0[0] + 2 * u * t * c[0] + t * t * p1[0],
    u * u * p0[1] + 2 * u * t * c[1] + t * t * p1[1],
  ];
}

// Recovers a connector's polyline from its d attribute, or null when any part
// of it cannot be fully interpreted - a connector the reader cannot read is
// skipped silently rather than approximated.
function readPolyline(d) {
  if (typeof d !== 'string') return null;
  const toks = tokenizePath(d);
  if (toks === null) return null;

  const points = [];
  let i = 0;
  let cx = 0;
  let cy = 0;
  let sx = 0;
  let sy = 0;
  let started = false;

  const isNumber = () => i < toks.length && toks[i].n !== undefined;
  const take = () => (isNumber() ? toks[i++].n : null);

  while (i < toks.length) {
    if (toks[i].cmd === undefined) return null; // coordinates with no command
    let cmd = toks[i++].cmd;

    if (cmd === 'Z' || cmd === 'z') {
      if (!started) return null;
      cx = sx;
      cy = sy;
      points.push([cx, cy]);
      continue;
    }
    if (!isNumber()) return null; // a command carrying no arguments

    let first = true;
    while (isNumber()) {
      const relative = cmd === cmd.toLowerCase();
      const kind = cmd.toUpperCase();

      if (kind === 'M' || kind === 'L') {
        const x = take();
        const y = take();
        if (y === null) return null;
        cx = relative ? cx + x : x;
        cy = relative ? cy + y : y;
        if (kind === 'M' && first) {
          sx = cx;
          sy = cy;
          started = true;
        }
        points.push([cx, cy]);
      } else if (kind === 'H') {
        const x = take();
        cx = relative ? cx + x : x;
        points.push([cx, cy]);
      } else if (kind === 'V') {
        const y = take();
        cy = relative ? cy + y : y;
        points.push([cx, cy]);
      } else if (kind === 'Q') {
        // The rounded elbow of 2.2. Its control point is the un-rounded corner,
        // a point the stroke never reaches, so the curve is subdivided and the
        // control point dropped: emitting it would bias a clearance measured on
        // the convex side inward and report a compliant mask.
        const x1 = take();
        const y1 = take();
        const x = take();
        const y = take();
        if (y === null) return null;
        const p0 = [cx, cy];
        const c = [relative ? cx + x1 : x1, relative ? cy + y1 : y1];
        const p1 = [relative ? cx + x : x, relative ? cy + y : y];
        for (const t of [0.25, 0.5, 0.75]) points.push(quadraticAt(p0, c, p1, t));
        points.push(p1);
        cx = p1[0];
        cy = p1[1];
      } else if (kind === 'A') {
        // The crossing hop of 2.3: a pass-through between the current point and
        // the arc endpoint, which is all the clearance checks need.
        take();
        take();
        take();
        take();
        take();
        const x = take();
        const y = take();
        if (y === null) return null;
        cx = relative ? cx + x : x;
        cy = relative ? cy + y : y;
        points.push([cx, cy]);
      } else {
        return null;
      }

      if (!started) return null; // a subpath that does not open with M
      first = false;
      if (kind === 'M') cmd = relative ? 'l' : 'L'; // implicit repeats are linetos
    }
  }

  return points.length >= 2 ? points : null;
}

function pointSegmentDistance(p, a, b) {
  const dx = b[0] - a[0];
  const dy = b[1] - a[1];
  const len2 = dx * dx + dy * dy;
  if (len2 === 0) return Math.hypot(p[0] - a[0], p[1] - a[1]);
  const t = Math.max(0, Math.min(1, ((p[0] - a[0]) * dx + (p[1] - a[1]) * dy) / len2));
  return Math.hypot(p[0] - (a[0] + t * dx), p[1] - (a[1] + t * dy));
}

function segmentsCross(a, b, c, d) {
  const side = (o, p, q) => (p[0] - o[0]) * (q[1] - o[1]) - (p[1] - o[1]) * (q[0] - o[0]);
  const d1 = side(a, b, c);
  const d2 = side(a, b, d);
  const d3 = side(c, d, a);
  const d4 = side(c, d, b);
  return ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) && ((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0));
}

function segmentDistance(a, b, c, d) {
  if (segmentsCross(a, b, c, d)) return 0;
  return Math.min(
    pointSegmentDistance(a, c, d),
    pointSegmentDistance(b, c, d),
    pointSegmentDistance(c, a, b),
    pointSegmentDistance(d, a, b),
  );
}

function rectContains(r, p) {
  return p[0] >= r.x && p[0] <= r.x + r.w && p[1] >= r.y && p[1] <= r.y + r.h;
}

function rectCorners(r) {
  return [
    [r.x, r.y],
    [r.x + r.w, r.y],
    [r.x + r.w, r.y + r.h],
    [r.x, r.y + r.h],
  ];
}

// Overlap or touching counts as distance 0, which is C2's crossing case.
function rectPolylineDistance(r, points) {
  const corners = rectCorners(r);
  let best = Infinity;
  for (let k = 1; k < points.length; k++) {
    const a = points[k - 1];
    const b = points[k];
    if (rectContains(r, a) || rectContains(r, b)) return 0;
    for (let e = 0; e < 4; e++) {
      best = Math.min(best, segmentDistance(a, b, corners[e], corners[(e + 1) % 4]));
      if (best === 0) return 0;
    }
  }
  return best;
}

function rectOf(el) {
  const w = num(el.attrs.width);
  const h = num(el.attrs.height);
  if (w === null || h === null || w <= 0 || h <= 0) return null;
  return { x: num(el.attrs.x) ?? 0, y: num(el.attrs.y) ?? 0, w, h };
}

// The four attach surfaces, ordered top / right / bottom / left so a tie
// between two equidistant edges resolves deterministically.
function edgesOf(r) {
  return [
    { horizontal: true, at: r.y, from: r.x, length: r.w },
    { horizontal: false, at: r.x + r.w, from: r.y, length: r.h },
    { horizontal: true, at: r.y + r.h, from: r.x, length: r.w },
    { horizontal: false, at: r.x, from: r.y, length: r.h },
  ];
}

// An endpoint reaches an edge only where its projection falls inside the edge
// span; the perpendicular distance is what the standoff is compared against.
function edgeReach(p, edge) {
  const offset = (edge.horizontal ? p[0] : p[1]) - edge.from;
  if (offset < 0 || offset > edge.length) return null;
  return { distance: Math.abs((edge.horizontal ? p[1] : p[0]) - edge.at), offset };
}

// The single nearest edge across every box, or null when none is in reach.
function bindEndpoint(p, boxes, standoff) {
  let best = null;
  for (let b = 0; b < boxes.length; b++) {
    const edges = boxes[b].edges;
    for (let e = 0; e < 4; e++) {
      const reach = edgeReach(p, edges[e]);
      if (reach === null || reach.distance > standoff) continue;
      if (best !== null && reach.distance >= best.distance) continue;
      best = { box: b, edge: e, offset: reach.offset, distance: reach.distance, length: edges[e].length };
    }
  }
  return best;
}

function touchesRect(p, edges, standoff) {
  for (const edge of edges) {
    const reach = edgeReach(p, edge);
    if (reach !== null && reach.distance <= standoff) return true;
  }
  return false;
}

// ---------------------------------------------------------------------------
// Checks
// ---------------------------------------------------------------------------

function lint(src, opts) {
  const { root, structural } = buildTree(src);
  const starts = lineIndex(src);
  const diagnostics = [];

  const report = (level, code, offset, message) => {
    const { line, column } = positionOf(starts, offset);
    diagnostics.push({ line, column, level, code, message });
  };

  for (const s of structural) {
    report(ERROR, 'SVG050', s.offset, s.message);
  }

  const svg = root.children.find((c) => c.name === 'svg');
  if (!svg) {
    report(ERROR, 'SVG001', 0, 'no root <svg> element was found');
    return diagnostics.sort((a, b) => a.line - b.line || a.column - b.column);
  }

  // SVG002 / SVG003 - viewBox sanity and aspect agreement.
  const viewBoxRaw = svg.attrs.viewBox;
  let viewBox = null;
  if (!viewBoxRaw) {
    report(ERROR, 'SVG002', svg.offset, 'root <svg> has no viewBox attribute');
  } else {
    const parts = viewBoxRaw.trim().split(/[\s,]+/).map(Number);
    if (parts.length !== 4 || parts.some((p) => !Number.isFinite(p))) {
      report(ERROR, 'SVG002', svg.offset, `viewBox "${viewBoxRaw}" is not four numbers`);
    } else if (parts[2] <= 0 || parts[3] <= 0) {
      report(ERROR, 'SVG002', svg.offset, `viewBox width and height must be positive, got ${parts[2]}x${parts[3]}`);
    } else {
      viewBox = { x: parts[0], y: parts[1], w: parts[2], h: parts[3] };
      const aw = num(svg.attrs.width);
      const ah = num(svg.attrs.height);
      if (aw !== null && ah !== null && aw > 0 && ah > 0) {
        const declared = aw / ah;
        const intrinsic = viewBox.w / viewBox.h;
        if (Math.abs(declared - intrinsic) / intrinsic > 0.01) {
          report(
            ERROR,
            'SVG003',
            svg.offset,
            `width/height ratio ${declared.toFixed(4)} disagrees with the viewBox ratio ${intrinsic.toFixed(4)}`,
          );
        }
      }
    }
  }

  // Collect ids for reference resolution.
  const ids = new Map();
  walk(svg, (el) => {
    const id = el.attrs.id;
    if (id === undefined) return;
    if (ids.has(id)) {
      report(ERROR, 'SVG010', el.offset, `duplicate id "${id}" (first declared on line ${positionOf(starts, ids.get(id)).line})`);
    } else {
      ids.set(id, el.offset);
    }
  });
  if (svg.attrs.id !== undefined && !ids.has(svg.attrs.id)) ids.set(svg.attrs.id, svg.offset);

  // SVG011 - every local reference must resolve.
  const collectRefs = (el) => {
    for (const [name, value] of Object.entries(el.attrs)) {
      const urlPattern = /url\(\s*#([^)\s]+)\s*\)/g;
      let m;
      while ((m = urlPattern.exec(value)) !== null) {
        if (!ids.has(m[1])) {
          report(ERROR, 'SVG011', el.offset, `attribute ${name} references "#${m[1]}" but no element declares that id`);
        }
      }
      if ((name === 'href' || name === 'xlink:href') && value.startsWith('#')) {
        const target = value.slice(1);
        if (!ids.has(target)) {
          report(ERROR, 'SVG011', el.offset, `attribute ${name} references "#${target}" but no element declares that id`);
        }
      }
    }
  };
  collectRefs(svg);
  walk(svg, collectRefs);

  // SVG020 / SVG021 - marker geometry must be explicit.
  walk(svg, (el) => {
    if (el.name !== 'marker') return;
    const required = ['markerWidth', 'markerHeight', 'refX', 'refY'];
    const missing = required.filter((k) => el.attrs[k] === undefined);
    if (missing.length > 0) {
      report(ERROR, 'SVG020', el.offset, `<marker> is missing required geometry: ${missing.join(', ')}`);
    }
    if (el.attrs.markerUnits === undefined) {
      report(
        ERROR,
        'SVG021',
        el.offset,
        '<marker> has no explicit markerUnits; the strokeWidth default rescales arrowheads with line width',
      );
    }
  });


  // SVG060 / SVG061 / SVG062 / SVG063 / SVG064 - the accessible-SVG contract.
  // An SVG carries no accessible name of its own: without these four pieces the
  // diagram is announced as an unlabelled graphic and none of its <text> is read.
  if (svg.attrs['aria-hidden'] === 'true') {
    // Decorative graphic: hidden from assistive technology by contract, so an
    // accessible name would only add noise. Nothing further to assert.
  } else {
    const role = svg.attrs.role;
    if (role === undefined) {
      report(
        ERROR,
        'SVG060',
        svg.offset,
        'root <svg> has no role; an informative diagram needs role="img" (a decorative graphic needs aria-hidden="true" instead)',
      );
    } else if (role !== 'img') {
      report(ERROR, 'SVG060', svg.offset, `root <svg> has role="${role}"; a diagram is a figure and needs role="img"`);
    }

    const titleEl = svg.children.find((c) => c.name === 'title');
    const descEl = svg.children.find((c) => c.name === 'desc');

    if (titleEl === undefined) {
      report(ERROR, 'SVG062', svg.offset, 'root <svg> has no <title>; the diagram has no accessible name');
    } else if (svg.children[0] !== titleEl) {
      report(
        ERROR,
        'SVG062',
        titleEl.offset,
        `<title> must be the first child of <svg>, but <${svg.children[0].name}> precedes it; a later <title> may be ignored`,
      );
    }

    if (descEl === undefined) {
      report(ERROR, 'SVG063', svg.offset, 'root <svg> has no <desc>; state in one sentence what the diagram shows');
    }

    const labelledBy = svg.attrs['aria-labelledby'];
    const refs = labelledBy === undefined ? [] : labelledBy.trim().split(/\s+/).filter(Boolean);
    if (refs.length === 0) {
      report(
        ERROR,
        'SVG061',
        svg.offset,
        'root <svg> has no aria-labelledby naming the <title> and <desc> ids; role="img" alone resolves no name',
      );
    } else {
      for (const ref of refs) {
        if (!ids.has(ref)) {
          report(ERROR, 'SVG061', svg.offset, `aria-labelledby names "${ref}" but no element declares that id`);
        }
      }
    }

    for (const [el, tag] of [[titleEl, 'title'], [descEl, 'desc']]) {
      if (el === undefined) continue;
      const id = el.attrs.id;
      if (id === undefined) {
        report(ERROR, 'SVG064', el.offset, `<${tag}> has no id, so aria-labelledby cannot name it`);
        continue;
      }
      if (id === tag) {
        report(
          ERROR,
          'SVG064',
          el.offset,
          `<${tag}> uses the bare id "${tag}"; prefix it per diagram (<slug>-${tag}) so two inlined diagrams cannot collide`,
        );
      }
      if (refs.length > 0 && !refs.includes(id)) {
        report(ERROR, 'SVG061', el.offset, `<${tag} id="${id}"> is not named by aria-labelledby`);
      }
    }
  }

  // SVG030 / SVG031 - heuristic text fit against the nearest preceding rect.
  walk(svg, (el) => {
    if (el.name !== 'text') return;
    const content = textContent(el);
    if (content === '') return;

    const siblings = el.parent ? el.parent.children : [];
    const index = siblings.indexOf(el);
    let container = null;
    for (let k = index - 1; k >= 0; k--) {
      if (siblings[k].name === 'rect') {
        container = siblings[k];
        break;
      }
    }
    if (container === null) return;

    const cw = num(container.attrs.width);
    const ch = num(container.attrs.height);
    if (cw === null || cw <= 0) return;

    const fontSize = inheritedFontSize(el);
    const estimated = estimateWidth(content, fontSize);
    const rx = num(container.attrs.rx);
    const isPill = rx !== null && ch !== null && ch > 0 && rx >= ch / 2 - 0.5;

    if (isPill) {
      const usable = cw - 2 * opts.pad - ch * 0.3;
      if (estimated > usable) {
        report(
          WARNING,
          'SVG031',
          el.offset,
          `pill label needs about ${estimated.toFixed(0)} units but the pill offers ${usable.toFixed(0)} after the round-cap inset`,
        );
      }
      return;
    }
    const usable = cw - 2 * opts.pad;
    if (estimated > usable) {
      report(
        WARNING,
        'SVG030',
        el.offset,
        `label needs about ${estimated.toFixed(0)} units but its container offers ${usable.toFixed(0)}`,
      );
    }
  });

  // SVG040 - geometry outside the viewBox (untransformed elements only).
  if (viewBox !== null) {
    walk(svg, (el) => {
      if (hasTransform(el)) return;
      let box = null;
      if (el.name === 'rect' || el.name === 'image') {
        const x = num(el.attrs.x) ?? 0;
        const y = num(el.attrs.y) ?? 0;
        const w = num(el.attrs.width);
        const h = num(el.attrs.height);
        if (w !== null && h !== null) box = { x, y, w, h };
      } else if (el.name === 'circle') {
        const cx = num(el.attrs.cx) ?? 0;
        const cy = num(el.attrs.cy) ?? 0;
        const r = num(el.attrs.r);
        if (r !== null) box = { x: cx - r, y: cy - r, w: 2 * r, h: 2 * r };
      }
      if (box === null) return;

      const overflow = [];
      if (box.x < viewBox.x) overflow.push('left');
      if (box.y < viewBox.y) overflow.push('top');
      if (box.x + box.w > viewBox.x + viewBox.w) overflow.push('right');
      if (box.y + box.h > viewBox.y + viewBox.h) overflow.push('bottom');
      if (overflow.length > 0) {
        report(WARNING, 'SVG040', el.offset, `<${el.name}> extends past the viewBox on the ${overflow.join(' and ')}`);
      }
    });
  }

  // SVG070 / SVG071 / SVG072 / SVG073 / SVG074 - connector geometry.
  // Mechanises C2 (a label mask clears its own stroke by 6-10 units), C4
  // (connector arrivals on one box edge stay apart) and C6 (a label mask may
  // not partially overlap a shape painted after it). Nothing new is asked of
  // the author: a connector, a box and a label mask are inferred from geometry
  // and sibling order alone, and every ambiguity resolves toward silence.
  {
    const candidates = [];
    walk(svg, (el) => {
      if (el.name !== 'path' && el.name !== 'rect' && el.name !== 'text') return;
      if (inNonRendered(el)) return;
      candidates.push(el);
    });
    const skipped = candidates.filter(hasTransform);
    const usable = candidates.filter((el) => !hasTransform(el));

    // Connectors: a stroked path, the documented idiom being fill="none". A
    // filled path is a shape; a d the reader cannot read is skipped silently.
    const connectors = [];
    for (const el of usable) {
      if (el.name !== 'path') continue;
      if (el.attrs.fill === undefined || el.attrs.fill.trim() !== 'none') continue;
      const points = readPolyline(el.attrs.d);
      if (points === null) continue;
      // An arrival point is an endpoint at which a marker resolves. Departures
      // are never compared: the fan-out trunk and the tree stem deliberately
      // emit every connector of a family from one identical point.
      const arrivals = [];
      if (el.attrs['marker-end'] !== undefined) arrivals.push(points[points.length - 1]);
      if (el.attrs['marker-start'] !== undefined) arrivals.push(points[0]);
      connectors.push({ el, points, arrivals, ends: [points[0], points[points.length - 1]] });
    }

    const boxes = [];
    for (const el of usable) {
      if (el.name !== 'rect') continue;
      const rect = rectOf(el);
      if (rect === null) continue;
      boxes.push({ el, rect, edges: edgesOf(rect) });
    }

    // SVG072 - C4: group arrivals by the edge they land on, then compare
    // neighbours against the floor that edge's length selects.
    const groups = new Map();
    for (const connector of connectors) {
      for (const point of connector.arrivals) {
        const bound = bindEndpoint(point, boxes, MARKER_LEN);
        if (bound === null) continue;
        const key = bound.box + ':' + bound.edge;
        if (!groups.has(key)) groups.set(key, []);
        groups.get(key).push({ offset: bound.offset, el: connector.el, bound });
      }
    }
    for (const arrivals of groups.values()) {
      arrivals.sort((a, b) => a.offset - b.offset);
      for (let k = 1; k < arrivals.length; k++) {
        const edgeLength = arrivals[k].bound.length;
        const floor = edgeLength >= C4_LONG_EDGE ? C4_FLOOR_LONG : C4_FLOOR_SHORT;
        const gap = arrivals[k].offset - arrivals[k - 1].offset;
        if (gap >= floor) continue;
        report(
          ERROR,
          'SVG072',
          Math.max(arrivals[k - 1].el.offset, arrivals[k].el.offset),
          `two connectors arrive ${gap.toFixed(1)} units apart on the ${EDGE_NAMES[arrivals[k].bound.edge]} edge of a ${edgeLength.toFixed(0)}-unit box; C4 fans arrivals at least ${floor} units apart there`,
        );
      }
    }

    // Label-mask candidates: a rect whose immediately following sibling is a
    // <text>, and which no connector endpoint attaches to. The attach test runs
    // at the association window rather than the markerLen standoff, so a node a
    // connector merely stops short of is never read as that connector's own
    // label. A mask wrapped away from its text is an accepted false negative.
    const masks = [];
    for (const box of boxes) {
      const siblings = box.el.parent ? box.el.parent.children : [];
      const next = siblings[siblings.indexOf(box.el) + 1];
      if (next === undefined || next.name !== 'text') continue;
      const attached = connectors.some((c) => c.ends.some((p) => touchesRect(p, box.edges, MASK_WINDOW)));
      if (attached) continue;
      masks.push(box);
    }
    const maskSet = new Set(masks.map((m) => m.el));

    // SVG070 / SVG073 - C2: measure against the nearest connector inside the
    // association window; a mask outside it labels nothing and is not checked.
    for (const mask of masks) {
      let nearest = Infinity;
      for (const connector of connectors) {
        const distance = rectPolylineDistance(mask.rect, connector.points);
        if (distance <= MASK_WINDOW && distance < nearest) nearest = distance;
      }
      if (nearest === Infinity) continue;
      if (nearest < C2_MIN_CLEARANCE) {
        report(
          ERROR,
          'SVG070',
          mask.el.offset,
          `label mask clears its connector by ${nearest.toFixed(1)} units; C2 needs ${C2_MIN_CLEARANCE} to ${C2_MAX_CLEARANCE}`,
        );
      } else if (nearest > C2_MAX_CLEARANCE) {
        report(
          WARNING,
          'SVG073',
          mask.el.offset,
          `label mask stands ${nearest.toFixed(1)} units off its connector; C2 keeps it within ${C2_MAX_CLEARANCE}`,
        );
      }
    }

    // SVG071 - C6: a mask may lie fully inside a rect painted after it (the
    // badge chip) or share no area with it. Anything between the two renders
    // the label as a fragment on that rect's border. Other masks are out of
    // scope: C6 constrains overlap with a node, not with a second label.
    for (const mask of masks) {
      for (const later of boxes) {
        if (later.el.offset <= mask.el.offset || maskSet.has(later.el)) continue;
        const a = mask.rect;
        const b = later.rect;
        const overlapX = Math.min(a.x + a.w, b.x + b.w) - Math.max(a.x, b.x);
        const overlapY = Math.min(a.y + a.h, b.y + b.h) - Math.max(a.y, b.y);
        if (overlapX <= 0 || overlapY <= 0) continue; // disjoint, touching included
        const contained = a.x >= b.x && a.y >= b.y && a.x + a.w <= b.x + b.w && a.y + a.h <= b.y + b.h;
        if (contained) continue; // the badge chip C6 allows
        report(
          ERROR,
          'SVG071',
          mask.el.offset,
          `label mask partially overlaps a <${later.el.name}> painted after it, so the label renders clipped on that shape's border`,
        );
      }
    }

    // SVG074 - one aggregate note per file. hasTransform walks ancestors, so a
    // single wrapping <g transform> excludes everything inside it: the count is
    // the transitively excluded population, stated against the whole candidate
    // set rather than the number of elements carrying the attribute.
    if (skipped.length > 0) {
      report(
        WARNING,
        'SVG074',
        svg.offset,
        `${skipped.length} of ${candidates.length} candidate elements carry a transform and were skipped; their geometry is unverified`,
      );
    }
  }

  return diagnostics.sort((a, b) => a.line - b.line || a.column - b.column);
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

function main() {
  const opts = parseArgs(process.argv.slice(2));
  if (opts.error) {
    process.stderr.write(`check-svg: ${opts.error}\n\n${USAGE}\n`);
    process.exit(2);
  }
  if (opts.help) {
    process.stdout.write(`${USAGE}\n`);
    process.exit(0);
  }

  let source;
  try {
    source = readFileSync(opts.file, 'utf8');
  } catch (err) {
    process.stderr.write(`check-svg: cannot read ${opts.file}: ${err.message}\n`);
    process.exit(2);
  }

  const diagnostics = lint(source, opts);
  const errors = diagnostics.filter((d) => d.level === ERROR).length;
  const warnings = diagnostics.length - errors;

  if (opts.json) {
    process.stdout.write(`${JSON.stringify({ file: opts.file, errors, warnings, diagnostics }, null, 2)}\n`);
  } else {
    for (const d of diagnostics) {
      process.stdout.write(`${opts.file}:${d.line}:${d.column}  ${d.level}  ${d.code}  ${d.message}\n`);
    }
    const noun = (count, word) => `${count} ${word}${count === 1 ? '' : 's'}`;
    process.stdout.write(`${noun(errors, 'error')}, ${noun(warnings, 'warning')}\n`);
  }

  if (errors > 0 || (opts.strict && warnings > 0)) process.exit(1);
  process.exit(0);
}

main();
