package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"maps"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// TODO: Make interfaced?
type HtmlEnvironment struct {
	AllTags []*TagValues
}

type TemplateData[T any] struct {
	Environment *HtmlEnvironment
	Data        T
}

type Numbering struct {
	Type               string
	Serial             int
	Random             int
	Century, Year, Day int
}

type Numberable interface {
	GetNumberingUrlString() string
	GetNumbering() Numbering
}

type Databaseable interface {
	Save()
}

// Series -> Work -> Contents
type Lengthable interface {
	GetLength() int
}

// Make Templatable and Editable
type Listable interface {
	Numberable
	Databaseable
	Lengthable
	GetId() uuid.UUID
	GetParentListing() Listable
	GetTitle() string
	GetTimeRequiredMinutes() int
	GetTags() []*Tag
	GetIsPublic() bool
	String() string
}

type Series struct {
	Id           uuid.UUID
	Numbering    Numbering
	ParentSeries *Series
	Title        string
	IsPublic     bool
	Listings     []Listable
	Tags         []*Tag
}

type Canvasable interface {
	AddContent(c Contentable)
	AddRowContent(vert int, c Contentable)
	RemoveContent(c Contentable)
	RemoveRow(vert int)
	RemoveInRow(vert int, hor int)
	MoveRowUp(vert int)
	MoveRowDown(vert int)
	MoveContentLeft(c Contentable)
	MoveContentRight(c Contentable)
}

type HandlerCanvasable interface {
	Canvasable
	HandleCanvasGetView(w http.ResponseWriter, r *http.Request)
	HandleCanvasGetEdit(w http.ResponseWriter, r *http.Request)
	HandleCanvasGetExisting(w http.ResponseWriter, r *http.Request)
	HandleCanvasExistingData(w http.ResponseWriter, r *http.Request)
	HandleCanvasAddText(w http.ResponseWriter, r *http.Request)
	HandleCanvasAddMedia(w http.ResponseWriter, r *http.Request)
	HandleCanvasAddCaption(w http.ResponseWriter, r *http.Request)
	HandleCanvasAddHorizontal(w http.ResponseWriter, r *http.Request)
	HandleCanvasContentUp(w http.ResponseWriter, r *http.Request)
	HandleCanvasContentDown(w http.ResponseWriter, r *http.Request)
	HandleCanvasHorizontalRight(w http.ResponseWriter, r *http.Request)
	HandleCanvasHorizontalLeft(w http.ResponseWriter, r *http.Request)
	HandleCanvasMediaUpload(w http.ResponseWriter, r *http.Request)
	HandleCanvasDeleteContent(w http.ResponseWriter, r *http.Request)
	HandleCanvasDeleteRow(w http.ResponseWriter, r *http.Request)
	HandleCanvasDeleteCaption(w http.ResponseWriter, r *http.Request)
	HandleCanvasDeleteHorizontal(w http.ResponseWriter, r *http.Request)
}

type TagValues struct {
	Name                  string
	PossibleValues        []string
	PossibleModeModifiers []string
	PossibleFilterModes   []string
}

type Tag struct {
	Id           uuid.UUID
	Name         string
	Value        string
	ModeModifier string
	FilterMode   string
}

type Templatable interface {
	Pathable
	ToHtml() template.HTML
}

type Editable interface {
	Pathable
	ToEditableHtml() template.HTML
}

// Contents referred to by indices are never empty or nil, the type of first Contentable in the Row defines the type of the ContentGroup
type ContentGroup struct {
	Row        *ContentRow
	Start, End int // Row[Start:End] -> [Start, End) on Row
}

type GroupedContent struct {
	Content Contentable
	Group   *ContentGroup
}

// A ContentRow must have at least one member to be valid, the vertical of the first content defines the vertical of the row
type ContentRow struct {
	Members []*GroupedContent
}

type Work struct {
	Id           uuid.UUID
	Numbering    Numbering
	ParentSeries *Series
	Title        string
	IsPublic     bool
	ContentRows  []*ContentRow
	Tags         []*Tag
}

type HandlerSearchable interface {
	HandleSearchGetExisting(w http.ResponseWriter, r *http.Request)
	HandleSearchAddFilter(w http.ResponseWriter, r *http.Request)
	HandleSearchRemoveFilter(w http.ResponseWriter, r *http.Request)
	HandleSearchDoSearch(w http.ResponseWriter, r *http.Request)
}

type Filter struct {
	Listables    []Listable
	FilterGroups map[string][]*Tag
	PreviousTag  string // Name of previous Tag added to the filter, or "" if a Tag was removed or none have been added
}

type Pathable interface {
	GetName(prefix string) string
	GetPath(fileName string) string
}

type Verticable interface {
	GetVertical() int
	SetVertical(vertical int)
}

type Position struct {
	Vertical   int // 0-indexed, 0 is the highest vertical position
	Horizontal int
}

type Positionable interface {
	Verticable
	GetHorizontal() int
	SetHorizontal(horizontal int)
	GetPosition() Position
	SetPosition(vertical int, horizontal int)
}
type Contentable interface {
	Templatable
	Editable
	Positionable
	Databaseable
	GetId() uuid.UUID
}

type Text struct {
	Id       uuid.UUID
	Text     string
	Position Position
}

type Caption struct {
	Text string
}

type Captionable interface {
	GetCaption() *Caption
	SetCaption(cap *Caption)
	GetCaptionHtml() template.HTML
	GetEditableCaptionHtml() template.HTML
}

type Mediable interface {
	Contentable
	Captionable
	GetSources() []string
	SetSources(sources []string)
}

type EmptyMedia struct {
	Id       uuid.UUID
	Position Position
}

type Sound struct {
	Id       uuid.UUID
	Work     *Work
	Sources  []string
	Position Position
	Caption  *Caption
}

type Video struct {
	Id       uuid.UUID
	Work     *Work
	Sources  []string
	Position Position
	Caption  *Caption
}

type Image struct {
	Id       uuid.UUID
	Work     *Work
	Sources  []string
	Position Position
	Caption  *Caption
}

type ContentType int

type MediaType int

const ContentTextType ContentType = 0

const (
	MediaEmptyType MediaType = iota
	MediaSoundType
	MediaVideoType
	MediaImageType
)

// ----------------------------
//
//	LISTINGS LISTINGS LISTINGS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (s *Series) GetId() uuid.UUID {
	return s.Id
}

func (s *Series) GetParentListing() Listable {
	return s.ParentSeries
}

func (s *Series) GetTitle() string {
	return s.Title
}

func (s *Series) GetTimeRequiredMinutes() int {
	totalTime := 0
	for _, elem := range s.Listings {
		totalTime += elem.GetTimeRequiredMinutes()
	}

	return totalTime
}

func (s *Series) GetTags() []*Tag {
	return s.Tags
}

func (s *Series) GetIsPublic() bool {
	return s.IsPublic
}

func (s *Series) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Series: %s\n", s.Title)

	if len(s.Listings) != 0 {
		b.WriteString("With Listings:\n")
	}

	for _, l := range s.Listings {
		b.WriteString(l.String() + "\n")
	}

	return b.String()
}

func (s *Series) GetNumberingUrlString() string {
	return s.Numbering.numberingToUrlString()
}

func (s *Series) GetNumbering() Numbering {
	return s.Numbering
}

// TODO:
func (s *Series) Save() {
	getMockDb().SaveSeries(s)
}

// TODO:
func (s *Series) GetLength() int {
	return -1
}

func (w *Work) GetId() uuid.UUID {
	return w.Id
}

func (w *Work) GetParentListing() Listable {
	return w.ParentSeries
}

