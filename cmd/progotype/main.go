package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// CONFIG:
// rootPath is your web server folder
// defaultFolder is the root index server
// src is the folder where your assets are

const (
	serverHost    string = "localhost"
	serverPort    int    = 8080
	rootPath      string = "../../server"
	defaultFolder string = "index"
	src           string = "src"
	rateLimit     int    = 20
	rateBurst     int    = 10
)

// Change nothing below this line
func rateLimitMiddleware(next http.Handler) http.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rateLimit), rateBurst)

	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.Header().Set("Retry-After", strconv.FormatInt(time.Now().Add(limiter.Reserve().Delay()).Unix(), 10))
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Vary", "User-Agent")
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func notFoundMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested file exists
		_, err := os.Stat(filepath.Join(rootPath, r.URL.Path))
		if os.IsNotExist(err) {
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func main() {

	// Parse the base template
	baseTemplate, err := template.ParseFiles(filepath.Join(rootPath, "layout.html"))
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	// Register the rate limit middleware for the root route
	mux.HandleFunc("/", notFoundMiddleware(rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmplPath := fmt.Sprintf("%s/%s.html", defaultFolder, defaultFolder)
		tmpl, err := template.Must(baseTemplate.Clone()).ParseFiles(filepath.Join(rootPath, tmplPath))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))))

	mux.Handle("/"+src+"/", http.StripPrefix("/"+src+"/", http.FileServer(http.Dir(filepath.Join(rootPath, src)))))

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != rootPath && filepath.Base(path) != src {
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}

			// Parse the template for this folder
			templatePath := filepath.Join(path, defaultFolder+".html")
			tmpl, err := template.Must(baseTemplate.Clone()).ParseFiles(templatePath)
			if err != nil {
				return err
			}

			// Serve the template
			mux.HandleFunc("/"+relPath+"/", func(w http.ResponseWriter, r *http.Request) {
				err := tmpl.Execute(w, nil)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})
		}
		return filepath.SkipDir
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Server started on http://%s:%d\n", serverHost, serverPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", serverHost, serverPort), mux))
}
