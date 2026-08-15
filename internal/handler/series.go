package handler

import (
	"WorksKeeper/internal/service"
	"net/http"
)

type SeriesHandler struct {
	seriesService *service.SeriesService
}

func (sh *SeriesHandler) Init(seriesService *service.SeriesService) {
	sh.seriesService = seriesService
}

func (sh *SeriesHandler) HandleRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sh.handleGet(rw, r)
	}
}

func (sh *SeriesHandler) handleGet(rw http.ResponseWriter, r *http.Request) {
	id := mustNumberingStringToId(r.PathValue("numbering"))
	sh.seriesService.GetTemplateData(r.Context(), id).ExecuteTemplate(rw)
}
