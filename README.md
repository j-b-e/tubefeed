# Tubefeed

Create a Podcast Feed from YouTube Videos

## How to run

~~~
docker build -t tubefeed .
docker run -v $PWD/config:/app/config tubefeed
~~~

Go to http://localhost:8091

## Features

* Creates an audio-only Podcast Feed from Youtube Videos
* Organize your Podcasts in multiple lists
* Uses htmx for a smooth and modern experience
