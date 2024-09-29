package rss

import (
	"encoding/xml"
	"fmt"
	"time"
	"tubefeed/internal/config"
	"tubefeed/internal/video"
	"tubefeed/internal/yt"
)

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

// Generates a podcast RSS feed with the given metadata
func GeneratePodcastRSSFeed(videos []video.VideoMetadata) string {
	channel := PodcastChannel{
		Title:       "tubefeed",
		Link:        config.ExternalURL,
		Description: "A collection of YouTube videos as podcast episodes.",
		Language:    "en-us",
		Author:      "tubefeed",
		Image:       PodcastImage{Href: fmt.Sprintf("http://%s/static/logo.png", config.ExternalURL)},
	}

	for _, video := range videos {
		// Dynamically generate the full YouTube URL using the video ID
		videoURL := yt.Yturl(video.VideoID)
		audioURL := fmt.Sprintf("http://%s/audio/%s", config.ExternalURL, video.VideoID) // Stub for audio files

		item := PodcastItem{
			Title:       fmt.Sprintf("%s - %s", video.Channel, video.Title),
			Description: "Dummy Description",
			PubDate:     time.Now().Format("Tue, 15 Sep 2023 19:00:00 GMT"), //"Tue, 15 Sep 2023 19:00:00 GMT",
			Link:        videoURL,
			GUID:        videoURL,
			Enclosure: PodcastEnclosure{
				URL:    audioURL,                        // Replace this with the actual audio file URL
				Length: fmt.Sprintf("%d", video.Length), // Stub for the length of the audio file
				Type:   "audio/mpeg",                    // The type of enclosure
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
	return xml.Header + string(output)
}
