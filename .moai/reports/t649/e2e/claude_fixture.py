#!/usr/bin/env python3
import os, pathlib, json, sys
cwd=pathlib.Path.cwd().resolve()
config=pathlib.Path(os.environ.get('CLAUDE_CONFIG_DIR','/nonexistent')).resolve()
bases=[(pathlib.Path.home()/('.moai/state/'+name+'/families')).resolve() for name in ['gateway-conversations','gateway-conversations-claude','gateway-conversations-glm']]
if cwd != pathlib.Path('/tmp/moai-t649-e2e-project').resolve() or not any(config.is_relative_to(base) for base in bases) or config.name != 'native':
    raise SystemExit('E2E fixture refuses non-test namespace: '+str(config))
f=config/'.claude.json'
if not f.exists():
    with f.open('x') as out:
        json.dump({'hasCompletedOnboarding':True,'theme':'dark','projects':{str(cwd):{'hasTrustDialogAccepted':True}}},out)
    f.chmod(0o600)
args=sys.argv[1:]
shape={'dangerously_skip_permissions': '--dangerously-skip-permissions' in args,
       'permission_mode': args[args.index('--permission-mode')+1] if '--permission-mode' in args else None}
(pathlib.Path(__file__).parent/('fixture-'+config.parent.name+'.json')).write_text(json.dumps(shape)+'\n')
os.execv('/Users/goos/.local/bin/claude' ,['/Users/goos/.local/bin/claude',*sys.argv[1:]])
