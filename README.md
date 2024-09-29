# Tubefeed

Create a Podcast Feed from YouTube Videos

## How to run

~~~
docker build -t tubefeed .
docker run -v $PWD/config:/app/config tubefeed
~~~

Go to http://localhost:8091

## Features

* Uses [yt-dlp](https://github.com/yt-dlp/yt-dlp) to create an Audio file from Youtube and adds it to the Feed
* Starts audio extraction on demand
