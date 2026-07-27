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
