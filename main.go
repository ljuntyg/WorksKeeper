package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	v2 "math/rand/v2"
	"net/http"
)

type Listing interface {
	GetTitle() string
	GetTimeRequiredMinutes() int
	GetUrl() string
}

type Series struct {
	Title string
	Works []Listing
}

type Work struct {
	Title    string
	Length   int
	Contents []Content
}

// TODO: add GetCaption() method
type Content interface {
	GetContent() []string
	GetType() ContentType
	ToHTML() template.HTML
}

type Text struct {
	Text string
}

type Sound struct {
	SoundPaths []string
}

type Video struct {
	VideoPaths []string
}

type Image struct {
	ImagePaths []string
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

const textTemplate = `<p>{{index .GetContent 0}}</p>`

const soundTemplate = `<audio controls>
	{{range .GetContent}}
	<source src={{.}}>
	{{end}}
	Your browser doesn't support this audio.
</audio>`

const videoTemplate = `<video controls>
	{{range .GetContent}}
	<source src={{.}}>
	{{end}}
	Your browser doesn't support this video.
</video>`

const imageTemplate = `<img src={{index .GetContent 0}} alt="Image">`

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

	return getBaseUrl() + "/series/with/id/MOCK"
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
	return getBaseUrl() + "/works/with/id/MOCK"
}

func (t *Text) GetContent() []string {
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

func (s *Sound) GetContent() []string {
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

func (v *Video) GetContent() []string {
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

func (i *Image) GetContent() []string {
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

func (ct ContentType) mediaDir() string {
	if ct.String() == "unknown" {
		panic("Unexpected ContentType when reading mediaDir()")
	}

	return "./resources/media/" + ct.String()
}

func (ct ContentType) urlPrefix() string {
	if ct.String() == "unknown" {
		panic("Unexpected ContentType when reading urlPrefix()")
	}

	return "/" + ct.String() + "/with/name/"
}

// TODO:
func getNWorkListings(n int) []Listing {
	return GetNMockListings(n)
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
	return GetMockBaseUrl()
}

// TODO:
func getWork(id string) *Work {
	return GetMockWork(99)
}

// TODO:
func getSeries(id string) *Series {
	return GetMockSeries(100)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
	}).ParseFiles("./resources/home.html"))

	listings := getNWorkListings(v2.IntN(10))
	templ.Execute(w, listings)
}

func workHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
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

func seriesHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	series := getSeries(id)

	templ := template.Must(template.New("series.html").Funcs(template.FuncMap{
		"formatTime": formatTimeRequired,
	}).ParseFiles("./resources/series.html"))

	templ.Execute(w, series)
}

func main() {
	for _, contentType := range AllContentTypes {
		if contentType == TextType {
			continue
		}

		mediaDir := contentType.mediaDir()
		urlPrefix := contentType.urlPrefix()

		fs := http.FileServer(http.Dir(mediaDir))
		http.Handle(urlPrefix, http.StripPrefix(urlPrefix, fs))
	}

	http.HandleFunc("/works/with/id/{id}", workHandler)
	http.HandleFunc("/series/with/id/{id}", seriesHandler)
	http.HandleFunc("/{$}", homeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
