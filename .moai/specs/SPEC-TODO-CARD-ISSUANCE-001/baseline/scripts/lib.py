"""Shared loader: read-only union of live + archived cards and findings."""
import sqlite3, json, re, os
DB = 'file:/Users/goos/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db?mode=ro'
LEGACY = '/Users/goos/.moai/db/moai-adk-go-1bd3d038/todo/backlog.json'
OUT = os.path.dirname(os.path.abspath(__file__))

def num(cid):
    return int(cid[1:])

def load():
    con = sqlite3.connect(DB, uri=True)
    con.row_factory = sqlite3.Row
    cards = {}
    for r in con.execute('select * from items'):
        d = dict(r); d['where'] = 'live'; d['archived_at'] = None; d['landing_verdict'] = None
        cards[d['id']] = d
    for r in con.execute('select * from archived_items'):
        d = dict(r); d['where'] = 'archived'; d['arch_seq'] = d['seq']
        cards[d['id']] = d
    findings = []
    for r in con.execute('select rowid,* from findings'):
        d = dict(r); d['where'] = 'live'; findings.append(d)
    for r in con.execute('select * from archived_findings'):
        d = dict(r); d['where'] = 'archived'; findings.append(d)
    meta = dict(con.execute('select key,value from meta').fetchall())
    assigns = [dict(r) for r in con.execute('select * from todo_runtime_assignments')]
    con.close()
    return cards, findings, meta, assigns

def fate(c):
    """done = archived; else live state."""
    if c['where'] == 'archived':
        return 'done(archived)'
    return c['state']

REF = re.compile(r'(?<![A-Za-z0-9])t(\d{1,4})(?![0-9A-Za-z])')
def refs(text, self_id):
    out = set()
    for m in REF.finditer(text):
        n = int(m.group(1))
        cid = 't%d' % n
        if cid != self_id and n >= 1:
            out.add(cid)
    return out
