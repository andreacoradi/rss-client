package client

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type customTime time.Time

func (c *customTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	err := d.DecodeElement(&v, &start)
	if err != nil {
		return err
	}

	parse, err := time.Parse(time.RFC1123Z, v)
	if err != nil {
		return err
	}

	*c = customTime(parse)
	return nil
}

type Channel struct {
	XMLName       xml.Name   `xml:"channel"`
	Title         string     `xml:"title"`
	Description   string     `xml:"description"`
	URL           string     `xml:"link"`
	LastBuildDate customTime `xml:"lastBuildDate"`
	PubDate       customTime `xml:"pubDate"`
	Items         []Item     `xml:"item"`
}

type Item struct {
	XMLName     xml.Name   `xml:"item"`
	Title       string     `xml:"title"`
	URL         string     `xml:"guid"`
	Description string     `xml:"description"`
	PubDate     customTime `xml:"pubDate"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

func (i Item) String() string {
	var ret string
	ret += fmt.Sprintf("%s (%v)\n", i.Title, time.Time(i.PubDate))
	//ret += fmt.Sprintf("%s\n", i.Description)
	ret += fmt.Sprintf("%s\n", i.URL)
	return ret
}

func (i Item) GetDate() string {
	delta := time.Now().Sub(time.Time(i.PubDate)).Round(time.Minute)
	// FIXME: This could be better
	hours := delta.Hours()
	minutes := delta.Minutes()
	if hours > 24 {
		return fmt.Sprintf("%d days ago", int(hours/24))
	}

	if hours == 0 {
		return fmt.Sprintf("%d minutes ago", int(minutes))
	}

	return fmt.Sprintf("%d hours ago", int(hours))
}

func (i Item) GetHost() string {
	r, _ := url.Parse(i.URL)
	return strings.TrimLeft(r.Host, "www.")
}

type client struct {
	urls     []string
	interval int
}

func parseRSS(r io.Reader) (RSS, error) {
	dec := xml.NewDecoder(r)
	var rss RSS
	err := dec.Decode(&rss)
	if err != nil {
		return RSS{}, err
	}

	return rss, nil
}

func (c *client) addSource(link string) {
	c.urls = append(c.urls, link)
}

func (c client) getLatest() []Item {
	type result struct {
		items []Item
		err   error
	}
	resultCh := make(chan result, len(c.urls))
	for _, link := range c.urls {
		go func(link string) {
			resp, err := http.Get(link)
			if err != nil {
				resultCh <- result{err: err}
				return
			}
			defer resp.Body.Close()

			rss, err := parseRSS(resp.Body)
			if err != nil {
				resultCh <- result{err: err}
				return
			}
			resultCh <- result{items: rss.Channel.Items}
		}(link)
	}

	var ret []Item
	for i := 0; i < len(c.urls); i++ {
		res := <-resultCh
		if res.err != nil {
			continue
		}

		ret = append(ret, res.items...)
	}

	return ret
}
