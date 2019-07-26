package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/PuerkitoBio/goquery"
)

var (
	url     string
	outPath string
)

func initFlag() {
	flag.StringVar(&url, "t", "", "-t target_url")
	flag.StringVar(&outPath, "o", "", "-o /path/to/target")
	flag.Parse()
}

func main() {
	initFlag()

	if outPath != "" {
		err := os.MkdirAll(outPath, os.ModeDir)
		if err != nil {
			log.Fatalln(err)
		}
	}

	i := 1
	var nextUrl = ""
	var curUrl = url
	for {
		resp, err := http.Get(curUrl)
		if err != nil {
			log.Println(err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Println(resp.Status, "retry!")
			continue
		}
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println("downloading:", curUrl)
		imgNode := doc.Find("#img")

		imgUrl, _ := imgNode.Attr("src")
		imgResp, err := http.Get(imgUrl)
		if err != nil {
			log.Println(err)
			continue
		}

		if imgResp.StatusCode != http.StatusOK {
			log.Println(imgResp.Status, "retry!")
			continue
		}
		out := path.Join(outPath, fmt.Sprintf("%03d.jpg", i))
		imgData, err := ioutil.ReadAll(imgResp.Body)
		if err != nil {
			log.Println(err)
			continue
		}

		err = ioutil.WriteFile(out, imgData, 0666)
		if err != nil {
			log.Println(err)
			continue
		}

		nextUrl, _ = imgNode.Parent().Attr("href")
		if nextUrl == curUrl {
			log.Println("Done")
			break
		}

		curUrl = nextUrl
		i++
	}

}
