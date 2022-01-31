package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cameronelliott/deadweb"
	"github.com/spf13/pflag"
)


var dir = pflag.StringP("dir","d",".","directory to serve")

func main() {

	pflag.Parse()
	
	log.SetFlags(log.Lshortfile)

	println(*dir)
	h, err := deadweb.FileServer(os.DirFS(*dir), true)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on :8080...")

	err = http.ListenAndServe(":8080", h)
	log.Fatal(err)

}
