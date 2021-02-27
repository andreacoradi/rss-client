package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"rssc/pkg/feed"
)

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func Add(args []string) {
	// Download fonte
	//f := feed.Feed{}
	url := "https://www.ansa.it/lombardia/notizie/lombardia_rss.xml"
	err := DownloadFile("sources/lombardia.xml", url)
	if err != nil {
		panic(err)
	}

}

// TODO: Embed a directory instead
//go:embed template.gohtml
var templateText string

func main() {
	//templateText, err := ioutil.ReadFile("template.gohtml")
	//if err != nil {
	//	log.Fatal("could not read file")
	//}
	tpl := template.Must(template.New("").Parse(string(templateText)))

	f, err := feed.NewFeed("sources")
	if err != nil {
		panic(err)
	}

	handler := feed.NewHandler(f, tpl)

	fmt.Println("Starting server...")
	log.Fatal(http.ListenAndServe("localhost:3000", handler))
	//if len(os.Args) > 1 {
	//	switch os.Args[1] {
	//	case "start":
	//		Start()
	//		return
	//	case "add":
	// 		Add(os.Args[1:])
	//		return
	//	}
	//}
	//
	//f := feed.Feed{}
	//for _, item := range f.GetItems() {
	//	fmt.Println(item)
	//}
}
