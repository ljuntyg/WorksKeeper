package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"iter"
	"log"
	"maps"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/google/uuid"
)

type Listing interface {
	GetTitle() string
	GetTimeRequiredMinutes() int
	GetUrl() string
	GetId() uuid.UUID
}

type Series struct {
	Title string
	Works []Listing
	Id    uuid.UUID
}

type Work struct {
	Title    string
	Length   int
	Contents []Content
	Id       uuid.UUID
}

// TODO: add GetCaption() method
type Content interface {
	GetSourceUrls() []string
	GetType() ContentType
	ToHTML() template.HTML
	GetIndex() int
	GetCaption() string
	GetId() uuid.UUID
	SetIndex(idx int)
}

type Text struct {
	Text  string
	Index int
	Id    uuid.UUID
}

type Sound struct {
	SoundPaths []string
	Index      int
	Caption    string
	Id         uuid.UUID
}

type Video struct {
	VideoPaths []string
	Index      int
	Caption    string
	Id         uuid.UUID
}

type Image struct {
	ImagePaths []string
	Index      int
	Caption    string
	Id         uuid.UUID
}

type ContentType int

const (
	TextType ContentType = iota
	SoundType
	VideoType
	ImageType
)

var AllContentTypes = []ContentType{
	TextType,
	SoundType,
	VideoType,
	ImageType,
}

var MimeToContentType = map[string]ContentType{
	"image/apng":      ImageType,
	"image/avif":      ImageType,
	"image/bmp":       ImageType,
	"image/gif":       ImageType,
	"image/jpeg":      ImageType,
	"image/png":       ImageType,
	"image/svg+xml":   ImageType,
	"image/tiff":      ImageType,
	"image/webp":      ImageType,
	"audio/aac":       SoundType,
	"audio/midi":      SoundType,
	"audio/x-midi":    SoundType,
	"audio/mpeg":      SoundType,
	"audio/ogg":       SoundType,
	"audio/wav":       SoundType,
	"audio/webm":      SoundType,
	"audio/3gpp":      SoundType,
	"audio/3gpp2":     SoundType,
	"application/ogg": SoundType, // TODO: ? some .ogg files
	"video/mp4":       VideoType,
	"video/mpeg":      VideoType,
	"video/ogg":       VideoType,
	"video/webm":      VideoType,
	"video/x-msvideo": VideoType,
	"video/mp2t":      VideoType,
	"video/3gpp":      VideoType,
	"video/3gpp2":     VideoType,
}

const textTemplate = `<p>{{index .GetSourceUrls 0}}</p>`

/* const soundTemplate = `<audio controls>
	{{range .GetSourceUrls}}
	<source src={{.}}>
	{{end}}
	Your browser doesn't support this audio.
</audio>` */

const soundTemplate = `<figure>
	<audio controls>
		{{range .GetSourceUrls}}
		<source src={{.}}>
		{{end}}
		Your browser doesn't support this audio.
	</audio>
	<figcaption>{{.GetCaption}}</figcaption>
</figure>`

/* const videoTemplate = `<video controls>
	{{range .GetSourceUrls}}
	<source src={{.}}>
	{{end}}
	Your browser doesn't support this video.
</video>` */

const videoTemplate = `<figure>
	<video controls>
		{{range .GetSourceUrls}}
		<source src={{.}}>
		{{end}}
		Your browser doesn't support this video.
	</video>
	<figcaption>{{.GetCaption}}</figcaption>
</figure>`

/* const imageTemplate = `<img src={{index .GetSourceUrls 0}} alt="Image">` */

const imageTemplate = `<figure>
	<img src={{index .GetSourceUrls 0}} alt="Image">
	<figcaption>{{.GetCaption}}</figcaption>
</figure>`

// ----------------------------
//
//	LISTINGS LISTINGS LISTINGS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (s *Series) GetTitle() string {
	return s.Title
}

func (s *Series) GetTimeRequiredMinutes() int {
	totalTime := 0
	for _, elem := range s.Works {
		totalTime += elem.GetTimeRequiredMinutes()
	}

	return totalTime
}

// TODO:
func (s *Series) GetUrl() string {
	log.Println("calling MOCK GetUrl()")

	return getBaseUrl() + "/series/number/" + s.Id.String()
}

