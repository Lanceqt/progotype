// don't change anything in this file if you don't know what you're doing
// It will break the server if you change anything in this file
package main

import (
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// CONFIG:
	// rootPath is your web server folder
	// defaultFolder is the root index server
	// src is the folder where your assets are
	rootPath := "../../server"
	defaultFolder := "index"
	src := "src"

	// Parse the base template
	baseTemplate, err := template.ParseFiles(filepath.Join(rootPath, "layout.html"))
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/"+src+"/", func(w http.ResponseWriter, r *http.Request) {
		// Get the requested file path
		filePath := filepath.Join(rootPath, r.URL.Path)

		// Check if the requested file exists
		fileInfo, err := os.Stat(filePath)
		if err != nil || fileInfo.IsDir() {
			// File does not exist or is a directory, return a 404 error
			http.NotFound(w, r)
			return
		}

		// Serve the requested file
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType != "" {
			// Set the content type for known file types
			w.Header().Set("Content-Type", contentType)
		}
		http.ServeFile(w, r, filePath)
	})

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
			templatePath := filepath.Join(path, filepath.Base(path)+".html")
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.Must(baseTemplate.Clone()).ParseFiles(filepath.Join(rootPath, defaultFolder+"/"+defaultFolder+".html"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", mux))
}
