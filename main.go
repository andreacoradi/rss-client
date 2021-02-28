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
	"rssc/rss"
)

// TODO: Embed a directory instead?
//go:embed template.gohtml
var templateText string

func main() {
	tpl := template.Must(template.New("").Parse(templateText))

	sourcesList, _ := os.ReadFile("sources.list")

	f := client.NewFeed()
	s := bufio.NewScanner(bytes.NewReader(sourcesList))
	for s.Scan() {
		f.AddSource(s.Text())
	}

	f.AddSource("https://www.ansa.it/sito/notizie/mondo/mondo_rss.xml")
	f.AddSource("https://omgubuntu.co.uk/feed")

	handler := client.NewHandler(f, tpl)

	fmt.Println("Starting server...")
	log.Fatal(http.ListenAndServe("localhost:3000", handler))
}
