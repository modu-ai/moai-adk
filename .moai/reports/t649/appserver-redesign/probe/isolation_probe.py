#!/usr/bin/env python3
import subprocess,os,json,time,selectors,signal,pathlib,tomllib
base=pathlib.Path(__file__).parent
result={'codex_version':'0.154.0','events':[],'errors':[],'tool_calls':[],'final_texts':[]}
cmd=['codex','app-server','--stdio','-c','approval_policy="never"','-c','sandbox_mode="read-only"','-c','web_search="disabled"','-c','features.shell_tool=false']
flags=['features.codex_hooks=false','features.hooks=false','features.plugin_hooks=false','features.plugins=false','features.apps=false','features.view_image=false','features.multi_agent=false','features.multi_agent_v2=false','tools.update_plan.enabled=false','tools.experimental_request_user_input.enabled=false']
usercfg=tomllib.loads((pathlib.Path(os.environ.get('CODEX_HOME',str(pathlib.Path.home()/'.codex')))/'config.toml').read_text())
for name in usercfg.get('mcp_servers',{}):flags.append('mcp_servers.'+name+'.enabled=false')
for flag in flags:cmd+=['-c',flag]
result['isolation_overrides']=flags;result['environments']=[]
env=os.environ.copy()
for k in ['OPENAI_API_KEY','ANTHROPIC_API_KEY']:env.pop(k,None)
p=subprocess.Popen(cmd,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.PIPE,start_new_session=True,env=env)
sel=selectors.DefaultSelector();sel.register(p.stdout,selectors.EVENT_READ);buf=b'';counter=0;thread=None;turn=None;started=time.monotonic()
def send(method,params):
 global counter
 counter+=1;p.stdin.write((json.dumps({'id':counter,'method':method,'params':params})+'\n').encode());p.stdin.flush();return counter
def reply(i,r):p.stdin.write((json.dumps({'id':i,'result':r})+'\n').encode());p.stdin.flush()
def read(timeout=20):
 global buf
 deadline=time.monotonic()+timeout
 while time.monotonic()<deadline:
  if b'\n' in buf:
   line,buf=buf.split(b'\n',1)
   try:return json.loads(line)
   except ValueError:continue
  if sel.select(.2):
   b=os.read(p.stdout.fileno(),65536)
   if not b:raise RuntimeError('server exited')
   buf+=b
 raise TimeoutError('bounded wait')
def rpc(method,params):
 i=send(method,params)
 while True:
  x=read()
  if x.get('id')==i and 'method' not in x:
   if 'error' in x:raise RuntimeError(method+': '+json.dumps(x['error']))
   return x.get('result',{})
try:
 rpc('initialize',{'clientInfo':{'name':'moai_t649_probe','version':'0.1.0'},'capabilities':{'experimentalApi':True}})
 p.stdin.write(b'{"method":"initialized"}\n');p.stdin.flush()
 account=rpc('account/read',{'refreshToken':False});result['account_type']=(account.get('account') or {}).get('type');result['requires_openai_auth']=account.get('requiresOpenaiAuth')
 models=rpc('model/list',{'includeHidden':False});result['models']=[{k:m.get(k) for k in ['id','model','displayName','isDefault']} for m in models.get('data',[]) if m.get('model') in ['gpt-5.6-sol','gpt-6-astra']]
 if result['account_type']!='chatgpt':raise RuntimeError('expected managed chatgpt auth; no inference')
 t=rpc('thread/start',{'model':'gpt-5.6-sol','cwd':str(base.resolve()),'ephemeral':True,'environments':[],'approvalPolicy':'never','sandbox':'read-only','baseInstructions':'Use only the moai_echo dynamic tool. No filesystem, shell, network or other tools. This is a synthetic integration test.','dynamicTools':[{'type':'function','name':'moai_echo','description':'Returns a fixed synthetic validation marker. Call exactly once.','inputSchema':{'type':'object','properties':{},'additionalProperties':False},'deferLoading':False}]})
 thread=t['thread']['id'];result['thread_started']=True;result['reported_model']=t.get('model')
 u=rpc('turn/start',{'threadId':thread,'environments':[],'effort':'low','input':[{'type':'text','text':'Call moai_echo exactly once. Then reply with exactly its returned marker and nothing else.'}]});turn=u['turn']['id']
 while time.monotonic()-started<90:
  x=read(30);method=x.get('method');params=x.get('params',{})
  if method:result['events'].append(method)
  if method=='item/tool/call':
   name=params.get('tool');result['tool_calls'].append({'tool':name,'arguments':params.get('arguments')})
   ok=name=='moai_echo' and len(result['tool_calls'])==1
   reply(x['id'],{'success':ok,'contentItems':[{'type':'inputText','text':'MOAI_APPSERVER_ECHO_OK' if ok else 'TOOL_REFUSED'}]})
  elif 'id' in x and method:
   p.stdin.write((json.dumps({'id':x['id'],'error':{'code':-32601,'message':'Synthetic probe refuses this request'}})+'\n').encode());p.stdin.flush()
  elif method=='item/completed':
   item=params.get('item',{})
   if item.get('type')=='agentMessage':result['final_texts'].append(item.get('text'))
  elif method=='turn/completed':result['turn_status']=params.get('turn',{}).get('status');break
  elif method=='error':result['errors'].append(params.get('error'))
 result['success']=result.get('turn_status')=='completed' and len(result['tool_calls'])==1 and result['tool_calls'][0]['tool']=='moai_echo' and 'MOAI_APPSERVER_ECHO_OK' in result['final_texts']
except Exception as e:result['errors'].append(str(e));result['success']=False
finally:
 if thread and turn and not result.get('turn_status'):
  try:send('turn/interrupt',{'threadId':thread,'turnId':turn});result['interrupt_sent']=True
  except Exception:pass
 try:os.killpg(p.pid,signal.SIGTERM)
 except ProcessLookupError:pass
 except PermissionError:result["cleanup_permission_error"]=True
 try:p.wait(timeout=3)
 except subprocess.TimeoutExpired:
  os.killpg(p.pid,signal.SIGKILL);p.wait(timeout=3)
 result['stderr_excerpt']=p.stderr.read(2048).decode('utf-8','replace');result['hook_events']=[m for m in result['events'] if m.startswith('hook/')];result['mcp_events']=[m for m in result['events'] if m.startswith('mcpServer/')];result['process_exit']=p.returncode;result['duration_seconds']=round(time.monotonic()-started,2)
 (base/'isolation-result.json').write_text(json.dumps(result,indent=2)+'\n')
 print(json.dumps(result));raise SystemExit(0 if result['success'] else 1)
