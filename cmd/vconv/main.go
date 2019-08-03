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
	configPath   string
	outPath      string
	x265Params   string
	tune         string
	shutdown     bool
	newConfig    bool
	paramsPreset string
)

const (
	preset1080p = "no-sao=1:ctu=32:qg-size=8:me=3:subme=3:me-range=38:ref=4:bframes=8:rc-lookahead=60:rd=3:psy-rdoq=1.0:cbqpoffs=-2:crqpoffs=-2"
	preset2kp   = "no-sao=1:ctu=32:qg-size=8:me=3:subme=3:me-range=57:ref=4:bframes=8:rc-lookahead=60:rd=3:psy-rdoq=1.0:cbqpoffs=-2:crqpoffs=-2"
)

type Entry struct {
	Input        string `json:"input"`
	Output       string `json:"output"`
	VideoEncoder string `json:"VideoEncoder"`
	AudioEncoder string `json:"AudioEncoder"`
	CRF          int    `json:"crf"`
	Preset       string `json:"preset"`
	Tune         string `json:"tune"`
	X265Params   string `json:"x265-params"`
	BV           string `json:"bv"`
	BA           string `json:"ba"`
}

func NewEntry(input string) Entry {
	ent := Entry{
		Input:        input,
		Output:       newOutName(input),
		VideoEncoder: "libx265",
		AudioEncoder: "copy",
		CRF:          28,
		Preset:       "faster",
	}
	if tune != "" {
		ent.Tune = tune
	}
	if paramsPreset != "" {
		switch paramsPreset {
		case "2k+":
			ent.X265Params = preset2kp
			break
		default:
			ent.X265Params = preset1080p
		}
	}
	if x265Params != "" {
		ent.X265Params = ent.X265Params + ":" + x265Params
	}
	return ent
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
	flag.BoolVar(&shutdown, "shutdown", false, "-shutdown (shutdown when finished)")
	flag.StringVar(&x265Params, "x265-params", "", "-x265-params (x265 params here)")
	flag.StringVar(&tune, "tune", "", "-tune animation (tune here)")
	flag.StringVar(&paramsPreset, "params-preset", "1080p", "-paramsPreset 1080p (now 1080p and 2k+)")
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
			"-c:a", entry.AudioEncoder,
		}
		// audio config
		if entry.BA != "" {
			args = append(args, "-b:a", entry.BA)
		}

		// video encoder
		args = append(args, "-c:v", entry.VideoEncoder)
		if entry.Preset != "" && entry.VideoEncoder != "copy" {
			args = append(args, "-preset", entry.Preset)
		}
		if entry.VideoEncoder != "hevc_nvenc" && entry.VideoEncoder != "copy" {
			args = append(args, "-crf", fmt.Sprintf("%d", entry.CRF))
		}
		if entry.X265Params != "" && entry.VideoEncoder == "libx265" {
			args = append(args, "-x265-params", entry.X265Params)
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
		fmt.Println(cmd.Args)

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
