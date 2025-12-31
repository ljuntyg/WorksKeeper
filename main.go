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
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

type HtmlEnvironment struct {
	AllTags []*TagValues
}

type TemplateData[T, V any] struct {
	Environment *HtmlEnvironment
	Data        T
}

type Numberable interface {
	GetNumberingUrlString() string
}

type Numbering struct {
	Century, Year, Day int
	Serial             int
	Random             int
}

type Listable interface {
	GetId() uuid.UUID
	GetNumbering() Numbering
	GetParentListing() Listable
	GetTitle() string
	GetTimeRequiredMinutes() int
	GetUrl() string
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

type Work struct {
	Id           uuid.UUID
	Numbering    Numbering
	ParentSeries *Series
	Title        string
	Length       int
	IsPublic     bool
	Contents     []Contentable
	Tags         []*Tag
}

type FilterGroup struct {
	Name string
	Tags []*Tag
}

type Filter struct {
	Id               uuid.UUID
	Numbering        Numbering
	FilterGroups     []*FilterGroup
	FilterVisibility string
	PreviousTag      string // Name of previous Tag added to the filter, or "" if a Tag was removed or none have been added
}

type Tag struct {
	Id           uuid.UUID
	Name         string
	Value        string
	ModeModifier string
	FilterMode   string
}

type TagValues struct {
	Name                  string
	PossibleValues        []string
	PossibleModeModifiers []string
	PossibleFilterModes   []string
}

type Contentable interface {
	GetId() uuid.UUID
	GetWork() *Work
	GetSources() []string
	GetType() ContentType
	ToHtml() template.HTML
	GetIndex() int
	GetCaption() string
	GetHtmlTemplateString() string
	SetIndex(idx int)
	SetSources(sources []string)
}

type Text struct {
	Id    uuid.UUID
	Work  *Work
	Text  string
	Index int
}

type Sound struct {
	Id         uuid.UUID
	Work       *Work
	SoundPaths []string
	Index      int
	Caption    string
}

type Video struct {
	Id         uuid.UUID
	Work       *Work
	VideoPaths []string
	Index      int
	Caption    string
}

type Image struct {
	Id         uuid.UUID
	Work       *Work
	ImagePaths []string
	Index      int
	Caption    string
}

type ContentType int

const (
	ContentTextType ContentType = iota
	ContentSoundType
	ContentVideoType
	ContentImageType
)

// ----------------------------
//
//	LISTINGS LISTINGS LISTINGS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (s *Series) GetId() uuid.UUID {
	return s.Id
}