func (w *Work) GetTitle() string {
	return w.Title
}

// TODO:
func (w *Work) GetTimeRequiredMinutes() int {
	log.Println("calling MOCK GetTimeRequiredMinutes()")

	return w.GetLength()
}

func (w *Work) GetTags() []*Tag {
	return w.Tags
}

func (w *Work) GetIsPublic() bool {
	return w.IsPublic
}

func (w *Work) String() string {
	return fmt.Sprintf("Work: %s", w.Title)
}

func (w *Work) GetNumberingUrlString() string {
	return w.Numbering.numberingToUrlString()
}

func (w *Work) GetNumbering() Numbering {
	return w.Numbering
}

// TODO:
func (w *Work) Save() {
	getMockDb().SaveWork(w)
}

// TODO:
func (w *Work) GetLength() int {
	return -1
}

func (w *Work) RemoveContent(c Contentable) {
	if c.GetHorizontal() == 0 && len(w.ContentRows[c.GetVertical()].Members) == 1 {
		w.RemoveRow(c.GetVertical())
	} else { // we need to remove the content and "fill" the gap by slicing
		w.RemoveInRow(c.GetVertical(), c.GetHorizontal())
	}
}

func (w *Work) RemoveRow(vert int) {
	w.ContentRows = append(w.ContentRows[:vert], w.ContentRows[vert+1:]...)

	// TODO: Only need to update verticals after the row removed
	for v := vert; v < len(w.ContentRows); v++ {
		w.ContentRows[v].SetVertical(v) // Update contents
	}

	w.Save()
}

func (w *Work) RemoveInRow(vert int, hor int) {
	w.ContentRows[vert].RemoveContent(hor)
	w.Save()
}

// Add a new row to the work
func (w *Work) AddContent(c Contentable) {
	vert := len(w.ContentRows)
	c.SetVertical(vert)
	newRow := &ContentRow{}
	newRow.AddContent(c)

	w.ContentRows = append(w.ContentRows, newRow)
	w.Save()
}

func (w *Work) AddRowContent(vert int, c Contentable) {
	w.ContentRows[vert].AddContent(c)
	w.Save()
}

func (w *Work) MoveRowUp(vert int) {
	if vert == 0 {
		return
	}

	w.ContentRows[vert].SetVertical(vert - 1)
	w.ContentRows[vert-1].SetVertical(vert)

	w.ContentRows[vert], w.ContentRows[vert-1] =
		w.ContentRows[vert-1], w.ContentRows[vert]

	w.Save()
}

func (w *Work) MoveRowDown(vert int) {
	if vert >= len(w.ContentRows)-1 {
		return
	}

	w.ContentRows[vert].SetVertical(vert + 1)
	w.ContentRows[vert+1].SetVertical(vert)

	w.ContentRows[vert], w.ContentRows[vert+1] =
		w.ContentRows[vert+1], w.ContentRows[vert]

	w.Save()
}

func (w *Work) MoveContentLeft(c Contentable) {
	pos := c.GetPosition()
	vert, hor := pos.Vertical, pos.Horizontal

	if hor <= 0 {
		return
	}

	w.ContentRows[vert].SwapContents(hor, hor-1)

	w.Save()
}

func (w *Work) MoveContentRight(c Contentable) {
	pos := c.GetPosition()
	vert, hor := pos.Vertical, pos.Horizontal

	row := w.ContentRows[vert]
	if hor >= len(row.Members)-1 {
		return
	}

	row.SwapContents(hor, hor+1)
	w.Save()
}

type WorkEditing struct {
	Work    *Work
	Editing bool
}

