package client

import (
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Feed struct {
	Sources []RSS
	client  client
}

func (f *Feed) AddSource(category, link string) {
	// FIXME: Invalidate cache to show articles from added source
	f.client.addSource(category, link)
}

func NewFeed(interval, maxAge int) Feed {
	return Feed{
		Sources: []RSS{},
		client: client{
			interval: interval,
			urls:     make(map[string][]string),
			maxAge:   time.Duration(maxAge),
		},
	}
}

func (f Feed) Items() map[string][]Item {
	ret := f.client.getLatest()
	for category, items := range ret {
		sort.Slice(items, func(i, j int) bool {
			return time.Time(items[i].PubDate).After(time.Time(items[j].PubDate))
		})

		var t []Item
		for _, item := range items {
			if time.Time(item.PubDate).Before(time.Now().Add(time.Hour * 24 * -f.client.maxAge)) {
				continue
			}
			t = append(t, item)
		}

		ret[category] = t
	}

	return ret
}

func (f Feed) Category(category string) []Item {
	items := f.CachedItems()
	return items[category]
}

var (
	cache           map[string][]Item
	cacheExpiration time.Time
	cacheMutex      sync.Mutex
)

func (f Feed) CachedItems() map[string][]Item {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if time.Now().Before(cacheExpiration) {
		return cache
	}

	cache = f.Items()
	cacheExpiration = time.Now().Add(time.Minute * time.Duration(f.client.interval))
	return cache
}

type handler struct {
	feed Feed
	tpl  *template.Template
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Path)
	start := time.Now()
	articles := make(map[string][]Item)
	if path != "" && path != "/" {
		path = path[1:]
		articles[path] = h.feed.Category(path)
	} else {
		articles = h.feed.CachedItems()
	}
	err := h.tpl.Execute(w, articles)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	log.Println("Refresh time:", time.Now().Sub(start))
}

func NewHandler(feed Feed, template *template.Template) http.Handler {
	return handler{feed, template}
}
