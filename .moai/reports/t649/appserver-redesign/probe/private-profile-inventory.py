#!/usr/bin/env python3
import subprocess,os,json,time,selectors,signal,pathlib,tempfile,re,shutil
base=pathlib.Path(__file__).parent
result={'codex_version':'0.154.0','events':[],'errors':[],'tool_calls':[],'final_texts':[]}
source=pathlib.Path('/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t650/internal/codexapp/client.go').read_text()
chunk=source[source.index('func Start('):source.index('func startProcess(')]
flags=re.findall(r'`([^`]+)`',chunk);result['flags']=flags
profile=tempfile.TemporaryDirectory(prefix='moai-private-inventory-');home=pathlib.Path(profile.name);home.chmod(0o700)
cmd=[shutil.which('codex'),'app-server','--stdio']
for flag in flags:cmd+=['-c',flag]
env={k:os.environ[k] for k in ['PATH','LANG','LC_ALL','SYSTEMROOT','SystemRoot','WINDIR','TEMP','TMP','TMPDIR'] if k in os.environ}
env.update(CODEX_HOME=str(home),HOME=str(home),USERPROFILE=str(home))
p=subprocess.Popen(cmd,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.PIPE,start_new_session=True,env=env,cwd=home)
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
  x=read(max(.1,min(10,30-(time.monotonic()-started))))
  if x.get('method')=='mcpServer/startupStatus/updated':result.setdefault('startup',[]).append({k:v for k,v in x.get('params',{}).items() if k in ['name','status','failureReason']})
  if x.get('id')==i and 'method' not in x:
   if 'error' in x:raise RuntimeError(method+': '+json.dumps(x['error']))
   return x.get('result',{})
try:
 rpc('initialize',{'clientInfo':{'name':'moai_private_inventory','version':'0.1.0'},'capabilities':{'experimentalApi':True}})
 p.stdin.write(b'{"method":"initialized"}\n');p.stdin.flush()
 account=rpc('account/read',{'refreshToken':False});result['account_type']=(account.get('account') or {}).get('type')
 t=rpc('thread/start',{'cwd':str(home),'ephemeral':True,'environments':[],'approvalPolicy':'never','sandbox':'read-only','baseInstructions':'Synthetic inventory only. No turn will start.'})
 thread=t['thread']['id'];result['thread_started']=True
 rows=[];cursor=None
 for page in range(5):
  q={'threadId':thread,'limit':100,'detail':'toolsAndAuthOnly'}
  if cursor:q['cursor']=cursor
  r=rpc('mcpServerStatus/list',q)
  for x in r.get('data',[]):rows.append({'name':x.get('name'),'tool_count':len(x.get('tools',{})),'authStatus':x.get('authStatus')})
  cursor=r.get('nextCursor')
  if not cursor:break
 result['servers']=rows;result['cursor_remaining']=bool(cursor);result['inference_started']=False
 result['success']=not rows and not cursor and result['account_type'] is None
except Exception as e:result['errors'].append(str(e));result['success']=False
finally:
 try:os.killpg(p.pid,signal.SIGTERM)
 except ProcessLookupError:pass
 except PermissionError:result['cleanup_permission_error']=True
 try:p.wait(timeout=3)
 except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait(timeout=3)
 result['stderr_excerpt']=p.stderr.read(2048).decode('utf-8','replace');result['process_exit']=p.returncode;result['duration_seconds']=round(time.monotonic()-started,2)
 profile.cleanup();result['profile_removed']=not home.exists()
 (base/'private-profile-inventory-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result));raise SystemExit(0 if result['success'] else 1)