func (s *Series) GetNumbering() Numbering {
	return s.Numbering
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

// TODO:
func (s *Series) GetUrl() string {
	log.Println("calling MOCK GetUrl()")

	return "/series/" + s.GetNumberingUrlString()
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

// TODO: add random
func (s *Series) GetNumberingUrlString() string {
	centuryString := fmt.Sprintf("%d%s", s.Numbering.Century, getEnglishNumberSuffix(s.Numbering.Century))
	yearString := fmt.Sprintf("%d%s", s.Numbering.Year, getEnglishNumberSuffix(s.Numbering.Year))
	dayString := fmt.Sprintf("%d%s", s.Numbering.Day, getEnglishNumberSuffix(s.Numbering.Day))

	return fmt.Sprintf("/%d/of/%s/century/%s/year/%s/day", s.Numbering.Serial, centuryString, yearString, dayString)
}

func (w *Work) GetId() uuid.UUID {
	return w.Id
}

func (w *Work) GetNumbering() Numbering {
	return w.Numbering
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

	return w.Length
}

// TODO:
func (w *Work) GetUrl() string {
	log.Println("calling MOCK GetUrl()")

	return "/work/" + w.GetNumberingUrlString()
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

// TODO: add random
func (w *Work) GetNumberingUrlString() string {
	centuryString := fmt.Sprintf("%d%s", w.Numbering.Century, getEnglishNumberSuffix(w.Numbering.Century))
	yearString := fmt.Sprintf("%d%s", w.Numbering.Year, getEnglishNumberSuffix(w.Numbering.Year))
	dayString := fmt.Sprintf("%d%s", w.Numbering.Day, getEnglishNumberSuffix(w.Numbering.Day))

	return fmt.Sprintf("/%d/of/%s/century/%s/year/%s/day", w.Numbering.Serial, centuryString, yearString, dayString)
}

// ----------------------------------------
//
//	TAGS/FILTERS TAGS/FILTERS TAGS/FILTERS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (f *Filter) GetNumberingUrlString() string {
	centuryString := fmt.Sprintf("%d%s", f.Numbering.Century, getEnglishNumberSuffix(f.Numbering.Century))
	yearString := fmt.Sprintf("%d%s", f.Numbering.Year, getEnglishNumberSuffix(f.Numbering.Year))
	dayString := fmt.Sprintf("%d%s", f.Numbering.Day, getEnglishNumberSuffix(f.Numbering.Day))

	return fmt.Sprintf("/%d/of/%s/century/%s/year/%s/day", f.Numbering.Serial, centuryString, yearString, dayString)
}

func (t *Tag) ToHtml() template.HTML {
	return t.filterTagToHtmlTemplate()
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

// TODO:
func (t *Tag) getHtmlTemplateString() string {
	switch t.Name {
	case "author":
		return tagDatalistTemplate
	case "date":
		return tagDateTemplate
	case "length":
		return tagLengthTemplate
	case "language":
		return tagDatalistTemplate
	case "media":
		return tagMediaTemplate
	default:
		panic("getHtmlTemplateString() not fully implemented")
		// TODO
		/* return toHtmlTemplate(tagTextTemplate, "default", nil) */
	}
}

func (t *Tag) filterTagToHtmlTemplate() template.HTML {
	log.Printf("tag %s to html template, possible values: %v", t.Name, t.GetPossibleValues())

	tmpl := template.Must(template.New(t.Name).Parse(t.getHtmlTemplateString()))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, t); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", t.Name))
	}

	return template.HTML(buf.String())
}

const tagDatalistTemplate = `<datalist id="tag-values-{{ .Name }}">
	{{ range .GetPossibleValues }}
	<option value="{{ . }}"></option>
	{{ end }}
</datalist>

<select name="mode-modifier">
	{{ range .GetPossibleModeModifiers}}
	<option value="{{ . }}"
		{{ if eq . $.ModeModifier }}selected{{ end }}>
		{{ . }}
	</option>
	{{ end }}
</select>

<input 
	list="tag-values-{{ .Name }}" 
	name="added-filters.value" 
	value="{{ .Value }}"
/>

<input type="hidden" name="added-filters.tag" value="{{ .Name }}" />
<input type="hidden" name="filter-mode" value="" />
<!-- TODO: add remove button, search button -->`

const tagDateTemplate = `
<select name="mode-modifier">
	{{ range .GetPossibleModeModifiers}}
	<option value="{{ . }}"
		{{ if eq . $.ModeModifier }}selected{{ end }}>
		{{ . }}
	</option>
	{{ end }}
</select>

<select name="filter-mode">
	{{ range .GetPossibleFilterModes }}
	<option value="{{ . }}"
		{{ if eq . $.FilterMode }}selected{{ end }}>
		{{ . }}
	</option>
	{{ end }}
</select>

<input 
	type="date"  
	name="added-filters.value" 
	value="{{ .Value }}"
/>

<input type="hidden" name="added-filters.tag" value="{{ .Name }}" />
<!-- TODO: add remove button, search button -->`

const tagLengthTemplate = `
<select name="mode-modifier">
	{{ range .GetPossibleModeModifiers}}
	<option value="{{ . }}"
		{{ if eq . $.FilterMode }}selected{{ end }}>
		{{ . }}
	</option>
	{{ end }}
</select>

<select name="filter-mode">
	{{ range .GetPossibleFilterModes }}
	<option value="{{ . }}"
		{{ if eq . $.FilterMode }}selected{{ end }}>
		{{ . }}
	</option>
	{{ end }}
</select>

<!-- assumes .PossibleValues for a length tag is a slice of ordered values (with min and max) -->
<datalist id="length-datalist">
	{{ range .GetPossibleValues }}
	<option value="{{ . }}"></option>
	{{ end }}
</datalist>

<input 
	type="range" 
	list="length-datalist" 
	name="added-filters.value" 
	value="{{ .Value }}"	
/>

<input type="hidden" name="added-filters.tag" value="{{ .Name }}" />
<!-- TODO: add remove button, search button -->`

const tagMediaTemplate = `<select name="mode-modifier">
	{{ range .GetPossibleModeModifiers}}
	<option value="{{ . }}"
		{{ if eq . $.FilterMode }}selected{{ end }}>
		{{ . }}
	</option>
	{{ end }}
</select>

<select name="added-filters.value">
	{{ range .GetPossibleValues }}
	<option value="{{ . }}"
		{{ if eq . $.Value }}selected{{ end }}>
		{{ . }}
	</option>
	{{ end }}
</select>

<input type="hidden" name="added-filters.tag" value="{{ .Name }}" />
<input type="hidden" name="filter-mode" value="" />
<!-- TODO: add remove button, search button -->`

// ----------------------------
//
//	CONTENTS CONTENTS CONTENTS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (t *Text) GetId() uuid.UUID {
	return t.Id
}

