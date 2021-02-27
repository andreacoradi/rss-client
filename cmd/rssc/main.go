package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"rssc/pkg/feed"
)

// TODO: Embed a directory instead?
//go:embed template.gohtml
var templateText string

func main() {
	tpl := template.Must(template.New("").Parse(templateText))

	//f, err := feed.NewFeed("sources")

	sourcesList, _ := os.ReadFile("sources.list")

	var links []string
	s := bufio.NewScanner(bytes.NewReader(sourcesList))
	for s.Scan() {
		links = append(links, s.Text())
	}

	f := feed.InitFeed(links)

	handler := feed.NewHandler(f, tpl)

	fmt.Println("Starting server...")
	log.Fatal(http.ListenAndServe("localhost:3000", handler))
}
