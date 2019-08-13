package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

var (
	encode string
	decode string
	file   string
)

func init() {
	flag.StringVar(&encode, "encode", "", "-encode HELLO")
	flag.StringVar(&decode, "decode", "", "-decode dm1lc3M6Ly9ZMmho")
	flag.StringVar(&file, "f", "", "-f FILE_NAME")
	flag.Parse()
}

func main() {
	b64 := base64.RawStdEncoding
	if encode != "" {
		fmt.Println("Encoding...")
		fmt.Println(b64.EncodeToString([]byte(encode)))
	}
	if decode != "" {
		fmt.Println("Decoding...")
		data, err := b64.DecodeString(decode)
		if err != nil {
			log.Panic(err)
		}
		if file != "" {
			err := ioutil.WriteFile(file, data, 0666)
			if err != nil {
				log.Panic(err)
			}
		} else {
			fmt.Println(string(data))
		}
	}
}
