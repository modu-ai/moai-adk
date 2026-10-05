const fs = require('node:fs');
const path = require('node:path');
const dir = __dirname;
const source = fs.readFileSync(path.join(dir, 'report.md'), 'utf8');
const previous = fs.readFileSync(path.join(dir, '../factory-llm-subscriptions-20260914-01a09dfb/report.html'), 'utf8');
const css = previous.match(/<style>([\s\S]*?)<\/style>/)[1];
const escape = s => s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
const inline = s => escape(s).replace(/\[([^\]]+)\]\(([^)]+)\)/g,'<a href="$2">$1</a>').replace(/`([^`]+)`/g,'<code>$1</code>');
const lines = source.split('\n');
const out = [], toc = [];
let opened = false;
for(let i=0;i<lines.length;) {
 const l=lines[i];
 if(!l.trim()||l.startsWith('# ')){i++;continue;}
 if(l.startsWith('## ')){if(opened)out.push('</section>');const id='s'+(toc.length+1);toc.push({id,title:l.slice(3)});out.push(`<section id="${id}"><h2>${inline(l.slice(3))}</h2>`);opened=true;i++;continue;}
 if(l.startsWith('```')){const c=[];i++;while(i<lines.length&&!lines[i].startsWith('```'))c.push(lines[i++]);i++;out.push('<pre><code>'+escape(c.join('\n'))+'</code></pre>');continue;}
 if(l.startsWith('|')){const r=[];while(i<lines.length&&lines[i].startsWith('|'))r.push(lines[i++].split('|').slice(1,-1));out.push('<div class="table-wrap" tabindex="0" role="region" aria-label="비교표"><table><thead><tr>'+r[0].map(c=>'<th scope="col">'+inline(c)+'</th>').join('')+'</tr></thead><tbody>'+r.slice(2).map(row=>'<tr>'+row.map(c=>'<td>'+inline(c)+'</td>').join('')+'</tr>').join('')+'</tbody></table></div>');continue;}
 if(l.startsWith('- ')){out.push('<ul>');while(i<lines.length&&lines[i].startsWith('- '))out.push('<li>'+inline(lines[i++].slice(2))+'</li>');out.push('</ul>');continue;}
 const p=[];while(i<lines.length&&lines[i].trim()&&!/^(## |```|\||- )/.test(lines[i]))p.push(lines[i++]);out.push('<p>'+inline(p.join(' '))+'</p>');
}
if(opened)out.push('</section>');
const flow=`<div class="flow-panel"><h3>같은 개발 흐름, 선택 가능한 실행 경로</h3><div class="flow-head">Claude Code main / Factory lead<small>사용자 대화 · 선택·승인 · 결과 확인</small></div><div class="flow-arrow">↓</div><div class="flow-head lane">일반 작업 / lane별 역할<small>plan → run → review · 필요할 때만 중첩</small></div><div class="flow-arrow">↓</div><div class="flow-head">공통 실행 정책<small>모델 · effort · 권한 · 예산 · 작업 소유권</small></div><div class="flow-branches"><div>Claude native<small>기존 실행 유지</small></div><div>MCP 위임<small>Codex · GLM API · 이미지</small></div><div>Gateway lane<small>lane 주 모델 선택</small></div></div><details><summary>흐름도 펼치기</summary><pre class="mermaid">flowchart TD
 U[사용자] --> M[Claude Code main 또는 Factory lead]
 M --> R[역할별 작업 및 lane]
 R --> P[공통 모델 권한 예산 정책]
 P --> N[Claude native]
 P --> C[MCP 외부 실행기]
 P --> G[Gateway lane]
 C --> I[이미지 artifact 또는 코드 결과]
 N --> V[검증 근거]
 G --> V
 I --> V
 V --> M
</pre><noscript>사용자 → 메인·lead → 역할·lane → 공통 정책 → native/MCP/Gateway → 산출물 검증 순서로 진행한다.</noscript></details></div>`;
const example=`<aside class="example"><b>적용 예: 이미지가 필요한 기능 개발</b><p>Claude main이 요구를 정리하고 GPT에 복잡한 설계를 맡겨. 명확한 구현은 선택한 모델이 담당하고, 이미지 요청은 별도 예산으로 실행해. 검증이 끝난 코드와 사용자가 채택한 이미지만 반영해. 짧은 수정이면 중간 역할을 만들지 않고 직접 처리해.</p></aside>`;
let body=out.join('\n').replace('<section id="s4">','<section id="s4">'+flow).replace('<section id="s9">','<section id="s9">'+example);
const html=`<!doctype html><html lang="ko"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>MoAI Hybrid Factory — 역할·모델·이미지 재설계</title><link rel="preconnect" href="https://fonts.googleapis.com"><link rel="preconnect" href="https://fonts.gstatic.com" crossorigin><link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;500;700&family=Noto+Serif+KR:wght@500;700&display=swap"><style>${css}</style></head><body><a class="skip" href="#main">본문으로 이동</a><header><div class="hero-tag">MoAI / HYBRID FACTORY REDESIGN</div><h1>역할은 유지하고,<br>모델은 작업에 맞게 고른다.</h1><p>기존 cc·glm·gpt와 Factory를 연결하는 공통 실행 정책.<br>3단계 sub-agent, MCP, Gateway, 이미지 생성을 함께 설계한다.</p><div class="meta"><span>2026-09-14</span><span>설계 제안 · 구현 없음</span><span>일반 모드 + Factory</span></div></header><div class="shell"><nav aria-label="목차">${toc.map(t=>`<a href="#${t.id}">${escape(t.title)}</a>`).join('')}</nav><main id="main"><div class="kpis"><div><b>얕게 시작</b>중첩은 필요한 작업만</div><div><b>한 정책</b>일반·lane·역할 공통</div><div><b>이미지 분리</b>파일·예산·채택 관리</div></div><p class="topline">plan / basic · MoAI-Easy 설정 기준 · <a href="report.md">Markdown 원문</a> · <a href="../factory-llm-subscriptions-20260914-01a09dfb/report.html">이전 조사 보고서</a></p>${body}</main></div><footer>설계 권고와 구현 완료를 구분한다 · 실계정·비용·3단계 실행은 미검증</footer><script type="module">try{const {default:mermaid}=await import('https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs');mermaid.initialize({startOnLoad:false,theme:'base',themeVariables:{primaryColor:'#FAF9F5',primaryTextColor:'#141413',primaryBorderColor:'#D97757',lineColor:'#788C5D'}});document.querySelectorAll('details').forEach(d=>d.addEventListener('toggle',()=>{if(d.open)mermaid.run({nodes:d.querySelectorAll('.mermaid')}).catch(()=>{});}));}catch{}</script></body></html>`;
fs.writeFileSync(path.join(dir,'report.html'),html);
const brokenLocalLinks=[...html.matchAll(/href="([^"#]+)"/g)].map(m=>m[1]).filter(h=>!/^https?:/.test(h)&&!fs.existsSync(path.resolve(dir,h)));
console.log(JSON.stringify({bytes:Buffer.byteLength(html),sections:toc.length,brokenLocalLinks}));
if(brokenLocalLinks.length||Buffer.byteLength(html)>120000)process.exitCode=1;
