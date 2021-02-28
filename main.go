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
	"os"
	"rssc/rss"
)

//go:embed template.gohtml
var templateText string

func main() {
	port := flag.Int("port", 3000, "port to run web server on")
	sourceFile := flag.String("sources", "sources.list", "provide a text file containing rss sources (links)")
	updateTime := flag.Int("updateInterval", 10, "update interval in minutes")
	flag.Parse()

	tpl := template.Must(template.New("").Parse(templateText))

	f := client.NewFeed(*updateTime)
	sourcesList, err := os.ReadFile(*sourceFile)
	if err == nil {
		s := bufio.NewScanner(bytes.NewReader(sourcesList))
		for s.Scan() {
			f.AddSource(s.Text())
		}
	}

	handler := client.NewHandler(f, tpl)

	fmt.Printf("Starting server on port %d...\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), handler))
}
