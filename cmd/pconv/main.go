package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	needDelete bool
	targetPath string
	quality    int
	threads    int
)

var (
	pool chan string
	wg   sync.WaitGroup
)

func initFlags() {
	flag.BoolVar(&needDelete, "delete", false, "-delete (switch on delete source mode)")
	flag.StringVar(&targetPath, "t", ".", "-t /path/to/in")
	flag.IntVar(&quality, "q", 100, "-q QUALITY")
	flag.IntVar(&threads, "th", 2, "-th THREADS")
	flag.Parse()
}

func main() {
	initFlags()

	pool = make(chan string, threads)

	go func() {
		wg.Add(1)
		//for _, t := range convList {
		//	pool <- t
		//}

		_ = filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
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

			pool <- path

			return nil
		})

		close(pool)
		wg.Done()
	}()

	for i := 0; i < threads; i++ {
		go func() {
			wg.Add(1)
			for t := range pool {
				convert(t)
			}
			wg.Done()
		}()
	}

	time.Sleep(time.Second)

	wg.Wait()
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
