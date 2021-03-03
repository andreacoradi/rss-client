package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"rssc/rss"
)

//go:embed template.gohtml
var templateText string

func main() {
	port := flag.Int("port", 3000, "port to run web server on")
	sourceFile := flag.String("sources", "sources.list", "provide a text file containing rss sources (links)")
	updateTime := flag.Int("updateInterval", 10, "update interval in minutes")
	maxAge := flag.Int("maxAge", 7, "get news that are no older than the value (days)")
	flag.Parse()

	tpl := template.Must(template.New("").Parse(templateText))

	f := client.NewFeed(*updateTime, *maxAge)
	sourcesList, err := os.ReadFile(*sourceFile)
	var category string
	if err == nil {
		s := bufio.NewScanner(bytes.NewReader(sourcesList))
		for s.Scan() {
			// FIXME: Find a better way to do this so we can also use relative paths
			if u, _ := url.Parse(s.Text()); !u.IsAbs() {
				category = s.Text()
				continue
			}
			f.AddSource(category, s.Text())
		}
	}

	handler := client.NewHandler(f, tpl)

	fmt.Printf("Starting server on port %d...\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), handler))
}
