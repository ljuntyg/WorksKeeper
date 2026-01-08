package main

import (
	"bytes"
	"fmt"
	"log"
	"maps"
	"math/rand"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Database interface {
	SaveWork(w *Work)
	SaveSeries(s *Series)

	SaveText(t *Text)
	SaveEmptyMedia(m *EmptyMedia)
	SaveSound(s *Sound)
	SaveVideo(v *Video)
	SaveImage(i *Image)

	GetWork(numbering Numbering) *Work
	GetSeries(numbering Numbering) *Series
	GetContent(id uuid.UUID) Contentable
	GetText(id uuid.UUID) *Text
	GetMedia(id uuid.UUID) Mediable
	GetEmptyMedia(id uuid.UUID) *EmptyMedia
	GetSound(id uuid.UUID) *Sound
	GetVideo(id uuid.UUID) *Video
	GetImage(id uuid.UUID) *Image

	GetAllListings() []Listable
}

type WorksKeeperDB struct {
	Listings map[Numbering]Listable
	Contents map[uuid.UUID]Contentable
	Tags     map[uuid.UUID]Tag
}

var mockDb = &WorksKeeperDB{
	Listings: make(map[Numbering]Listable),
	Contents: make(map[uuid.UUID]Contentable),
	Tags:     make(map[uuid.UUID]Tag),
}

var mockTags = []*Tag{
	// author
	{Id: uuid.New(), Name: "author", Value: "Funny Duck", FilterMode: ""},
	{Id: uuid.New(), Name: "author", Value: "Slow Man", FilterMode: ""},
	{Id: uuid.New(), Name: "author", Value: "Rolling Chen", FilterMode: ""},
	{Id: uuid.New(), Name: "author", Value: "Coffee Monster", FilterMode: ""},

	// date
	{Id: uuid.New(), Name: "date", Value: "2023-01-15", FilterMode: ""},
	{Id: uuid.New(), Name: "date", Value: "2023-06-03", FilterMode: ""},
	{Id: uuid.New(), Name: "date", Value: "2024-02-27", FilterMode: ""},
	{Id: uuid.New(), Name: "date", Value: "2025-01-10", FilterMode: ""},

	// length
	{Id: uuid.New(), Name: "length", Value: "100", FilterMode: ""},
	{Id: uuid.New(), Name: "length", Value: "200", FilterMode: ""},
	{Id: uuid.New(), Name: "length", Value: "5000", FilterMode: ""},

	// language
	{Id: uuid.New(), Name: "language", Value: "en", FilterMode: ""},
	{Id: uuid.New(), Name: "language", Value: "sv", FilterMode: ""},
	{Id: uuid.New(), Name: "language", Value: "ja", FilterMode: ""},
	{Id: uuid.New(), Name: "language", Value: "fr", FilterMode: ""},

	// media
	{Id: uuid.New(), Name: "media", Value: "audio", FilterMode: ""},
	{Id: uuid.New(), Name: "media", Value: "video", FilterMode: ""},
	{Id: uuid.New(), Name: "media", Value: "text", FilterMode: ""},
	{Id: uuid.New(), Name: "media", Value: "image", FilterMode: ""},
}

var mockTagValues map[string]*TagValues
var mockTagNames []string
var mockHtmlEnvironment *HtmlEnvironment

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

var currentWork = 0
var currentSeries = 0
var currentFilter = 0

var maxNestedSeries = 2
var maxWorkContents = 10
var maxSeriesListings = 2
var maxWorks = 2
var maxSeries = 2

func init() {
	log.Println("calling MOCK init()")

	mockTagNames = make([]string, len(mockTags))
	for i, tag := range mockTags {
		mockTagNames[i] = tag.Name
	}

	mockHtmlEnvironment = &HtmlEnvironment{
		AllTags: initMockTagValues(),
	}

	for range maxSeries {
		getMockSeries().Save()
	}

	for range maxWorks {
		getMockWork().Save()
	}
}

func initMockTagValues() []*TagValues {
	tagValuesMap := make(map[string]*TagValues)

	for _, tag := range mockTags {
		tv, exists := tagValuesMap[tag.Name]
		if !exists {
			var filterModes []string
			switch tag.Name {
			case "length":
				filterModes = []string{"exactly", "below", "above"}
			case "date":
				filterModes = []string{"exactly", "before", "after"}
			}

			tv = &TagValues{
				Name:                  tag.Name,
				PossibleValues:        []string{},
				PossibleModeModifiers: []string{"being", "not"},
				PossibleFilterModes:   filterModes,
			}
			tagValuesMap[tag.Name] = tv
		}
		tv.PossibleValues = append(tv.PossibleValues, tag.Value)
	}

	mockTagValues = tagValuesMap

	return slices.Collect(maps.Values(tagValuesMap))
}

func (db *WorksKeeperDB) SaveWork(w *Work) {
	db.Listings[w.GetNumbering()] = w

	for _, row := range w.Contents {
		for _, c := range row {
			c.Save()
		}
	}
}

func (db *WorksKeeperDB) SaveSeries(s *Series) {
	db.Listings[s.GetNumbering()] = s

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

func (db *WorksKeeperDB) SaveContent(c Contentable) {
	db.Contents[c.GetId()] = c
}

func (db *WorksKeeperDB) SaveText(t *Text) {
	db.Contents[t.GetId()] = t
}

func (db *WorksKeeperDB) SaveEmptyMedia(m *EmptyMedia) {
	db.Contents[m.GetId()] = m
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

func (db *WorksKeeperDB) GetWork(numbering Numbering) *Work {
	v, ok := db.Listings[numbering]
	if !ok {
		panic("GetWork: no listing with that ID")
	}

	w, ok := v.(*Work)
	if !ok {
		panic(fmt.Sprintf("GetWork: listing %T is not a Work", v))
	}

	return w
}

func (db *WorksKeeperDB) GetSeries(numbering Numbering) *Series {
	v, ok := db.Listings[numbering]
	if !ok {
		panic("GetSeries: no listing with that ID")
	}

	s, ok := v.(*Series)
	if !ok {
		panic(fmt.Sprintf("GetSeries: listing %T is not a Series", v))
	}

	return s
}

func (db *WorksKeeperDB) GetContent(id uuid.UUID) Contentable {
	c, ok := db.Contents[id]
	if !ok {
		panic("GetContent: no content with that ID")
	}

	return c
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

func (db *WorksKeeperDB) GetMedia(id uuid.UUID) Mediable {
	c, ok := db.Contents[id]
	if !ok {
		panic("GetMedia: no content with that ID")
	}

	m, ok := c.(Mediable)
	if !ok {
		panic(fmt.Sprintf("GetMedia: content %T is not media", c))
	}

	return m
}

func (db *WorksKeeperDB) GetEmptyMedia(id uuid.UUID) *EmptyMedia {
	c, ok := db.Contents[id]
	if !ok {
		panic("GetEmptyMedia: no content with that ID")
	}

	m, ok := c.(*EmptyMedia)
	if !ok {
		panic(fmt.Sprintf("GetEmptyMedia: content %T is not EmptyMedia", c))
	}

	return m
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

func getMockTagValues(t *Tag) *TagValues {
	log.Println("calling MOCK getMockTagValues()")

	return mockTagValues[t.Name]
}

func getMockDb() *WorksKeeperDB {
	log.Println("calling MOCK GetMockDb()")

	return mockDb
}

func getMockText(pos Position) *Text {
	log.Println("calling MOCK GetMockText()")

	return &Text{
		Text: "Mock text!",
		Id:   uuid.New(),
	}
}

func getMockSound(pos Position) *Sound {
	log.Println("calling MOCK GetMockSound()")

	return &Sound{
		Sources:  []string{getBaseUrl() + MediaSoundType.urlPrefix() + "609562_migfus20_background-music.ogg"},
		Caption:  getMockCaption(),
		Id:       uuid.New(),
		Position: pos,
	}
}

func getMockVideo(pos Position) *Video {
	log.Println("calling MOCK GetMockVideo()")

	return &Video{
		Sources:  []string{getBaseUrl() + MediaVideoType.urlPrefix() + "14044733_1080_1920_48fps(2).mp4"},
		Caption:  getMockCaption(),
		Id:       uuid.New(),
		Position: pos,
	}
}

func getMockImage(pos Position) *Image {
	log.Println("calling MOCK GetMockImage()")

	return &Image{
		Sources:  []string{getBaseUrl() + MediaImageType.urlPrefix() + "IMG20250819173509~2.jpg"},
		Caption:  getMockCaption(),
		Id:       uuid.New(),
		Position: pos,
	}
}

// TODO:
func getMockCaption() *Caption {
	return &Caption{"Test caption"}
}

func (mt MediaType) GetMock(pos Position) Contentable {
	switch mt {
	case MediaSoundType:
		return getMockSound(pos)
	case MediaVideoType:
		return getMockVideo(pos)
	case MediaImageType:
		return getMockImage(pos)
	default:
		panic("unexpected ContentType when getting mock")
	}
}

func (ct ContentType) GetMock(pos Position) Contentable {
	switch ct {
	case ContentTextType:
		return getMockText(pos)
	default:
		panic("unexpected ContentType when getting mock")
	}
}

func getMockContents() [][]Contentable {
	n := maxWorkContents

	contents := make([][]Contentable, n)

	for i := range n {
		contents[i] = make([]Contentable, 1)

		pickMedia := rng.Intn(2)
		if pickMedia == 1 {
			t := AllMediaTypes[rng.Intn(len(AllMediaTypes))]
			contents[i][0] = t.GetMock(Position{
				Vertical: i,
			})
		} else {
			t := AllContentTypes[rng.Intn(len(AllContentTypes))]
			contents[i][0] = t.GetMock(Position{
				Vertical: i,
			})
		}
	}

	return contents
}

func getMockWork() *Work {
	log.Println("calling MOCK GetMockWork()")

	currentWork++

	contents := getMockContents()

	work := &Work{
		Title:     fmt.Sprintf("Mock work #%d", currentWork),
		Length:    1030,
		Id:        uuid.New(),
		Numbering: getMockNumbering(&currentWork, "work"),
		Tags:      getMockTags(),
		IsPublic:  true,
	}

	for _, c := range contents {
		work.AddContent(c[0])
	}

	return work
}

func getMockEmptyWork() *Work {
	log.Println("calling MOCK GetMockEmptyWork()")

	return &Work{
		Title:     "Untitled",
		Length:    0,
		Contents:  [][]Contentable{},
		Id:        uuid.New(),
		Numbering: getMockNumbering(&currentWork, "work"),
	}
}

var nestedSeries = 0

func getMockListings(n int) []Listable {
	listings := make([]Listable, n)

	for i := range n {
		randVal := rng.Intn(3)

		if randVal > 0 || maxNestedSeries-nestedSeries <= 0 {
			listings[i] = getMockWork()
		} else {
			nestedSeries++
			listings[i] = getMockSeries()
		}

		if listings[i] == nil {
			log.Println("NIL listing created")
		}
	}

	nestedSeries = 0

	return listings
}

func getMockSeries() *Series {
	log.Println("calling MOCK GetMockSeries()")

	currentSeries++

	return &Series{
		Title:     fmt.Sprintf("Mock series #%d", currentSeries),
		Listings:  getMockListings(maxSeriesListings),
		Id:        uuid.New(),
		Numbering: getMockNumbering(&currentSeries, "series"),
		Tags:      getMockTags(),
		IsPublic:  true,
	}
}

func getMockNumbering(typeCounter *int, typeString string) Numbering {
	*typeCounter++

	timestamp := time.Now()
	year := timestamp.Year()

	century := year/100 + 1
	centuryYear := year % 100
	day := timestamp.YearDay()

	return Numbering{
		Type:    typeString,
		Century: century,
		Year:    centuryYear,
		Day:     day,
		Serial:  *typeCounter,
		Random:  getMockNumberingRandom(),
	}
}

func getMockFilter() *Filter {
	log.Println("calling MOCK getMockFilter()")

	listings, _ := getNListings(10)

	return &Filter{
		Listables:    listings,
		FilterGroups: createFilterGroupsFromTags(getMockTags()),
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

func getMockTag() *Tag {
	log.Println("calling MOCK getMockTag()")

	return mockTags[rng.Intn(len(mockTags))]
}

func getMockTagOf(name string) *Tag {
	log.Println("calling MOCK getMockTagOf()")

	for _, tag := range mockTags {
		if tag.Name == name {
			return tag
		}
	}

	return nil
}

func getMockProtocol() string {
	log.Println("calling MOCK getMockProtocol()")

	return "http://"
}

func getMockHost() string {
	log.Println("calling MOCK getMockHost()")

	return "localhost"
}

func getMockPort() string {
	log.Println("calling MOCK getMockPort()")

	return ":8080"
}

func getMockBaseUrl() string {
	log.Println("calling MOCK GetMockBaseUrl()")

	return getMockProtocol() + getMockHost() + getMockPort()
}

func getMockHtmlEnvironment() *HtmlEnvironment {
	log.Println("calling MOCK getMockHtmlEnvironment()")
	return mockHtmlEnvironment
}

func getMockNumberingRandom() int {
	log.Println("calling MOCK getMockNumberingRandom()")

	var buf bytes.Buffer
	for range 3 {
		fmt.Fprintf(&buf, "%d", rng.Intn(9)+1)
	}

	ret, err := strconv.Atoi(buf.String())
	if err != nil {
		panic("unexpected error creating random int")
	}

	return ret
}
