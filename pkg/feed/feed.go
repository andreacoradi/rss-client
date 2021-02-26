package feed

import (
	"encoding/xml"
	"fmt"
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
	Link          string     `xml:"link"`
	LastBuildDate customTime `xml:"lastBuildDate"`
	PubDate       customTime `xml:"pubDate"`
	Items         []Item     `xml:"item"`
}

type Item struct {
	XMLName     xml.Name   `xml:"item"`
	Title       string     `xml:"title"`
	Link        string     `xml:"guid"`
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
	ret += fmt.Sprintf("%s\n", i.Link)
	return ret
}

//func (i *Item) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
//	var s string
//	if err := d.DecodeElement(&s, &start); err != nil {
//		return err
//	}
//
//	fmt.Println(s)
//
//	return nil
//}

type Feed []RSS

func (f *Feed) AddSource(rss RSS) {
	*f = append(*f, rss)
}

func (f Feed) GetItems() []Item {
	var ret []Item
	for _, source := range f {
		for _, item := range source.Channel.Items {
			ret = append(ret, item)
		}
	}
	return ret
}

func NewRSS(c []byte) (RSS, error) {
	var feed RSS
	if err := xml.Unmarshal(c, &feed); err != nil {
		return RSS{}, err
	}

	return feed, nil
}
