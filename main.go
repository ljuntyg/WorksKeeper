package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"maps"
	"math/rand/v2"
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
	Data2       V
}

type Listable interface {
	GetId() uuid.UUID
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
	ParentSeries *Series
	Title        string
	IsPublic     bool
	Listings     []Listable // Go only
	Tags         []*Tag     // Go only?
}

type Work struct {
	Id           uuid.UUID
	ParentSeries *Series
	Title        string
	Length       int
	IsPublic     bool
	Contents     []Contentable // Go only
	Tags         []*Tag        // Go only?
}

type Numberable interface {
	GetNumberingUrlString() string
}

type Filter struct {
	Id        uuid.UUID
	Numbering Numbering
	Tags      []*Tag
}

type Numbering struct {
	Century, Year, Day int
	Serial             int
	Random             int
}

type Tag struct {
	Name       string
	Value      string
	FilterMode string
}

type TagValues struct {
	Name        string
	Values      []string
	FilterModes []string
}

// TODO: add GetCaption() method
type Contentable interface {
	GetId() uuid.UUID
	GetWork() *Work
	GetSourceUrls() []string
	GetType() ContentType
	ToHtml() template.HTML
	GetIndex() int
	GetCaption() string
	SetIndex(idx int)
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

const textTemplate = `<p>{{index .GetSourceUrls 0}}</p>`

const soundTemplate = `<figure>
	<audio controls>
		{{range .GetSourceUrls}}
		<source src={{.}}>
		{{end}}
		Your browser doesn't support this audio.
	</audio>
	<figcaption>{{.GetCaption}}</figcaption>
</figure>`

const videoTemplate = `<figure>
	<video controls>
		{{range .GetSourceUrls}}
		<source src={{.}}>
		{{end}}
		Your browser doesn't support this video.
	</video>
	<figcaption>{{.GetCaption}}</figcaption>
</figure>`

const imageTemplate = `<figure>
	<img src={{index .GetSourceUrls 0}} alt="Image">
	<figcaption>{{.GetCaption}}</figcaption>
</figure>`

const tagDatalistTemplate = `<li>
	{{ .Name }}

	<datalist id="tag-values">
		{{ range .Values }}
		<option value="{{ . }}"></option>
		{{ end }}
	</datalist>
	
	<input list="tag-values" name="added-filters.value" id="added-filters.value" />
	<input type="hidden" name="added-filters.tag" value="{{ .Name }}" />
	<input type="hidden" name="filter-mode" value="" />
	<!-- TODO: add remove button, search button -->
</li>`

const tagDateTemplate = `<li>
	{{ .Name }}

	<select name="filter-mode">
		{{ range .FilterModes }}
		<option value="{{ . }}">
		{{ . }}
		</option>
		{{ end }}
	</select>

	<input type="date" id="added-filters.value" name="added-filters.value" />

	<input type="hidden" name="added-filters.tag" value="{{ .Name }}" />
	<!-- TODO: add remove button, search button -->
</li>`

const tagLengthTemplate = `<li>
	{{ .Name }}

	<select name="filter-mode">
		{{ range .FilterModes }}
		<option value="{{ . }}">
		{{ . }}
		</option>
		{{ end }}
	</select>

	<!-- assumes .Values for a length tag is a slice of ordered values (with min and max) -->
	<datalist id="length-datalist">
		{{ range .Values }}
		<option value="{{ . }}" label="{{ . }}" ></option>
		{{ end }}
	</datalist>

	<input type="range" list="length-datalist" id="added-filters.value" name="added-filters.value" />
	<input type="hidden" name="added-filters.tag" value="{{ .Name }}" />
	<!-- TODO: add remove button, search button -->
</li>`

const tagMediaTemplate = `<li>
	{{ .Name }}

	<select name="added-filters.value">
		{{ range .Values }}
		<option value="{{ . }}">
		{{ . }}
		</option>
		{{ end }}
	</select>

	<input type="hidden" name="added-filters.tag" value="{{ .Name }}" />
	<input type="hidden" name="filter-mode" value="" />
	<!-- TODO: add remove button, search button -->
</li>`

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

// TODO:
func (s *Series) GetUrl() string {
	log.Println("calling MOCK GetUrl()")

	return /* getBaseUrl() +  */ "/series/number/" + s.Id.String()
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

// TODO:
func (w *Work) GetUrl() string {
	log.Println("calling MOCK GetUrl()")

	// TODO: the returned URL should be a sudbomain, the proxy will reverse it
	return /* getBaseUrl() + */ "/work/number/" + w.Id.String()
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
func (t *Tag) ToHtml() template.HTML {
	switch t.Name {
	case "author":
		return toHtmlTemplate(tagDatalistTemplate, "author", t.getPossibleValues())
	case "date":
		return toHtmlTemplate(tagDateTemplate, "date", t.getPossibleValues())
	case "length":
		return toHtmlTemplate(tagLengthTemplate, "length", t.getPossibleValues())
	case "language":
		return toHtmlTemplate(tagDatalistTemplate, "language", t.getPossibleValues())
	case "media":
		return toHtmlTemplate(tagMediaTemplate, "media", t.getPossibleValues())
	default:
		panic("not implemented")
		// TODO
		/* return toHtmlTemplate(tagTextTemplate, "default", nil) */
	}
}

// TODO:
func (t *Tag) getPossibleValues() *TagValues {
	return getMockTagValues(t)
}

func (t *Tag) getFilterModes() []string {
	return t.getPossibleValues().FilterModes
}

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

func (t *Text) GetSourceUrls() []string {
	return []string{t.Text}
}

func (t *Text) GetType() ContentType {
	return ContentTextType
}

func (t *Text) ToHtml() template.HTML {
	return toHtmlTemplate(textTemplate, "text", t)
}

func (t *Text) GetIndex() int {
	return t.Index
}

func (t *Text) GetCaption() string {
	return ""
}

func (t *Text) SetIndex(idx int) {
	t.Index = idx
}

func (s *Sound) GetId() uuid.UUID {
	return s.Id
}

func (s *Sound) GetWork() *Work {
	return s.Work
}

func (s *Sound) GetSourceUrls() []string {
	return s.SoundPaths
}

func (s *Sound) GetType() ContentType {
	return ContentSoundType
}

func (s *Sound) ToHtml() template.HTML {
	return toHtmlTemplate(soundTemplate, "sound", s)
}

func (s *Sound) GetIndex() int {
	return s.Index
}

func (s *Sound) GetCaption() string {
	return s.Caption
}

func (s *Sound) SetIndex(idx int) {
	s.Index = idx
}

func (v *Video) GetWork() *Work {
	return v.Work
}

func (v *Video) GetId() uuid.UUID {
	return v.Id
}

func (v *Video) GetSourceUrls() []string {
	return v.VideoPaths
}

func (v *Video) GetType() ContentType {
	return ContentVideoType
}

func (v *Video) ToHtml() template.HTML {
	return toHtmlTemplate(videoTemplate, "video", v)
}

func (v *Video) GetIndex() int {
	return v.Index
}

func (v *Video) GetCaption() string {
	return v.Caption
}

func (v *Video) SetIndex(idx int) {
	v.Index = idx
}

func (i *Image) GetId() uuid.UUID {
	return i.Id
}

func (i *Image) GetWork() *Work {
	return i.Work
}

func (i *Image) GetSourceUrls() []string {
	return i.ImagePaths
}

func (i *Image) GetType() ContentType {
	return ContentImageType
}

func (i *Image) ToHtml() template.HTML {
	return toHtmlTemplate(imageTemplate, "image", i)
}

func (i *Image) GetIndex() int {
	return i.Index
}

func (i *Image) GetCaption() string {
	return i.Caption
}

func (i *Image) SetIndex(idx int) {
	i.Index = idx
}

// --------------------------------------
//
//	CONTENTTYPE CONTENTTYPE CONTENTTYPE
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/

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
	log.Println("calling MOCK createNewWork()")

	work := &Work{
		Title:  strconv.Itoa(rand.IntN(100)),
		Length: rand.IntN(100),
		Id:     uuid.New(),
	}

	saveWork(work)
	return work
}

// TODO:
func getWork(id uuid.UUID) *Work {
	return getMockDb().GetWork(id)
}

// TODO:
func getSeries(id uuid.UUID) *Series {
	return getMockDb().GetSeries(id)
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
		log.Println(err)
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

func createTemplateDataWithData2[T, V any](data T, data2 V) TemplateData[T, V] {
	return TemplateData[T, V]{
		Environment: getHtmlEnvironment(),
		Data:        data,
		Data2:       data2,
	}
}

// Request must be made to path matched by pattern like
// /foo/boo/goo/{id} and id must refer to an existing work
func getWorkFromRequest(r *http.Request) *Work {
	id := getUuidFromId(r.PathValue("id"))
	return getWork(id)
}

// Request must be made to path matched by pattern like
// /foo/boo/goo/{id} and id must refer to an existing series
func getSeriesFromRequest(r *http.Request) *Series {
	id := getUuidFromId(r.PathValue("id"))
	return getSeries(id)
}

// TODO:
func getNewFilter() *Filter {
	return getMockFilter()
}

// TODO:
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

func getFilterByNumbering(numbering Numbering) *Filter {
	return getMockDb().GetFilter(numbering)
}

func toHtmlTemplate[T any](templateString string, name string, data T) template.HTML {
	tmpl := template.Must(template.New(name).Parse(templateString))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", name))
	}

	return template.HTML(buf.String())
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
		return
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
		return
	}
}

func searchWorksNewFilterHandler(w http.ResponseWriter, r *http.Request) {
	newFilter := getNewFilter()
	saveFilter(newFilter)

	http.Redirect(w, r, r.URL.Path+"/with/filter/number/"+newFilter.GetNumberingUrlString(), http.StatusSeeOther)
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
	}).ParseFiles("./resources/search.html"))

	listings := getAllListings()
	numbering, err := urlStringToNumbering(r.PathValue("numbering"))
	if err != nil {
		panic("unexpected error when trying to get filter")
	}

	filter := getFilterByNumbering(numbering)
	templateData := createTemplateDataWithData2(listings, filter)

	templ.Execute(w, templateData)
}

