import json,pathlib,os,subprocess
root=pathlib.Path(__file__).resolve().parent
wt=root.parents[5]
# Resolve repository by the known overlay virtual path, without branch mutations.
wt=pathlib.Path('/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified')
for model,slug,lo,hi in [('gpt-5.6-sol','sol',921000,921928),('gpt-6-astra','astra',921000,922000)]:
 for iteration in range(4):
  n=(lo+hi)//2;p=root/f'{slug}-{n}.json';env=os.environ.copy();env.update(T649_MODEL=model,T649_WORDS=str(n),T649_RESULT=str(p))
  with (root/'.runs'/f'{slug}-{n}.log').open('w') as log:
   proc=subprocess.run(['go','test','-overlay',str(root/'overlay.json'),'./internal/cli','-run','^TestT649ContextLive$','-count=1','-v'],cwd=wt,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=190)
  if proc.returncode or not p.exists():raise SystemExit('HARNESS_FAILURE')
  d=json.loads(p.read_text());u=d.get('usage') or {};print(json.dumps({'model':model,'n':n,'event':d.get('event'),'usage_input':u.get('input_tokens'),'error':d.get('server_error_code'),'needle':d.get('needle_recall')}),flush=True)
  if d.get('event')=='response.completed' and d.get('needle_recall'):lo=n
  elif d.get('server_error_code')=='context_length_exceeded':hi=n
  else:raise SystemExit('NON_CONTEXT_FAILURE_STOP')
 # Independent repeat retained under distinct filename.
 p=root/f'{slug}-{lo}-repeat.json';env=os.environ.copy();env.update(T649_MODEL=model,T649_WORDS=str(lo),T649_RESULT=str(p))
 with (root/'.runs'/f'{slug}-{lo}-repeat.log').open('w') as log:
  proc=subprocess.run(['go','test','-overlay',str(root/'overlay.json'),'./internal/cli','-run','^TestT649ContextLive$','-count=1','-v'],cwd=wt,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=190)
 if proc.returncode or not p.exists():raise SystemExit('REPEAT_HARNESS_FAILURE')
 d=json.loads(p.read_text());print(json.dumps({'model':model,'repeat_n':lo,'event':d.get('event'),'usage_input':(d.get('usage') or {}).get('input_tokens'),'needle':d.get('needle_recall'),'bracket_repetitions':[lo,hi]}),flush=True)
 if d.get('event')!='response.completed' or not d.get('needle_recall'):raise SystemExit('REPEAT_FAILURE_STOP')
