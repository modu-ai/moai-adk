from pathlib import Path
import subprocess,json
O=Path(__file__).parent;session='hooks-audit-01a08e35'
def ab(*args):
 p=subprocess.run(['agent-browser','--session',session,*args],text=True,capture_output=True,timeout=45)
 if p.returncode:raise RuntimeError(p.stderr+p.stdout)
 return p.stdout.strip()
js='''JSON.stringify({width:innerWidth,documentWidth:document.documentElement.scrollWidth,findings:document.querySelectorAll("article.finding").length,duplicateIds:[...document.querySelectorAll("[id]")].map(e=>e.id).filter((id,i,a)=>a.indexOf(id)!==i),brokenAnchors:[...document.querySelectorAll('a[href^="#"]')].map(a=>a.getAttribute("href")).filter(h=>!document.querySelector(h)),title:document.title})'''
records=[]
ab('reload')
for width in [1440,768,390]:
 ab('set','viewport',str(width),'1000');ab('eval','window.scrollTo(0,0)');records.append(json.loads(ab('eval',js).strip('"').replace('\\"','"')) if False else ab('eval',js));ab('screenshot',str(O/f'screen-{width}.png'))
ab('set','offline','on');ab('reload');records.append('OFFLINE '+ab('eval',js));ab('screenshot',str(O/'screen-offline.png'));ab('set','offline','off');ab('set','viewport','1440','1000');ab('pdf',str(O/'hooks-audit-print.pdf'));errors=ab('errors');(O/'browser-checks.json').write_text(json.dumps(dict(measurements=records,page_errors=errors),ensure_ascii=False,indent=2));print(json.dumps(dict(measurements=records,page_errors=errors),ensure_ascii=False,indent=2))
