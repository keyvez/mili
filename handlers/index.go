package handlers

import (
	"disaster/pages"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	component := pages.Index()
	component.Render(ctx, w)
}