func (t *Text) GetWork() *Work {
	return t.Work
}

func (t *Text) GetSources() []string {
	return []string{t.Text}
}

func (t *Text) GetType() ContentType {
	return ContentTextType
}

func (t *Text) ToHtml() template.HTML {
	return contentToHtmlTemplate(t)
}

func (t *Text) GetIndex() int {
	return t.Index
}

func (t *Text) GetCaption() string {
	return ""
}

func (t *Text) GetHtmlTemplateString() string {
	return textTemplate
}

func (t *Text) SetIndex(idx int) {
	t.Index = idx
}

func (t *Text) SetSources(sources []string) {
	t.Text = sources[0]
}

func (s *Sound) GetId() uuid.UUID {
	return s.Id
}

func (s *Sound) GetWork() *Work {
	return s.Work
}

func (s *Sound) GetSources() []string {
	return s.SoundPaths
}

func (s *Sound) GetType() ContentType {
	return ContentSoundType
}

func (s *Sound) ToHtml() template.HTML {
	return contentToHtmlTemplate(s)
}

func (s *Sound) GetIndex() int {
	return s.Index
}

func (s *Sound) GetCaption() string {
	return s.Caption
}

func (s *Sound) GetHtmlTemplateString() string {
	return soundTemplate
}

func (s *Sound) SetIndex(idx int) {
	s.Index = idx
}

func (s *Sound) SetSources(sources []string) {
	s.SoundPaths = sources
}

func (v *Video) GetWork() *Work {
	return v.Work
}

func (v *Video) GetId() uuid.UUID {
	return v.Id
}

func (v *Video) GetSources() []string {
	return v.VideoPaths
}

func (v *Video) GetType() ContentType {
	return ContentVideoType
}

func (v *Video) ToHtml() template.HTML {
	return contentToHtmlTemplate(v)
}

func (v *Video) GetIndex() int {
	return v.Index
}

func (v *Video) GetCaption() string {
	return v.Caption
}

func (v *Video) GetHtmlTemplateString() string {
	return videoTemplate
}

func (v *Video) SetIndex(idx int) {
	v.Index = idx
}

func (v *Video) SetSources(sources []string) {
	v.VideoPaths = sources
}

func (i *Image) GetId() uuid.UUID {
	return i.Id
}

func (i *Image) GetWork() *Work {
	return i.Work
}

func (i *Image) GetSources() []string {
	return i.ImagePaths
}

func (i *Image) GetType() ContentType {
	return ContentImageType
}

func (i *Image) ToHtml() template.HTML {
	return contentToHtmlTemplate(i)
}

func (i *Image) GetIndex() int {
	return i.Index
}

func (i *Image) GetCaption() string {
	return i.Caption
}

