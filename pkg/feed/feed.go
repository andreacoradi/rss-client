package feed

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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

func (i Item) GetDate() string {
	delta := time.Now().Sub(time.Time(i.PubDate))
	return fmt.Sprintf("%d hours ago", int(delta.Hours()))
}

type Feed []RSS

func (f *Feed) AddSource(rss RSS) {
	// TODO: Implement the concept of categories
	*f = append(*f, rss)
}

func InitFeed(links []string) Feed {
	start := time.Now()
	c := Client{sources: links}
	sources := c.GetLatest()

	log.Println("Init time:", time.Now().Sub(start))
	return sources
}

func (f Feed) GetItems() ([]Item, error) {
	// TODO: Order by latest
	var ret []Item
	for _, source := range f {
		for _, item := range source.Channel.Items {
			ret = append(ret, item)
		}
	}
	return ret, nil
}

var (
	cache           []Item
	cacheExpiration time.Time
	cacheMutex      sync.Mutex
)

func (f Feed) GetCachedItems() ([]Item, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if time.Now().Before(cacheExpiration) {
		return cache, nil
	}
	items, err := f.GetItems()
	if err != nil {
		return nil, err
	}
	cache = items
	cacheExpiration = time.Now().Add(time.Minute * 10)
	return cache, nil
}

func NewRSS(c []byte) (RSS, error) {
	var feed RSS
	if err := xml.Unmarshal(c, &feed); err != nil {
		return RSS{}, err
	}

	return feed, nil
}

// Useless for now
func NewFeed(sourceDir string) (Feed, error) {
	feed := Feed{}
	//sourceDir := "sources"
	dir, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}

	for _, file := range dir {
		c, err := os.ReadFile(fmt.Sprintf("%s/%s", sourceDir, file.Name()))
		if err != nil {
			return nil, err
		}

		rss, err := NewRSS(c)
		if err != nil {
			return nil, err
		}
		feed.AddSource(rss)
	}

	return feed, nil
}

type handler struct {
	feed Feed
	tpl  *template.Template
}

func (h handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	//path := strings.TrimSpace(r.URL.Path)
	//if path == "" || path == "/" {
	//	path = "/intro"
	//}
	//path = path[1:]
	start := time.Now()
	// FIXME: Better error handling
	items, err := h.feed.GetCachedItems()
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "Something went wrong...", http.StatusBadRequest)
		return
	}

	err = h.tpl.Execute(w, items)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "Something went wrong...", http.StatusBadRequest)
		return
	}

	log.Println("Refresh time:", time.Now().Sub(start))

	// TODO: Categories
	//http.Error(w, fmt.Sprintf("Could not find category %q", path), http.StatusNotFound)
}

func NewHandler(feed Feed, template *template.Template) http.Handler {
	return handler{feed, template}
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

type Client struct {
	sources []string
}

func (c *Client) AddSource(link string) {
	c.sources = append(c.sources, link)
}

func (c Client) GetLatest() []RSS {
	type result struct {
		rss RSS
		err error
	}
	resultCh := make(chan result, len(c.sources))
	for _, link := range c.sources {
		go func(link string) {
			resp, err := http.Get(link)
			if err != nil {
				resultCh <- result{err: err}
				return
			}
			defer resp.Body.Close()

			rss, err := parseRSS(resp.Body)
			resultCh <- result{rss: rss}
		}(link)
	}

	var ret []RSS
	for i := 0; i < len(c.sources); i++ {
		res := <-resultCh
		if res.err != nil {
			continue
		}

		ret = append(ret, res.rss)
	}

	sort.Slice(ret, func(i, j int) bool {
		return time.Time(ret[i].Channel.PubDate).Before(time.Time(ret[j].Channel.PubDate))
	})

	return ret
}