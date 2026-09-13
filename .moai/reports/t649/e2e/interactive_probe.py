#!/usr/bin/env python3
"""One synthetic paid interactive probe; only safe assertions persist."""
import importlib.util,pathlib,json,os,hashlib,re
base=pathlib.Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('gateway_probe',base/'gateway_probe.py')
p=importlib.util.module_from_spec(spec);spec.loader.exec_module(p)
binary=pathlib.Path('/tmp/moai-v3.2.0-rc.8-t649')
os.environ['MOAI_CLAUDE_BIN']=str(base/'claude_fixture.py')
before=set(base.glob('fixture-*.json'))
marker='T649_INTERACTIVE_OK'
argv=[str(binary),'gpt','--permission-mode','default','--settings','{"disableAllHooks":true}','--model','gpt-5.6-sol','--strict-mcp-config','--mcp-config','{"mcpServers":{}}']
t,meta=p.capture(argv,'/tmp/moai-t649-e2e-project',55,picker=True,interactive_marker=marker)
r=p.summarize(t,meta,marker,True)
r['visible_assistant_marker']=('⏺'+marker) in re.sub(r'\s','',t)
new=set(base.glob('fixture-*.json'))-before
r['new_fixture_count']=len(new)
r['native_assistant_match']=False
r['native_models']=[]
if len(new)==1:
 family=next(iter(new)).stem.removeprefix('fixture-')
 native=pathlib.Path.home()/'.moai/state/gateway-conversations/families'/family/'native/projects/-private-tmp-moai-t649-e2e-project'
 r['family_id']=family
 for f in native.glob('*.jsonl'):
  for line in f.read_text().splitlines():
   try:v=json.loads(line)
   except ValueError:continue
   msg=v.get('message',{})
   if v.get('type')=='assistant' and msg.get('role')=='assistant':
    model=msg.get('model');r['native_models'].append(model)
    if model=='gpt-6-astra' and any(b.get('type')=='text' and b.get('text','').strip()==marker for b in msg.get('content',[])):
     r['native_assistant_match']=True
r['native_models']=sorted(set(r['native_models']))
r['binary_sha256']=hashlib.sha256(binary.read_bytes()).hexdigest()
r['journey']='interactive_picker_astra_answer'
r['success']=r['session_only_selection_confirmed'] and 'gpt-6-astra' in r['selected_model_ids'] and r['visible_assistant_marker'] and r['native_assistant_match'] and not r['error_categories']
(base/'interactive-rc8.json').write_text(json.dumps(r,indent=2)+'\n')
print(json.dumps(r,indent=2))
raise SystemExit(0 if r['success'] else 1)