// TODO: handle search
func searchWorksPostHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	numbering, err := urlStringToNumbering(r.PathValue("numbering"))
	if err != nil {
		panic("unexpected error when trying to get filter")
	}

	filter := getFilterByNumbering(numbering)

	log.Println("POST!")

	switch action {
	case "add-filter":
		err := r.ParseForm()
		if err != nil {
			panic("unexpected error parsing form")
		}

		existingFilterTags := r.Form["added-filters.tag"]
		existingFilterValues := r.Form["added-filters.value"]
		existingFilterModes := r.Form["filter-mode"]
		existingTags := make([]*Tag, len(existingFilterTags))

		for i, name := range existingFilterTags {
			existingTags[i] = &Tag{
				Name:       name,
				Value:      existingFilterValues[i],
				FilterMode: existingFilterModes[i],
			}
		}

		filter.Tags = existingTags

		newFilterTag := r.FormValue("filter-tag")
		filter.Tags = append(filter.Tags, &Tag{
			Name: newFilterTag,
		})

		saveFilter(filter)

		//log.Println("redirecting on post to filterString: " + newFilterString)
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	default:
	}

	/* action := r.FormValue("action")
	numbering := r.PathValue("numbering")

	getFilterByNumbering(id) */

	return

	/* action := r.FormValue("action")
	filterString := r.PathValue("filters")
	log.Println("extracted existing filterString from url: " + filterString)
	existingTags := extractTagsFromFilter(filterString)
	log.Println("EXISTING TAGS EXTRACTED FROM URL:")
	for _, tag := range existingTags {
		log.Println(tag.String())
	}

	switch action {
	case "add-filter":
		newFilter := r.FormValue("filter-tag")
		err := r.ParseForm()
		if err != nil {
			panic("unexpected error parsing form")
		}

		newTag := &Tag{
			Name:   newFilter,
			Values: []string{},
		}

		newFilterString := tagsToFilterString(append(existingTags, newTag))
		log.Println("redirecting on post to filterString: " + newFilterString)
		http.Redirect(w, r, "/works"+newFilterString, http.StatusSeeOther)
	default:
	} */
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

