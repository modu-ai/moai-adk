import os, json, pty, select, time, signal, subprocess, pathlib, re, fcntl, termios, struct, sys
mode=sys.argv[1] if len(sys.argv)>1 else 'gpt'
rows={'gpt':[('gpt-6-astra','GPT-6 Astra PROBE'),('gpt-5.6-sol','GPT Sol PROBE')],'glm':[('glm-5.3','GLM 5.3 PROBE'),('glm-5.3-flash','GLM Flash PROBE')],'cc':[('claude-opus-5','Claude Opus PROBE'),('claude-sonnet-5','Claude Sonnet PROBE')]}[mode]
root = pathlib.Path('/tmp/moai-model-research-20260912/picker-probe-'+mode).resolve()
root.mkdir(exist_ok=True)
config = root / 'config'
config.mkdir(exist_ok=True)
(config / '.claude.json').write_text(json.dumps({'hasCompletedOnboarding':True,'theme':'dark','projects':{str(root):{'hasTrustDialogAccepted':True}}}))
settings={'model':rows[-1][0],'modelPicker':{'replaceBuiltInOptions':True,'options':[{'model':m,'label':l} for m,l in rows]},'env':{k:rows[0][0] for k in ['ANTHROPIC_DEFAULT_OPUS_MODEL','ANTHROPIC_DEFAULT_SONNET_MODEL','ANTHROPIC_DEFAULT_HAIKU_MODEL','ANTHROPIC_DEFAULT_FABLE_MODEL']},'disableAllHooks':True}
env={k:v for k,v in os.environ.items() if not any(x in k for x in ['ANTHROPIC','CLAUDE','MOAI','OPENAI','Z_AI'])}
env.update({'CLAUDE_CONFIG_DIR':str(config),'ANTHROPIC_BASE_URL':'http://127.0.0.1:9','ANTHROPIC_AUTH_TOKEN':'local-probe-not-a-secret','CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC':'1','TERM':'xterm-256color'})
master,slave=pty.openpty()
fcntl.ioctl(slave,termios.TIOCSWINSZ,struct.pack('HHHH',44,150,0,0))
p=subprocess.Popen(['/Users/goos/.local/bin/claude','--settings',json.dumps(settings),'--strict-mcp-config','--mcp-config','{"mcpServers":{}}','--model',rows[-1][0]],cwd=root,env=env,stdin=slave,stdout=slave,stderr=slave,start_new_session=True)
os.close(slave)
out=b''; start=time.monotonic(); sent=False; selected=False
try:
 while time.monotonic()-start<16:
  ready,_,_=select.select([master],[],[],0.2)
  if ready:
   try:chunk=os.read(master,65536)
   except OSError:break
   out+=chunk
  elapsed=time.monotonic()-start
  if elapsed>4 and not sent:
   os.write(master,b'/model\r');sent=True
  if elapsed>8 and not selected:
   os.write(master,b'\x1b[As');selected=True
  if p.poll() is not None:break
finally:
 if p.poll() is None:
  os.killpg(p.pid,signal.SIGTERM)
  try:p.wait(timeout=3)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
 os.close(master)
text=re.sub(r'\x1b\[[0-?]*[ -/]*[@-~]','',out.decode('utf8','replace'))
(root/'screen.txt').write_text(text)
print(text[-11000:])
print('PROBE_PROCESS_CLEANED',p.poll() is not None)
