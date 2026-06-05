package web

import (
	"embed"
	"io/fs"
)

// assetsFS bundles the Console's static assets (the 모두의AI design-system
// console CSS, the self-hosted Pretendard + Noto Sans CJK woff2 subsets + their
// OFL-1.1 licenses, the progressive-enhancement JS, the interface-i18n
// dictionary) into the binary (REQ-WC-005, REQ-WC4-012, REQ-WC5-011,
// REQ-WC6-006/011). No separate JS runtime, build toolchain, icon CDN, or
// network fetch of frontend dependencies is required — fonts, the i18n
// dictionary, and icons are served offline from this embed, preserving the
// loopback-only zero-network invariant. The assets/fonts glob covers both the
// Pretendard Latin+Hangul subset (004) and the Noto Sans CJK ja/zh subset (005).
//
// SPEC-WEB-CONSOLE-006: page.html.tmpl was dropped from this embed — the page is
// now rendered by the compiled-in Templ root component page(view) (the generated
// *_templ.go is compiled Go, NOT an embedded data asset). htmx.min.js is added at
// M4 for the HTMX foundation.
//
//go:embed assets/console.css assets/app.js assets/i18n.js assets/fonts
var assetsFS embed.FS

// errDictKey was the error sentinel for the retired html/template "dict" FuncMap
// helper. It is declared in validate.go, which SPEC-WEB-CONSOLE-006 keeps
// BYTE-UNCHANGED (REQ-WC6-004 / AC-WC6-004) — so the now-unused sentinel cannot be
// deleted from there. This blank reference keeps it live (the `unused` linter would
// otherwise flag it) without modifying the byte-unchanged validator file. The
// sentinel itself is dead with the dict helper; the reference is the minimal
// keep-alive that honors the byte-unchanged constraint.
var _ = errDictKey

// staticFS exposes the CSS/JS/font assets under their bare paths so the
// /static/ handler can serve assets/console.css as /static/console.css and
// assets/fonts/Pretendard-Regular.subset.woff2 as
// /static/fonts/Pretendard-Regular.subset.woff2.
func staticFS() fs.FS {
	sub, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		// Unreachable: the embed directive guarantees the assets/ subtree exists.
		panic(err)
	}
	return sub
}