func (work *Work) HandleCanvasGetView(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.New("canvas.html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFiles("./resources/canvas.html"))
	templ.Execute(w, &WorkEditing{
		work,
		false,
	})
}

func (work *Work) HandleCanvasGetEditable(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.New("canvas.html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFiles("./resources/canvas.html"))
	templ.Execute(w, &WorkEditing{
		work,
		true,
	})
}

func (work *Work) HandleCanvasGetExisting(w http.ResponseWriter, r *http.Request) {
	editing, err := strconv.ParseBool(r.FormValue("editing"))
	if err != nil {
		// By default, edit
		work.HandleCanvasGetEditable(w, r)
	} else if !editing {
		work.HandleCanvasGetView(w, r)
	} else {
		work.HandleCanvasGetEditable(w, r)
	}
}

// Only call on a post request
func (work *Work) HandleCanvasExistingData(w http.ResponseWriter, r *http.Request) {
	editing, err := strconv.ParseBool(r.FormValue("editing"))
	if err != nil { // editing should've been passed in and set as a hidden input when getting the canvas before posting
		panic("unexpected canvas state")
	}

	// Handle data on canvas if editing
	if editing {

		err := r.ParseMultipartForm(10 << 20) // TODO: increase?
		if err != nil {
			// TODO: not always this error
			http.Error(w, "content uploaded is too large", http.StatusBadRequest)
			return
		}

		// Only when not all contents have been deleted, but a ContentRow should not exist when length is 0, but maybe this will change in the future
		if len(work.ContentRows) != 0 {
			for k, v := range r.Form {
				if strings.HasPrefix(k, "text[") {
					id, err := uuid.Parse(k[5:41])
					if err != nil {
						panic("unexpected error parsing id for text")
					}

					text := getText(id)
					text.Text = v[0]

					pos := text.GetPosition()
					vert, hor := pos.Vertical, pos.Horizontal
					work.ContentRows[vert].SetContent(hor, text)
					work.Save()
				}

				if strings.HasPrefix(k, "caption[") {
					id, err := uuid.Parse(k[8:44])
					if err != nil {
						panic("unexpected error parsing id for content")
					}

					media := getMedia(id)
					media.SetCaption(&Caption{v[0]})

					pos := media.GetPosition()
					vert, hor := pos.Vertical, pos.Horizontal
					work.ContentRows[vert].SetContent(hor, media)

					work.Save()
				}
			}
		}

		title := r.Form["title-text"]
		work.Title = title[0]
	}
}

func (work *Work) HandleCanvasAddText(w http.ResponseWriter, r *http.Request) {
	newText := ContentTextType.CreateNew()
	work.AddContent(newText)
}

func (work *Work) HandleCanvasAddMedia(w http.ResponseWriter, r *http.Request) {
	media := MediaEmptyType.CreateNew(nil)
	work.AddContent(media)
}

func (work *Work) HandleCanvasAddCaption(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	media := getMedia(id)
	text := r.FormValue("added-caption")
	media.SetCaption(&Caption{text})

	pos := media.GetPosition()
	vert, hor := pos.Vertical, pos.Horizontal
	work.ContentRows[vert].SetContent(hor, media)
	work.Save()
}

func (work *Work) HandleCanvasAddHorizontal(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	content := getContent(id)
	var newContent Contentable

	switch content.(type) {
	case Mediable:
		newContent = MediaEmptyType.CreateNew(nil)
	case Contentable:
		newContent = ContentTextType.CreateNew()
	}

	work.AddRowContent(content.GetVertical(), newContent)
}

func (work *Work) HandleCanvasContentUp(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	content := getContent(id)
	work.MoveRowUp(content.GetVertical())
}

func (work *Work) HandleCanvasContentDown(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	content := getContent(id)
	work.MoveRowDown(content.GetVertical())
}

func (work *Work) HandleCanvasHorizontalRight(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	content := getContent(id)
	work.MoveContentRight(content)
}

func (work *Work) HandleCanvasHorizontalLeft(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	content := getContent(id)
	work.MoveContentLeft(content)
}

func (work *Work) HandleCanvasMediaUpload(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	empty := getEmptyMedia(id)

	file, header, err := r.FormFile("upload")
	if err != nil {
		log.Println(err)
		http.Error(w, "error during content upload", http.StatusInternalServerError)
	} else {
		mt, validMt := getFileMediaType(file)
		if !validMt {
			http.Error(w, "invalid file type uploaded", http.StatusBadRequest)
		}

		defer file.Close()

		localPath := mt.toLocalPath(header.Filename)

		dst, err := os.OpenFile(localPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
		if err != nil {
			// TODO: better handling of dupe names
			for idx := 0; os.IsExist(err); idx++ {
				log.Println("Dupe file name")
				ext := filepath.Ext(localPath)
				localPath = fmt.Sprintf("%s(%d)%s", strings.TrimSuffix(localPath, ext), idx, ext)
				log.Printf("%s(%d).%s", strings.TrimSuffix(localPath, ext), idx, ext)
				log.Println(idx)

				dst, err = os.OpenFile(localPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)

				if idx >= 10 {
					panic("too many attempts at storing file")
				}
			}

			if err != nil {
				log.Println("err")
				http.Error(w, "unexpected error during file upload", http.StatusInternalServerError)
			}
		}

		defer dst.Close()
		io.Copy(dst, file)

		media := mt.CreateNew([]string{mt.toSourceUrl(header.Filename)})
		media.SetPosition(empty.GetVertical(), empty.GetHorizontal())

		work.ContentRows[media.GetVertical()].SetContent(media.GetHorizontal(), media)
		work.Save()
	}
}

func (work *Work) HandleCanvasDeleteContent(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	content := getContent(id)
	work.RemoveContent(content)
}

func (work *Work) HandleCanvasDeleteRow(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	vert, err := strconv.Atoi(actionValue)
	if err != nil {
		panic("unexpecte error parsing row to remove")
	}

	work.RemoveRow(vert)
}

func (work *Work) HandleCanvasDeleteCaption(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	media := getMedia(id)
	media.SetCaption(nil)

	pos := media.GetPosition()
	vert, hor := pos.Vertical, pos.Horizontal
	work.ContentRows[vert].SetContent(hor, media)
	work.Save()
}

func (work *Work) HandleCanvasDeleteHorizontal(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	id := uuid.MustParse(actionValue)
	content := getContent(id)
	work.RemoveContent(content)

}

func (w *Work) ToHtml() template.HTML {
	templ, err := template.ParseFiles(w.getWorkTemplatePath(false))
	if err != nil {
		log.Println(err)
	}
	var buf bytes.Buffer
	if err := templ.Execute(&buf, w); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", "work"))
	}

	return template.HTML(buf.String())
}

func (w *Work) ToEditableHtml() template.HTML {
	name := w.getWorkTemplateName(true)
	path := w.getWorkTemplatePath(true)
	templ := template.Must(template.New(strings.TrimPrefix(name, "editable/") + ".html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFiles(path))

	var buf bytes.Buffer
	if err := templ.Execute(&buf, w); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", "work"))
	}

	return template.HTML(buf.String())
}

// GetName(
func (w *Work) getWorkTemplateName(editable bool) string {
	name := "work"
	if editable {
		name = "editable/work"
	}

	return name
}

// GetPath(
func (w *Work) getWorkTemplatePath(editable bool) string {
	ret := fmt.Sprintf("./resources/templates/listings/%s.html", w.getWorkTemplateName(editable))
	return ret
}

func swap(row []Contentable, i, j, v int) {
	row[i], row[j] = row[j], row[i]
	row[i].SetPosition(v, i)
	row[j].SetPosition(v, j)
}

// ----------------------------------------
//
//	TAGS/FILTERS TAGS/FILTERS TAGS/FILTERS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (filter *Filter) HandleSearchGetExisting(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.New("search.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
		"add":        func(a, b int) int { return a + b },
	}).ParseFiles("./resources/search.html"))

	templateData := createTemplateData(filter)
	templ.Execute(w, templateData)
}

func (filter *Filter) HandleSearchAddFilter(w http.ResponseWriter, r *http.Request) {
	// Extract the tags from the request, because the values in the request might have changed
	// from the values in the retrieved filter
	existingTags := extractTagsFromRequest(r)
	newFilterTag := &Tag{
		Name: r.FormValue("filter-tag"),
	}

	existingTags = append(existingTags, newFilterTag)
	filter.FilterGroups = createFilterGroupsFromTags(existingTags)
	filter.PreviousTag = newFilterTag.Name
}

func (filter *Filter) HandleSearchRemoveFilter(w http.ResponseWriter, r *http.Request) {
	_, actionValue := extractActionAndValueFromRequest(r)

	idx, err := strconv.Atoi(actionValue)
	if err != nil {
		panic("unexpected error parsing index for filter to remove")
	}

	existingTags := extractTagsFromRequest(r)
	existingTags = append(existingTags[:idx], existingTags[idx+1:]...)
	filter.FilterGroups = createFilterGroupsFromTags(existingTags)
	filter.PreviousTag = ""
}

func (filter *Filter) HandleSearchDoSearch(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented HandleSearchDoSearch")
}

func (t *Tag) ToHtml() template.HTML {
	return t.filterTagToHtmlTemplate()
}

// Name of template file
func (t *Tag) GetName(prefix string) string {
	switch t.Name {
	case "author":
		return "datalist"
	case "date":
		return "date"
	case "length":
		return "length"
	case "language":
		return "datalist"
	case "media":
		return "media"
	default:
		panic("getHtmlTemplateString() not fully implemented")
		// TODO
		/* return toHtmlTemplate(tagTextTemplate, "default", nil) */
	}
}

// Path to template file
func (t *Tag) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/tags/%s.html", fileName)
}

// TODO:
func (t *Tag) GetPossibleValues() []string {
	return getMockTagValues(t).PossibleValues
}

// TODO:
func (t *Tag) GetPossibleModeModifiers() []string {
	return getMockTagValues(t).PossibleModeModifiers
}

// TODO:
func (t *Tag) GetPossibleFilterModes() []string {
	return getMockTagValues(t).PossibleFilterModes
}

func (t *Tag) filterTagToHtmlTemplate() template.HTML {
	templ := template.Must(template.ParseFiles(t.GetPath(t.GetName(""))))
	var buf bytes.Buffer
	if err := templ.Execute(&buf, t); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", t.GetName("")))
	}

	return template.HTML(buf.String())
}

// ----------------------------
//
//	CONTENTS CONTENTS CONTENTS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (t *Text) GetId() uuid.UUID {
	return t.Id
}

func (t *Text) GetName(prefix string) string {
	return prefix + "text"
}

// TODO: ???
func (t *Text) GetPath(fileName string) string {
	return getContentTemplatePath(fileName)
}

func (t *Text) ToHtml() template.HTML {
	return contentableToHtmlTemplate(t, false)
}

func (t *Text) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(t, true)
}

func (t *Text) GetVertical() int {
	return t.Position.Vertical
}

func (t *Text) SetVertical(vertical int) {
	t.Position.Vertical = vertical
}

func (t *Text) GetHorizontal() int {
	return t.Position.Horizontal
}

func (t *Text) SetHorizontal(horizontal int) {
	t.Position.Horizontal = horizontal
}

func (t *Text) GetPosition() Position {
	return t.Position
}

func (t *Text) SetPosition(vertical int, horizontal int) {
	t.Position.Vertical, t.Position.Horizontal = vertical, horizontal
}

