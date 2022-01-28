package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cameronelliott/deadweb"
)

func main() {

	h, err := deadweb.FileServer(os.DirFS("."), true)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on :8080...")

	err = http.ListenAndServe(":8080", h)
	log.Fatal(err)

}
