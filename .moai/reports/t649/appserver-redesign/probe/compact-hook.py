#!/usr/bin/env python3
import os,json,pathlib,tempfile,subprocess,threading,http.server,time,signal,sys,shlex
base=pathlib.Path(__file__).parent
if '--hook' in sys.argv:
 data=json.load(sys.stdin)
 row={k:data.get(k) for k in ['hook_event_name','session_id','trigger','source'] if k in data};row['time_ns']=time.time_ns();row['kind']='hook';row['transcript_path_present']=bool(data.get('transcript_path'))
 with open(os.environ['MOAI_COMPACT_FIXTURE_LOG'],'a') as f:f.write(json.dumps(row)+'\n')
 raise SystemExit
if '--mcp' in sys.argv:
 for line in sys.stdin:
  try:x=json.loads(line)
  except ValueError:continue
  if 'id' not in x:continue
  method=x.get('method')
  r={'protocolVersion':'2024-11-05','capabilities':{'tools':{}},'serverInfo':{'name':'fixture','version':'1'}} if method=='initialize' else {'tools':[{'name':'echo','description':'Synthetic echo fixture. Return the input marker.','inputSchema':{'type':'object','properties':{'marker':{'type':'string'}},'required':['marker'],'additionalProperties':False}}]} if method=='tools/list' else {'content':[{'type':'text','text':'NO_EXECUTION_EXPECTED'}]}
  print(json.dumps({'jsonrpc':'2.0','id':x['id'],'result':r}),flush=True)
 raise SystemExit
result={'claude_version':'2.1.269','requests':[],'errors':[],'invocations':[]};lock=threading.Lock()
def refs(v):
 out=[]
 if isinstance(v,dict):
  if v.get('type')=='tool_reference':out.append(v.get('tool_name'))
  for x in v.values():out+=refs(x)
 elif isinstance(v,list):
  for x in v:out+=refs(x)
 return out
