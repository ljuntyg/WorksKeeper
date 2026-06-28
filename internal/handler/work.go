package handler

import (
	"WorksKeeper/internal/service"
	"net/http"
)

type WorkHandler struct {
	workService *service.WorkService
}

func (wh *WorkHandler) Init(workService *service.WorkService) {
	wh.workService = workService
}

func (wh *WorkHandler) HandleRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		wh.handleGet(rw, r)
	}
}

func (wh *WorkHandler) handleGet(rw http.ResponseWriter, r *http.Request) {
	id := mustNumberingStringToId(r.PathValue("numbering"))
	wh.workService.GetTemplateData(id).ExecuteTemplate(rw)
}
