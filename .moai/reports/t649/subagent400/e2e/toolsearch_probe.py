#!/usr/bin/env python3
import argparse,importlib.util,pathlib,os,json,hashlib
p=argparse.ArgumentParser();p.add_argument('--binary',required=True);p.add_argument('--model',required=True);p.add_argument('--output',required=True);a=p.parse_args()
base=pathlib.Path(__file__).resolve().parents[2]/'e2e'
spec=importlib.util.spec_from_file_location('probe',base/'gateway_probe.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
os.environ['MOAI_CLAUDE_BIN']=str(base/'claude_fixture.py')
marker='T649_TOOLSEARCH_COMPLETE'
argv=[a.binary,'gpt','--permission-mode','default','--settings','{"disableAllHooks":true}','--model',a.model,'--strict-mcp-config','--mcp-config','{"mcpServers":{}}','--print','--output-format','stream-json','--verbose','--max-turns','4','--allowedTools','ToolSearch','--','Use ToolSearch to discover WebFetch. Do not call WebFetch or any network tools. After ToolSearch returns, reply with exactly '+marker+'.']
raw,meta=m.capture(argv,'/tmp/moai-t649-e2e-project',90)
rows=[]
for line in m.ANSI.sub('',raw).splitlines():
 try: rows.append(json.loads(line))
 except (ValueError,TypeError): pass
uses=[];refs=[];texts=[];results=[]
def visit(x):
 if isinstance(x,dict):
  if x.get('type')=='tool_use':uses.append(x.get('name'))
  if x.get('type')=='tool_reference':refs.append(x.get('tool_name'))
  if x.get('type')=='text':texts.append(x.get('text',''))
  for v in x.values():visit(v)
 elif isinstance(x,list):
  for v in x:visit(v)
for row in rows:
 visit(row)
 if row.get('type')=='result':results.append({k:row.get(k) for k in ['subtype','is_error','num_turns','result']})
r={'model':a.model,'binary_sha256':hashlib.sha256(pathlib.Path(a.binary).read_bytes()).hexdigest(),'meta':meta,'tool_uses':uses,'tool_references':refs,'marker_seen':any(marker in t for t in texts),'errors':[e for e in m.ERRORS if e in raw],'results':results,'event_types':[r.get('type') for r in rows]}
r['success']='ToolSearch' in uses and 'WebFetch' in refs and r['marker_seen'] and not r['errors'] and any(not x.get('is_error') for x in results)
pathlib.Path(a.output).write_text(json.dumps(r,indent=2)+'\n');print(json.dumps(r));raise SystemExit(0 if r['success'] else 1)
