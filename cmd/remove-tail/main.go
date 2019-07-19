package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	fileOnly := flag.Bool("f", false, "-f (file only mode on)")
	splitter := flag.String("s", "-", `-s "-" (splitter you wang to use)`)
	target := flag.Int("t", -1, "-t -1 (target section you wan't to remove)")

	fis, err := ioutil.ReadDir("")
	if err != nil {
		log.Fatalln(err)
	}

	for _, fi := range fis {
		if *fileOnly && fi.IsDir() {
			continue
		}

		strs := strings.Split(fi.Name(), *splitter)
		strlength := len(strs)

		//append(arr[:length+t],arr[length+t+1:length]...)
		if *target < -1 {
			strs = append(strs[:strlength+*target], strs[strlength+*target:]...)
		} else {
			strs = strs[:strlength-1]
		}

		newFname := strings.Join(strs, *splitter)
		err := os.Rename(fi.Name(), newFname)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
