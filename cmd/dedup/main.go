package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/blake2b"
)

var (
	targetPath string
	threads    int
	del        bool
	load       string

	taskPool    chan string
	wg          sync.WaitGroup
	fileIndex   map[string]string
	dupFileList map[string][]string
	dupFLLock   sync.Mutex
	fIndexLock  sync.Mutex
)

func init() {
	flag.StringVar(&targetPath, "t", ".", "-t PATH")
	flag.IntVar(&threads, "th", 1, "-th THREADS")
	flag.BoolVar(&del, "del", false, "-del (delete switch)")
	flag.StringVar(&load, "l", "", "-l path/to/del.json")
	flag.Parse()

	taskPool = make(chan string, threads)
	fileIndex = make(map[string]string, 0)
	dupFileList = make(map[string][]string, 0)
}

func main() {
	if load == "" {
		findDup()
		data, err := json.Marshal(dupFileList)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println(string(data))
	} else {
		data, err := ioutil.ReadFile(load)
		if err != nil {
			log.Panic(err)
		}
		err = json.Unmarshal(data, &dupFileList)
		if err != nil {
			log.Panic(err)
		}
	}

	if del {
		removeDupFile()
	}
}

func findDup() {
	go func() {
		_ = filepath.Walk(targetPath, walk)
		close(taskPool)
	}()

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go worker()
	}

	wg.Wait()
	log.Println("Done")
}

func walk(fPath string, fInfo os.FileInfo, err error) error {
	if fInfo.IsDir() {
		return nil
	}
	if err != nil {
		log.Println(err)
		return nil
	}

	taskPool <- fPath

	return nil
}

func worker() {
	for taskPath := range taskPool {
		log.Println("hashing: " + taskPath)
		fHash, e := hashFile(taskPath)
		if e != nil {
			log.Println(e)
			continue
		}

		fHashStr := hex.EncodeToString(fHash)

		wg.Add(1)
		go func() {
			fIndexLock.Lock()
			if _, ok := fileIndex[fHashStr]; ok {
				dupFLLock.Lock()
				dupFileList[fileIndex[fHashStr]] = append(dupFileList[fileIndex[fHashStr]], taskPath)
				dupFLLock.Unlock()
			} else {
				fileIndex[fHashStr] = taskPath
			}
			fIndexLock.Unlock()
			wg.Done()
		}()

	}
	wg.Done()
}

func hashFile(fPath string) ([]byte, error) {
	file, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	fReader := bufio.NewReader(file)

	buffer := make([]byte, 4<<10) //4k buffer
	hashes := make([]byte, 0)

	for {
		n, err := fReader.Read(buffer)
		if err == io.EOF || n == 0 {
			break
		}
		if err != io.EOF && err != nil {
			return nil, err
		}

		h := blake2b.Sum256(buffer[:n])

		hashes = append(hashes, h[:]...)

	}

	fHash := blake2b.Sum256(hashes)

	return fHash[:], file.Close()
}

func removeDupFile() {
	for _, list := range dupFileList {
		for _, v := range list {
			log.Println("deleting: " + v)
			err := os.Remove(v)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
