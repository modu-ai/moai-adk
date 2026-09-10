# t610 baseline — exact commands

Tree: worktree `.claude/worktrees/t610`, branch WT-go-1266, HEAD d3b7d438d (go.mod:3 `go 1.26.4`, unmodified).
Run by lane-8 on 2026-09-10. `$S` = session scratchpad; outputs copied here afterwards.

## govulncheck cells (sequential, one background invocation)

```bash
for tc in auto go1.26.6 go1.26.8; do
  GOTOOLCHAIN=$tc go -C $W version > $S/vuln-$tc.goversion 2>&1
  GOTOOLCHAIN=$tc GOMAXPROCS=2 timeout 900 govulncheck -C $W ./... > $S/vuln-$tc.log 2>&1
  echo "$tc exit=$?" >> $S/vuln-exits.txt
done
```

`$W` = `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610`. govulncheck binary: `~/go/bin/govulncheck`.
Files: `goversion-<tc>.txt`, `govulncheck-<tc>.log`, `govulncheck-exits.txt`.

## Toolchain probes

```bash
command -v go                      # /opt/homebrew/bin/go
go env GOTOOLCHAIN GOROOT          # auto / ~/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.4.darwin-arm64
GOTOOLCHAIN=local /opt/homebrew/bin/go version   # go1.26.0 (run from /private/tmp)
GOTOOLCHAIN=go1.26.6 go version    # go: downloading go1.26.6 … go version go1.26.6
curl -s 'https://go.dev/dl/?mode=json'           # stable: go1.26.8, go1.27.1
```

## Version references

```bash
/usr/bin/grep -rnE 'go-version|GO_VERSION|GOTOOLCHAIN|golang:1\.|setup-go' ./.github ./Makefile ./.goreleaser.yaml ./.goreleaser.yml ./Dockerfile   # ci-go-version-refs.txt (exit 2: .goreleaser.yaml and Dockerfile absent)
/usr/bin/grep -rnF '1.26.4' --include=… --exclude-dir=node_modules --exclude-dir=reports --exclude-dir=specs .   # doc-1264-refs.txt
/usr/bin/grep -rnE 'CGO_ENABLED' ./.goreleaser.yml ./Makefile ./.github/workflows/release.yml   # cgo-enabled-refs.txt
```

## 1.26.6 → 1.26.8 delta

```bash
curl -sL 'https://go.dev/doc/devel/release'      # go-release-history.html (entries go1.26.6..go1.26.8)
gh issue list -R golang/go --state all --limit 100 --search 'milestone:Go1.26.7 label:CherryPickApproved' --json number,title   # go1.26.7-milestone-issues.json
gh issue list -R golang/go --state all --limit 100 --search 'milestone:Go1.26.8 label:CherryPickApproved' --json number,title   # go1.26.8-milestone-issues.json
```
