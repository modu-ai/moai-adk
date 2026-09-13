import json, hashlib, base64

MAN = '/Users/goos/.moai/state/gateway-conversations/families/1f14d174-9f5c-4626-a0d0-82a2bc9afa62/receipt/manifest.json'
TX = '/Users/goos/.moai/state/gateway-conversations/families/1f14d174-9f5c-4626-a0d0-82a2bc9afa62/native/projects/-Users-goos-MoAI-mo-ai-kr/1f14d174-9f5c-4626-a0d0-82a2bc9afa62.jsonl'
FAMILY_ID = '1f14d174-9f5c-4626-a0d0-82a2bc9afa62'
SCOPE = 'e9581c233ffef46aa6dcee042ab7b3c354aaa57353ac8fd63c12a0a60b8b2059'

# --- canonical mirroring internal/gateway/receipt/projection.go ---
def canon(b, v):
    if v is None:
        b.write(b'n')
    elif v is True:
        b.write(b't')
    elif v is False:
        b.write(b'f')
    elif isinstance(v, str):
        b.write(b's'); b.write(str(len(v)).encode()); b.write(b':'); b.write(v.encode())
    elif isinstance(v, Num):
        canon_num(b, v)
    elif isinstance(v, list):
        b.write(b'[')
        for x in v: canon(b, x)
        b.write(b']')
    elif isinstance(v, dict):
        b.write(b'{')
        for k in sorted(v.keys()):
            canon(b, k); canon(b, v[k])
        b.write(b'}')
    else:
        raise TypeError('unsupported %r' % (v,))

# json.Number semantics: preserve the literal string like Go big.Rat normalization
class Num:
    def __init__(self, s): self.s = s

def canon_num(b, v):
    from fractions import Fraction
    r = Fraction(v.s)
    b.write(b'd'); b.write(str(r.numerator).encode() if r.denominator == 1 else ('%d/%d' % (r.numerator, r.denominator)).encode()); b.write(b';')

def strict_load(text):
    def pairs(items):
        d = {}
        for k, v in items:
            if k in d: raise ValueError('dup')
            d[k] = v
        return d
    def hook(s): return Num(s)
    return json.loads(text, object_pairs_hook=pairs, parse_float=hook, parse_int=hook, parse_constant=hook)

def decode_id(id_, ids):
    if id_ in ids: return ids[id_]
    if id_.startswith('toolu_moai_v1_'):
        raise ValueError('unrestorable marker')
    return id_

def normalize_content(v, ids, tool_result):
    if isinstance(v, str):
        return [{'text': v, 'type': 'text'}]
    out = []
    for block in v:
        t = block['type']
        if t in ('thinking', 'redacted_thinking'):
            continue
        nb = dict(block)
        nb.pop('cache_control', None)
        key = 'id' if t == 'tool_use' else ('tool_use_id' if t == 'tool_result' else '')
        if key:
            nb[key] = decode_id(nb[key], ids)
        if t == 'tool_result' and 'content' in nb:
            nb['content'] = normalize_content(nb['content'], ids, True)
        out.append(nb)
    return out

def canonical_prefixes(messages, ids):
    import copy, io
    h = hashlib.sha256()
    h.update(b'[')
    out = []
    prev = bytes(32)
    for msg in messages:
        assert set(msg.keys()) == {'role', 'content'}, msg.keys()
        content = normalize_content(copy.deepcopy(msg['content']), ids, False)
        b = io.BytesIO()
        canon(b, {'role': msg['role'], 'content': content})
        h.update(b.getvalue())
        if msg['role'] == 'assistant':
            state = h.copy()
            state.update(b']')
            out.append((state.digest(), prev))
            prev = state.digest()
    return out

# --- build ids map: marker -> call_id (per assistant msg envelope) ---
def build_ids(messages):
    ids = {}
    for msg in messages:
        if msg['role'] != 'assistant': continue
        env = None
        for blk in (msg['content'] if isinstance(msg['content'], list) else []):
            if blk.get('type') == 'redacted_thinking':
                s = blk['data']
                for pfx in ('moai_opaque_v1_', 'moai_opaque_v2_'):
                    if s.startswith(pfx):
                        raw = base64.urlsafe_b64decode(s[len(pfx):] + '=' * (-len(s[len(pfx):]) % 4))
                        env = raw
        for blk in (msg['content'] if isinstance(msg['content'], list) else []):
            if blk.get('type') == 'tool_use':
                i = blk['id']
                if i.startswith('toolu_moai_v1_'):
                    b = json.loads(base64.urlsafe_b64decode(i[len('toolu_moai_v1_'):] + '=' * (-len(i) % 4)))
                    ids[i] = b['call_id']
                else:
                    ids[i] = i
    return ids

# --- transcript -> messages ---
rows = [json.loads(l) for l in open(TX) if l.strip()]
messages = []
pending = {}
for r in rows:
    t = r.get('type')
    if t not in ('user', 'assistant'): continue
    m = r.get('message', {})
    if m.get('model') == '<synthetic>': continue
    content = m.get('content')
    if t == 'assistant':
        mid = m.get('id', '')
        if mid in pending:
            pending[mid]['content'].extend(content)
        else:
            mm = {'role': 'assistant', 'content': list(content)}
            pending[mid] = mm
            messages.append(mm)
    else:
        if isinstance(content, list):
            content = [dict(b) for b in content]
        messages.append({'role': 'user', 'content': content})

print('messages:', len(messages))
messages = strict_load(json.dumps(messages))
assistants = sum(1 for m in messages if m['role'] == 'assistant')
print('assistant messages:', assistants)

man = json.load(open(MAN))
cands = man['candidates']

ids = build_ids(messages)
prefixes = canonical_prefixes(messages, ids)

def bind(u):
    domain = b'moai-history-v1\x00' + FAMILY_ID.encode() + b'\x00' + SCOPE.encode() + b'\x00' + b'gpt-5.6\x00'
    return hashlib.sha256(domain + u).digest()

def h2s(b): return b.hex()

print('\nidx | computed(bound) vs manifest prefix | prev match')
for i, (u, prev) in enumerate(prefixes):
    bu = bind(u)
    if i < len(cands):
        cp = bytes.fromhex(cands[i]['prefix']) if isinstance(cands[i]['prefix'], str) else cands[i]['prefix']
        cpv = bytes.fromhex(cands[i]['previous']) if isinstance(cands[i]['previous'], str) else cands[i]['previous']
        print(i, 'PREFIX-MATCH' if cp == bu else 'PREFIX-DIVERGE', 'prev-' + ('match' if cpv == bind(prev) else 'DIVERGE'))
    else:
        print(i, 'no candidate (extra boundary)')
if len(prefixes) != len(cands):
    print('COUNT MISMATCH: boundaries', len(prefixes), 'candidates', len(cands))
