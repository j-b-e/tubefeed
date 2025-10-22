package rss

import (
	"encoding/xml"
	"errors"
	"fmt"
	"time"
	"tubefeed/internal/meta"

	"github.com/google/uuid"
)

var ErrRSS = errors.New("rss error")

const rfc2822 = "Mon Jan 02 2006 15:04:05 MST"

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
func (r *RSS) GeneratePodcastFeed(videos []meta.Source, playlist uuid.UUID) (string, error) {
	channel := PodcastChannel{
		Title:       fmt.Sprintf("%s - Tubefeed", playlist.String()), // TODO use playlist  struct to retrieve name
		Link:        r.ExternalUrl,
		Description: "A collection of videos as podcast episodes.",
		Language:    "en-us",
		Author:      "Tubefeed",
		Image:       PodcastImage{Href: fmt.Sprintf("http://%s/static/logo.png", r.ExternalUrl)},
	}

	for _, video := range videos {
		if video.Status != meta.StatusReady {
			continue
		}
		audioURL := fmt.Sprintf("http://%s/audio/%s", r.ExternalUrl, video.ID)

		// https://help.apple.com/itc/podcasts_connect/#/itcb54353390
		item := PodcastItem{
			Title:       fmt.Sprintf("%s - %s", video.Meta.Channel, video.Meta.Title),
			Description: fmt.Sprintf("created with Tubefeed on playlist %s", playlist.String()),
			PubDate:     time.Now().Format(rfc2822),
			Link:        video.Meta.URL,
			GUID:        video.ID.String(),
			Enclosure: PodcastEnclosure{
				URL:    audioURL,
				Length: fmt.Sprintf("%d", 0), // TODO: size in bytes of the audio file
				Type:   "audio/mpeg",
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