// TODO:
func (t *Text) Save() {
	getMockDb().SaveText(t)
}

func (m *EmptyMedia) GetId() uuid.UUID {
	return m.Id
}

func (m *EmptyMedia) GetSources() []string {
	return nil
}

func (m *EmptyMedia) SetSources(sources []string) {
}

func (m *EmptyMedia) GetName(prefix string) string {
	// panic("empty media is not templatable")
	return prefix + "empty"
}

func (m *EmptyMedia) GetPath(fileName string) string {
	return getContentTemplatePath(fileName)
}

func (m *EmptyMedia) ToHtml() template.HTML {
	return contentableToHtmlTemplate(m, false)
}

func (m *EmptyMedia) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(m, true)
}

func (m *EmptyMedia) GetVertical() int {
	return m.Position.Vertical
}

func (m *EmptyMedia) SetVertical(vertical int) {
	m.Position.Vertical = vertical
}

func (m *EmptyMedia) GetHorizontal() int {
	return m.Position.Horizontal
}

func (m *EmptyMedia) SetHorizontal(horizontal int) {
	m.Position.Horizontal = horizontal
}

func (m *EmptyMedia) GetPosition() Position {
	return m.Position
}

func (m *EmptyMedia) SetPosition(vertical int, horizontal int) {
	m.Position.Vertical, m.Position.Horizontal = vertical, horizontal
}

func (m *EmptyMedia) GetCaption() *Caption {
	return nil
}

func (m *EmptyMedia) SetCaption(cap *Caption) {}

func (m *EmptyMedia) GetCaptionHtml() template.HTML {
	return mediaToCaptionHtml(m, nil, false)
}

func (m *EmptyMedia) GetEditableCaptionHtml() template.HTML {
	return mediaToCaptionHtml(m, nil, true)
}

// TODO:
func (m *EmptyMedia) Save() {
	getMockDb().SaveEmptyMedia(m)
}

func (s *Sound) GetId() uuid.UUID {
	return s.Id
}

func (s *Sound) GetSources() []string {
	return s.Sources
}

func (s *Sound) SetSources(sources []string) {
	s.Sources = sources
}

func (s *Sound) GetName(prefix string) string {
	return prefix + "sound"
}

func (s *Sound) GetPath(fileName string) string {
	return getContentTemplatePath(fileName)
}

func (s *Sound) ToHtml() template.HTML {
	return contentableToHtmlTemplate(s, false)
}

func (s *Sound) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(s, true)
}

func (s *Sound) GetVertical() int {
	return s.Position.Vertical
}

func (s *Sound) SetVertical(vertical int) {
	s.Position.Vertical = vertical
}

func (s *Sound) GetHorizontal() int {
	return s.Position.Horizontal
}

func (s *Sound) SetHorizontal(horizontal int) {
	s.Position.Horizontal = horizontal
}

func (s *Sound) GetPosition() Position {
	return s.Position
}

func (s *Sound) SetPosition(vertical int, horizontal int) {
	s.Position.Vertical, s.Position.Horizontal = vertical, horizontal
}

// TODO:
func (s *Sound) Save() {
	getMockDb().SaveSound(s)
}

func (s *Sound) GetCaption() *Caption {
	return s.Caption
}

func (s *Sound) SetCaption(cap *Caption) {
	s.Caption = cap
}

func (s *Sound) GetCaptionHtml() template.HTML {
	return mediaToCaptionHtml(s, s.Caption, false)
}

func (s *Sound) GetEditableCaptionHtml() template.HTML {
	return mediaToCaptionHtml(s, s.Caption, true)
}

func (v *Video) GetId() uuid.UUID {
	return v.Id
}

func (v *Video) GetSources() []string {
	return v.Sources
}

func (v *Video) SetSources(sources []string) {
	v.Sources = sources
}

func (v *Video) GetName(prefix string) string {
	return prefix + "video"
}

func (v *Video) GetPath(fileName string) string {
	return getContentTemplatePath(fileName)
}

func (v *Video) ToHtml() template.HTML {
	return contentableToHtmlTemplate(v, false)
}

func (v *Video) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(v, true)
}

func (v *Video) GetCaption() *Caption {
	return v.Caption
}

func (v *Video) SetCaption(cap *Caption) {
	v.Caption = cap
}

func (v *Video) GetCaptionHtml() template.HTML {
	return mediaToCaptionHtml(v, v.Caption, false)
}

func (v *Video) GetEditableCaptionHtml() template.HTML {
	return mediaToCaptionHtml(v, v.Caption, true)
}

func (v *Video) GetVertical() int {
	return v.Position.Vertical
}

func (v *Video) SetVertical(vertical int) {
	v.Position.Vertical = vertical
}

func (v *Video) GetHorizontal() int {
	return v.Position.Horizontal
}

func (v *Video) SetHorizontal(horizontal int) {
	v.Position.Horizontal = horizontal
}

func (v *Video) GetPosition() Position {
	return v.Position
}

func (v *Video) SetPosition(vertical int, horizontal int) {
	v.Position.Vertical, v.Position.Horizontal = vertical, horizontal
}

// TODO:
func (v *Video) Save() {
	getMockDb().SaveVideo(v)
}

func (i *Image) GetId() uuid.UUID {
	return i.Id
}

func (i *Image) GetSources() []string {
	return i.Sources
}

func (i *Image) SetSources(sources []string) {
	i.Sources = sources
}

func (i *Image) GetName(prefix string) string {
	return prefix + "image"
}

func (i *Image) GetPath(fileName string) string {
	return getContentTemplatePath(fileName)
}

func (i *Image) ToHtml() template.HTML {
	return contentableToHtmlTemplate(i, false)
}

func (i *Image) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(i, true)
}

func (i *Image) GetVertical() int {
	return i.Position.Vertical
}

func (i *Image) SetVertical(vertical int) {
	i.Position.Vertical = vertical
}

func (i *Image) GetHorizontal() int {
	return i.Position.Horizontal
}

func (i *Image) SetHorizontal(horizontal int) {
	i.Position.Horizontal = horizontal
}

func (i *Image) GetPosition() Position {
	return i.Position
}

func (i *Image) SetPosition(vertical int, horizontal int) {
	i.Position.Vertical, i.Position.Horizontal = vertical, horizontal
}

// TODO:
func (i *Image) Save() {
	getMockDb().SaveImage(i)
}

func (i *Image) GetCaption() *Caption {
	return i.Caption
}

func (i *Image) SetCaption(cap *Caption) {
	i.Caption = cap
}

func (i *Image) GetCaptionHtml() template.HTML {
	return mediaToCaptionHtml(i, i.Caption, false)
}

func (i *Image) GetEditableCaptionHtml() template.HTML {
	return mediaToCaptionHtml(i, i.Caption, true)
}

// Filename can be prefixed with "editable/"" via .GetName(prefix string)
func getContentTemplatePath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/%s.html", fileName)
}

func contentableToHtmlTemplate(c Contentable, editable bool) template.HTML {
	fileNamePrefix := ""
	if editable {
		fileNamePrefix = "editable/"
	}

	templ := template.Must(template.ParseFiles(c.GetPath(c.GetName(fileNamePrefix))))
	var buf bytes.Buffer
	if err := templ.Execute(&buf, c); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", c.GetName("")))
	}

	return template.HTML(buf.String())
}

