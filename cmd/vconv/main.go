package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	util "github.com/blinkspark/go-blink-util"
)

var (
	configPath string
	outPath    string
	pools      int
	shutdown   bool
	newConfig  bool
)

type Entry struct {
	Input        string `json:"input"`
	Output       string `json:"output"`
	VideoEncoder string `json:"VideoEncoder"`
	AudioEncoder string `json:"AudioEncoder"`
	CRF          int    `json:"crf"`
	Preset       string `json:"preset"`
	X265Params   string `json:"x265-params"`
	BV           string `json:"bv"`
}

func NewEntry(input string) Entry {
	return Entry{
		Input:        input,
		Output:       newOutName(input),
		VideoEncoder: "libx265",
		AudioEncoder: "copy",
		CRF:          28,
		Preset:       "medium",
		X265Params:   fmt.Sprintf("pools=%d:ssim-rd=1", pools),
	}
}

func newOutName(input string) string {
	strs := strings.Split(input, ".")
	length := len(strs)

	strs = append(strs[:length-1], "mkv")
	fname := strings.Join(strs, ".")

	return path.Join(outPath, fname)
}

type Config struct {
	Entries []Entry `json:"entries"`
}

func isVideoFmt(fname string) bool {
	if strings.HasSuffix(fname, ".mkv") || strings.HasSuffix(fname, ".mp4") ||
		strings.HasSuffix(fname, ".rmvb") || strings.HasSuffix(fname, ".ts") ||
		strings.HasSuffix(fname, ".flv") || strings.HasSuffix(fname, ".mpg") ||
		strings.HasSuffix(fname, ".mpeg") || strings.HasSuffix(fname, ".rm") ||
		strings.HasSuffix(fname, ".wmv") || strings.HasSuffix(fname, ".mov") ||
		strings.HasSuffix(fname, ".webm") {
		return true
	}
	return false
}

func main() {
	flag.StringVar(&configPath, "c", "config.json", "-c /path/to/config.json")
	flag.StringVar(&outPath, "o", "tmp", "-o tmp (path of the output)")
	flag.BoolVar(&newConfig, "n", false, "-n (create a new config.json file)")
	flag.IntVar(&pools, "p", 8, "-p 8 (x265 param thread pool)")
	flag.BoolVar(&shutdown, "shutdown", false, "-shutdown (shutdown when finished)")
	flag.Parse()

	config := &Config{}
	if newConfig {

		config.Entries = []Entry{}
		files, err := ioutil.ReadDir(".")
		if err != nil {
			log.Fatal(err)
		}
		for i := 0; i < len(files); i++ {
			f := files[i]
			if f.IsDir() || !isVideoFmt(f.Name()) {
				continue
			}

			config.Entries = append(config.Entries, NewEntry(f.Name()))
		}
		data, err := json.Marshal(config)
		util.CheckErr(err)

		err = ioutil.WriteFile(configPath, data, 0666)
		util.CheckErr(err)
		return
	}

	data, err := ioutil.ReadFile(configPath)
	util.CheckErr(err)

	err = json.Unmarshal(data, config)
	util.CheckErr(err)

	entryCount := len(config.Entries)
	for i := 0; i < entryCount; i++ {
		entry := config.Entries[i]
		dir := path.Dir(entry.Output)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.ModeDir)
			if err != nil {
				log.Fatal(err)
			}
		}
		args := []string{
			"-y", "-i",
			entry.Input,
			"-c:a",
			entry.AudioEncoder,
			"-c:v", entry.VideoEncoder,
		}
		if entry.Preset != "" {
			args = append(args, "-preset", entry.Preset)
		}
		if entry.VideoEncoder != "hevc_nvenc" {
			args = append(args, "-crf", fmt.Sprintf("%d", entry.CRF))
			if entry.X265Params != "" {
				args = append(args, "-x265-params", entry.X265Params)
			}
		}
		if entry.BV != "" {
			args = append(args, "-b:v", entry.BV)
		}
		args = append(args, entry.Output)
		cmd := exec.Command("ffmpeg", args...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Println("---------------------------------------------")
		fmt.Printf("%d/%d %s->%s\n", i+1, entryCount, entry.Input, entry.Output)

		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}
		fmt.Println("---------------------------------------------")
	}

	if shutdown {
		switch runtime.GOOS {
		case "windows":
			cmd := exec.Command("shutdown", "-s")
			err := cmd.Run()
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}
