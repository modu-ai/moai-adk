#!/usr/bin/env python3
import os,json,pathlib,tempfile,subprocess,threading,http.server,time,signal,sys
base=pathlib.Path(__file__).parent
if '--mcp' in sys.argv:
 for line in sys.stdin:
  try:x=json.loads(line)
  except ValueError:continue
  if 'id' not in x:continue
  method=x.get('method')
  r={'protocolVersion':'2024-11-05','capabilities':{'tools':{}},'serverInfo':{'name':'fixture','version':'1'}} if method=='initialize' else {'tools':[{'name':'echo','description':'Synthetic echo fixture. Return the input marker.','inputSchema':{'type':'object','properties':{'marker':{'type':'string'}},'required':['marker'],'additionalProperties':False}}]} if method=='tools/list' else {'content':[{'type':'text','text':'NO_EXECUTION_EXPECTED'}]}
  print(json.dumps({'jsonrpc':'2.0','id':x['id'],'result':r}),flush=True)
 raise SystemExit
result={'claude_version':'2.1.269','requests':[],'errors':[],'cli_events':[]};lock=threading.Lock()
def refs(v):
 out=[]
 if isinstance(v,dict):
  if v.get('type')=='tool_reference':out.append(v.get('tool_name'))
  for x in v.values():out+=refs(x)
 elif isinstance(v,list):
  for x in v:out+=refs(x)
 return out
def identity(d,headers):
 out={'metadata_keys':sorted(d.get('metadata',{})),'identifiers':{},'header_keys':sorted(k.lower() for k in headers if k.lower() not in ['authorization','x-api-key'])}
 user=d.get('metadata',{}).get('user_id')
 if isinstance(user,str):
  try:
   u=json.loads(user)
   out['user_id_keys']=sorted(u)
   out['identifiers']={k:v for k,v in u.items() if any(x in k.lower() for x in ['session','agent','parent'])}
  except ValueError:
   import re
   out['identifiers']={'session_ids':re.findall(r'session[_:]([a-f0-9-]{36})',user)}
 for k,v in headers.items():
  if any(x in k.lower() for x in ['session','agent','parent']) and k.lower() not in ['user-agent']:
   out.setdefault('identity_headers',{})[k]=v
 return out
class Handler(http.server.BaseHTTPRequestHandler):
 def log_message(self,*a):pass
 def do_POST(self):
  n=int(self.headers.get('Content-Length','0'));d=json.loads(self.rfile.read(n))
  if self.path.endswith('count_tokens'):
   b=b'{"input_tokens":1000}';self.send_response(200);self.send_header('Content-Type','application/json');self.end_headers();self.wfile.write(b);return
  tools=[{k:t[k] for k in ['name','type','defer_loading','input_schema'] if k in t} for t in d.get('tools',[])]
  with lock:result['requests'].append({'path':self.path,'tools':tools,'tool_references':refs(d.get('messages',[])),'identity':identity(d,self.headers),'messages_full':d.get('messages',[])});idx=len(result['requests'])
  if idx==1 and 'fork' in json.dumps(next((t.get('input_schema') for t in tools if t.get('name')=='Agent'),{})):
   result['schema_mentions_fork']=True
   block={'type':'tool_use','id':'toolu_fixture_fork_RANDOM_UNIQUE','name':'Agent','input':{'description':'Synthetic identity probe','prompt':'Reply only CHILD_IDENTITY_OK. Do not use tools.','subagent_type':'fork'}};stop='tool_use'
  else:block={'type':'text','text':'CHILD_IDENTITY_OK' if idx==2 else 'PARENT_IDENTITY_OK'};stop='end_turn'
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
 (home/'.claude.json').write_text(json.dumps({'hasCompletedOnboarding':True,'projects':{str(cwd):{'hasTrustDialogAccepted':True}}}))
 env={k:v for k,v in os.environ.items() if not k.startswith(('ANTHROPIC_','CLAUDE_','MOAI_')) and k not in ['CLAUDECODE','TMUX','TMUX_PANE']}
 env.update(HOME=str(home),CLAUDE_CONFIG_DIR=str(home),ANTHROPIC_API_KEY='synthetic-local-fixture',ANTHROPIC_BASE_URL='http://127.0.0.1:'+str(server.server_port),CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC='1',ENABLE_TOOL_SEARCH='true')
 cmd=['/Users/goos/.local/bin/claude','--print','--verbose','--output-format','stream-json','--max-turns','3','--permission-mode','default','--settings','{"disableAllHooks":true}','--strict-mcp-config','--mcp-config',json.dumps({'mcpServers':{'fixture':{'command':sys.executable,'args':[str(pathlib.Path(__file__).resolve()),'--mcp']}}}),'--','Delegate to one fork Agent to reply CHILD_IDENTITY_OK without using tools. Then reply PARENT_IDENTITY_OK.']
 start=time.monotonic();p=subprocess.Popen(cmd,cwd=cwd,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,start_new_session=True)
 try:
  out,err=p.communicate(timeout=45);result['exit_code']=p.returncode;result['final_marker_seen']=b'PARENT_IDENTITY_OK' in out
  for line in out.decode('utf-8','replace').splitlines():
   try:x=json.loads(line)
   except ValueError:continue
   result['cli_events'].append({k:x.get(k) for k in ['type','subtype','session_id','parent_tool_use_id'] if k in x})
 except subprocess.TimeoutExpired:result['errors'].append('timeout');out=b''
 finally:
  try:os.killpg(p.pid,signal.SIGTERM)
  except ProcessLookupError:pass
  except PermissionError:result['cleanup_permission_error']=True
  try:p.wait(timeout=3)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait(timeout=3)
 result['duration_seconds']=round(time.monotonic()-start,2)
server.shutdown()
result['success']=len(result['requests'])>=3 and result.get('final_marker_seen',False)
(base/'fork-history-shape-result.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='requests'}));print('requests='+str(len(result['requests'])));raise SystemExit(0 if result['success'] else 1)
