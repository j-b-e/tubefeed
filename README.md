# Tubefeed

Create Podcast Feeds from YouTube Videos

## How to run

~~~
docker build -t tubefeed .
docker run -v $PWD/config:/app/config -v $PWD/audio:/app/audio -e EXTERNAL_URL=localhost:8091 tubefeed
~~~

Go to http://localhost:8091

## Features

* Create audio-only Podcast Feeds from Youtube Videos
* Organize your Podcasts in multiple playlist
* Uses htmx for a smooth and modern experience

## Development

* Generate sqlc queries with `make generate`
