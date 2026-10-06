import os,json,pathlib,subprocess,tempfile
OUT=pathlib.Path(__file__).parent
S=pathlib.Path((OUT/'snapshot.txt').read_text())
SCRIPT=S/'.claude/hooks/moai/sync-phase-quality-gate.sh'
results=[]
def run(args,cwd,env=None):
 return subprocess.run(args,cwd=cwd,env=env,capture_output=True,text=True,timeout=20)
def fixture(marker,source,tool,code):
 root=pathlib.Path(tempfile.mkdtemp(prefix='hook-shell-probe-')); bindir=root/'bin';bindir.mkdir()
 (root/marker).write_text('');(root/source).write_text('intentionally invalid source')
 (bindir/tool).write_text('#!/bin/sh\n'+code+'\n');(bindir/tool).chmod(0o755)
 for args in [['git','init','-q'],['git','add',marker,source],['git','-c','user.name=Audit','-c','user.email=audit@example.invalid','-c','core.hooksPath=/dev/null','commit','-qm','docs: sync audit fixture']]:
  v=run(args,root)
  if v.returncode: raise RuntimeError(v.stderr)
 env=os.environ.copy();env.update(PATH=str(bindir)+':/usr/bin:/bin:/usr/sbin:/sbin',CLAUDE_PROJECT_DIR=str(root),MOAI_SYNC_GATE_BLOCKING='1',MOAI_AUTONOMY_TIER='semi-auto')
 return root,env
root,env=fixture('go.mod','main.go','go','echo failed >&2; exit 7')
a=run(['bash',str(SCRIPT)],root,env);b=run(['bash',str(SCRIPT)],root,env)
results.append(dict(id='sync_sentinel',first_exit=a.returncode,first_stdout=a.stdout.strip(),second_exit=b.returncode,second_stdout=b.stdout.strip()))
for name,marker,src,tool in [('java','pom.xml','Main.java','javac'),('ruby','Gemfile','bad.rb','ruby'),('cpp','CMakeLists.txt','bad.cpp','g++')]:
 root,env=fixture(marker,src,tool,'echo invoked >> "$CLAUDE_PROJECT_DIR/invoked"; echo syntax-error >&2; exit 7')
 v=run(['bash',str(SCRIPT)],root,env)
 log=root/'.moai/logs/sync-quality-gate.log'
 results.append(dict(id='sync_'+name,exit=v.returncode,stdout=v.stdout.strip(),compiler_invoked=(root/'invoked').exists(),log=log.read_text().strip() if log.exists() else 'absent'))
root,env=fixture('build.gradle.kts','Main.kt','kotlinc','echo invoked >> "$CLAUDE_PROJECT_DIR/invoked"; exit 7')
v=run(['bash',str(SCRIPT)],root,env)
results.append(dict(id='sync_kotlin',exit=v.returncode,stdout=v.stdout.strip(),compiler_invoked=(root/'invoked').exists(),log_exists=(root/'.moai/logs/sync-quality-gate.log').exists()))
(OUT/'probes-shell.json').write_text(json.dumps(results,ensure_ascii=False,indent=2));print(json.dumps(results,ensure_ascii=False,indent=2))
