const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
let renderer = fs.readFileSync(path.join(__dirname, '../hybrid-factory-redesign-20260914/render.cjs'), 'utf8');
const start = renderer.indexOf('const flow=');
const end = renderer.indexOf('let body=');
if (start < 0 || end < start) throw Error('renderer contract changed');
const flow = '<div class="flow-panel"><h3>Claude를 거치지 않는 GPT 메인 추론</h3><div class="flow-head">Claude Code UI · MoAI<small>사용자 입력 · 승인 · hooks · 도구 실행</small></div><div class="flow-arrow">↕</div><div class="flow-head lane">MoAI Messages Bridge<small>세션·history·streaming·tool ID 대응</small></div><div class="flow-arrow">↕</div><div class="flow-head">공식 Codex App Server → GPT<small>본인 구독 인증 · 추론 · tool 요청 · 대화 상태</small></div><p>GPT가 요청한 도구는 Claude Code가 실행하고, 결과는 대기 중인 같은 Codex 요청으로 돌아간다. Claude 모델의 중개 추론은 추가하지 않는다.</p><details><summary>도구 왕복 흐름</summary><pre class="mermaid">flowchart TD\n C[Claude Code] --> B[Messages Bridge]\n B --> A[공식 Codex App Server]\n A --> G[GPT main]\n G --> T[도구 요청]\n T --> B\n B --> X[Claude 승인과 도구 실행]\n X --> R[같은 RPC에 결과 반환]\n R --> A</pre><noscript>Claude Code 요청을 Bridge가 Codex로 전달하고, GPT의 도구 요청을 Claude Code가 실행한 뒤 동일 RPC에 결과를 반환한다.</noscript></details></div>';
const example = '<aside class="example"><b>속도 판정 예시</b><p>답변이 끝난 뒤 SSE로 한 번에 전송되면 형식은 streaming이어도 사용자는 기다리게 된다. 첫 upstream delta, Bridge flush, 화면 표시를 따로 측정해야 한다. 이 예시는 측정 방법이며 지연 실측값은 아니다.</p></aside>';
renderer = renderer.slice(0,start) + 'const flow=' + JSON.stringify(flow) + ';\nconst example=' + JSON.stringify(example) + ';\n' + renderer.slice(end);
renderer = renderer.replaceAll('s9','s7')
 .replaceAll('MoAI Hybrid Factory — 역할·모델·이미지 재설계','moai gpt — GPT 메인 연결의 허용성과 성능')
 .replace('MoAI / HYBRID FACTORY REDESIGN','MoAI / GPT MAIN FEASIBILITY')
 .replace('역할은 유지하고,<br>모델은 작업에 맞게 고른다.','메인 모델은 GPT.<br>허용성과 성능은 따로 검증한다.')
 .replace('기존 cc·glm·gpt와 Factory를 연결하는 공통 실행 정책.<br>3단계 sub-agent, MCP, Gateway, 이미지 생성을 함께 설계한다.','Claude Code를 유지하는 구독 연결 후보와 API 대안.<br>기존 이력 오류·도구 실행·실시간 응답을 다시 검토한다.')
 .replace('일반 모드 + Factory','GPT main 필수')
 .replace('<b>얕게 시작</b>중첩은 필요한 작업만','<b>GPT main</b>Claude 중개 추론 없음')
 .replace('<b>한 정책</b>일반·lane·역할 공통','<b>구독 후보</b>공식 App Server')
 .replace('<b>이미지 분리</b>파일·예산·채택 관리','<b>출시 미완료</b>속도·품질 실측 필요')
 .replace('실계정·비용·3단계 실행은 미검증','실계정 GPT main·품질·속도는 미검증');
vm.runInNewContext(renderer, {require,__dirname,console,process,Buffer}, {filename:'report-renderer.cjs'});