func (i *Image) GetHtmlTemplateString() string {
	return imageTemplate
}

func (i *Image) SetIndex(idx int) {
	i.Index = idx
}

func (i *Image) SetSources(sources []string) {
	i.ImagePaths = sources
}

func contentToHtmlTemplate(c Contentable) template.HTML {
	tmpl := template.Must(template.New(c.GetType().SingularString()).Parse(c.GetHtmlTemplateString()))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", c.GetType().SingularString()))
	}

	return template.HTML(buf.String())
}

const textTemplate = `{{index .GetSources 0}}`

const soundTemplate = `<figure>
	<audio controls>
		{{range .GetSources}}
		<source src={{.}}>
		{{end}}
		Your browser doesn't support this audio.
	</audio>
	<figcaption>{{.GetCaption}}</figcaption>
</figure>`

const videoTemplate = `<figure>
	<video controls>
		{{range .GetSources}}
		<source src={{.}}>
		{{end}}
		Your browser doesn't support this video.
	</video>
	<figcaption>{{.GetCaption}}</figcaption>
</figure>`

const imageTemplate = `<figure>
	<img src={{index .GetSources 0}} alt="Image">
	<figcaption>{{.GetCaption}}</figcaption>
</figure>`

// --------------------------------------
//
//	CONTENTTYPE CONTENTTYPE CONTENTTYPE
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/

// Careful changing this, html templates might use this to check type
func (ct ContentType) String() string {
	switch ct {
	case ContentTextType:
		return "texts"
	case ContentSoundType:
		return "sounds"
	case ContentVideoType:
		return "videos"
	case ContentImageType:
		return "images"
	default:
		return "unknown"
	}
}

// Careful changing this, html templates might use this to check type
func (ct ContentType) SingularString() string {
	str := ct.String()
	if str[len(str)-1:] == "s" {
		return str[:len(str)-1]
	} else {
		return str
	}
}

// This should NOT be used for content source links
func (ct ContentType) mediaDir() string {
	if ct.String() == "unknown" {
		panic("Unexpected ContentType when reading mediaDir()")
	}

	return "/resources/media/" + ct.String() + "/"
}

// This should be used for content source links
func (ct ContentType) urlPrefix() string {
	if ct.String() == "unknown" {
		panic("Unexpected ContentType when reading urlPrefix()")
	}

	return "/" + ct.SingularString() + "/with/name/"
}

func (ct ContentType) toLocalPath(fileName string) string {
	return "." + ct.mediaDir() + fileName
}

func (ct ContentType) toSourceUrl(fileName string) string {
	return getBaseUrl() + ct.urlPrefix() + fileName
}

func (ct ContentType) createNew(sources []string) Contentable {
	switch ct {
	case ContentTextType:
		return &Text{
			Text: sources[0],
			Id:   uuid.New(),
		}
	case ContentSoundType:
		return &Sound{
			SoundPaths: sources,
			Id:         uuid.New(),
		}
	case ContentImageType:
		return &Image{
			ImagePaths: sources,
			Id:         uuid.New(),
		}
	case ContentVideoType:
		return &Video{
			VideoPaths: sources,
			Id:         uuid.New(),
		}
	default:
		panic("unknown ContentType when creating empty")
	}
}

var AllContentTypes = []ContentType{
	ContentTextType,
	ContentSoundType,
	ContentVideoType,
	ContentImageType,
}

