import os,subprocess,json,pathlib,signal,sys
r=pathlib.Path(__file__).parent
binary=sys.argv[1];name=sys.argv[2]
env={k:v for k,v in os.environ.items() if not k.startswith(('ANTHROPIC_','CLAUDE_CODE_','MOAI_SESSION','MOAI_KANBAN','MOAI_FACTORY')) and k not in ['CLAUDECODE','CLAUDE_SESSION_ID','TMUX','TMUX_PANE','MOAI_CLAUDE_BIN']}
cmd=[binary,'gpt','-p','mo.ai.kr','--model','gpt-5.6-sol','--effort','low','--print','--verbose','--output-format','stream-json','--max-turns','3','--settings','{"disableAllHooks":true}','--strict-mcp-config','--mcp-config','{"mcpServers":{}}','--tools','Bash','--allowedTools','Bash(exit 5)','--','Call Bash once with command exactly: exit 5 . This deliberate failure is a test. After seeing the failed tool result, do not retry or use other tools; reply exactly MOAI_TOOL_FAILURE_HANDLED.']
p=subprocess.Popen(cmd,cwd='/Users/goos/MoAI/mo.ai.kr',env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,start_new_session=True);result={'command':cmd,'tool_calls':[],'tool_results':[],'results':[]}
try:
 out,err=p.communicate(timeout=90); result['exit_code']=p.returncode
 for line in out.decode(errors='replace').splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  for c in e.get('message',{}).get('content',[]):
   if not isinstance(c,dict):continue
   if c.get('type')=='tool_use':result['tool_calls'].append({k:c.get(k) for k in ['name','input']})
   if c.get('type')=='tool_result':result['tool_results'].append({k:c.get(k) for k in ['is_error','content']})
  if e.get('type')=='result':result['results'].append({k:e.get(k) for k in ['is_error','subtype','result','num_turns','permission_denials','api_error_status']})
 result['stderr']=err.decode(errors='replace')[-500:]
except subprocess.TimeoutExpired:result['timeout']=True
finally:
 try:os.killpg(p.pid,signal.SIGTERM)
 except ProcessLookupError:pass
 try:p.wait(timeout=3)
 except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
result['failed_tool_observed']=any(x.get('is_error') is True for x in result['tool_results'])
result['recovered']=any(x.get('result','').strip()=='MOAI_TOOL_FAILURE_HANDLED' and x.get('is_error') is False for x in result['results'])
(r/(name+'.json')).write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print(json.dumps(result,ensure_ascii=False))
