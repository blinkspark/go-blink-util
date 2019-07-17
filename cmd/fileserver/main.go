package main

import (
	"flag"
	"fmt"
	"net/http"

	util "github.com/blinkspark/go-blink-util"
)

func main() {
	path := flag.String("path", "", "absolute path you want to serve")
	port := flag.Int("port", 8080, "port of file server")
	flag.Parse()
	http.Handle("/", http.FileServer(http.Dir(*path)))
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	err := http.ListenAndServe(addr, nil)
	util.CheckErr(err)
}
