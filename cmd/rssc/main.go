package main

import (
	"fmt"
	"github.com/andreacoradi/rssc/pkg/feed"
	"io"
	"net/http"
	"os"
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

func main() {
	f := feed.Feed{}

	// Download fonte
	//url := "https://www.ansa.it/lombardia/notizie/lombardia_rss.xml"
	//err := DownloadFile("sources/lombardia.xml", url)
	//if err != nil {
	//	panic(err)
	//}

	sourceDir := "sources"
	dir, err := os.ReadDir(sourceDir)
	if err != nil {
		panic(err)
	}

	for _, file := range dir {
		c, err := os.ReadFile(fmt.Sprintf("%s/%s", sourceDir, file.Name()))
		if err != nil {
			panic(err)
		}

		rss, err := feed.NewRSS(c)
		if err != nil {
			panic(err)
		}
		f.AddSource(rss)
	}

	for _, item := range f.GetItems() {
		fmt.Println(item)
	}
}
