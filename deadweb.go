package deadweb

import (
	"bytes"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type fileServer struct {
	xfs      fs.FS
	tmp      *template.Template
	markdown goldmark.Markdown
}

var masterHtmlPath = filepath.Join("dead-down", "master.html")

func FileServer(fs fs.FS, parseEvery bool) (http.Handler, error) {

	var tmpl *template.Template
	var err error

	if !parseEvery {
		tmpl, err = template.ParseFS(fs, masterHtmlPath)
		if err != nil {
			return nil, err
		}
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
			//html.WithXHTML(),
		),
	)

	x := &fileServer{xfs: fs, tmp: tmpl, markdown: md}
	return x, nil
}

func plainFileExists(xfs fs.FS, path string) bool {
	info, err := fs.Stat(xfs, path)
	if os.IsNotExist(err) || err != nil {
		return false
	}

	return !info.IsDir()
}

func (x *fileServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	fp := filepath.Clean(filepath.Join(".", filepath.Clean(r.URL.Path)))

	//println(fp)

	var tmpl *template.Template
	var err error

	if x.tmp == nil {

		tmpl, err = template.ParseFS(x.xfs, masterHtmlPath)
		if err != nil {
			log.Println(err.Error())
			http.Error(rw, "failed template.ParseFS", 500)
			return
		}

	} else {
		tmpl = x.tmp
	}

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

	if info.IsDir() {
		a := filepath.Join(fp, "readme.md") // priority
		b := filepath.Join(fp, "index.html")
		if plainFileExists(x.xfs, a) {
			fp = a
		} else if plainFileExists(x.xfs, b) {
			fp = b
		} else {
			http.NotFound(rw, r)
			return
		}
	}

	mdExt := strings.HasSuffix(fp, ".md")
	htmlExt := strings.HasSuffix(fp, ".html")
	if mdExt || htmlExt {
		raw, err := fs.ReadFile(x.xfs, fp)
		if err != nil {
			log.Println(err.Error())
			http.Error(rw, http.StatusText(500), 500)
			return
		}
		if mdExt {
			var buf bytes.Buffer

			if err := x.markdown.Convert(raw, &buf); err != nil {
				log.Println(err.Error())
				http.Error(rw, http.StatusText(500), 500)
			}
			raw = buf.Bytes()
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

		mt := mime.TypeByExtension(filepath.Ext(fp))
		//			println(222,fp,mt)
		if mt != "" {
			rw.Header().Set("Content-Type", mt)
		}

		_, err = io.Copy(rw, f)
		if err != nil {
			log.Println(err.Error())
			http.Error(rw, http.StatusText(500), 500)
			return
		}

	}

}
