package handler

import (
	"WorksKeeper/internal/service"
	"log"
	"net/http"
	"strconv"
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
	workId := mustNumberingStringToId(r.PathValue("numbering"))
	ch.canvasService.GetTemplateData(workId, editing).ExecuteTemplate(rw)
}

func (ch *CanvasHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
	/*
		view-work
		add-text
		add-media
		edit-work
		add-group

		delete content
		move content */

	action, value := extractActionAndValueFromRequest(r)
	log.Println(action, value)

	switch action {
	case "edit-work":
		ch.handleGet(rw, r, true)
	case "view-work":
		ch.handleGet(rw, r, false)
	case "add-group":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		workId := mustNumberingStringToId(r.PathValue("numbering"))
		ch.handleAddGroup(rw, r, workId, int64(groupId))
	}
}

func (ch *CanvasHandler) handleGetNew(rw http.ResponseWriter, r *http.Request) {
	work := ch.canvasService.GetNewWork()
	http.Redirect(rw, r, "/compose"+work.GetNumberingUrlString(), http.StatusFound)
}

func (ch *CanvasHandler) handleAddGroup(rw http.ResponseWriter, r *http.Request, workId int64, groupId int64) {
	ch.canvasService.GetNewGroupForGroup(workId, groupId)
	ch.handleGet(rw, r, true)
}
