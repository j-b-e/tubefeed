package rss

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"time"
	"tubefeed/internal/provider"
)

var ErrRSS = errors.New("rss error")

// PodcastRSS defines the structure for the podcast RSS XML feed
type PodcastRSS struct {
	XMLName     xml.Name       `xml:"rss"`
	Version     string         `xml:"version,attr"`
	XmlnsItunes string         `xml:"xmlns:itunes,attr"`
	Channel     PodcastChannel `xml:"channel"`
}

// PodcastChannel is the rss feed
type PodcastChannel struct {
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Description string        `xml:"description"`
	Language    string        `xml:"language"`
	Author      string        `xml:"itunes:author"`
	Image       PodcastImage  `xml:"itunes:image"`
	Items       []PodcastItem `xml:"item"`
}

// PodcastImage for the podcast
type PodcastImage struct {
	Href string `xml:"href,attr"`
}

// PodcastItem is an Item
type PodcastItem struct {
	Title       string           `xml:"title"`
	Description string           `xml:"description"`
	PubDate     string           `xml:"pubDate"`
	Link        string           `xml:"link"`
	GUID        string           `xml:"guid"`
	Enclosure   PodcastEnclosure `xml:"enclosure"`
}

// PodcastEnclosure is enclosure
type PodcastEnclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type RSS struct {
	ExternalUrl string
}

func NewRSS(externalUrl string) *RSS {
	return &RSS{
		ExternalUrl: externalUrl,
	}
}

// Generates a podcast RSS feed with the given metadata
func (r *RSS) GeneratePodcastFeed(videos []provider.VideoProvider, tabname string) (string, error) {
	channel := PodcastChannel{
		Title:       fmt.Sprintf("%s - Tubefeed", tabname),
		Link:        r.ExternalUrl,
		Description: "A collection of videos as podcast episodes.",
		Language:    "en-us",
		Author:      "Tubefeed",
		Image:       PodcastImage{Href: fmt.Sprintf("http://%s/static/logo.png", r.ExternalUrl)},
	}

	for _, video := range videos {
		metadata, err := video.LoadMetadata()
		if err != nil {
			log.Println(err)
			continue
		}
		audioURL := fmt.Sprintf("http://%s/audio/%s", r.ExternalUrl, metadata.VideoID) // Stub for audio files

		item := PodcastItem{
			Title:       fmt.Sprintf("%s - %s", metadata.Channel, metadata.Title),
			Description: fmt.Sprintf("created with Tubefeed on playlist %s", tabname),
			PubDate:     time.Now().Format("Tue, 15 Sep 2023 19:00:00 GMT"), //"Tue, 15 Sep 2023 19:00:00 GMT",
			Link:        video.Url(),
			GUID:        metadata.VideoID.String(),
			Enclosure: PodcastEnclosure{
				URL:    audioURL,                                     // Replace this with the actual audio file URL
				Length: fmt.Sprintf("%f", metadata.Length.Seconds()), // size in bytes of the audio file
				Type:   "audio/mpeg",                                 // The type of enclosure
			},
		}
		channel.Items = append(channel.Items, item)
	}

	rss := PodcastRSS{
		Version:     "2.0",
		XmlnsItunes: "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Channel:     channel,
	}

	output, _ := xml.MarshalIndent(rss, "", "  ")
	return xml.Header + string(output), nil
}