func mediaToCaptionHtml(m Mediable, cap *Caption, editable bool) template.HTML {
	fileName := "caption"
	if editable {
		fileName = "editable/" + fileName
	}

	path := fmt.Sprintf("./resources/templates/content/%s.html", fileName)

	type MediableCaption struct {
		Media   Mediable
		Caption *Caption
	}

	templ := template.Must(template.ParseFiles(path))
	var buf bytes.Buffer
	if err := templ.Execute(&buf, &MediableCaption{
		Media:   m,
		Caption: cap,
	}); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", m.GetName("")))
	}

	return template.HTML(buf.String())
}

// ----------------------------------------------------
//
//	CONTENTROW/GROUP CONTENTROW/GROUP CONTENTROW/GROUP
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (cg *ContentGroup) GetVertical() int {
	return cg.Row.GetVertical()
}

// Returns true if the content was the same type as that of the group and was added to the row, or false otherwise
func (cg *ContentGroup) addContent(c Contentable) bool {
	// TODO: check for Sound also? so they end up in their own groups
	if !canMatchGroup(cg.Row.Members[cg.End-1].Content, c) {
		return false
	}

	cg.Row.Members = append(cg.Row.Members, createGroupedContent(c, cg))
	cg.End++

	c.SetHorizontal(cg.End - 1)

	return true
}

// Removes content at position hor
func (cg *ContentGroup) removeContent(hor int) {
	cg.Row.Members = append(cg.Row.Members[:hor], cg.Row.Members[hor+1:]...)

	// Reindex horiozontally Row[hor:]
	for _, member := range cg.Row.Members[hor:] {
		member.Content.SetHorizontal(member.Content.GetHorizontal() - 1)
	}

	cg.End--
}

// i and j can be of same type
// or they can split up a group, like X V V V -> V X V V
func (cg *ContentGroup) swapContents(i, j int) {
	// this method was called on the group of Row[i], so content i must belong to this group
	if i < cg.Start || i >= cg.End {
		panic("unexpected index error when swapping contents")
	}

	row := cg.Row.Members
	iContent := row[i].Content
	jContent := row[j].Content
	if canMatchGroup(iContent, jContent) {
		// CASE 1: i.type == j.type => X X X X -> X X X X

		iContent.SetHorizontal(j)
		jContent.SetHorizontal(i)

		row[i], row[j] = row[j], row[i]
	} else {
		// CASE 2: i.type != j.type => X V V V -> V X V V
		cg.swapAcrossGroups(i, j)
	}

}

func (cg *ContentGroup) swapAcrossGroups(i, j int) {
	row := cg.Row.Members
	rowLen := len(row)

	row[i].Content.SetHorizontal(j)
	row[j].Content.SetHorizontal(i)

	row[i], row[j] = row[j], row[i]
	// i and j have swapped position, i,j -> j,i

	// A. (maybe ... V X V X X X -> V V X X X X, must check if i and j can join their groups)
	// B. (maybe ... V X V V V -> V V X V V, must check if j can join the group)
	// C. (maybe ... X V X X X -> V X X X X, must check if i can join the group)
	caseB := canMatchGroup(row[j].Content, row[j-1].Content)
	caseC := canMatchGroup(row[i].Content, row[i+1].Content)
	caseA := caseB && caseC

	// Adjust boundaries of content group after swap.
	// Only call if types of content group match, only call from swapAcrossGroups.
	// Assumes the content itself has already been moved
	appendFirstFromTo := func(cg *ContentGroup, cg2 *ContentGroup) {
		cg.Start++
		cg2.End++
	}

	prependLastFromTo := func(cg *ContentGroup, cg2 *ContentGroup) {
		cg.End--
		cg2.Start--
	}

	if i > 0 && j < rowLen && caseA {
		// check if both j can join j-1, and i can join i+1, requires i_idx was larger than 0 and j_idx was less than len(row)

		// j must join (j-1).Group and i must join (i+1).Group
		appendFirstFromTo(row[j].Group, row[j-1].Group)
		prependLastFromTo(row[i].Group, row[i+1].Group)
	} else if i > 0 && caseB {
		// check if j can join j-1, requires i_idx was not 0

		appendFirstFromTo(row[j].Group, row[j-1].Group)
	} else if j < rowLen && caseC {
		// check if i can join i+1, requires j_idx was less than len(row)

		prependLastFromTo(row[i].Group, row[i+1].Group)
	} else {
		panic("unexpected case when moving content between groups")
	}

}

// TODO: use some kind of type enum?
func canMatchGroup(c1 Contentable, c2 Contentable) bool {
	switch c1.(type) {
	case *Text:
		if _, ok := c2.(*Text); !ok {
			return false
		}
	case Mediable:
		if _, ok := c2.(Mediable); !ok {
			return false
		}
	default:
		panic("unexpected content type when matching content groups")
	}

	return true
}

func (cr *ContentRow) GetVertical() int {
	return cr.Members[0].Content.GetVertical()
}

func (cr *ContentRow) SetVertical(vertical int) {
	for _, member := range cr.Members {
		member.Content.SetVertical(vertical)
	}
}

// Adds c to the final ContentGroup if types match, if not it appends the content in a new ContentGroup to the row.
// If the content added is the first added to the row, then the vertical position of the content forms the basis for the vertical of the whole row.
func (cr *ContentRow) AddContent(c Contentable) {
	rowLen := len(cr.Members)
	if rowLen == 0 {
		cr.Members = append(cr.Members, cr.getNewGroupedContent(c))
		c.SetHorizontal(0)
		return
	}

	c.SetVertical(cr.GetVertical())
	final := cr.Members[rowLen-1]
	if ok := final.Group.addContent(c); !ok {
		cr.Members = append(cr.Members, createGroupedContent(c, cr.getNewContentGroup(rowLen, rowLen+1)))
		c.SetHorizontal(rowLen)
	}
}

func (cr *ContentRow) RemoveContent(horizontal int) {
	cr.Members[horizontal].Group.removeContent(horizontal)
}

// Returns a new *ContentRow with a vertical index of the Contentable passed.
// If there are multiple rows, only call on the final row
func (cr *ContentRow) GetNewRow(c Contentable) *ContentRow {
	return &ContentRow{
		Members: []*GroupedContent{
			cr.getNewGroupedContent(c),
		},
	}
}

// i must be less than j when calling this method
func (cr *ContentRow) SwapContents(i, j int) {
	cr.Members[i].Group.swapContents(i, j)
}

func (cr *ContentRow) SetContent(hor int, c Contentable) {
	c.SetHorizontal(hor)
	cr.Members[hor].Content = c
}

func (cr *ContentRow) GetContentGroups() []*ContentGroup {
	if len(cr.Members) == 0 {
		return nil
	}

	groups := make([]*ContentGroup, 0)

	var last *ContentGroup
	for _, m := range cr.Members {
		if m.Group != last {
			groups = append(groups, m.Group)
			last = m.Group
		}
	}

	return groups
}

func (cr *ContentRow) getNewContentGroup(startIdx int, endIdx int) *ContentGroup {
	return &ContentGroup{
		Row:   cr,
		Start: startIdx,
		End:   endIdx,
	}
}

func createGroupedContent(c Contentable, cg *ContentGroup) *GroupedContent {
	return &GroupedContent{
		Content: c,
		Group:   cg,
	}
}

// Only use to create the first member on a row
func (cr *ContentRow) getNewGroupedContent(c Contentable) *GroupedContent {
	return &GroupedContent{
		Content: c,
		Group:   cr.getNewContentGroup(0, 1),
	}
}

func (cg *ContentGroup) GetName(prefix string) string {
	members := cg.Row.Members[cg.Start:cg.End]
	switch members[0].Content.(type) {
	case *Text:
		return prefix + "text-group"
	/* case *Sound:
	return prefix + "sound" */
	case *Sound, *Video, *Image, *EmptyMedia:
		return prefix + "media-group"
	default:
		panic("unexpected content type when getting name")
	}
}

