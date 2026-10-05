const fs = require('node:fs');
const path = require('node:path');
const dir = __dirname;
const md = fs.readFileSync(path.join(dir, 'report.md'), 'utf8');
const esc = s => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
function inline(s) {
  const held = [];
  s = s.replace(/`([^`]+)`/g, (_, v) => `@@${held.push(`<code>${esc(v)}</code>`) - 1}@@`);
  s = esc(s).replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, label, href) => `<a href="${href}">${label}</a>`).replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  return s.replace(/@@(\d+)@@/g, (_, n) => held[+n]);
}
const lines = md.split('\n');
const chunks = [];
const toc = [];
let sectionOpen = false;
for (let i = 0; i < lines.length;) {
  const l = lines[i];
  if (!l.trim() || l.startsWith('# ')) { i++; continue; }
  if (l.startsWith('## ')) {
    if (sectionOpen) chunks.push('</section>');
    const name = l.slice(3), id = `s${toc.length + 1}`;
    toc.push({id, name});
    chunks.push(`<section id="${id}"><h2>${inline(name)}</h2>`);
    sectionOpen = true; i++; continue;
  }
  if (l.startsWith('```')) {
    const code = []; i++;
    while (i < lines.length && !lines[i].startsWith('```')) code.push(lines[i++]);
    i++;
    chunks.push(`<pre><code>${esc(code.join('\n'))}</code></pre>`); continue;
  }
  if (l.startsWith('|')) {
    const rows = [];
    while (i < lines.length && lines[i].startsWith('|')) rows.push(lines[i++].split('|').slice(1, -1).map(s => s.trim()));
    chunks.push('<div class="table-wrap" tabindex="0" role="region" aria-label="가로로 스크롤할 수 있는 비교표"><table><thead><tr>' + rows[0].map(c => `<th scope="col">${inline(c)}</th>`).join('') + '</tr></thead><tbody>' + rows.slice(2).map(row => '<tr>' + row.map(c => `<td>${inline(c)}</td>`).join('') + '</tr>').join('') + '</tbody></table></div>'); continue;
  }
  if (/^(- |\d+\. )/.test(l)) {
    const ordered = /^\d/.test(l), tag = ordered ? 'ol' : 'ul';
    chunks.push(`<${tag}>`);
    while (i < lines.length && (ordered ? /^\d+\. / : /^- /).test(lines[i])) chunks.push('<li>' + inline(lines[i++].replace(/^(- |\d+\. )/, '')) + '</li>');
    chunks.push(`</${tag}>`); continue;
  }
  const paragraph = [];
  while (i < lines.length && lines[i].trim() && !/^(## |```|\||- |\d+\. )/.test(lines[i])) paragraph.push(lines[i++]);
  chunks.push('<p>' + inline(paragraph.join(' ')) + '</p>');
}
if (sectionOpen) chunks.push('</section>');
const flow = `<div class="flow-panel"><p class="eyebrow">권고 구조 · 구현 전 제안</p><h3>Claude Code에서 시작하고, 전문 실행기로 나눈다</h3><div class="fallback-flow"><div class="flow-head">Claude Code Lead <small>사용자 지시 · 배정 · 결과 확인</small></div><div class="flow-arrow">↓</div><div class="flow-head lane">Claude Code Lanes <small>각 lane의 에이전트가 모델·effort를 선택해 위임</small></div><div class="flow-arrow">↓</div><div class="flow-head">moai-mcp <small>owner · job · thread · 모델 정책 · 사용 한도</small></div><div class="flow-branches"><div>GPT<small>공식 Codex 실행부<br>본인 ChatGPT 로그인</small></div><div>GLM / 외부 모델<small>허용된 API·플랜<br>공급자별 조건 확인</small></div><div>이미지 작업<small>별도 생성 능력<br>파일 결과 검증</small></div></div></div><details><summary>세션 모델 교체 경로까지 함께 보기</summary><pre class="mermaid">flowchart TD
 U[사용자] --> C[Claude Code lead와 lanes]
 C --> M[moai-mcp 작업 위임]
 M --> X[공식 Codex App Server]
 M --> G[허용된 GLM 및 외부 API]
 X --> A[산출물과 검증 근거]
 G --> A
 A --> C
 C -. 세션 자체 모델 교체를 선택할 때 .-> W[MoAI Messages Gateway]
 W -. Anthropic 비지원 경계 .-> X
</pre><noscript><p>작업 위임은 Claude Code → moai-mcp → 공식 Codex 또는 외부 API → 산출물 확인 순서다. 세션 자체 모델 교체는 별도 Gateway 경로다.</p></noscript></details><p class="caption">정책·작업 접점은 공통으로 관리하고, 모델 추론과 도구 실행 권한은 공급자별로 검증한다.</p></div>`;
let body = chunks.join('\n');
body = body.replace('<section id="s7">', '<section id="s7">' + flow);
body = body.replace('<section id="s8">', '<section id="s8"><aside class="example"><b>적용 예시</b><p>lane-1이 GPT에 설계를 맡기고, lane-2가 허용된 GLM API로 테스트 후보를 만든다. GPT는 설계 문서, GLM은 검토할 텍스트를 반환한다. Claude Code가 결과를 읽고 실제 변경을 수행하면 초기에는 외부 실행기에 쓰기 권한을 줄 필요가 없다. 독립 구현을 맡길 때만 해당 worktree의 쓰기를 허용한다.</p></aside>');
const html = `<!doctype html>
<html lang="ko"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="color-scheme" content="light"><title>Claude Code 다중 LLM Factory — 구독·MCP·Gateway 최종 비교</title>
<link rel="preconnect" href="https://fonts.googleapis.com"><link rel="preconnect" href="https://fonts.gstatic.com" crossorigin><link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;500;700&family=Noto+Serif+KR:wght@500;700&display=swap">
<style>
:root{--ivory:#FAF9F5;--paper:#fff;--slate:#141413;--clay:#D97757;--clay-d:#9b4228;--oat:#E3DACC;--olive:#788C5D;--g100:#F0EEE6;--g300:#D1CFC5;--g500:#68675f;--g700:#3D3D3A;--sans:'Noto Sans KR',-apple-system,sans-serif;--serif:'Noto Serif KR',serif;--mono:ui-monospace,SFMono-Regular,monospace;--max-width:1180px;--radius-panel:16px;--radius-row:8px;--border:1px solid var(--g300)}
*{box-sizing:border-box}html{scroll-behavior:smooth;scroll-padding-top:28px}body{margin:0;background:var(--ivory);color:var(--g700);font-family:var(--sans);font-size:15px;line-height:1.85}a{color:var(--clay-d);text-underline-offset:4px;overflow-wrap:anywhere}a:hover{color:#652a17}a:focus-visible,summary:focus-visible,.table-wrap:focus-visible{outline:3px solid var(--clay);outline-offset:4px}header{background:#19241e;color:#f9f7ef;padding:72px max(28px,calc((100vw - 1180px)/2)) 55px}.eyebrow{font-size:12px;letter-spacing:.1em;font-weight:700;color:#b66345}.hero-tag{color:#d5cabb;font-size:12px;letter-spacing:.12em}.hero-grid{display:grid;grid-template-columns:1fr 250px;gap:40px;align-items:end}h1{font-family:var(--serif);font-size:clamp(32px,4.2vw,52px);line-height:1.4;letter-spacing:-.035em;font-weight:700;max-width:830px;margin:20px 0}header p{max-width:750px;color:#dce0d8;margin:10px 0;font-size:17px}.hero-note{border-top:2px solid var(--clay);padding-top:18px;color:#ddd4c3;font-size:13px}.hero-note b{display:block;color:#fff;font-size:20px;margin-bottom:8px}.meta{display:flex;flex-wrap:wrap;gap:8px;margin-top:30px}.meta span{padding:4px 12px;border:1px solid #536053;border-radius:20px;font-size:12px;color:#e2e7dd}.shell{max-width:var(--max-width);margin:0 auto;padding:42px 24px 100px;display:grid;grid-template-columns:215px minmax(0,1fr);gap:38px}nav{position:sticky;top:24px;align-self:start;max-height:calc(100vh - 45px);overflow:auto;font-size:12px}nav strong{display:block;color:var(--slate);margin-bottom:16px;font-size:13px}nav a{display:block;text-decoration:none;color:#68675f;padding:7px 0;border-bottom:1px solid #e7e2d8;line-height:1.65}main{min-width:0}section{margin-bottom:48px;background:var(--paper);border:var(--border);padding:30px;border-radius:var(--radius-panel)}h2{font-family:var(--serif);color:var(--slate);font-size:24px;line-height:1.55;letter-spacing:-.02em;margin:0 0 22px;padding-bottom:15px;border-bottom:2px solid var(--oat)}h3{font-size:20px;line-height:1.55;color:var(--slate)}p{margin:15px 0;overflow-wrap:anywhere}strong{color:var(--slate)}li{margin:10px 0;padding-left:3px}ul,ol{padding-left:24px}code{font-family:var(--mono);font-size:.9em;overflow-wrap:anywhere}p code,li code,td code{background:var(--g100);padding:2px 5px;border-radius:4px}pre{white-space:pre-wrap;overflow-wrap:anywhere;background:#f2f1ea;padding:18px;border:1px solid #e0ddd3;border-radius:9px;line-height:1.7;font-size:12px}.table-wrap{overflow:auto;margin:24px 0;border:var(--border);border-radius:9px}table{border-collapse:collapse;width:100%;font-size:13px;min-width:620px}th,td{padding:14px 15px;text-align:left;vertical-align:top;border-bottom:1px solid #e8e4dc;line-height:1.8}th{background:#ecece3;color:#303c30;font-weight:700}tr:last-child td{border-bottom:0}td:first-child{font-weight:500;color:#27372b}tbody tr:nth-child(even){background:#fafaf6}.flow-panel{background:#f2f3ea;border:1px solid #d8ddcd;border-radius:12px;padding:24px;margin-bottom:32px}.flow-panel h3{margin:8px 0 22px}.flow-head{text-align:center;padding:14px;background:#fff;border:1px solid #b9c4b0;border-radius:10px;color:#203b2a;font-weight:700}.flow-head.lane{background:#e0e8d8}.flow-arrow{text-align:center;font-size:24px;color:#788C5D;line-height:1.5}small{display:block;font-size:11px;line-height:1.8;font-weight:400;margin-top:5px}.flow-branches{display:grid;grid-template-columns:repeat(3,1fr);gap:8px;margin-top:18px}.flow-branches>div{background:#fff;padding:14px 8px;border:1px solid #d4d9c8;border-top:3px solid var(--olive);border-radius:8px;text-align:center;font-size:13px;font-weight:700}.caption{font-size:12px;color:#66695c}details{margin-top:20px;border-top:1px solid #ced5c2;padding-top:15px}summary{cursor:pointer;font-size:13px;color:#425638}.mermaid{background:white}.mermaid svg{max-width:100%;height:auto}.example{border-left:4px solid var(--clay);padding:18px 22px;background:#fcf2ec;margin-bottom:25px}.example b{color:#9b4228}.example p{font-size:14px;margin-bottom:0}.topline{font-size:12px;color:var(--g500);margin:0 0 18px}.kpis{display:grid;grid-template-columns:repeat(3,1fr);gap:12px;margin:0 0 30px}.kpis div{background:#eee9de;border-radius:12px;padding:20px 15px;line-height:1.6;font-size:12px}.kpis b{font-size:19px;display:block;color:#273b2b;margin-bottom:4px}footer{border-top:var(--border);padding:24px;text-align:center;font-size:12px;color:var(--g500)}.skip{position:absolute;left:20px;top:-100px;background:white;padding:10px;z-index:3}.skip:focus{top:12px}
@media(max-width:960px){.shell{grid-template-columns:1fr;gap:20px}nav{position:static;max-height:none;display:flex;gap:12px;overflow:auto;padding-bottom:12px}nav a{min-width:130px;max-width:160px;flex-shrink:0}nav strong{display:none}.hero-grid{grid-template-columns:1fr}.hero-note{max-width:480px}.hero-grid{gap:15px}}
@media(max-width:600px){header{padding:38px 20px}.shell{padding:24px 12px 50px}section{padding:21px 16px;margin-bottom:24px}h2{font-size:21px}.kpis{gap:6px}.kpis div{padding:14px 9px;font-size:11px}.kpis b{font-size:16px}.flow-panel{padding:15px}.flow-branches{grid-template-columns:1fr}.flow-branches>div{padding:10px}body{font-size:14px}header p{font-size:15px}.hero-note b{font-size:18px}}
@media print{body{background:white;font-size:10pt}header{background:white;color:black;padding:0 0 20px}header p,.hero-tag,.hero-note,.hero-note b{color:black}.meta,nav,.skip{display:none}.shell{display:block;padding:0}section{padding:18px 0;border:0;border-radius:0;margin-bottom:15px}h1{font-size:26pt}h2{font-size:16pt;break-after:avoid}table{min-width:0;font-size:8pt}.table-wrap{overflow:visible}tr{break-inside:avoid}pre{font-size:8pt}a{color:#222}details{display:none}.flow-panel{break-inside:avoid}.hero-grid{display:block}.kpis{break-inside:avoid}footer{padding:12px}}
</style></head><body><a class="skip" href="#main">본문으로 이동</a><header><div class="hero-tag">MoAI RESEARCH / ARCHITECTURE DECISION</div><div class="hero-grid"><div><h1>Claude Code 안에서,<br>모델마다 잘하는 일을 맡긴다.</h1><p>GPT·Claude 구독, GLM과 외부 코딩 플랜.<br>MCP·Gateway·OMP·OpenCode를 비교한 Factory 통합안.</p></div><div class="hero-note"><b>권고: MCP 중심의 분업</b>기존 MoAI task 도구를 확장하고,<br>세션 모델 교체는 Gateway로 선택.<br>공식 선례와 운영 미검증을 구분했다.</div></div><div class="meta"><span>2026-09-14 조사</span><span>Claude Code 필수</span><span>lead + lanes</span><span>설계 보고 · 제품 변경 없음</span></div></header><div class="shell"><nav aria-label="보고서 목차"><strong>조사와 결정</strong>${toc.map(t => `<a href="#${t.id}">${esc(t.name)}</a>`).join('')}</nav><main id="main"><div class="kpis"><div><b>공식 선례 있음</b>OpenAI Claude Code 플러그인</div><div><b>기존 도구 재사용</b>codex_task · glm_task</div><div><b>조건별 경로 분리</b>구독 · 자동화 · 모델 교체</div></div><p class="topline">plan / basic · MoAI-Easy 설정 기준 · <a href="report.md">Markdown 원문</a> · <a href="../gpt-final-design-20260914-b40ff9f4/report.html">기존 기획안</a></p>${body}</main></div><footer>MoAI · source session 01a09dfb-5908-7e50-ac3a-bde14d9af79f · 공개 문서 조사와 로컬 선택 테스트 / 실계정 운영 검증은 별도</footer><script type="module">try { const {default:mermaid}=await import('https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs'); mermaid.initialize({startOnLoad:false,theme:'base',themeVariables:{primaryColor:'#FAF9F5',primaryTextColor:'#141413',primaryBorderColor:'#D97757',lineColor:'#788C5D',fontFamily:'Noto Sans KR, sans-serif'}}); document.querySelectorAll('details').forEach(d=>d.addEventListener('toggle',()=>{if(d.open) mermaid.run({nodes:d.querySelectorAll('.mermaid')}).catch(()=>{});})); } catch {} </script></body></html>`;
fs.writeFileSync(path.join(dir, 'report.html'), html);
console.log(JSON.stringify({output:path.join(dir,'report.html'),bytes:Buffer.byteLength(html),sections:toc.length,markdownBytes:Buffer.byteLength(md)}));
