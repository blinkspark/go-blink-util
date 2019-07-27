package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	util "github.com/blinkspark/go-blink-util"
)

var (
	needDelete bool
	targetPath string
	quality    int
)

func initFlags() {
	flag.BoolVar(&needDelete, "delete", false, "-delete (switch on delete source mode)")
	flag.StringVar(&targetPath, "t", ".", "-t /path/to/in")
	flag.IntVar(&quality, "q", 100, "-q QUALITY")
	flag.Parse()
}

func main() {
	initFlags()

	convList := make([]string, 0)

	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if !isImg(path) {
			return nil
		}

		convList = append(convList, path)

		return nil
	})
	util.CheckErr(err)

	for _, t := range convList {
		convert(t)
	}
}

func isImg(path string) bool {
	if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") ||
		strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".bmp") {
		return true
	}
	return false
}

func getOutName(path string) string {
	dirPath, fName := filepath.Split(path)
	fNames := strings.Split(fName, ".")
	fNames = fNames[:len(fNames)-1]
	fNames = append(fNames, "webp")
	fName = strings.Join(fNames, ".")
	return filepath.Join(dirPath, fName)
}

func convert(path string) {
	args := []string{"-y", "-i", path, "-quality", fmt.Sprintf("%d", quality), getOutName(path)}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdErr := cmd.Run()
	if cmdErr != nil {
		log.Println(cmdErr)
	}

	if needDelete {
		err := os.Remove(path)
		if err != nil {
			log.Println(err)
		}
	}
}
