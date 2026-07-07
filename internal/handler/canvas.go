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
		edit-work

		add-text
		add-media
		add-group

		delete-content
		content-neg-dir
		content-pos-dir

		delete caption
		add caption

		upload
	*/

	action, value := extractActionAndValueFromRequest(r)
	log.Println(action, value)

	// TODO: save texts in textboxes

	switch action {
	case "edit-work":
		ch.handleGet(rw, r, true)
	case "view-work":
		ch.handleGet(rw, r, false)
	case "add-text":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		ch.handleAddText(rw, r, int64(groupId))
	case "add-media":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		ch.handleAddMedia(rw, r, int64(groupId))
	case "add-group":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		ch.handleAddGroup(rw, r, int64(groupId))
	case "content-increase-position":
		contentId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Content id")
		}

		ch.handleContentIncreasePosition(rw, r, int64(contentId))
	case "content-decrease-position":
		contentId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Content id")
		}

		ch.handleContentDecreasePosition(rw, r, int64(contentId))
	}
}

func (ch *CanvasHandler) handleGetNew(rw http.ResponseWriter, r *http.Request) {
	work := ch.canvasService.MustInsertNewWorkInBaseCollection()
	http.Redirect(rw, r, "/compose"+work.GetNumberingUrlString(), http.StatusFound)
}

func (ch *CanvasHandler) handleAddText(rw http.ResponseWriter, r *http.Request, groupId int64) {
	ch.canvasService.MustInsertNewTextInGroup(groupId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleAddMedia(rw http.ResponseWriter, r *http.Request, groupId int64) {
	ch.canvasService.MustInsertNewMediaInGroup(groupId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleAddGroup(rw http.ResponseWriter, r *http.Request, groupId int64) {
	ch.canvasService.MustInsertNewGroupInGroup(groupId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleContentIncreasePosition(rw http.ResponseWriter, r *http.Request, contentId int64) {
	ch.canvasService.MustIncreaseContentPosition(contentId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleContentDecreasePosition(rw http.ResponseWriter, r *http.Request, contentId int64) {
	ch.canvasService.MustDecreaseContentPosition(contentId)
	ch.handleGet(rw, r, true)
}
