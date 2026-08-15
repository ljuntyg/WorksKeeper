package handler

import (
	"WorksKeeper/internal/service"
	"net/http"
)

type HomeHandler struct {
	homeService *service.HomeService
}

func (h *HomeHandler) Init(homeService *service.HomeService) {
	h.homeService = homeService
}

func (h *HomeHandler) HandleRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(rw, r)
	}
}

func (h *HomeHandler) handleGet(rw http.ResponseWriter, r *http.Request) {
	/* templ := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
	}).ParseFiles("./resources/home.html")) */

	// TODO:
	/* listings, hasMore := getNewListings(5, nil) */

	/* templ.Execute(w, &ListingsHasMore{
		Listings: listings,
		HasMore:  hasMore,
	}) */
	h.homeService.GetTemplateData(r.Context()).ExecuteTemplate(rw)
}
