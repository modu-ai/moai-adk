package web

import "net/http"

// registerLegacyRoutes keeps the retired screen path as a redirect and nothing else: an old bookmark or a
// browser history entry lands on the screen that replaced it. The path has no handler logic and renders no
// page; the redirect is the whole of what survives of the old name.
func registerLegacyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /kanban", func(w http.ResponseWriter, r *http.Request) {
		target := "/factory"
		if q := r.URL.RawQuery; q != "" {
			target += "?" + q
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}
