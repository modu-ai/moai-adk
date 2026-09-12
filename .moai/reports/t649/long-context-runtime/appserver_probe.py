import os,json,subprocess,selectors,time,signal,pathlib
root=pathlib.Path(__file__).parent
results=[]
for model in ['gpt-5.6-sol','gpt-6-astra']:
 env=os.environ.copy();env.pop('OPENAI_API_KEY',None)
 cmd=['codex','app-server','--stdio','-c','model_context_window=872000','-c','model_auto_compact_token_limit=850000','-c','approval_policy="never"','-c','sandbox_mode="read-only"','-c','web_search="disabled"','-c','features.shell_tool=false']
 p=subprocess.Popen(cmd,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,start_new_session=True,env=env)
 sel=selectors.DefaultSelector();sel.register(p.stdout,selectors.EVENT_READ);buf=b'';counter=0;started=time.monotonic();thread=None;turn=None
 result={'model':model,'configured_window':872000,'repetitions':800000,'route':'official Codex App Server','text':[],'usage':None,'errors':[]}
 def send(method,params):
  global counter
  counter+=1;p.stdin.write((json.dumps({'id':counter,'method':method,'params':params})+'\n').encode());p.stdin.flush();return counter
 def read():
  global buf
  while time.monotonic()-started<230:
   if b'\n' in buf:
    line,buf=buf.split(b'\n',1);return json.loads(line)
   if sel.select(.2):
    chunk=os.read(p.stdout.fileno(),65536)
    if not chunk:raise RuntimeError('app-server exited')
    buf+=chunk
  raise TimeoutError('probe deadline')
 def rpc(method,params):
  i=send(method,params)
  while True:
   x=read()
   if x.get('id')==i:
    if 'error' in x:raise RuntimeError(str(x['error'])[:300])
    return x.get('result',{})
 try:
  rpc('initialize',{'clientInfo':{'name':'moai_context_probe','version':'0.1'},'capabilities':{'experimentalApi':True}});p.stdin.write(b'{"method":"initialized"}\n');p.stdin.flush()
  account=rpc('account/read',{'refreshToken':False});result['account_type']=(account.get('account') or {}).get('type')
  if result['account_type']!='chatgpt':raise RuntimeError('subscription required')
  t=rpc('thread/start',{'model':model,'cwd':str(root.resolve()),'ephemeral':True,'approvalPolicy':'never','sandbox':'read-only','environments':[],'baseInstructions':'Read synthetic input only. No tools. Return exactly the three marker values in order, comma separated.','dynamicTools':[]});thread=t['thread']['id']
  text='START_MARKER=MAPLE71\n'+' x'*400000+'\nMIDDLE_MARKER=OTTER83\n'+' x'*400000+'\nEND_MARKER=CEDAR29\nReturn the three marker values only.'
  result['input_text_bytes']=len(text.encode());second=text[len(text)//2:];text=text[:len(text)//2]+'\nRemember this segment. Reply ACK only.';result['segments']=2;u=rpc('turn/start',{'threadId':thread,'environments':[],'effort':'low','input':[{'type':'text','text':text}]});turn=u['turn']['id']
  while True:
   x=read();method=x.get('method');params=x.get('params',{})
   if method=='thread/tokenUsage/updated':result['usage']=params.get('tokenUsage')
   if method=='item/completed' and params.get('item',{}).get('type')=='agentMessage':result['text'].append(params['item'].get('text',''))
   if method=='error':result['errors'].append(str(params.get('error'))[:500])
   if method=='turn/completed':
    result['status']=params.get('turn',{}).get('status')
    if second and result['status']=='completed':
     result['first_turn_usage']=result.get('usage');result['text']=[]
     u=rpc('turn/start',{'threadId':thread,'environments':[],'effort':'low','input':[{'type':'text','text':second}]});turn=u['turn']['id'];second='';continue
    break
   if method and 'id' in x:
    p.stdin.write((json.dumps({'id':x['id'],'error':{'code':-32601,'message':'No tools in context probe'}})+'\n').encode());p.stdin.flush()
  result['success']=result.get('status')=='completed' and all(v in ''.join(result['text']) for v in ['MAPLE71','OTTER83','CEDAR29'])
 except Exception as e:result['errors'].append(str(e));result['success']=False
 finally:
  if thread and turn and not result.get('status'):
   try:send('turn/interrupt',{'threadId':thread,'turnId':turn})
   except Exception:pass
  try:os.killpg(p.pid,signal.SIGTERM)
  except ProcessLookupError:pass
  try:p.wait(timeout=3)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
  sel.close()
 result['elapsed_seconds']=round(time.monotonic()-started,2);results.append(result);(root/(model+'.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
 if not result['success']:break
