package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

func handle(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
}

type Config struct {
	Dir string
}

func main() {
	rawConfig, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatal(err)
	}
	var c Config
	if err := yaml.Unmarshal(rawConfig, &c); err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.ParseGlob("*.html")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handle(func(w http.ResponseWriter, r *http.Request) error {
		var message string
		if r.Method == http.MethodPost {

			file, handler, err := r.FormFile("file")
			if err != nil {
				return err
			}
			defer file.Close()

			basename := time.Now().Format(time.RFC3339) + "-" + filepath.Base(handler.Filename)
			filename := filepath.Join(c.Dir, basename)
			f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				return err
			}
			defer f.Close()
			io.Copy(f, file)

			message = fmt.Sprintf("Uploaded %s!", handler.Filename)
		}

		if err := tmpl.ExecuteTemplate(w, "index.html", message); err != nil {
			return err
		}

		return nil
	}))

	if err := cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bits := strings.Split(r.URL.Path, "elections.cgi")
		if len(bits) == 2 {
			r.URL.Path = bits[1]
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
		}
		mux.ServeHTTP(w, r)
	})); err != nil {
		fmt.Println(err)
	}
}