func (cg *ContentGroup) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/groups/%s.html", fileName)
}

func (cg *ContentGroup) ToHtml() template.HTML {
	return cg.contentGroupToHtml(false)
}

func (cg *ContentGroup) ToEditableHtml() template.HTML {
	return cg.contentGroupToHtml(true)
}

func (cg *ContentGroup) contentGroupToHtml(editable bool) template.HTML {
	fileNamePrefix := ""
	if editable {
		fileNamePrefix = "editable/"
	}

	name := cg.GetName("")
	path := cg.GetPath(cg.GetName(fileNamePrefix))

	templ, err := template.New(name + ".html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFiles(path)
	if err != nil {
		log.Println(err)
	}

	type MembersLengthVertical struct {
		Members  []*GroupedContent
		Length   int
		Vertical int
	}

	data := &MembersLengthVertical{
		cg.Row.Members[cg.Start:cg.End],
		cg.End - cg.Start,
		cg.Row.GetVertical(),
	}

	var buf bytes.Buffer
	if err := templ.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", name))
	}

	return template.HTML(buf.String())
}

// --------------------------------------
//
//	CONTENTTYPE CONTENTTYPE CONTENTTYPE
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (ct ContentType) CreateNew() Contentable {
	switch ct {
	case ContentTextType:
		return &Text{
			Id: uuid.New(),
		}
	default:
		panic("unexpected content type when creating new")
	}
}

func (ct ContentType) String() string {
	switch ct {
	case ContentTextType:
		return "sound"
	default:
		panic("unexpected content type when getting string")
	}
}

func (mt MediaType) CreateNew(sources []string) Mediable {
	switch mt {

	// TODO: remove first case?
	case MediaEmptyType:
		return &EmptyMedia{
			Id: uuid.New(),
		}
	case MediaSoundType:
		return &Sound{
			Sources: sources,
			Id:      uuid.New(),
		}
	case MediaVideoType:
		return &Video{
			Sources: sources,
			Id:      uuid.New(),
		}
	case MediaImageType:
		return &Image{
			Sources: sources,
			Id:      uuid.New(),
		}
	default:
		panic("unexpected media type when creating new")
	}
}

func (mt MediaType) String() string {
	switch mt {
	case MediaSoundType:
		return "sound"
	case MediaVideoType:
		return "video"
	case MediaImageType:
		return "image"
	default:
		panic("unexpected media type when getting string")
	}
}

// This should NOT be used for content source links
func (mt MediaType) mediaDir() string {
	return "/resources/media/" + mt.String() + "s/"
}

// This should be used for content source links
func (mt MediaType) urlPrefix() string {
	if mt.String() == "unknown" {
		panic("Unexpected ContentType when reading urlPrefix()")
	}

	return "/" + mt.String() + "/with/name/"
}

func (mt MediaType) toLocalPath(fileName string) string {
	return "." + mt.mediaDir() + fileName
}

func (mt MediaType) toSourceUrl(fileName string) string {
	return getBaseUrl() + mt.urlPrefix() + fileName
}

// TODO: improve matching mime -> ContentType
func getFileMediaType(file multipart.File) (MediaType, bool) {
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	file.Seek(0, io.SeekStart)
	mt, ok := MimeToMediaType[mimeType]

	if !ok {
		log.Println("mime type " + mimeType + " not found in mime map")
	}

	return mt, ok
}

var AllContentTypes = []ContentType{
	ContentTextType,
}

var AllMediaTypes = []MediaType{
	MediaSoundType,
	MediaVideoType,
	MediaImageType,
}

var MimeToMediaType = map[string]MediaType{
	"image/apng":      MediaImageType,
	"image/avif":      MediaImageType,
	"image/bmp":       MediaImageType,
	"image/gif":       MediaImageType,
	"image/jpeg":      MediaImageType,
	"image/png":       MediaImageType,
	"image/svg+xml":   MediaImageType,
	"image/tiff":      MediaImageType,
	"image/webp":      MediaImageType,
	"audio/aac":       MediaSoundType,
	"audio/midi":      MediaSoundType,
	"audio/x-midi":    MediaSoundType,
	"audio/mpeg":      MediaSoundType,
	"audio/ogg":       MediaSoundType,
	"audio/wav":       MediaSoundType,
	"audio/webm":      MediaSoundType,
	"audio/3gpp":      MediaSoundType,
	"audio/3gpp2":     MediaSoundType,
	"application/ogg": MediaSoundType, // TODO: ? some .ogg files
	"video/mp4":       MediaVideoType,
	"video/mpeg":      MediaVideoType,
	"video/ogg":       MediaVideoType,
	"video/webm":      MediaVideoType,
	"video/x-msvideo": MediaVideoType,
	"video/mp2t":      MediaVideoType,
	"video/3gpp":      MediaVideoType,
	"video/3gpp2":     MediaVideoType,
}

// --------------------------
//
//	HELPERS HELPERS HELPERS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/

