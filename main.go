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

type TemplateData[T any] struct {
	Environment *HtmlEnvironment
	Data        T
}

type Numbering struct {
	Century, Year, Day int
	Serial             int
	Random             int
}

type Numberable interface {
	GetHrefString() string
	GetNumberingUrlString() string
	GetNumbering() Numbering
}

type Saveable interface {
	Save()
}

type Listable interface {
	Numberable
	Saveable
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

type Emptiable interface {
	Remove(id uuid.UUID)
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

type Templatable interface {
	ToHtml() template.HTML
	GetHtmlTemplateString() string
	GetTemplateName() string
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

type Editable interface {
	ToEditableHtml() template.HTML
	GetEditableHtmlTemplateString() string
}

type Contentable interface {
	Templatable
	Editable
	Saveable
	GetId() uuid.UUID
	GetWork() *Work
	GetIndex() int
	SetIndex(idx int)
}

type Text struct {
	Id    uuid.UUID
	Work  *Work
	Text  string
	Index int
}

type Caption struct {
	Text string
}

type Captionable interface {
	GetCaptionHtml() template.HTML
	GetEditableCaptionHtml() template.HTML
}

type Mediable interface {
	Contentable
	Captionable
	GetSources() []string
	GetCaption() *Caption
	SetSources(sources []string)
	SetCaption(caption *Caption)
}

type EmptyMedia struct {
	Id    uuid.UUID
	Work  *Work
	Index int
}

type Sound struct {
	Id      uuid.UUID
	Work    *Work
	Sources []string
	Index   int
	Caption *Caption
}

type Video struct {
	Id      uuid.UUID
	Work    *Work
	Sources []string
	Index   int
	Caption *Caption
}

type Image struct {
	Id      uuid.UUID
	Work    *Work
	Sources []string
	Index   int
	Caption *Caption
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

func (s *Series) GetHrefString() string {
	return "/series" + s.GetNumberingUrlString()
}

// TODO: add random
func (s *Series) GetNumberingUrlString() string {
	centuryString := fmt.Sprintf("%d%s", s.Numbering.Century, getEnglishNumberSuffix(s.Numbering.Century))
	yearString := fmt.Sprintf("%d%s", s.Numbering.Year, getEnglishNumberSuffix(s.Numbering.Year))
	dayString := fmt.Sprintf("%d%s", s.Numbering.Day, getEnglishNumberSuffix(s.Numbering.Day))

	return fmt.Sprintf("/%d/of/%s/century/%s/year/%s/day", s.Numbering.Serial, centuryString, yearString, dayString)
}

func (s *Series) GetNumbering() Numbering {
	return s.Numbering
}

// TODO:
func (s *Series) Save() {
	getMockDb().SaveSeries(s)
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

	return w.Length
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

func (w *Work) GetHrefString() string {
	return "/work" + w.GetNumberingUrlString()
}

// TODO: add random
func (w *Work) GetNumberingUrlString() string {
	centuryString := fmt.Sprintf("%d%s", w.Numbering.Century, getEnglishNumberSuffix(w.Numbering.Century))
	yearString := fmt.Sprintf("%d%s", w.Numbering.Year, getEnglishNumberSuffix(w.Numbering.Year))
	dayString := fmt.Sprintf("%d%s", w.Numbering.Day, getEnglishNumberSuffix(w.Numbering.Day))

	return fmt.Sprintf("/%d/of/%s/century/%s/year/%s/day", w.Numbering.Serial, centuryString, yearString, dayString)
}

func (w *Work) GetNumbering() Numbering {
	return w.Numbering
}

// TODO:
func (w *Work) Save() {
	getMockDb().SaveWork(w)
}

// Assume Contents is always sorted by index with no gaps
// TODO: must reorder every time?
func (w *Work) Remove(id uuid.UUID) {
	content := getContent(id)
	idx := content.GetIndex()

	w.Contents = append(w.Contents[:idx], w.Contents[idx+1:]...)

	for i := idx; i < len(w.Contents); i++ {
		w.Contents[i].SetIndex(i)
	}

	w.Save()
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

// TODO:
func (f *Filter) Save() {
	getMockDb().SaveFilter(f)
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
func (t *Tag) GetHtmlTemplateString() string {
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

	tmpl := template.Must(template.New(t.Name).Parse(t.GetHtmlTemplateString()))

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

func (t *Text) GetIndex() int {
	return t.Index
}

func (t *Text) SetIndex(idx int) {
	t.Index = idx
}

func (t *Text) ToHtml() template.HTML {
	return contentableToHtmlTemplate(t, false)
}

func (t *Text) GetHtmlTemplateString() string {
	return textTemplate
}

func (t *Text) GetTemplateName() string {
	return "text"
}

func (t *Text) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(t, true)
}

func (t *Text) GetEditableHtmlTemplateString() string {
	return editableTextTemplate
}

// TODO:
func (t *Text) Save() {
	getMockDb().SaveText(t)
}

func (m *EmptyMedia) GetId() uuid.UUID {
	return m.Id
}

func (m *EmptyMedia) GetWork() *Work {
	return m.Work
}

func (m *EmptyMedia) GetSources() []string {
	return []string{""}
}

func (m *EmptyMedia) GetIndex() int {
	return m.Index
}

func (m *EmptyMedia) GetCaption() *Caption {
	return nil
}

func (m *EmptyMedia) SetIndex(idx int) {
	m.Index = idx
}

func (m *EmptyMedia) SetSources(sources []string) {}

func (m *EmptyMedia) SetCaption(caption *Caption) {}

func (m *EmptyMedia) ToHtml() template.HTML {
	return contentableToHtmlTemplate(m, false)
}

func (m *EmptyMedia) GetHtmlTemplateString() string {
	return ""
}

func (m *EmptyMedia) GetTemplateName() string {
	return "empty"
}

func (m *EmptyMedia) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(m, true)
}

func (m *EmptyMedia) GetEditableHtmlTemplateString() string {
	return editableMediaTemplate
}

func (m *EmptyMedia) GetCaptionHtml() template.HTML {
	return mediaToCaptionHtml(m, false)
}

func (m *EmptyMedia) GetEditableCaptionHtml() template.HTML {
	return mediaToCaptionHtml(m, true)
}

// TODO:
func (m *EmptyMedia) Save() {
	getMockDb().SaveEmptyMedia(m)
}

func (s *Sound) GetId() uuid.UUID {
	return s.Id
}

func (s *Sound) GetWork() *Work {
	return s.Work
}

func (s *Sound) GetSources() []string {
	return s.Sources
}

func (s *Sound) GetIndex() int {
	return s.Index
}

func (s *Sound) GetCaption() *Caption {
	return s.Caption
}

func (s *Sound) SetIndex(idx int) {
	s.Index = idx
}

func (s *Sound) SetSources(sources []string) {
	s.Sources = sources
}

func (s *Sound) SetCaption(caption *Caption) {
	s.Caption = caption
}

func (s *Sound) ToHtml() template.HTML {
	return contentableToHtmlTemplate(s, false)
}

func (s *Sound) GetHtmlTemplateString() string {
	return soundTemplate
}

func (s *Sound) GetTemplateName() string {
	return "sound"
}

func (s *Sound) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(s, true)
}

func (s *Sound) GetEditableHtmlTemplateString() string {
	return editableSoundTemplate
}

func (s *Sound) GetCaptionHtml() template.HTML {
	return mediaToCaptionHtml(s, false)
}

func (s *Sound) GetEditableCaptionHtml() template.HTML {
	return mediaToCaptionHtml(s, true)
}

// TODO:
func (s *Sound) Save() {
	getMockDb().SaveSound(s)
}

func (v *Video) GetId() uuid.UUID {
	return v.Id
}

func (v *Video) GetWork() *Work {
	return v.Work
}

func (v *Video) GetSources() []string {
	return v.Sources
}

func (v *Video) GetIndex() int {
	return v.Index
}

func (v *Video) GetCaption() *Caption {
	return v.Caption
}

func (v *Video) SetIndex(idx int) {
	v.Index = idx
}

func (v *Video) SetSources(sources []string) {
	v.Sources = sources
}

func (v *Video) SetCaption(caption *Caption) {
	v.Caption = caption
}

func (v *Video) ToHtml() template.HTML {
	return contentableToHtmlTemplate(v, false)
}

func (v *Video) GetHtmlTemplateString() string {
	return videoTemplate
}

func (v *Video) GetTemplateName() string {
	return "video"
}

func (v *Video) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(v, true)
}

func (v *Video) GetEditableHtmlTemplateString() string {
	return editableVideoTemplate
}

func (v *Video) GetCaptionHtml() template.HTML {
	return mediaToCaptionHtml(v, false)
}

func (v *Video) GetEditableCaptionHtml() template.HTML {
	return mediaToCaptionHtml(v, true)
}

// TODO:
func (v *Video) Save() {
	getMockDb().SaveVideo(v)
}

func (i *Image) GetId() uuid.UUID {
	return i.Id
}

func (i *Image) GetWork() *Work {
	return i.Work
}

func (i *Image) GetSources() []string {
	return i.Sources
}

func (i *Image) GetIndex() int {
	return i.Index
}

func (i *Image) GetCaption() *Caption {
	return i.Caption
}

func (i *Image) SetIndex(idx int) {
	i.Index = idx
}

func (i *Image) SetCaption(caption *Caption) {
	i.Caption = caption
}

func (i *Image) SetSources(sources []string) {
	i.Sources = sources
}

func (i *Image) ToHtml() template.HTML {
	return contentableToHtmlTemplate(i, false)
}

func (i *Image) GetHtmlTemplateString() string {
	return imageTemplate
}

func (i *Image) GetTemplateName() string {
	return "image"
}

func (i *Image) ToEditableHtml() template.HTML {
	return contentableToHtmlTemplate(i, true)
}

func (i *Image) GetEditableHtmlTemplateString() string {
	return editableImageTemplate
}

func (i *Image) GetCaptionHtml() template.HTML {
	return mediaToCaptionHtml(i, false)
}

func (i *Image) GetEditableCaptionHtml() template.HTML {
	return mediaToCaptionHtml(i, true)
}

// TODO:
func (i *Image) Save() {
	getMockDb().SaveImage(i)
}

func contentableToHtmlTemplate(c Contentable, editable bool) template.HTML {
	var tmplString string
	if editable {
		tmplString = c.GetEditableHtmlTemplateString()
	} else {
		tmplString = c.GetHtmlTemplateString()
	}

	tmpl := template.Must(template.New(c.GetTemplateName()).Parse(tmplString))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", c.GetTemplateName()))
	}

	return template.HTML(buf.String())
}

func mediaToCaptionHtml(m Mediable, editable bool) template.HTML {
	var tmplString string
	if editable {
		tmplString = editableCaptionTemplate
	} else {
		tmplString = captionTemplate
	}

	tmpl := template.Must(template.New(m.GetTemplateName()).Parse(tmplString))

	type MediableCaption struct {
		Media   Mediable
		Caption *Caption
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, &MediableCaption{
		Media:   m,
		Caption: m.GetCaption(),
	}); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", m.GetTemplateName()))
	}

	return template.HTML(buf.String())
}

// MediableCaption is passed in
const captionTemplate = `{{ with .Media.GetCaption }}
<figcaption>
	{{ .Text }}
</figcaption>
{{ end }}`

const textTemplate = `<p>{{ .Text }}</p>`

const soundTemplate = `<figure>
	<audio controls>
		{{ range .GetSources }}
		<source src={{ . }}>
		{{ end }}
		Your browser doesn't support this audio.
	</audio>
	
	{{ .GetCaptionHtml }}
</figure>`

const videoTemplate = `<figure>
	<video controls>
		{{ range .GetSources }}
		<source src={{ . }}>
		{{ end }}
		Your browser doesn't support this video.
	</video>
	
	{{ .GetCaptionHtml }}
</figure>`

const imageTemplate = `<figure>
	<img src={{ index .GetSources 0 }} alt="Image">

	{{ .GetCaptionHtml }}
</figure>`

// MediableCaption is passed in
const editableCaptionTemplate = `{{ if .Caption }}
<fieldset>
	<legend>
		<button type="submit" name="action" value="delete-caption:{{ .Media.GetId }}">
			✕
		</button>

		caption
	</legend>

	<textarea name="caption[{{ .Media.GetId }}]">
		{{ .Caption.Text }}
	</textarea>
</fieldset>
{{ else }}
<br>
<button type="submit" name="action" value="add-caption:{{ .Media.GetId }}">
	&
</button>

caption
{{ end }}`

const editableTextTemplate = `<textarea name=text[{{ .GetId }}]>{{index .Text }}</textarea>`

const editableMediaTemplate = `<fieldset>
	<legend>	
		<button type="submit" name="action" value="upload:{{ .GetId }}">
			Upload
		</button>

		sound, video or image
	</legend>

	<input type="file" name="upload" />
</fieldset>`

const editableSoundTemplate = `<figure>
	<audio controls>
		{{ range .GetSources }}
		<source src={{ . }}>
		{{ end }}
		Your browser doesn't support this audio.
	</audio>

	{{ .GetEditableCaptionHtml }}
</figure>`

const editableVideoTemplate = `<figure>
	<video controls>
		{{ range .GetSources }}
		<source src={{ . }}>
		{{ end }}
		Your browser doesn't support this video.
	</video>

	{{ .GetEditableCaptionHtml }}
</figure>`

const editableImageTemplate = `<figure>
	<img src={{ index .GetSources 0 }} alt="Image">
	
	{{ .GetEditableCaptionHtml }}
</figure>`

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
func getAllListings() []Listable {
	return slices.Collect(
		maps.Values(getMockDb().Listings),
	)
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
func getFilter(numbering Numbering) *Filter {
	return getMockDb().GetFilter(numbering)
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

func storeUploadedFormFile(r *http.Request) (Mediable, error) {
	file, header, err := r.FormFile("upload")
	if err != nil {
		return nil, err
	} else {
		mt, validCt := getFileMediaType(file)
		if !validCt {
			return nil, errors.New("invalid file type uploaded")
		}

		defer file.Close()

		localPath := mt.toLocalPath(header.Filename)

		dst, err := os.Create(localPath)
		if err != nil {
			log.Println(err)
			return nil, err
		} else {
			defer dst.Close()
			io.Copy(dst, file)
		}

		return mt.CreateNew([]string{mt.toSourceUrl(header.Filename)}), nil
	}
}

func handleContentUpload(r *http.Request, work *Work, media Mediable) error {
	idx := media.GetIndex()
	media, err := storeUploadedFormFile(r)
	if err != nil {
		log.Println(err)
		return err
	}

	work.Contents[idx] = media
	work.Save()

	return nil
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
	newFilter.Save()

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

	filter.Save()

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
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
	newWork := getMockEmptyWork()
	newWork.Save()
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

	// Only when not all contents have been deleted
	if len(work.Contents) != 0 {
		for k, v := range r.Form {
			log.Println(k)
			if strings.HasPrefix(k, "text[") {
				id, err := uuid.Parse(k[5:41])
				if err != nil {
					panic("unexpected error parsing id for text")
				}

				text := getText(id)
				text.Text = v[0]

				log.Printf("text with id %v setting text to %s", text.GetId(), v)

				work.Contents[text.GetIndex()] = text
				work.Save()
			}

			if strings.HasPrefix(k, "caption[") {
				id, err := uuid.Parse(k[8:44])
				if err != nil {
					panic("unexpected error parsing id for content")
				}

				media := getMedia(id)
				media.SetCaption(&Caption{v[0]})
				work.Contents[media.GetIndex()] = media

				work.Save()
			}
		}
	}

	title := r.Form["title-text"]
	work.Title = title[0]

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

		log.Printf("id requested to be deleted %s", actionValue)

		work.Remove(id)

	case "add-media":
		media := MediaEmptyType.CreateNew(nil)
		log.Printf("contents nil? %t", work.Contents == nil)
		media.SetIndex(len(work.Contents))

		work.Contents = append(work.Contents, media)
		work.Save()

	case "upload":
		id, err := uuid.Parse(actionValue)
		if err != nil {
			panic("unexpected error parsing uuid for content")
		}

		media := getMedia(id)
		err = handleContentUpload(r, work, media)
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
		newText := ContentTextType.CreateNew()
		newText.SetIndex(len(work.Contents))
		work.Contents = append(work.Contents, newText)
		work.Save()

	case "add-caption":
		id, err := uuid.Parse(actionValue)
		if err != nil {
			panic("unexpected error parsing uuid for content")
		}

		content := getMedia(id)
		text := r.FormValue("added-caption")
		content.SetCaption(&Caption{text})
		work.Contents[content.GetIndex()] = content
		work.Save()

	case "delete-caption":
		id, err := uuid.Parse(actionValue)
		if err != nil {
			panic("unexpected error parsing uuid for content")
		}

		content := getMedia(id)
		content.SetCaption(nil)
		work.Contents[content.GetIndex()] = content
		work.Save()
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

	for _, mt := range AllMediaTypes {
		mediaDir := "." + mt.mediaDir() // ./{resourcefolder}/{media}/{contentType (plural)}/
		urlPrefix := mt.urlPrefix()     // /{contentType (singular)}/{with}/{name}/{fileName}/

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
