package deadweb

import (
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"github.com/yuin/goldmark"
)

type fileServer struct {
	xfs fs.FS
	tmp *template.Template
}

var lp = filepath.Join("templates", "master.html")

func FileServer(fs fs.FS, parseEvery bool) (http.Handler, error) {

	var tmpl *template.Template
	var err error

	if !parseEvery {
		tmpl, err = template.ParseFS(fs, lp)
		if err != nil {
			return nil, err
		}
	}

	x := &fileServer{xfs: fs, tmp: tmpl}

	return x, nil
}

func (x *fileServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	var tmpl *template.Template
	var err error

	if x.tmp == nil {

		tmpl, err = template.ParseFS(x.xfs, lp)
		if err != nil {
			log.Println(err.Error())
			http.Error(rw, "failed template.ParseFS", 500)
			return
		}

	} else {
		tmpl = x.tmp
	}

	fp := filepath.Join(".", filepath.Clean(r.URL.Path))

	// Return a 404 if the template doesn't exist
	info, err := fs.Stat(x.xfs, fp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(rw, r)
			return
		}
		log.Println(err.Error())
		http.Error(rw, http.StatusText(500), 500)
		return
	}

	// Return a 404 if the request is for a directory
	if info.IsDir() {
		http.NotFound(rw, r)
		return
	}

	if strings.HasSuffix(fp, ".md") {
		raw, err := fs.ReadFile(x.xfs, fp)
		if err != nil {
			log.Println(err.Error())
			http.Error(rw, http.StatusText(500), 500)
			return
		}
		if err := goldmark.Convert(raw, rw); err != nil {
			panic(err)
		}
	} else if strings.HasSuffix(fp, ".html") {
		raw, err := fs.ReadFile(x.xfs, fp)
		if err != nil {
			log.Println(err.Error())
			http.Error(rw, http.StatusText(500), 500)
			return
		}

		err = tmpl.Execute(rw, string(raw))
		if err != nil {
			log.Println(err.Error())
			http.Error(rw, http.StatusText(500), 500)
			return
		}

	} else {
		f, err := x.xfs.Open(fp)
		if err != nil {
			log.Println(err.Error())
			http.Error(rw, http.StatusText(500), 500)
			return
		}

		_, err = io.Copy(rw, f)
		if err != nil {
			log.Println(err.Error())
			http.Error(rw, http.StatusText(500), 500)
			return
		}

	}

}
