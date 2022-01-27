package deadweb

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

//https://www.alexedwards.net/blog/serving-static-sites-with-go

func Server() {

	// fs := http.FileServer(http.Dir("./static"))
	// http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", serveMarkdownOrStatic)

	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func serveMarkdownOrStatic(rw http.ResponseWriter, r *http.Request) {
	//lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join(".", filepath.Clean(r.URL.Path))

	// Return a 404 if the template doesn't exist
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(rw, r)
			return
		}
	}

	// Return a 404 if the request is for a directory
	if info.IsDir() {
		http.NotFound(rw, r)
		return
	}

	raw, err := ioutil.ReadFile(fp)
	if err != nil {
		log.Println(err.Error())
		http.Error(rw, http.StatusText(500), 500)
		return
	}

	if strings.HasSuffix(fp, ".md") {
		var md bytes.Buffer
		if err := goldmark.Convert(raw, &md); err != nil {
			panic(err)
		}
		_, _ = rw.Write(md.Bytes())
	} else {
		_, _ = rw.Write(raw)
	}

}
