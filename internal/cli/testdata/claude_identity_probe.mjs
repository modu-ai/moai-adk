import http from 'node:http';
import {spawn} from 'node:child_process';
import {mkdtempSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {createHash} from 'node:crypto';
// Run with a PTY: node internal/cli/testdata/claude_identity_probe.mjs
// This starts only a loopback mock, with a fresh Claude config and fake API key.
// Confirm theme, the fake key, and trust ONLY the printed temporary directory.
// The 38s synthetic child delay lets Claude's 30s agent-summary timer fire.
// --print is intentionally not used: it does not issue agent-summary requests.
const dir=process.env.PROBE_DIR||mkdtempSync(join(tmpdir(),'moai-identity-'));
let count=0;
const server=http.createServer(async(req,res)=>{
 let chunks=[],size=0; for await(const c of req){size+=c.length;if(size>1<<20){res.writeHead(413);res.end();return;}chunks.push(c);}
 let b;try{b=JSON.parse(Buffer.concat(chunks))}catch{res.writeHead(200);res.end('{}');return;}
 const n=++count;
 const sys=JSON.stringify(b.system);
 console.log(JSON.stringify({n,url:req.url,headers:Object.fromEntries(Object.entries(req.headers).filter(([k])=>['user-agent','x-claude-code-session-id','x-claude-code-agent-id','content-type','content-length'].includes(k))),keys:Object.keys(b),systemHash:createHash('sha256').update(sys||'').digest('hex'),systemTail:sys?.slice(-512),messages:b.messages?.map(m=>({role:m.role,contentPrefix:JSON.stringify(m.content).slice(0,512),content:JSON.stringify(m.content).slice(-4096)})),tools:b.tools?.map(t=>t.name)}));
 if(req.url.includes('count_tokens')){res.writeHead(200);res.end('{"input_tokens":100}');return;}
 const child=sys?.includes('CHILD_PROBE');
 const childTools=process.env.PROBE_TOOLS&&child&&b.tools?.some(t=>t.name==='Bash');
 const hasChildResult=b.messages?.some(m=>JSON.stringify(m.content).includes('toolu_child_probe'));
 if(child&&(!childTools||hasChildResult)) await new Promise(resolve=>setTimeout(resolve,38000));
 const isResult=b.messages?.some(m=>JSON.stringify(m.content).includes('tool_result'));
 const content=(childTools&&!hasChildResult)?[{type:'tool_use',id:'toolu_child_probe',name:'Bash',input:{command:'printf CHILD_TOOL_OK',description:'Local fixture output'}}]:(!child&&!isResult&&b.tools?.some(t=>t.name==='Agent'))?[{type:'tool_use',id:'toolu_probe_one',name:'Agent',input:{description:'probe child',prompt:'Reply CHILD_OK',subagent_type:'probe'}}]:[{type:'text',text:child?'CHILD_OK':'MAIN_OK'}];
 const msg={id:'msg_probe_'+n,type:'message',role:'assistant',model:b.model,content,stop_reason:content[0].type==='tool_use'?'tool_use':'end_turn',stop_sequence:null,usage:{input_tokens:100,output_tokens:10}};
 if(!b.stream){res.writeHead(200,{'content-type':'application/json'});res.end(JSON.stringify(msg));return;}
 res.writeHead(200,{'content-type':'text/event-stream'});
 const send=(type,v)=>res.write(`event: ${type}\ndata: ${JSON.stringify({type,...v})}\n\n`);
 send('message_start',{message:{...msg,content:[],stop_reason:null}});
 content.forEach((c,index)=>{send('content_block_start',{index,content_block:c.type==='text'?{type:'text',text:''}:{...c,input:{}}});send('content_block_delta',{index,delta:c.type==='text'?{type:'text_delta',text:c.text}:{type:'input_json_delta',partial_json:JSON.stringify(c.input)}});send('content_block_stop',{index});});
 send('message_delta',{delta:{stop_reason:msg.stop_reason,stop_sequence:null},usage:{output_tokens:10}});send('message_stop',{});res.end();
});
server.listen(0,'127.0.0.1',()=>{
 const env=Object.fromEntries(['PATH','HOME','TMPDIR','TERM','LANG'].filter(k=>process.env[k]).map(k=>[k,process.env[k]]));
 Object.assign(env,{ANTHROPIC_API_KEY:'local-probe',ANTHROPIC_BASE_URL:`http://127.0.0.1:${server.address().port}`,CLAUDE_CONFIG_DIR:dir,CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC:'1'});
 const args=['--model','claude-sonnet-4-5','--tools',process.env.PROBE_COMPACT?'':process.env.PROBE_TOOLS?'Agent,Bash':'Agent','--allowedTools',process.env.PROBE_TOOLS?'Agent,Bash':'Agent','--agents',JSON.stringify({probe:{description:'probe child',prompt:'CHILD_PROBE: answer briefly',tools:process.env.PROBE_TOOLS?['Bash']:[]}}),process.env.PROBE_COMPACT?'Remember COLOR=blue and reply MAIN_OK':'Use the probe agent then reply MAIN_OK'];
 const child=spawn(process.env.CLAUDE_BINARY||'claude',args,{cwd:dir,env,stdio:'inherit'});
 const timer=setTimeout(()=>child.kill('SIGTERM'),55000);
 child.on('exit',code=>{clearTimeout(timer);server.closeAllConnections();server.close();console.error('exit='+code)});
});
