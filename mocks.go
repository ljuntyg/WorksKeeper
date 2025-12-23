package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var mockDb = &WorksKeeperDB{
	Listings: make(map[uuid.UUID]Listing),
	Contents: make(map[uuid.UUID]Content),
	Tags:     make(map[uuid.UUID]Tag),
}

var mockTags = map[string][]string{
	"created-by": {
		"Funny Duck",
		"Slow Man",
		"Rolling Chen",
		"Coffee Monster",
	},
	"date-created": {
		"2023-01-15",
		"2023-06-03",
		"2024-02-27",
		"2025-01-10",
	},
	"length": {
		"100",
		"200",
		"5000",
	},
	"language": {
		"en",
		"sv",
		"ja",
		"fr",
	},
	"media-types": {
		"audio",
		"video",
		"text",
		"image",
	},
}

var mockTagNames []string

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

var currentWork = 0

var maxNestedSeries = 2
var maxWorkContents = 10
var maxSeriesListings = 3
var maxWorks = 3

func init() {
	log.Println("calling MOCK init()")

	keys := make([]string, 0, len(mockTags))
	for k := range mockTags {
		keys = append(keys, k)
	}

	mockTagNames = keys

	for range maxWorks / 3 {
		mockDb.SaveSeries(getMockSeries())
	}

	for range 2 * maxWorks / 3 {
		mockDb.SaveWork(getMockWork())
	}
}

type Database interface {
	SaveWork(w *Work)
	SaveSeries(s *Series)
	SaveText(t *Text)
	SaveSound(s *Sound)
	SaveVideo(v *Video)
	SaveImage(i *Image)

	GetWork(id uuid.UUID) *Work
	GetSeries(id uuid.UUID) *Series
	GetText(id uuid.UUID) *Text
	GetSound(id uuid.UUID) *Sound
	GetVideo(id uuid.UUID) *Video
	GetImage(id uuid.UUID) *Image

	GetAllListings() []Listing
}

type WorksKeeperDB struct {
	Listings map[uuid.UUID]Listing
	Contents map[uuid.UUID]Content
	Tags     map[uuid.UUID]Tag
}

func (db *WorksKeeperDB) SaveWork(w *Work) {
	db.Listings[w.GetId()] = w

	for _, c := range w.Contents {
		saveContent(c)
	}
}

func (db *WorksKeeperDB) SaveSeries(s *Series) {
	db.Listings[s.GetId()] = s

	for _, l := range s.Listings {
		switch l := l.(type) {
		case *Work:
			db.SaveWork(l)
		case *Series:
			db.SaveSeries(l)
		default:
			panic("unexpected Listing type when saving")
		}
	}
}

func (db *WorksKeeperDB) SaveText(t *Text) {
	db.Contents[t.GetId()] = t
}

func (db *WorksKeeperDB) SaveSound(s *Sound) {
	db.Contents[s.GetId()] = s
}

func (db *WorksKeeperDB) SaveVideo(v *Video) {
	db.Contents[v.GetId()] = v
}

func (db *WorksKeeperDB) SaveImage(i *Image) {
	db.Contents[i.GetId()] = i
}

func (db *WorksKeeperDB) GetWork(id uuid.UUID) *Work {
	v, ok := db.Listings[id]
	if !ok {
		panic("GetWork: no listing with that ID")
	}
	w, ok := v.(*Work)
	if !ok {
		panic(fmt.Sprintf("GetWork: listing %T is not a Work", v))
	}
	return w
}

func (db *WorksKeeperDB) GetSeries(id uuid.UUID) *Series {
	v, ok := db.Listings[id]
	if !ok {
		panic("GetSeries: no listing with that ID")
	}
	s, ok := v.(*Series)
	if !ok {
		panic(fmt.Sprintf("GetSeries: listing %T is not a Series", v))
	}
	return s
}

func (db *WorksKeeperDB) GetText(id uuid.UUID) *Text {
	v, ok := db.Contents[id]
	if !ok {
		panic("GetText: no content with that ID")
	}
	t, ok := v.(*Text)
	if !ok {
		panic(fmt.Sprintf("GetText: content %T is not a Text", v))
	}
	return t
}

func (db *WorksKeeperDB) GetImage(id uuid.UUID) *Image {
	v, ok := db.Contents[id]
	if !ok {
		panic("GetImage: no content with that ID")
	}
	img, ok := v.(*Image)
	if !ok {
		panic(fmt.Sprintf("GetImage: content %T is not an Image", v))
	}
	return img
}

func (db *WorksKeeperDB) GetSound(id uuid.UUID) *Sound {
	v, ok := db.Contents[id]
	if !ok {
		panic("GetSound: no content with that ID")
	}
	s, ok := v.(*Sound)
	if !ok {
		panic(fmt.Sprintf("GetSound: content %T is not a Sound", v))
	}
	return s
}