func getEnglishNumberSuffix(nbr int) string {
	if nbr%100 >= 11 && nbr%100 <= 13 {
		return "th"
	}

	switch nbr % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func formatTimeRequired(min int) string {
	if min >= 60 {
		hours := min / 60

		return fmt.Sprintf("%d hr", hours)
	}

	return fmt.Sprintf("%d min", min)
}

// TODO:
func getBaseUrl() string {
	return getMockBaseUrl()
}

// TODO:
func getNewFilter() *Filter {
	return getMockFilter()
}

// TODO:
func getNewWork() *Work {
	work := getMockEmptyWork()
	work.Save()
	return work
}

// TODO:
func getAllListings() []Listable {
	return slices.Collect(
		maps.Values(getMockDb().Listings),
	)
}

// TODO:
// returns bool true if n+1 new were available
func getNListings(n int) ([]Listable, bool) {
	listings := getAllListings()

	available := len(listings)
	if available < n {
		return listings, false
	} else if available >= n+1 {
		return listings[:n], true
	} else { // n <= available < n+1 -> available == n, n+1 not available
		return listings[:n], false
	}
}

// TODO:
// returns bool true if n+1 new were available
func getNewListings(n int, existing []Listable) ([]Listable, bool) {
	if existing == nil {
		return getNListings(n)
	}

	log.Printf("all listings size: %d", len(getAllListings()))

	existingSet := make(map[uuid.UUID]struct{}, len(existing))
	for _, e := range existing {
		existingSet[e.GetId()] = struct{}{}
	}

	result := make([]Listable, 0, n)
	hasMore := false
	nAdded := 0
	for _, l := range getAllListings() {
		if _, found := existingSet[l.GetId()]; found {
			continue
		}

		nAdded++
		if nAdded == n+1 {
			hasMore = true
			break
		}

		result = append(result, l)
	}

	return append(existing, result...), hasMore
}

// TODO:
func getListing(numbering Numbering) Listable {
	return getMockDb().Listings[numbering]
}

// TODO:
func getWork(numbering Numbering) *Work {
	return getMockDb().GetWork(numbering)
}

// TODO:
func getSeries(numbering Numbering) *Series {
	return getMockDb().GetSeries(numbering)
}

// TODO:
func getContent(id uuid.UUID) Contentable {
	return getMockDb().GetContent(id)
}

// TODO:
func getText(id uuid.UUID) *Text {
	return getMockDb().GetText(id)
}

// TODO:
func getMedia(id uuid.UUID) Mediable {
	return getMockDb().GetMedia(id)
}

// TODO:
func getEmptyMedia(id uuid.UUID) *EmptyMedia {
	return getMockDb().GetEmptyMedia(id)
}

// TODO:
func getSound(id uuid.UUID) *Sound {
	return getMockDb().GetSound(id)
}

// TODO:
func getVideo(id uuid.UUID) *Video {
	return getMockDb().GetVideo(id)
}

// TODO:
func getImage(id uuid.UUID) *Image {
	return getMockDb().GetImage(id)
}

// TODO:
/* func getFilter(numbering Numbering) *Filter {
	return getMockDb().GetFilter(numbering)
} */

// TODO:
func getHtmlEnvironment() *HtmlEnvironment {
	return getMockHtmlEnvironment()
}

func createTemplateData[T any](data T) TemplateData[T] {
	return TemplateData[T]{
		Environment: getHtmlEnvironment(),
		Data:        data,
	}
}

// Request must be made to path matched by pattern like
// /foo/boo/goo/{id} and id must refer to an existing work
func getWorkFromRequest(r *http.Request) *Work {
	numberingString := r.PathValue("numbering")
	numbering := mustUrlStringToNumbering(numberingString, "work")

	return getWork(numbering)
}

// Request must be made to path matched by pattern like
// /foo/boo/goo/{id} and id must refer to an existing series
func getSeriesFromRequest(r *http.Request) *Series {
	numberingString := r.PathValue("numbering")
	numbering := mustUrlStringToNumbering(numberingString, "series")

	return getSeries(numbering)
}

func (n Numbering) numberingToUrlString() string {
	lenRand := len(strconv.Itoa(n.Random))
	randString := fmt.Sprintf("%d%d", lenRand, n.Random)

	lenSerial := len(strconv.Itoa(n.Serial))
	serialString := fmt.Sprintf("%d%d", lenSerial, n.Serial)

	centuryString := strconv.Itoa(n.Century)
	lenCentury := len(centuryString)
	centuryString = fmt.Sprintf("%d%s", lenCentury, centuryString)

	yearString := strconv.Itoa(n.Year)
	lenYear := len(yearString)
	yearString = fmt.Sprintf("%d%s", lenCentury, yearString)

	dayString := strconv.Itoa(n.Day)
	lenDay := len(dayString)
	dayString = fmt.Sprintf("%d%s", lenDay, dayString)

	if lenRand > 4 || lenSerial > 6 || lenCentury > 3 || lenYear > 3 || lenDay > 3 {
		panic("invalid numbering provided when creating url")
	}

	return fmt.Sprintf("/%s/%s%s%s%s%s",
		n.Type,
		randString,
		serialString,
		centuryString,
		yearString,
		dayString)
}

func isDigitsOnly(s string) bool {
	if s == "" {
		return false
	}

	for _, i := range s {
		if i < '0' || i > '9' {
			return false
		}
	}

	return true
}

func urlStringToNumbering(s string, typeString string) (Numbering, error) {
	// lenRand[1-4]|Random|lenSerial[1-6]|Serial|lenCentury[1-3]|Century|lenYear[1-3]|Year|lenDay[1-3]|Day, Type
	if !isDigitsOnly(s) {
		return Numbering{}, fmt.Errorf("invalid format: %q", s)
	}

	n := 5
	vals := make([]int, n)
	pos := 0
	for i := range n {
		if pos+1 > len(s) {
			log.Printf("out of bounds with pos %d", pos)
			return Numbering{}, errors.New("unexpected end of string")
		}

		lenI := int(s[pos] - '0')
		pos++

		if pos+lenI > len(s) {
			log.Printf("out of bounds with pos %d, lenI %d", pos, lenI)
			return Numbering{}, errors.New("unexpected end of string")
		}

		val, err := strconv.Atoi(s[pos : pos+lenI])
		if err != nil {
			return Numbering{}, err
		}

		pos += lenI
		vals[i] = val
	}

	return Numbering{
		Random:  vals[0],
		Serial:  vals[1],
		Century: vals[2],
		Year:    vals[3],
		Day:     vals[4],
		Type:    typeString,
	}, nil
}

func mustUrlStringToNumbering(s string, typeString string) Numbering {
	numbering, err := urlStringToNumbering(s, typeString)
	if err != nil {
		panic(err)
	}

	return numbering
}

// Assumes all inputs (which should be buttons) with name "action" are of the format
// "action-name: value"
func extractActionAndValueFromRequest(r *http.Request) (string, string) {
	actionString := r.FormValue("action")
	action, value, found := strings.Cut(actionString, ":")
	if !found {
		panic("unexpected action name when handling action")
	}

	return action, value
}

// Assumes post request made to search
func extractTagsFromRequest(r *http.Request) []*Tag {
	err := r.ParseForm()
	if err != nil {
		panic("unexpected error parsing form")
	}

	existingFilterTags := r.Form["added-filters.tag"]
	existingFilterValues := r.Form["added-filters.value"]
	existingFilterModifiers := r.Form["mode-modifier"]
	existingFilterModes := r.Form["filter-mode"]
	existingTags := make([]*Tag, len(existingFilterTags))

	for i, name := range existingFilterTags {
		existingTags[i] = &Tag{
			Name:         name,
			Value:        existingFilterValues[i],
			ModeModifier: existingFilterModifiers[i],
			FilterMode:   existingFilterModes[i],
		}
	}

	return existingTags
}

func createFilterGroupsFromTags(tags []*Tag) map[string][]*Tag {
	tagsNamed := make(map[string][]*Tag)
	groupOrder := []string{}

	for _, tag := range tags {
		if _, exists := tagsNamed[tag.Name]; !exists {
			groupOrder = append(groupOrder, tag.Name)
		}
		tagsNamed[tag.Name] = append(tagsNamed[tag.Name], tag)
	}

	return tagsNamed
}

func extractListingsFromRequest(r *http.Request) []Listable {
	err := r.ParseForm()
	if err != nil {
		panic("unexpected error parsing form")
	}

	listingIds := r.Form["listing-numberings"]
	listings := make([]Listable, len(listingIds))
	for i, numberingString := range listingIds {
		listingType, numbering, found := strings.Cut(strings.TrimPrefix(numberingString, "/"), "/")
		if !found {
			panic("unexpected error extracting listing")
		}

		listings[i] = getListing(mustUrlStringToNumbering(numbering, listingType))
	}

	return listings
}

func extractFilterFromRequest(r *http.Request) *Filter {
	return &Filter{
		Listables:    extractListingsFromRequest(r),
		FilterGroups: createFilterGroupsFromTags(extractTagsFromRequest(r)),
	}
}

func mustCombineNumbers(numbers []int) int {
	numStrings := make([]string, len(numbers))
	for i, num := range numbers {
		numStrings[i] = strconv.Itoa(num)
	}

	numString := strings.Join(numStrings, "")
	combined, err := strconv.Atoi(numString)
	if err != nil {
		panic("unexpected error combining numbers")
	}

	return combined
}

// ----------------------------
//
//	HANDLERS HANDLERS HANDLERS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/

// Case 1 (localhost):
// request made to localhost -> Nginx proxies to localhost:8080 with X-Suffix-Subdomain = "" -> localhost:8080 redirects to localhost:8080/works

// Case 2 (localhost/resource):
// redirect request made to localhost/works -> Nginx proxies to localhost:8080 with X-Suffix-Subdomain = "works"
// -> Go subdomainPeriodReplacer redirects to works.at.localhost -> Nginx proxies to localhost:8080 with X-Prefix-Subdomain = "works.at"
// -> Go subdomainPeriodReplacer handles and responds to request to works.at.localhost by replacing the request url with localhost/works internally

// Case 3 (subdomain.localhost/resource):
// request made to works.at.localhost/work/2760 -> Nginx proxies to localhost:8080 with X-Prefix-Subdomain = "works.at" and X-Suffix-Subdomain = "work/2760"
// -> Go subdomainPeriodReplacer redirects to work.2760.at.localhost -> Nginx proxies to localhost:8080 with X-Prefix-Subdomain = "work.2760.at"
// -> Go subdomainPeriodReplacer handles and responds to request made to work.2760.at.localhost by replacing the request url with localhost/work/2760 internally
func subdomainPeriodReplacer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suffixSubdomain := r.Header.Get("X-Suffix-Subdomain")
		prefixSubdomain := r.Header.Get("X-Prefix-Subdomain")
		log.Printf("suffix: %s, prefix: %s", suffixSubdomain, prefixSubdomain)
		log.Println("path: " + r.URL.Path)

		if suffixSubdomain != "" {
			subdomain := strings.ReplaceAll(suffixSubdomain, "/", ".")
			target := getMockProtocol() + subdomain + ".at." + getMockHost()
			log.Println("Redirecting to URL: " + target)
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
			return
		} else if prefixSubdomain != "" {
			r.URL.Path = "/" + strings.ReplaceAll(prefixSubdomain, ".", "/")[0:len(prefixSubdomain)-len(".at.")+1]
		}

		log.Println("Continuing to URL: " + r.Host + r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/works", http.StatusSeeOther)
}

func viewWorksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		viewNewWorksHandler(w, r)
	case http.MethodPost:
		viewWorksPostHandler(w, r)
	}
}

