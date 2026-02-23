package view

import (
	"WorksKeeper/internal/entity"
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strconv"
)

type WorkView struct {
	work   *entity.Work
	canvas *CanvasView
}

func NewEmptyWorkView(work *entity.Work) *WorkView {
	return &WorkView{
		work: work,
	}
}

func NewWorkView(work *entity.Work, canvasView *CanvasView) *WorkView {
	return &WorkView{
		work:   work,
		canvas: canvasView,
	}
}

func (wv *WorkView) GetId() int64 {
	return wv.work.Id
}

func (wv *WorkView) GetNumberingUrlString() string {
	lenSerial := len(strconv.Itoa(int(wv.work.Id)))
	serialString := fmt.Sprintf("%d%d", lenSerial, int(wv.work.Id))

	centuryString := strconv.Itoa((wv.work.CreatedAt.Year()-1)/100 + 1)
	lenCentury := len(centuryString)
	centuryString = fmt.Sprintf("%d%s", lenCentury, centuryString)

	yearString := strconv.Itoa(wv.work.CreatedAt.Year() % 100)
	lenYear := len(yearString)
	yearString = fmt.Sprintf("%d%s", lenCentury, yearString)

	dayString := strconv.Itoa(wv.work.CreatedAt.YearDay())
	lenDay := len(dayString)
	dayString = fmt.Sprintf("%d%s", lenDay, dayString)

	if lenSerial > 6 || lenCentury > 3 || lenYear > 3 || lenDay > 3 {
		panic("invalid numbering provided when creating url")
	}

	return fmt.Sprintf("/work/%s%s%s%s",
		serialString,
		centuryString,
		yearString,
		dayString)
}

func (wv *WorkView) GetTimeRequiredString() string {
	return "-1"
}

func (wv *WorkView) GetTitle() string {
	return wv.work.Title
}

func (wv *WorkView) ToHtml() template.HTML {
	var buf bytes.Buffer
	err := entity.WorkTemplate.Execute(&buf, wv)
	if err != nil {
		log.Panicln(err)
	}

	return template.HTML(buf.String())
}

func (wv *WorkView) ToEditableHtml() template.HTML {
	var buf bytes.Buffer
	err := entity.EditableWorkTemplate.Execute(&buf, wv)
	if err != nil {
		log.Panicln(err)
	}

	return template.HTML(buf.String())
}