func (s *Series) GetId() uuid.UUID {
	return s.Id
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
	return getBaseUrl() + "/work/number/" + w.Id.String()
}

func (w *Work) GetId() uuid.UUID {
	return w.Id
}

// TODO:
func (w *Work) addContent(c Content) {
	w.Contents = append(w.Contents, c)
	c.SetIndex(len(w.Contents) - 1)

	sort.Slice(w.Contents, func(i, j int) bool {
		return w.Contents[i].GetIndex() < w.Contents[j].GetIndex()
	})

	GetExistingMockDb().Contents[c.GetId()] = c
}

// ----------------------------
//
//	CONTENTS CONTENTS CONTENTS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/

func (t *Text) GetSourceUrls() []string {
	return []string{t.Text}
}

func (t *Text) GetType() ContentType {
	return TextType
}

func (t *Text) ToHTML() template.HTML {
	tmpl := template.Must(template.New("text").Parse(textTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, t); err != nil {
		panic("Unexpected error executing text HTML template")
	}

	return template.HTML(buf.String())
}

func (t *Text) GetIndex() int {
	return t.Index
}

func (t *Text) GetCaption() string {
	return ""
}

func (t *Text) GetId() uuid.UUID {
	return t.Id
}

func (t *Text) SetIndex(idx int) {
	t.Index = idx
}

func (s *Sound) GetSourceUrls() []string {
	return s.SoundPaths
}

func (s *Sound) GetType() ContentType {
	return SoundType
}

func (s *Sound) ToHTML() template.HTML {
	tmpl := template.Must(template.New("sound").Parse(soundTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, s); err != nil {
		panic("Unexpected error executing sound HTML template")
	}

	return template.HTML(buf.String())
}

func (s *Sound) GetIndex() int {
	return s.Index
}

func (s *Sound) GetId() uuid.UUID {
	return s.Id
}

func (s *Sound) GetCaption() string {
	return s.Caption
}

func (s *Sound) SetIndex(idx int) {
	s.Index = idx
}

func (v *Video) GetSourceUrls() []string {
	return v.VideoPaths
}

func (v *Video) GetType() ContentType {
	return VideoType
}

func (v *Video) ToHTML() template.HTML {
	tmpl := template.Must(template.New("video").Parse(videoTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, v); err != nil {
		panic("Unexpected error executing video HTML template")
	}

	return template.HTML(buf.String())
}

func (v *Video) GetIndex() int {
	return v.Index
}

func (v *Video) GetCaption() string {
	return v.Caption
}

func (v *Video) GetId() uuid.UUID {
	return v.Id
}

func (v *Video) SetIndex(idx int) {
	v.Index = idx
}

func (i *Image) GetSourceUrls() []string {
	return i.ImagePaths
}

func (i *Image) GetType() ContentType {
	return ImageType
}

func (i *Image) ToHTML() template.HTML {
	tmpl := template.Must(template.New("image").Parse(imageTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, i); err != nil {
		panic("Unexpected error executing image HTML template")
	}

	return template.HTML(buf.String())
}

func (i *Image) GetIndex() int {
	return i.Index
}

func (i *Image) GetCaption() string {
	return i.Caption
}

func (i *Image) GetId() uuid.UUID {
	return i.Id
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
	case TextType:
		return "texts"
	case SoundType:
		return "sounds"
	case VideoType:
		return "videos"
	case ImageType:
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

func (ct ContentType) createNew(sources []string) Content {
	switch ct {
	case TextType:
		return &Text{
			Text: sources[0],
			Id:   uuid.New(),
		}
	case SoundType:
		return &Sound{
			SoundPaths: sources,
			Id:         uuid.New(),
		}
	case ImageType:
		return &Image{
			ImagePaths: sources,
			Id:         uuid.New(),
		}
	case VideoType:
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

func formatTimeRequired(min int) string {
	if min >= 60 {
		hours := min / 60

		return fmt.Sprintf("%d hr", hours)
	}

	return fmt.Sprintf("%d min", min)
}

// TODO:
func addListing(l Listing) {
	GetExistingMockDb().Listings[l.GetId()] = l
}

// TODO:
func getBaseUrl() string {
	return GetMockBaseUrl()
}

// TODO:
func createNewWork() *Work {
	log.Println("calling MOCK createNewWork()")

	work := &Work{
		Title:  strconv.Itoa(rand.IntN(100)),
		Length: rand.IntN(100),
		Id:     uuid.New(),
	}

	addListing(work)
	return work
}

// TODO:
func getWork(id uuid.UUID) *Work {
	return GetExistingMockDb().GetWork(id)
}

// TODO:
func getSeries(id uuid.UUID) *Series {
	return GetExistingMockDb().GetSeries(id)
}

// TODO:
func getContent(id uuid.UUID) Content {
	return GetExistingMockDb().Contents[id]
}

// TODO:
func GetImage(id uuid.UUID) *Image {
	return GetExistingMockDb().GetImage(id)
}

func getUuidFromId(id string) uuid.UUID {
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

func storeUploadedFormFile(r *http.Request) (Content, error) {
	file, header, err := r.FormFile("upload")
	if err != nil {
		log.Println(err)
		return nil, err
	} else {
		ct, validCt := getFileContentType(file)
		if !validCt || ct == TextType {
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

	work.addContent(content)

	return nil
}

// TODO:
func getAllListings() iter.Seq[Listing] {
	return maps.Values(GetExistingMockDb().Listings)
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

	delete(GetExistingMockDb().Contents, contentId)
}

// ----------------------------
//
//	HANDLERS HANDLERS HANDLERS
//
// \/\/\/\/\/\/\/\/\/\/\/\/\/\/

func homeHandler(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
	}).ParseFiles("./resources/home.html"))

	listings := getAllListings()

	templ.Execute(w, listings)
}

func viewWorkHandler(w http.ResponseWriter, r *http.Request) {
	id := getUuidFromId(r.PathValue("id"))
	work := getWork(id)
	templ := template.Must(template.New("work.html").Funcs(template.FuncMap{
		"isText":             func(ct ContentType) bool { return ct == TextType },
		"isSound":            func(ct ContentType) bool { return ct == SoundType },
		"isVideo":            func(ct ContentType) bool { return ct == VideoType },
		"isImage":            func(ct ContentType) bool { return ct == ImageType },
		"getAllContentTypes": func() []ContentType { return AllContentTypes },
	}).ParseFiles("./resources/work.html"))

	templ.Execute(w, work)
}

func viewSeriesHandler(w http.ResponseWriter, r *http.Request) {
	id := getUuidFromId(r.PathValue("id"))
	series := getSeries(id)

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
		return
	}
}

func createWorkGetHandler(w http.ResponseWriter, r *http.Request) {
	newWork := createNewWork()
	addListing(newWork)

	templ := template.Must(template.ParseFiles("./resources/canvas.html"))
	templ.Execute(w, newWork)
}

func createWorkPostHandler(w http.ResponseWriter, r *http.Request) {

	action := r.FormValue("action")
	id := getUuidFromId(r.FormValue("work-id"))
	work := getWork(id)

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

	templ := template.Must(template.ParseFiles("./resources/canvas.html"))
	templ.Execute(w, work)
}

func editWorkHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		editWorkGetHandler(w, r)
	case http.MethodPost:
		createWorkPostHandler(w, r)
	default:
		return
	}
}

func editWorkGetHandler(w http.ResponseWriter, r *http.Request) {
	id := getUuidFromId(r.FormValue("id"))
	work := getWork(id)
	templ := template.Must(template.ParseFiles("./resources/canvas.html"))
	templ.Execute(w, work)
}

// TODO:
func organizeWorksHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("hi")
}

func main() {
	for _, contentType := range AllContentTypes {
		if contentType == TextType {
			continue
		}

		mediaDir := "." + contentType.mediaDir() // ./{resourcefolder}/{media}/{contentType (plural)}/
		urlPrefix := contentType.urlPrefix()     // /{contentType (singular)}/{with}/{name}/{fileName}/

		contentTypeFileServer := http.FileServer(http.Dir(mediaDir))
		http.Handle(urlPrefix, http.StripPrefix(urlPrefix, contentTypeFileServer))
	}

	http.HandleFunc("/work/number/{id}", viewWorkHandler)
	http.HandleFunc("/series/number/{id}", viewSeriesHandler)

	http.HandleFunc("/compose/work", createWorkHandler)
	http.HandleFunc("/compose/work/number/{id}", editWorkHandler)

	http.HandleFunc("/organize/works/by/{name}", organizeWorksHandler)
	http.HandleFunc("/{$}", homeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
