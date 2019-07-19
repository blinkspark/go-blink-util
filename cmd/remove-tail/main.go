package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	fileOnly := flag.Bool("f", true, "-f (file only mode on)")
	splitter := flag.String("s", "-", `-s "-" (splitter you wang to use)`)
	target := flag.Int("t", -1, "-t -1 (target section you wan't to remove)")
	flag.Parse()

	fis, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatalln(err)
	}

	for _, fi := range fis {
		if *fileOnly && fi.IsDir() {
			continue
		}

		ext := getExt(fi.Name())
		strs := strings.Split(fi.Name(), *splitter)
		strlength := len(strs)

		//append(arr[:length+t],arr[length+t+1:length]...)
		if *target < -1 {
			strs = append(strs[:strlength+*target], strs[strlength+*target:]...)
		} else {
			strs = strs[:strlength-1]
		}

		newFname := strings.Join(strs, *splitter)
		if *target == -1 {
			newFname = newFname + "." + ext
		}
		log.Println(fi.Name())
		log.Println(newFname)
		err := os.Rename(fi.Name(), newFname)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func getExt(fname string) (ext string) {
	strs := strings.Split(fname, ".")
	ext = strs[len(strs)-1]
	return
}