var MimeToContentType = map[string]ContentType{
	"image/apng":      ContentImageType,
	"image/avif":      ContentImageType,
	"image/bmp":       ContentImageType,
	"image/gif":       ContentImageType,
	"image/jpeg":      ContentImageType,
	"image/png":       ContentImageType,
	"image/svg+xml":   ContentImageType,
	"image/tiff":      ContentImageType,
	"image/webp":      ContentImageType,
	"audio/aac":       ContentSoundType,
	"audio/midi":      ContentSoundType,
	"audio/x-midi":    ContentSoundType,
	"audio/mpeg":      ContentSoundType,
	"audio/ogg":       ContentSoundType,
	"audio/wav":       ContentSoundType,
	"audio/webm":      ContentSoundType,
	"audio/3gpp":      ContentSoundType,
	"audio/3gpp2":     ContentSoundType,
	"application/ogg": ContentSoundType, // TODO: ? some .ogg files
	"video/mp4":       ContentVideoType,
	"video/mpeg":      ContentVideoType,
	"video/ogg":       ContentVideoType,
	"video/webm":      ContentVideoType,
	"video/x-msvideo": ContentVideoType,
	"video/mp2t":      ContentVideoType,
	"video/3gpp":      ContentVideoType,
	"video/3gpp2":     ContentVideoType,
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
func saveWork(w *Work) {
	getMockDb().SaveWork(w)
}

// TODO:
func saveSeries(s *Series) {
	getMockDb().SaveSeries(s)
}

// TODO:
func saveContent(c Contentable) {
	switch c.GetType() {
	case ContentTextType:
		getMockDb().SaveText(c.(*Text))
	case ContentSoundType:
		getMockDb().SaveSound(c.(*Sound))
	case ContentVideoType:
		getMockDb().SaveVideo(c.(*Video))
	case ContentImageType:
		getMockDb().SaveImage(c.(*Image))
	default:
		panic("unknown ContentType when saving content")
	}
}

func saveFilter(f *Filter) {
	getMockDb().SaveFilter(f)
}

// TODO:
func getBaseUrl() string {
	return getMockBaseUrl()
}

// TODO:
func createNewWork() *Work {
	return getMockEmptyWork()
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
	return getMockDb().Contents[id]
}

// TODO:
func getImage(id uuid.UUID) *Image {
	return getMockDb().GetImage(id)
}

// Panics if id provided is not parsable as uuid
func getUuidFromId(id string) uuid.UUID {
	log.Println("trying to parse id: " + id)

	retId, err := uuid.Parse(id)
	if err != nil {
		panic(err)
	}

	return retId
}

// TODO: improve matching mime -> ContentType
func getFileContentType(file multipart.File) (ContentType, bool) {
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	file.Seek(0, io.SeekStart)
	ct, ok := MimeToContentType[mimeType]

	if !ok {
		log.Println("mime type " + mimeType + " not found in mime map")
	}

	return ct, ok
}

func storeUploadedFormFile(r *http.Request) (Contentable, error) {
	file, header, err := r.FormFile("upload")
	if err != nil {
		return nil, err
	} else {
		ct, validCt := getFileContentType(file)
		if !validCt || ct == ContentTextType {
			return nil, errors.New("invalid file type uploaded")
		}

		defer file.Close()

		localPath := ct.toLocalPath(header.Filename)

		dst, err := os.Create(localPath)
		if err != nil {
			log.Println(err)
			return nil, err
		} else {
			defer dst.Close()
			io.Copy(dst, file)
		}

		return ct.createNew([]string{ct.toSourceUrl(header.Filename)}), nil
	}
}

func handleContentUpload(r *http.Request, work *Work) error {
	content, err := storeUploadedFormFile(r)
	if err != nil {
		log.Println(err)
		return err
	}

	content.SetIndex(len(work.Contents))
	work.Contents = append(work.Contents, content)
	saveWork(work)

	return nil
}

// TODO:
func getAllListings() []Listable {
	return slices.Collect(
		maps.Values(getMockDb().Listings),
	)
}

// Assume Contents is always sorted by index
func moveContentUp(w *Work, idx int) {
	if idx == 0 {
		return
	}

	// TODO: remove
	for i, c := range w.Contents {
		if c.GetIndex() != i {
			panic("contents slice is not sorted by index")
		}
	}

	w.Contents[idx].SetIndex(idx - 1)
	w.Contents[idx-1].SetIndex(idx)
	w.Contents[idx], w.Contents[idx-1] = w.Contents[idx-1], w.Contents[idx]
}

// Assume Contents is always sorted by index
func moveContentDown(w *Work, idx int) {
	if idx >= len(w.Contents)-1 {
		return
	}

	// TODO: remove
	for i, c := range w.Contents {
		if c.GetIndex() != i {
			panic("contents slice is not sorted by index")
		}
	}

	w.Contents[idx].SetIndex(idx + 1)
	w.Contents[idx+1].SetIndex(idx)
	w.Contents[idx], w.Contents[idx+1] = w.Contents[idx+1], w.Contents[idx]
}

// Assume Contents is always sorted by index with no gaps
// TODO:
func deleteContent(w *Work, contentId uuid.UUID) {
	idx := getContent(contentId).GetIndex()
	cs := w.Contents

	w.Contents = append(cs[:idx], cs[idx+1:]...)
	for i := idx; i < len(w.Contents); i++ {
		w.Contents[i].SetIndex(i)
	}

	delete(getMockDb().Contents, contentId)
}

func getHtmlEnvironment() *HtmlEnvironment {
	return getMockHtmlEnvironment()
}

func createTemplateData[T any, V []any](data T) TemplateData[T, V] {
	return TemplateData[T, V]{
		Environment: getHtmlEnvironment(),
		Data:        data,
	}
}

// Request must be made to path matched by pattern like
// /foo/boo/goo/{id} and id must refer to an existing work
func getWorkFromRequest(r *http.Request) *Work {
	numberingString := r.PathValue("numbering")
	numbering, err := urlStringToNumbering(numberingString)
	if err != nil {
		panic("unexpected error parsing numbering for work")
	}

	return getWork(numbering)
}

// Request must be made to path matched by pattern like
// /foo/boo/goo/{id} and id must refer to an existing series
func getSeriesFromRequest(r *http.Request) *Series {
	numberingString := r.PathValue("numbering")
	numbering, err := urlStringToNumbering(numberingString)
	if err != nil {
		panic("unexpected error parsing numbering for series")
	}

	return getSeries(numbering)
}

// TODO:
func getNewFilter() *Filter {
	return getMockFilter()
}

// TODO: for each Numberable?
func urlStringToNumbering(s string) (Numbering, error) {
	// 1/of/21st/century/25th/year/362nd/day
	parts := strings.Split(s, "/")
	if len(parts) < 7 {
		return Numbering{}, fmt.Errorf("invalid format: %q", s)
	}

	serial, err := strconv.Atoi(parts[0])
	if err != nil {
		return Numbering{}, err
	}

	century, err := strconv.Atoi(strings.TrimRight(parts[2], "stndrh"))
	if err != nil {
		return Numbering{}, err
	}

	year, err := strconv.Atoi(strings.TrimRight(parts[4], "stndrh"))
	if err != nil {
		return Numbering{}, err
	}

	day, err := strconv.Atoi(strings.TrimRight(parts[6], "stndrh"))
	if err != nil {
		return Numbering{}, err
	}

	return Numbering{
		Century: century,
		Year:    year,
		Day:     day,
		Serial:  serial,
	}, nil
}

// TODO:
func getFilterByNumbering(numbering Numbering) *Filter {
	return getMockDb().GetFilter(numbering)
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

func createFilterGroupsFromTags(tags []*Tag) []*FilterGroup {
	tagsNamed := make(map[string][]*Tag)
	groupOrder := []string{}

	for _, tag := range tags {
		if _, exists := tagsNamed[tag.Name]; !exists {
			groupOrder = append(groupOrder, tag.Name)
		}
		tagsNamed[tag.Name] = append(tagsNamed[tag.Name], tag)
	}

	filterGroups := make([]*FilterGroup, 0, len(groupOrder))
	for _, tagName := range groupOrder {
		filterGroups = append(filterGroups, &FilterGroup{
			Name: tagName,
			Tags: tagsNamed[tagName],
		})
	}

	return filterGroups
}

// ----------------------------
//
//	HANDLERS HANDLERS HANDLERS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/

func subdomainPeriodReplacer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suffixSubdomain := r.Header.Get("X-Suffix-Subdomain")
		prefixSubdomain := r.Header.Get("X-Prefix-Subdomain")

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
		viewWorksGetHandler(w, r)
	case http.MethodPost:
		searchWorksNewFilterHandler(w, r)
	default:
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	}
}