func (db *WorksKeeperDB) GetVideo(id uuid.UUID) *Video {
	v, ok := db.Contents[id]
	if !ok {
		panic("GetVideo: no content with that ID")
	}
	video, ok := v.(*Video)
	if !ok {
		panic(fmt.Sprintf("GetVideo: content %T is not a Video", v))
	}
	return video
}

func getMockDb() *WorksKeeperDB {
	log.Println("calling MOCK GetMockDb()")

	return mockDb
}

func getMockText() *Text {
	log.Println("calling MOCK GetMockText()")

	return &Text{
		Text: "Mock text!",
		Id:   uuid.New(),
	}
}

func getMockSound() *Sound {
	log.Println("calling MOCK GetMockSound()")

	return &Sound{
		SoundPaths: []string{getBaseUrl() + SoundType.urlPrefix() + "609562_migfus20_background-music.ogg"},
		Caption:    "Sound caption",
		Id:         uuid.New(),
	}
}

func getMockVideo() *Video {
	log.Println("calling MOCK GetMockVideo()")

	return &Video{
		VideoPaths: []string{getBaseUrl() + VideoType.urlPrefix() + "14044733_1080_1920_48fps(2).mp4"},
		Caption:    "Testing a video caption",
		Id:         uuid.New(),
	}
}

func getMockImage() *Image {
	log.Println("calling MOCK GetMockImage()")

	return &Image{
		ImagePaths: []string{getBaseUrl() + ImageType.urlPrefix() + "IMG20250819173509~2.jpg"},
		Caption:    "An image caption",
		Id:         uuid.New(),
	}
}

func (ct ContentType) getMock() Content {
	switch ct {
	case TextType:
		return getMockText()
	case SoundType:
		return getMockSound()
	case VideoType:
		return getMockVideo()
	case ImageType:
		return getMockImage()
	default:
		panic("unexpected ContentType when getting mock")
	}
}

func getMockContents() []Content {
	n := maxWorkContents

	contents := make([]Content, n)

	for i := range n {
		ct := AllContentTypes[rng.Intn(len(AllContentTypes))]

		contents[i] = ct.getMock()
	}

	return contents
}

func getMockWork() *Work {
	log.Println("calling MOCK GetMockWork()")

	currentWork++

	return &Work{
		Title:    fmt.Sprintf("Mock work #%d", currentWork),
		Length:   1030,
		Contents: getMockContents(),
		Id:       uuid.New(),
		Tags:     getMockTags(),
		IsPublic: true,
	}
}

func getMockEmptyWork() *Work {
	log.Println("calling MOCK GetMockEmptyWork()")

	return &Work{
		Title:    "Untitled",
		Length:   0,
		Contents: []Content{},
		Id:       uuid.New(),
	}
}

var nestedSeries = 0

func getMockListings(n int) []Listing {
	listings := make([]Listing, n)

	for i := range n {
		randVal := rng.Intn(3)

		if randVal > 0 || maxNestedSeries-nestedSeries <= 0 {
			listings[i] = getMockWork()
		} else {
			nestedSeries++
			listings[i] = getMockSeries()
		}
	}

	nestedSeries = 0

	return listings
}

func getMockSeries() *Series {
	log.Println("calling MOCK GetMockSeries()")

	currentWork++

	return &Series{
		Title:    fmt.Sprintf("Mock series #%d", currentWork),
		Listings: getMockListings(maxSeriesListings),
		Id:       uuid.New(),
		Tags:     getMockTags(),
		IsPublic: true,
	}
}

func getMockTags() []*Tag {
	log.Println("calling MOCK GetMockTags()")

	n1 := rng.Intn(len(mockTagNames)-1) + 1
	n := rng.Intn(n1)
	tags := make([]*Tag, n)

	for i := range n {
		tags[i] = getMockTag()
	}

	return tags
}

func pickRandomValues(values []string, min, max int) []string {
	if len(values) == 0 {
		return nil
	}

	n := min + rng.Intn(max-min+1)
	if n > len(values) {
		n = len(values)
	}

	perm := rng.Perm(len(values))
	out := make([]string, 0, n)

	for i := 0; i < n; i++ {
		out = append(out, values[perm[i]])
	}

	return out
}

func getMockTag() *Tag {
	name := mockTagNames[rng.Intn(len(mockTagNames))]
	possibleValues := mockTags[name]

	return &Tag{
		Name:   name,
		Values: pickRandomValues(possibleValues, 1, 3),
	}
}

func GetMockBaseUrl() string {
	log.Println("calling MOCK GetMockBaseUrl()")

	return "http://localhost:8080"
}
