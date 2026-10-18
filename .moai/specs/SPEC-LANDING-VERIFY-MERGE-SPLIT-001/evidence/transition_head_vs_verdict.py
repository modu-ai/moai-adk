import os,re,glob,sys
root="/Users/goos/MoAI/moai-adk-go"
ids=[l.strip() for l in open(sys.argv[1]) if l.strip()]
def head_of(tree):
    g=os.path.join(tree,".git")
    if not os.path.exists(g): return None,"no .git"
    if os.path.isfile(g):
        gd=open(g).read().strip().split("gitdir:")[1].strip()
    else: gd=g
    h=open(os.path.join(gd,"HEAD")).read().strip()
    if h.startswith("ref:"):
        ref=h.split()[1]
        cd=open(os.path.join(gd,"commondir")).read().strip() if os.path.exists(os.path.join(gd,"commondir")) else None
        base=os.path.normpath(os.path.join(gd,cd)) if cd else gd
        for b in (gd,base):
            p=os.path.join(b,ref)
            if os.path.exists(p): return open(p).read().strip(),ref
        pr=os.path.join(base,"packed-refs")
        for l in open(pr):
            if l.strip().endswith(" "+ref): return l.split()[0],ref
        return None,"ref unresolved"
    return h,"detached"
print("card\ttree\tHEAD\tverdict_file\taudited_sha\tmatch")
n=m=0
for c in ids:
    tree=None
    for cand in (f"{root}/.moai/worktrees/{c}",f"{root}/.claude/worktrees/{c}"):
        if os.path.isdir(cand): tree=cand;break
    if not tree: print(c,"no tree",sep="\t"); continue
    hd,why=head_of(tree)
    files=sorted(glob.glob(f"{tree}/.moai/reports/{c}/sync-audit*.md")+glob.glob(f"{tree}/.moai/reports/{c}/verdict.md"),key=os.path.getmtime)
    if not files: print(c,tree.split('/')[-1],(hd or "?")[:12],"no verdict file","-","-",sep="\t"); n+=1; continue
    f=files[-1]; txt=open(f).read()
    m_=re.search(r"^audited_sha: ([0-9a-f]{7,40})",txt,re.M)
    a=m_.group(1) if m_ else None
    ok = bool(a and hd and hd.startswith(a)) 
    n+=1; m+=ok
    print(c,tree.split('/')[-1],(hd or "?")[:12],os.path.basename(f),a[:12] if a else "none",ok,sep="\t")
print(f"cards_with_tree={n} match_current_head={m}")
