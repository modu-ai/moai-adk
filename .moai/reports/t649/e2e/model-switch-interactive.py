#!/usr/bin/env python3
import importlib.util,pathlib,json,os,hashlib,re,inspect
base=pathlib.Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('gateway_probe',base/'gateway_probe.py')
p=importlib.util.module_from_spec(spec);spec.loader.exec_module(p)
source=inspect.getsource(p.capture).replace('sent = False','sent = False\n    seed_sent = False',1)
source=source.replace('if picker and not sent and time.monotonic() - started > 7:', 'if picker and not seed_sent and time.monotonic() - started > 7:\n                os.write(master, b"Reply with exactly T649_SWITCH_SOL_OK.\\r")\n                seed_sent = True\n            if picker and seed_sent and not sent and "⏺T649_SWITCH_SOL_OK" in screen_compact:')
source=source.replace('if picker and sent and not selected and time.monotonic() - started > 12:', 'if picker and sent and not selected and "Selectmodel" in screen_compact:')
source=source.replace('prompt_sent = False', 'prompt_sent = False\n    switch_confirmed_at = None',1)
source=source.replace('if interactive_marker and selected and not prompt_sent and "forthissessiononly" in screen_compact:', 'if selected and switch_confirmed_at is None and "meansthefullhistorygetsre-readonyournextmessage" in screen_compact:\n                os.write(master,b"\\r")\n                switch_confirmed_at = time.monotonic()\n            if interactive_marker and selected and not prompt_sent and ("forthissessiononly" in screen_compact or (switch_confirmed_at is not None and time.monotonic()-switch_confirmed_at > 2)):')
exec(source,p.__dict__)
binary=pathlib.Path('/tmp/moai-v3.2.0-rc.8-t649-model-switch')
os.environ['MOAI_CLAUDE_BIN']=str(base/'claude_fixture.py')
before=set(base.glob('fixture-*.json'))
marker='T649_SWITCH_ASTRA_OK'
argv=[str(binary),'gpt','--permission-mode','default','--settings','{"disableAllHooks":true}','--model','gpt-5.6-sol','--strict-mcp-config','--mcp-config','{"mcpServers":{}}']
t,meta=p.capture(argv,'/tmp/moai-t649-e2e-project',100,picker=True,interactive_marker=marker)
r=p.summarize(t,meta,marker,True)
r['visible_sol_answer']='⏺T649_SWITCH_SOL_OK' in re.sub(r'\s','',t)
r['visible_astra_answer']=('⏺'+marker) in re.sub(r'\s','',t)
new=set(base.glob('fixture-*.json'))-before
r['new_fixture_count']=len(new);r['native_matches']=[]
if len(new)==1:
 family=next(iter(new)).stem.removeprefix('fixture-');r['family_id']=family
 native=pathlib.Path.home()/'.moai/state/gateway-conversations/families'/family/'native/projects'
 for f in native.rglob('*.jsonl'):
  for line in f.read_text().splitlines():
   try:v=json.loads(line)
   except ValueError:continue
   msg=v.get('message',{})
   if v.get('type')=='assistant' and msg.get('role')=='assistant':
    for b in msg.get('content',[]):
     if b.get('type')=='text' and b.get('text','').strip().removesuffix('.') in ['T649_SWITCH_SOL_OK',marker]:r['native_matches'].append({'model':msg.get('model'),'marker':b['text'].strip().removesuffix('.'),'trailing_period':b['text'].strip().endswith('.')})
r['binary_sha256']=hashlib.sha256(binary.read_bytes()).hexdigest();r['journey']='sol_answer_then_interactive_model_astra_answer'
r['success']=r['visible_sol_answer'] and r['visible_astra_answer'] and [{k:v for k,v in m.items() if k!='trailing_period'} for m in r['native_matches']]==[{'model':'gpt-5.6-sol','marker':'T649_SWITCH_SOL_OK'},{'model':'gpt-6-astra','marker':marker}] and not r['error_categories']
(base/'model-switch-interactive.json').write_text(json.dumps(r,indent=2)+'\n')
print(json.dumps(r,indent=2));raise SystemExit(0 if r['success'] else 1)
