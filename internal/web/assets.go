package web

import (
	"embed"
	"html/template"
	"io/fs"
)

// assetsFS bundles the Console's static assets (the 모두의AI design-system
// console CSS, the self-hosted Pretendard woff2 subset + its OFL-1.1 license,
// the progressive-enhancement JS) and the HTML page template into the binary
// (REQ-WC-005, REQ-WC4-012). No separate JS runtime, build toolchain, icon CDN,
// or network fetch of frontend dependencies is required — fonts and icons are
// served offline from this embed, preserving the loopback-only zero-network
// invariant.
//
//go:embed assets/console.css assets/app.js assets/page.html.tmpl assets/fonts
var assetsFS embed.FS

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

// pageTemplate parses the embedded HTML page template once at startup. The
// "dict" helper builds a map for passing keyed values into the nested
// langSelect template (html/template has no built-in map constructor).
func pageTemplate() (*template.Template, error) {
	funcs := template.FuncMap{
		"dict": func(pairs ...any) (map[string]any, error) {
			m := make(map[string]any, len(pairs)/2)
			for i := 0; i+1 < len(pairs); i += 2 {
				key, ok := pairs[i].(string)
				if !ok {
					return nil, errDictKey
				}
				m[key] = pairs[i+1]
			}
			return m, nil
		},
	}
	return template.New("page.html.tmpl").Funcs(funcs).ParseFS(assetsFS, "assets/page.html.tmpl")
}