func createNewWorkHandler(w http.ResponseWriter, r *http.Request) {

	newWork := createNewWork()
	log.Println("newWork id: " + newWork.Id.String())
	saveWork(newWork)

	http.Redirect(w, r, r.URL.Path+"/number/"+newWork.Id.String(), http.StatusSeeOther)
}

func composeWorkPostHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	work := getWorkFromRequest(r)

	switch action {
	case "move-up":
		idx, err := strconv.Atoi(r.FormValue("content-index"))
		if err != nil {
			log.Println(err)
			http.Error(w, "Unexpected error moving content.", http.StatusInternalServerError)
			return
		}

		moveContentUp(work, idx)

	case "move-down":
		idx, err := strconv.Atoi(r.FormValue("content-index"))
		if err != nil {
			log.Println(err)
			http.Error(w, "Unexpected error moving content.", http.StatusInternalServerError)
			return
		}

		moveContentDown(work, idx)

	case "delete-content":
		contentId := getUuidFromId(r.FormValue("content-id"))
		deleteContent(work, contentId)

	case "upload":
		err := r.ParseMultipartForm(10 << 20) // TODO: increase?
		if err != nil {
			log.Println(err)
			http.Error(w, "Content uploaded is too large.", http.StatusBadRequest)
			return
		}

		err = handleContentUpload(r, work)
		if err != nil {
			http.Error(w, "Error during content upload.", http.StatusInternalServerError)
			return
		}

	default:
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

func createWorkHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		composeWorkGetHandler(w, r)
	case http.MethodPost:
		composeWorkPostHandler(w, r)
	default:
		return
	}
}

func composeWorkGetHandler(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/search/works/with/filter/number/{numbering...}", searchWorksHandler)

	mux.HandleFunc("/work/number/{id}", viewWorkHandler)
	mux.HandleFunc("/series/number/{id}", viewSeriesHandler)

	mux.HandleFunc("/compose/work", createNewWorkHandler)
	mux.HandleFunc("/compose/work/number/{id}", createWorkHandler)

	mux.HandleFunc("/organize/works/by/{name}", organizeWorksHandler)
	log.Fatal(http.ListenAndServe(":8080", subdomainPeriodReplacer(mux)))
}
