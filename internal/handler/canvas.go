package handler

import (
	"WorksKeeper/internal/service"
	"log"
	"net/http"
)

type CanvasHandler struct {
	canvasService *service.CanvasService
}

func (ch *CanvasHandler) Init(canvasService *service.CanvasService) {
	ch.canvasService = canvasService
}

func (ch *CanvasHandler) HandleRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ch.handleGet(rw, r, true)
	case http.MethodPost:
		ch.handlePost(rw, r)
	}
}

func (ch *CanvasHandler) HandleRequestNew(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ch.handleGetNew(rw, r)
	}
}

func (ch *CanvasHandler) handleGet(rw http.ResponseWriter, r *http.Request, editing bool) {
	id := mustNumberingStringToId(r.PathValue("numbering"))
	ch.canvasService.GetTemplateData(id, editing).ExecuteTemplate(rw)
}

func (ch *CanvasHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
	/*
		view-work
		add-text
		add-media
		edit-work

		delete content
		move content */

	action, _ := extractActionAndValueFromRequest(r)
	log.Println(action)

	switch action {
	case "edit-work":
		ch.handleGet(rw, r, true)
	case "view-work":
		ch.handleGet(rw, r, false)
	}
}

func (ch *CanvasHandler) handleGetNew(rw http.ResponseWriter, r *http.Request) {
	work := ch.canvasService.GetNewWork()
	http.Redirect(rw, r, "/compose"+work.GetNumberingUrlString(), http.StatusFound)
}
