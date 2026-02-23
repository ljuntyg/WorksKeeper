package handler

import (
	"WorksKeeper/internal/frontend"
	"WorksKeeper/internal/service"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type CanvasHandler struct {
	service  *service.CanvasService
	template *template.Template
}

type CanvasState struct {
	Canvasable frontend.Canvasable
	IsEditing  bool
}

func NewCanvasHandler(service *service.CanvasService, template *template.Template) *CanvasHandler {
	return &CanvasHandler{
		service:  service,
		template: template,
	}
}

func (ch *CanvasHandler) HandleRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ch.handleGet(rw, r)
		/* case http.MethodPost:
		ch.handlePost(rw, r) */
	}
}

func (ch *CanvasHandler) HandleRequestNew(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ch.handleGetNew(rw, r)
	}
}

func (ch *CanvasHandler) handleGet(rw http.ResponseWriter, r *http.Request) {
	numbering := r.PathValue("numbering")
	if numbering == "" {
		log.Panicln("no numbering in path")
	}

	idLen, err := strconv.Atoi(numbering[:1])
	if err != nil {
		log.Panicln("invalid numbering")
	}

	id, idErr := strconv.Atoi(numbering[1 : 1+idLen])
	if idErr != nil {
		log.Panicln("unable to parse id")
	}

	work := ch.service.GetWorkView(int64(id))

	state := CanvasState{
		Canvasable: work,
		IsEditing:  false,
	}

	ch.writeHtml(rw, r, state)
}

func (ch *CanvasHandler) handleGetNew(rw http.ResponseWriter, r *http.Request) {
	work := ch.service.GetNewWorkView()
	http.Redirect(rw, r, "/compose"+work.GetNumberingUrlString(), http.StatusFound)
}

func (ch *CanvasHandler) writeHtml(rw http.ResponseWriter, r *http.Request, state CanvasState) {
	ch.template.Execute(rw, state)
}