func viewWorksGetHandler(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
		"capitalize": func(s string) string {
			if s == "" {
				return s
			}

			r := []rune(s)
			r[0] = unicode.ToUpper(r[0])
			return string(r)
		},
	}).ParseFiles("./resources/home.html"))

	listings := getAllListings()
	templateData := createTemplateData(listings)

	templ.Execute(w, templateData)
}

func searchWorksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		searchWorksGetHandler(w, r)
	case http.MethodPost:
		searchWorksPostHandler(w, r)
	default:
		searchWorksNewFilterHandler(w, r)
	}
}

func searchWorksNewFilterHandler(w http.ResponseWriter, r *http.Request) {
	newFilter := getNewFilter()
	saveFilter(newFilter)

	http.Redirect(w, r, r.URL.Path+"/with/filter/"+newFilter.GetNumberingUrlString(), http.StatusSeeOther)
}

func searchWorksGetHandler(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.New("search.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
		"capitalize": func(s string) string {
			if s == "" {
				return s
			}

			r := []rune(s)
			r[0] = unicode.ToUpper(r[0])
			return string(r)
		},
		"add": func(a, b int) int { return a + b },
	}).ParseFiles("./resources/search.html"))

	listings := getAllListings()
	numbering, err := urlStringToNumbering(r.PathValue("numbering"))
	if err != nil {
		panic("unexpected error when trying to get filter")
	}

	filter := getFilterByNumbering(numbering)

	type ListablesFilter struct {
		Listables []Listable
		Filter    *Filter
	}

	templateData := createTemplateData(&ListablesFilter{
		Listables: listings,
		Filter:    filter,
	})

	templ.Execute(w, templateData)
}

