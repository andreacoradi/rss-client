# rss-client
A very simple rss aggregator

## Build
`go build -o rssc main.go`

## Usage
Run `rssc -h` to see all the configurable flags and their defaults.

Go to e.g. `localhost:3000` on your browser to see all your articles. 

You'll need a text file containing your rss sources:
### Example
(Categories are optional)
```
Linux
https://www.phoronix.com/rss.php
https://www.gamingonlinux.com/news_rss.php
https://omgubuntu.co.uk/feed
Technology
https://www.ansa.it/sito/notizie/tecnologia/tecnologia_rss.xml
https://feeds.feedburner.com/hd-blog
World
https://www.ansa.it/sito/notizie/mondo/mondo_rss.xml
http://rss.cnn.com/rss/edition_world.rss
https://rss.nytimes.com/services/xml/rss/nyt/World.xml
```

Tip: You can click on the category names to get a more focused view on them
