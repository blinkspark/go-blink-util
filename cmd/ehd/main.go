package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	util "github.com/blinkspark/go-blink-util"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
)

var (
	url      string
	outPath  string
	retry    int
	override bool
)

func initFlag() {
	flag.StringVar(&url, "t", "", "-t target_url")
	flag.StringVar(&outPath, "o", "", "-o /path/to/target")
	flag.IntVar(&retry, "re", 10, "-re N (retry times)")
	flag.BoolVar(&override, "override", false, "-override (override folder)")
	flag.Parse()
}

func main() {
	initFlag()

	doc, err := newDoc(url, retry)
	if err != nil {
		util.CheckErr(err)
	}

	handlePage(doc)
}

func handleGDoc(doc *goquery.Document) {
	href, _ := doc.Find(".gdtm").First().Find("a").Attr("href")
	gn := doc.Find("#gn").Text()

	download(href, path.Join(outPath, gn))
}

func handlePage(doc *goquery.Document) {
	doc.Find(".glname").Find("a").Each(func(i int, selection *goquery.Selection) {
		ghref, _ := selection.Attr("href")
		gDoc, err := newDoc(ghref, retry)
		if err != nil {
			log.Println(err)
			return
		}

		handleGDoc(gDoc)
	})

	// next page
	nextDoc, err := findNextPage(doc)
	if err != nil {
		log.Println(err)
		return
	}
	handlePage(nextDoc)
}

func findNextPage(doc *goquery.Document) (*goquery.Document, error) {
	lastPL := doc.Find(".ptb").Find("a").Last()
	if lastPL.Text() != ">" {
		return nil, errors.New("this is last page")
	}

	href, _ := lastPL.Attr("href")
	nextDoc, err := newDoc(href, retry)
	if err != nil {
		return nil, err
	}
	return nextDoc, nil
}

func newDoc(url string, retry int) (*goquery.Document, error) {
	for i := 0; i < retry; i++ {
		resp, err := http.Get(url)
		if err != nil {
			log.Println(err, "retry!")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Println("Getting " + url + "failed, retry!")
			continue
		}

		return goquery.NewDocumentFromReader(resp.Body)
	}
	return nil, errors.New("max retry reached")
}

func download(url string, oPath string) {
	if oPath != "" {
		reg := regexp.MustCompile(`[ /\\:*?;'"<>|@#$]+`)
		oPath = reg.ReplaceAllString(oPath, "_")
		if util.Exists(oPath) && !override {
			return
		}
		log.Println("oPath:", oPath)
		err := os.MkdirAll(oPath, os.ModeDir)
		if err != nil {
			log.Fatalln(err)
		}
	}

	i := 1
	var nextUrl = ""
	var curUrl = url
	for {
		doc, err := newDoc(curUrl, retry)
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

		//fName := getFName(doc)
		out := path.Join(oPath, fmt.Sprintf("%04d.jpg", i))
		log.Println(out)
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

// TODO bug always get same name
//func getFName(doc *goquery.Document) string {
//	info := doc.Find("#i2").Last().Text()
//	log.Println(info)
//	return strings.Split(info, " ")[0]
//}
