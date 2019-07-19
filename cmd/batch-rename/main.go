package main

import (
	"flag"
	"fmt"
	util "github.com/blinkspark/go-blink-util"
	"log"
	"os"
)

func main() {
	in := flag.String("in", "", "-in target%2d.mp4 (target file name you want to rename)")
	out := flag.String("out", "", "-out to%2d.mp4 (target file name you want to rename)")
	min := flag.Int("min", 0, "-min 1 (min value of template)")
	max := flag.Int("max", 0, "-max 10 (min value of template)")
	flag.Parse()

	util.Ignore(in, out, min, max)

	for n := *min; n <= *max; n++ {
		inFname := fmt.Sprintf(*in, n)
		outFname := fmt.Sprintf(*out, n)
		err := os.Rename(inFname, outFname)
		if err != nil {
			log.Println(err)
		}
	}
}
