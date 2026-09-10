# AC-GTS2-008 Evidence — M5 sync-phase doc version bump

Tree pin (pre-commit measurement, this working tree): HEAD `403bac94b339c19ebcee29e12b1a722cfd3d69ee`.

Working-tree files were already edited to `1.26.8` before this measurement (M5 doc bump). Commands
below were run individually, each output redirected to its own sibling file, exit code captured
unpiped from the command's own `$?`.

## 1.26.4 counts (expect 0 each, post-edit)

| File | Command | stdout | exit |
|---|---|---|---|
| `.moai/project/product.md` | `grep -c '1\.26\.4' .moai/project/product.md` | `0` | 1 (no-match grep; judged on stdout) |
| `.moai/project/structure.md` | `grep -c '1\.26\.4' .moai/project/structure.md` | `0` | 1 (no-match grep; judged on stdout) |
| `.moai/project/codemaps/overview.md` | `grep -c '1\.26\.4' .moai/project/codemaps/overview.md` | `0` | 1 (no-match grep; judged on stdout) |
| `.moai/project/codemaps/modules.md` | `grep -c '1\.26\.4' .moai/project/codemaps/modules.md` | `0` | 1 (no-match grep; judged on stdout) |

Sum: 0. Raw output + exit files: `ac008-product-1264.{txt,exit}`, `ac008-structure-1264.{txt,exit}`,
`ac008-overview-1264.{txt,exit}`, `ac008-modules-1264.{txt,exit}`.

## 1.26.8 counts (expect total 5)

| File | Command | stdout | exit |
|---|---|---|---|
| `.moai/project/product.md` | `grep -c '1\.26\.8' .moai/project/product.md` | `2` | 0 |
| `.moai/project/structure.md` | `grep -c '1\.26\.8' .moai/project/structure.md` | `1` | 0 |
| `.moai/project/codemaps/overview.md` | `grep -c '1\.26\.8' .moai/project/codemaps/overview.md` | `1` | 0 |
| `.moai/project/codemaps/modules.md` | `grep -c '1\.26\.8' .moai/project/codemaps/modules.md` | `1` | 0 |

Sum: 2 + 1 + 1 + 1 = 5. Matches AC-GTS2-008's expected total. Raw output + exit files:
`ac008-product-1268.{txt,exit}`, `ac008-structure-1268.{txt,exit}`, `ac008-overview-1268.{txt,exit}`,
`ac008-modules-1268.{txt,exit}`.

## Control — grep reads the token (pre-edit tree, base commit `403bac94b`)

Proves the grep instrument actually reads the `1.26.4` token, rather than passing vacuously on an
empty or wrong file.

```
command : git show 403bac94b339c19ebcee29e12b1a722cfd3d69ee:.moai/project/product.md > control-product-at-403bac94b.md
exit    : 0

command : grep -c '1\.26\.4' control-product-at-403bac94b.md
stdout  : 2
exit    : 0
```

Files: `control-product-at-403bac94b.md` (the pre-edit product.md content at HEAD `403bac94b`),
`control-product-1264.txt` (`2`), `control-product-1264.exit` (`0`). The `git show` exit code above
was not saved to a file (sync-audit F2); the 326-line capture it produced is the observable result.

The control's `2` matches acceptance.md § D.0 E-08's baseline stdout for `product.md` (2 mentions: lines
244, 300), confirming the grep pattern is live against a tree that still carries the old token.

## Verdict

AC-GTS2-008: PASS. 0 `1.26.4` mentions across the 4 project documents (working tree, post-edit);
5 `1.26.8` mentions total (2 + 1 + 1 + 1); control confirms the grep instrument reads the token
against a tree known to carry it.
