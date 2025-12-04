package main

import (
	"fmt"
	"log"
)

func GetMockText() *Text {
	log.Println("calling MOCK GetMockText()")

	return &Text{
		Text: "Mock text!",
	}
}

func GetMockSound() *Sound {
	log.Println("calling MOCK GetMockSound()")

	return &Sound{
		SoundPaths: []string{"/sounds/with/name/609562_migfus20_background-music.ogg"},
	}
}

func GetMockVideo() *Video {
	log.Println("calling MOCK GetMockVideo()")

	return &Video{
		VideoPaths: []string{"/videos/with/name/14044733_1080_1920_48fps(2).mp4"},
	}
}

func GetMockImage() *Image {
	log.Println("calling MOCK GetMockImage()")

	return &Image{
		ImagePaths: []string{"/images/with/name/IMG20250819173509~2.jpg"},
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
	}
}

func GetNonNestedMockSeries(i int) *Series {
	log.Println("calling MOCK GetNonNestedMockSeries()")

	return &Series{
		Title: fmt.Sprintf("Mock series #%d", i),
		Works: []Listing{
			GetMockWork(30),
		},
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
	}
}

func GetNMockListings(n int) []Listing {
	log.Println("calling MOCK GetMockListings()")

	list := make([]Listing, 0, n)

	for j := range n {
		i := j * 10
		if i%3 == 0 {
			list = append(list, GetMockSeries(i+1))
		} else {
			list = append(list, GetMockWork(i+1))
		}
	}

	return list
}

func GetMockBaseUrl() string {
	log.Println("calling MOCK GetMockBaseUrl()")

	return "http://localhost:8080"
}