// TODO: handle search
func searchWorksPostHandler(w http.ResponseWriter, r *http.Request) {
	action, actionValue := extractActionAndValueFromRequest(r)
	numbering, err := urlStringToNumbering(r.PathValue("numbering"))
	if err != nil {
		panic("unexpected error when trying to get filter")
	}

	filter := getFilterByNumbering(numbering)
	filter.FilterVisibility = r.FormValue("filter-visibility")

	switch action {
	case "add-filter":

		// Extract the tags from the request, because the values in the request might have changed
		// from the values in the retrieved filter
		existingTags := extractTagsFromRequest(r)
		newFilterTag := &Tag{
			Name: r.FormValue("filter-tag"),
		}

		existingTags = append(existingTags, newFilterTag)
		filter.FilterGroups = createFilterGroupsFromTags(existingTags)
		filter.PreviousTag = newFilterTag.Name
	case "remove-filter":
		idx, err := strconv.Atoi(actionValue)
		if err != nil {
			panic("unexpected error parsing index for filter to remove")
		}

		existingTags := extractTagsFromRequest(r)
		existingTags = append(existingTags[:idx], existingTags[idx+1:]...)
		filter.FilterGroups = createFilterGroupsFromTags(existingTags)
		filter.PreviousTag = ""
	}

	saveFilter(filter)

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

func viewWorkHandler(w http.ResponseWriter, r *http.Request) {
	work := getWorkFromRequest(r)
	templ := template.Must(template.New("work.html").Funcs(template.FuncMap{
		"isText":             func(ct ContentType) bool { return ct == ContentTextType },
		"isSound":            func(ct ContentType) bool { return ct == ContentSoundType },
		"isVideo":            func(ct ContentType) bool { return ct == ContentVideoType },
		"isImage":            func(ct ContentType) bool { return ct == ContentImageType },
		"getAllContentTypes": func() []ContentType { return AllContentTypes },
	}).ParseFiles("./resources/work.html"))

	templ.Execute(w, work)
}

func viewSeriesHandler(w http.ResponseWriter, r *http.Request) {
	series := getSeriesFromRequest(r)

	templ := template.Must(template.New("series.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
	}).ParseFiles("./resources/series.html"))

	templ.Execute(w, series)
}

func createWorkHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		createWorkGetHandler(w, r)
	case http.MethodPost:
		createWorkPostHandler(w, r)
	default:
		createNewWorkHandler(w, r)
	}
}

func createNewWorkHandler(w http.ResponseWriter, r *http.Request) {

	newWork := createNewWork()
	saveWork(newWork)

	http.Redirect(w, r, r.URL.Path+newWork.GetNumberingUrlString(), http.StatusSeeOther)
}

func createWorkPostHandler(w http.ResponseWriter, r *http.Request) {
	action, actionValue := extractActionAndValueFromRequest(r)
	work := getWorkFromRequest(r)

	err := r.ParseMultipartForm(10 << 20) // TODO: increase?
	if err != nil {
		log.Println(err)
		http.Error(w, "content uploaded is too large", http.StatusBadRequest)
		return
	}

	switch action {
	case "move-up":
		id, err := uuid.Parse(actionValue)
		if err != nil {
			panic("unexpected error parsing uuid for content")
		}

		content := getContent(id)
		moveContentUp(work, content.GetIndex())

	case "move-down":
		id, err := uuid.Parse(actionValue)
		if err != nil {
			panic("unexpected error parsing uuid for content")
		}

		content := getContent(id)
		moveContentDown(work, content.GetIndex())

	case "delete-content":
		id, err := uuid.Parse(actionValue)
		if err != nil {
			panic("unexpected error parsing uuid for content")
		}

		deleteContent(work, id)

	case "upload":
		err = handleContentUpload(r, work)
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				// TODO: handle better
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				break
			}

			http.Error(w, "error during content upload", http.StatusInternalServerError)
			return
		}

	case "add-text":
		newText := ContentTextType.createNew([]string{""})
		newText.SetIndex(len(work.Contents))
		work.Contents = append(work.Contents, newText)
		saveWork(work)
	}

	title := r.Form["title-text"]
	work.Title = title[0]

	for k, v := range r.Form {
		log.Println(k)
		if strings.HasPrefix(k, "text[") {
			id, err := uuid.Parse(k[5:41])
			if err != nil {
				panic("unexpected error parsing id for text")
			}

			text := getContent(id)
			text.SetSources(v)

			log.Printf("text with id %v setting text to %s", text.GetId(), v)

			work.Contents[text.GetIndex()] = text
			saveWork(work)
		}
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

func createWorkGetHandler(w http.ResponseWriter, r *http.Request) {
	work := getWorkFromRequest(r)
	templ := template.Must(template.ParseFiles("./resources/canvas.html"))
	templ.Execute(w, work)
}

// TODO:
func organizeWorksHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("hi")
}

func main() {
	mux := http.NewServeMux()

	for _, contentType := range AllContentTypes {
		if contentType == ContentTextType {
			continue
		}

		mediaDir := "." + contentType.mediaDir() // ./{resourcefolder}/{media}/{contentType (plural)}/
		urlPrefix := contentType.urlPrefix()     // /{contentType (singular)}/{with}/{name}/{fileName}/

		contentTypeFileServer := http.FileServer(http.Dir(mediaDir))
		mux.Handle(urlPrefix, http.StripPrefix(urlPrefix, contentTypeFileServer))
	}

	mux.HandleFunc("/{$}", homeHandler)
	mux.HandleFunc("/works", viewWorksHandler)

	mux.HandleFunc("/search/works", searchWorksNewFilterHandler)
	mux.HandleFunc("/search/works/with/filter/{numbering...}", searchWorksHandler)

	mux.HandleFunc("/work/{numbering...}", viewWorkHandler)
	mux.HandleFunc("/series/{numbering...}", viewSeriesHandler)

	mux.HandleFunc("/compose/work", createNewWorkHandler)
	mux.HandleFunc("/compose/work/{numbering...}", createWorkHandler)

	mux.HandleFunc("/organize/works/by/{name}", organizeWorksHandler)
	log.Fatal(http.ListenAndServe(":8080", subdomainPeriodReplacer(mux)))
}
