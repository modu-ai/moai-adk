import os,pty,fcntl,termios,struct,subprocess,select,signal,time,json,re,pathlib
root=pathlib.Path(__file__).parent;build=json.loads((root/'build.json').read_text());env=os.environ.copy()
for k in list(env):
 if k.startswith(('ANTHROPIC_','CLAUDE_CODE_','MOAI_SESSION','MOAI_KANBAN','MOAI_FACTORY')) or k in ['CLAUDECODE','CLAUDE_SESSION_ID','TMUX','TMUX_PANE','MOAI_CLAUDE_BIN']:env.pop(k,None)
env['TERM']='xterm-256color';master,slave=pty.openpty();fcntl.ioctl(slave,termios.TIOCSWINSZ,struct.pack('HHHH',45,160,0,0));args=[build['binary'],'gpt','-p','mo.ai.kr','--model','gpt-5.6-sol','--permission-mode','default','--settings','{"disableAllHooks":true}','--strict-mcp-config','--mcp-config','{"mcpServers":{}}','--tools',''];p=subprocess.Popen(args,cwd='/Users/goos/MoAI/mo.ai.kr',env=env,stdin=slave,stdout=slave,stderr=slave,start_new_session=True);os.close(slave);start=time.monotonic();raw=bytearray();sent=False;result={'command':args,'api_key_prompt':False,'context_warning':False,'login_prompt':False,'reply_seen':False,'scope':'Actual Claude interactive UI and one synthetic GPT subscription answer; no tools/hooks'}
ansi=re.compile(r'\x1b(?:\[[0-?]*[ -/]*[@-~]|\][^\x07]*(?:\x07|\x1b\\))')
try:
 while time.monotonic()-start<55:
  if select.select([master],[],[],.2)[0]:
   try:b=os.read(master,65536)
   except OSError:break
   if not b:break
   raw.extend(b)
  if len(raw)>4<<20:raise RuntimeError('output limit')
  screen=ansi.sub('',raw.decode('utf-8','replace'));compact=re.sub(r'\s','',screen)
  result['api_key_prompt']='DoyouwanttousethisAPIkey?' in compact
  result['context_warning']='CLAUDE_CODE_DISABLE_1M_CONTEXTisset' in compact
  result['login_prompt']='Selectloginmethod:' in compact or 'Notloggedin' in compact
  if result['api_key_prompt'] or result['login_prompt']:break
  if not sent and time.monotonic()-start>7:
   os.write(master,b'Reply with exactly MOAI_BEARER_READY and nothing else.\r');sent=True
  if '⏺MOAI_BEARER_READY' in compact:result['reply_seen']=True;break
 result['input_sent']=sent;result['success']=result['reply_seen'] and not any(result[k] for k in ['api_key_prompt','context_warning','login_prompt']);result['elapsed_seconds']=round(time.monotonic()-start,2)
 result['other_prompt_categories']=[s for s in ['Yes,Itrustthisfolder','Choosethetextstyle','BypassPermissions','trust'] if s in compact]
finally:
 try:os.killpg(p.pid,signal.SIGTERM)
 except ProcessLookupError:pass
 try:p.wait(timeout=3)
 except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
 os.close(master)
(root/'interactive-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
