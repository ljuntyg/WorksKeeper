package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

func init() {
	log.Println("calling MOCK init()")

	for _, l := range GetNMockListings(10) {
		mockExistingDb.AddListing(l)
	}

	for _, c := range GetNMockContents(20) {
		mockExistingDb.AddContent(c)
	}
}

type Database interface {
	AddContent(c Content)
	AddListing(l Listing)

	GetWork(id uuid.UUID) *Work
	GetSeries(id uuid.UUID) *Series
	GetText(id uuid.UUID) *Text
	GetImage(id uuid.UUID) *Image
	GetSound(id uuid.UUID) *Sound
	GetVideo(id uuid.UUID) *Video
}

type WorksKeeperDB struct {
	Listings map[uuid.UUID]Listing
	Contents map[uuid.UUID]Content
}

var mockExistingDb = &WorksKeeperDB{
	Listings: make(map[uuid.UUID]Listing),
	Contents: make(map[uuid.UUID]Content),
}

func (db *WorksKeeperDB) AddContent(c Content) {
	db.Contents[c.GetId()] = c
}

func (db *WorksKeeperDB) AddListing(l Listing) {
	db.Listings[l.GetId()] = l

	switch x := l.(type) {

	case *Work:
		for _, c := range x.Contents {
			db.AddContent(c)
		}

	case *Series:
		for _, item := range x.Works {
			db.AddListing(item)
		}
	}
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

func GetExistingMockDb() *WorksKeeperDB {
	log.Println("calling MOCK GetExistingMockDb()")

	return mockExistingDb
}

func GetMockText() *Text {
	log.Println("calling MOCK GetMockText()")

	return &Text{
		Text: "Mock text!",
		Id:   uuid.New(),
	}
}

func GetMockSound() *Sound {
	log.Println("calling MOCK GetMockSound()")

	return &Sound{
		SoundPaths: []string{getBaseUrl() + SoundType.urlPrefix() + "609562_migfus20_background-music.ogg"},
		Caption:    "Sound caption",
		Id:         uuid.New(),
	}
}

func GetMockVideo() *Video {
	log.Println("calling MOCK GetMockVideo()")

	return &Video{
		VideoPaths: []string{getBaseUrl() + VideoType.urlPrefix() + "14044733_1080_1920_48fps(2).mp4"},
		Caption:    "Testing a video caption",
		Id:         uuid.New(),
	}
}

func GetMockImage() *Image {
	log.Println("calling MOCK GetMockImage()")

	return &Image{
		ImagePaths: []string{getBaseUrl() + ImageType.urlPrefix() + "IMG20250819173509~2.jpg"},
		Caption:    "An image caption",
		Id:         uuid.New(),
	}
}

func GetMockWork(i int) *Work {
	log.Println("calling MOCK GetMockWork()")

	return &Work{
		Title:  fmt.Sprintf("Mock work %d", i),
		Length: 1030,
		Contents: []Content{
			GetMockText(),
			GetMockSound(),
			GetMockImage(),
			GetMockVideo(),
			GetMockText(),
			GetMockText(),
			GetMockText(),
			GetMockVideo(),
			GetMockVideo(),
			GetMockImage(),
		},
		Id: uuid.New(),
	}
}

func GetMockEmptyWork() *Work {
	log.Println("calling MOCK GetMockEmptyWork()")

	return &Work{
		Title:    "Untitled",
		Length:   0,
		Contents: []Content{},
		Id:       uuid.New(),
	}
}

func GetNonNestedMockSeries(i int) *Series {
	log.Println("calling MOCK GetNonNestedMockSeries()")

	return &Series{
		Title: fmt.Sprintf("Mock series #%d", i),
		Works: []Listing{
			GetMockWork(30),
		},
		Id: uuid.New(),
	}
}

func GetMockSeries(i int) *Series {
	log.Println("calling MOCK GetMockSeries()")

	return &Series{
		Title: fmt.Sprintf("Mock series #%d", i),
		Works: []Listing{
			GetMockWork(1),
			GetMockWork(2),
			GetNonNestedMockSeries(29),
		},
		Id: uuid.New(),
	}
}

func GetNMockListings(n int) []Listing {
	log.Println("calling MOCK GetNMockListings()")

	list := make([]Listing, 0, n)

	for j := range n {
		i := j + 1
		if i%3 == 0 {
			list = append(list, GetMockSeries(i))
		} else {
			list = append(list, GetMockWork(i))
		}
	}

	return list
}

func GetNMockContents(n int) []Content {
	log.Println("calling MOCK GetNMockContents()")

	contents := make([]Content, 0, n)

	for i := range n {
		switch i % 4 {
		case 0:
			contents = append(contents, GetMockText())
		case 1:
			contents = append(contents, GetMockSound())
		case 2:
			contents = append(contents, GetMockImage())
		case 3:
			contents = append(contents, GetMockVideo())
		}
	}

	return contents
}
func GetMockBaseUrl() string {
	log.Println("calling MOCK GetMockBaseUrl()")

	return "http://localhost:8080"
}