class Handler(http.server.BaseHTTPRequestHandler):
 def log_message(self,*a):pass
 def do_POST(self):
  n=int(self.headers.get('Content-Length','0'));d=json.loads(self.rfile.read(n))
  if self.path.endswith('count_tokens'):
   b=b'{"input_tokens":1000}';self.send_response(200);self.send_header('Content-Type','application/json');self.end_headers();self.wfile.write(b);return
  with open(eventlog,'a') as f:f.write(json.dumps({'kind':'http_request','time_ns':time.time_ns(),'path':self.path})+'\n')
  shape={'path':self.path,'keys':sorted(d),'headers':{k:v for k,v in self.headers.items() if k.lower() in ['anthropic-beta','anthropic-version','content-type','x-app']},'model':d.get('model'),'max_tokens':d.get('max_tokens'),'stream':d.get('stream'),'context_management':d.get('context_management'),'metadata_keys':sorted(d.get('metadata',{})),'system_blocks':[{'type':x.get('type'),'text_length':len(x.get('text',''))} for x in d.get('system',[]) if isinstance(x,dict)],'messages':[{'role':m.get('role'),'content_types':[x.get('type') for x in m.get('content',[]) if isinstance(x,dict)] if isinstance(m.get('content'),list) else ['string']} for m in d.get('messages',[])]}
  with lock:result['requests'].append(shape);idx=len(result['requests'])
  block={'type':'text','text':'SYNTHETIC_SUMMARY_OK'};stop='end_turn'
  msg={'id':'msg_fixture_'+str(idx),'type':'message','role':'assistant','model':d.get('model','claude-sonnet-4-6'),'content':[],'stop_reason':None,'stop_sequence':None,'usage':{'input_tokens':1000,'output_tokens':1}}
  events=[('message_start',{'type':'message_start','message':msg}),('content_block_start',{'type':'content_block_start','index':0,'content_block':dict(block,**({'input':{}} if block['type']=='tool_use' else {'text':''}))})]
  delta={'type':'input_json_delta','partial_json':json.dumps(block['input'])} if block['type']=='tool_use' else {'type':'text_delta','text':block['text']}
  events += [('content_block_delta',{'type':'content_block_delta','index':0,'delta':delta}),('content_block_stop',{'type':'content_block_stop','index':0}),('message_delta',{'type':'message_delta','delta':{'stop_reason':stop,'stop_sequence':None},'usage':{'output_tokens':20}}),('message_stop',{'type':'message_stop'})]
  self.send_response(200);self.send_header('Content-Type','text/event-stream');self.end_headers()
  try:
   for name,e in events:self.wfile.write(('event: '+name+'\ndata: '+json.dumps(e)+'\n\n').encode());self.wfile.flush()
  except BrokenPipeError:pass
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Handler);threading.Thread(target=server.serve_forever,daemon=True).start()
with tempfile.TemporaryDirectory(prefix='moai-claude-tool-surface-') as temp:
 root=pathlib.Path(temp);home=root/'home';cwd=root/'cwd';home.mkdir();cwd.mkdir()
 eventlog=root/'events.jsonl';eventlog.touch()
 (home/'.claude.json').write_text(json.dumps({'hasCompletedOnboarding':True,'projects':{str(cwd):{'hasTrustDialogAccepted':True}}}))
 env={k:v for k,v in os.environ.items() if not k.startswith(('ANTHROPIC_','CLAUDE_','MOAI_')) and k not in ['CLAUDECODE','TMUX','TMUX_PANE']}
 env.update(HOME=str(home),MOAI_COMPACT_FIXTURE_LOG=str(eventlog),CLAUDE_CONFIG_DIR=str(home),ANTHROPIC_API_KEY='synthetic-local-fixture',ANTHROPIC_BASE_URL='http://127.0.0.1:'+str(server.server_port),CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC='1',ENABLE_TOOL_SEARCH='true')
 hookcmd=shlex.quote(sys.executable)+' '+shlex.quote(str(pathlib.Path(__file__).resolve()))+' --hook'
 settings={'hooks':{name:[{'matcher':matcher,'hooks':[{'type':'command','command':hookcmd,'timeout':5}]}] for name,matcher in [('PreCompact','manual'),('PostCompact','manual'),('SessionStart','compact')]}}
 cmd=['/Users/goos/.local/bin/claude','--print','--verbose','--output-format','stream-json','--max-turns','3','--permission-mode','default','--settings',json.dumps(settings),'--strict-mcp-config','--mcp-config',json.dumps({'mcpServers':{'fixture':{'command':sys.executable,'args':[str(pathlib.Path(__file__).resolve()),'--mcp']}}}),'--','Discover the synthetic fixture echo tool with ToolSearch, then reply CLAUDE_TOOL_SURFACE_OK. Do not execute the echo tool.']
 start=time.monotonic();session=None
 for phase in ['seed','compact']:
  call=list(cmd)
  call[-1]='Remember the synthetic fixture marker BLUEBIRD42. Reply briefly.' if phase=='seed' else '/compact'
  if phase=='compact':
   if not session:result['errors'].append('no_seed_session_id');break
   call=call[:-2]+['--resume',session]+call[-2:]
  child=subprocess.Popen(call,cwd=cwd,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,start_new_session=True)
  before=len(result['requests'])
  try:
   out,err=child.communicate(timeout=45);events=[];messages=[]
   for line in out.decode('utf-8','replace').splitlines():
    try:x=json.loads(line)
    except ValueError:continue
    events.append({'type':x.get('type'),'subtype':x.get('subtype'),'is_error':x.get('is_error')})
    if x.get('session_id'):session=x['session_id']
    if x.get('type')=='result':messages.append(str(x.get('result',''))[:500])
   result['invocations'].append({'phase':phase,'exit_code':child.returncode,'events':events,'result_texts':messages,'new_request_count':len(result['requests'])-before,'stderr_excerpt':err.decode('utf-8','replace')[:500]})
  except subprocess.TimeoutExpired:result['errors'].append(phase+'_timeout')
  finally:
   try:os.killpg(child.pid,signal.SIGTERM)
   except ProcessLookupError:pass
   except PermissionError:result.setdefault('cleanup_permission_errors',[]).append(phase)
   try:child.wait(timeout=3)
   except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait(timeout=3)
 result['ordered_events']=sorted([json.loads(line) for line in eventlog.read_text().splitlines()],key=lambda x:x['time_ns'])
 result['duration_seconds']=round(time.monotonic()-start,2)
server.shutdown()
result['success']=len(result['invocations'])==2 and result['invocations'][1]['exit_code']==0 and all(any(e.get('hook_event_name')==name for e in result.get('ordered_events',[])) for name in ['PreCompact','PostCompact','SessionStart'])
(base/'compact-hook-result.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='requests'}));print('requests='+str(len(result['requests'])));raise SystemExit(0 if result['success'] else 1)