type ListingsHasMore struct {
	Listings []Listable
	HasMore  bool
}

func viewNewWorksHandler(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
	}).ParseFiles("./resources/home.html"))

	// TODO:
	listings, hasMore := getNewListings(5, nil)

	templ.Execute(w, &ListingsHasMore{
		Listings: listings,
		HasMore:  hasMore,
	})
}

func viewHandlerGetExisting(w http.ResponseWriter, r *http.Request, listings *ListingsHasMore) {
	templ := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
	}).ParseFiles("./resources/home.html"))
	templ.Execute(w, listings)
}

// TODO:
func viewWorksPostHandler(w http.ResponseWriter, r *http.Request) {
	action, _ := extractActionAndValueFromRequest(r)
	listings := extractListingsFromRequest(r)

	hasMore := false
	switch action {
	case "more-works":
		listings, hasMore = getNewListings(5, listings)
	}

	viewHandlerGetExisting(w, r, &ListingsHasMore{
		Listings: listings,
		HasMore:  hasMore,
	})
}

func searchWorksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		searchWorksNewFilterHandler(w, r)
	case http.MethodPost:
		searchWorksPostHandler(w, r)
	}
}

func searchWorksNewFilterHandler(w http.ResponseWriter, r *http.Request) {
	newFilter := getNewFilter()
	newFilter.HandleSearchGetExisting(w, r)
}

// TODO: handle search
func searchWorksPostHandler(w http.ResponseWriter, r *http.Request) {
	action, _ := extractActionAndValueFromRequest(r)
	filter := extractFilterFromRequest(r)

	switch action {
	case "add-filter":
		filter.HandleSearchAddFilter(w, r)
	case "remove-filter":
		filter.HandleSearchRemoveFilter(w, r)
	}

	filter.HandleSearchGetExisting(w, r)
}

func viewWorkHandler(w http.ResponseWriter, r *http.Request) {
	work := getWorkFromRequest(r)
	templ := template.Must(template.New("work.html").ParseFiles("./resources/work.html"))
	templ.Execute(w, work)
}

func viewSeriesHandler(w http.ResponseWriter, r *http.Request) {
	series := getSeriesFromRequest(r)
	templ := template.Must(template.New("series.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
	}).ParseFiles("./resources/series.html"))

	templ.Execute(w, series)
}

func createNewWorkHandler(w http.ResponseWriter, r *http.Request) {
	newWork := getNewWork()
	http.Redirect(w, r, "/compose"+newWork.GetNumberingUrlString(), http.StatusMovedPermanently)
}

func createWorkHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		createWorkGetHandler(w, r)
	case http.MethodPost:
		createWorkPostHandler(w, r)
	}
}

func createWorkGetHandler(w http.ResponseWriter, r *http.Request) {
	work := getWorkFromRequest(r)
	work.HandleCanvasGetExisting(w, r)
}

func createWorkPostHandler(w http.ResponseWriter, r *http.Request) {
	action, _ := extractActionAndValueFromRequest(r)
	work := getWorkFromRequest(r)
	work.HandleCanvasExistingData(w, r)

	switch action {
	case "view-work":
		work.HandleCanvasGetView(w, r)
		return
	case "edit-work":
		work.HandleCanvasGetEditable(w, r)
		return
	case "add-text":
		work.HandleCanvasAddText(w, r)
	case "add-media":
		work.HandleCanvasAddMedia(w, r)
	case "add-caption":
		work.HandleCanvasAddCaption(w, r)
	case "add-horizontal":
		work.HandleCanvasAddHorizontal(w, r)
	case "move-up":
		work.HandleCanvasContentUp(w, r)
	case "move-down":
		work.HandleCanvasContentDown(w, r)
	case "horizontal-right":
		work.HandleCanvasHorizontalRight(w, r)
	case "horizontal-left":
		work.HandleCanvasHorizontalLeft(w, r)
	case "upload":
		work.HandleCanvasMediaUpload(w, r)
	case "delete-content":
		work.HandleCanvasDeleteContent(w, r)
	case "delete-row":
		work.HandleCanvasDeleteRow(w, r)
	case "delete-caption":
		work.HandleCanvasDeleteCaption(w, r)
	case "delete-horizontal":
		work.HandleCanvasDeleteHorizontal(w, r)
	}

	work.HandleCanvasGetExisting(w, r)
}

// TODO:
func organizeWorksHandler(w http.ResponseWriter, r *http.Request) {
	panic("organize handler not implemented")
}

func main() {
	var _ Mediable = (*EmptyMedia)(nil)

	mux := http.NewServeMux()

	// TODO: using mux to handle means .jpg, . will get replaced with / from subdomainPeriodReplacer
	for _, mt := range AllMediaTypes {
		mediaDir := "." + mt.mediaDir() // ./{resourcefolder}/{media}/{contentType (plural)}/
		urlPrefix := mt.urlPrefix()     // /{contentType (singular)}/{with}/{name}/{fileName}/
		log.Printf("mediaDir: %s, urlPrefix: %s", mediaDir, urlPrefix)

		contentTypeFileServer := http.FileServer(http.Dir(mediaDir))
		mux.Handle(urlPrefix, http.StripPrefix(urlPrefix, contentTypeFileServer))
	}

	cssFileServer := http.FileServer(http.Dir("./resources/static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", cssFileServer))

	mux.HandleFunc("/{$}", homeHandler)
	mux.HandleFunc("/works", viewWorksHandler)

	mux.HandleFunc("/search/works", searchWorksHandler)

	mux.HandleFunc("/work/{numbering}", viewWorkHandler)
	mux.HandleFunc("/series/{numbering}", viewSeriesHandler)

	mux.HandleFunc("/compose/work", createNewWorkHandler)
	mux.HandleFunc("/compose/work/{numbering}", createWorkHandler)

	mux.HandleFunc("/organize/works/by/{name}", organizeWorksHandler)

	log.Fatal(http.ListenAndServe(":8080", subdomainPeriodReplacer(mux)))
}
