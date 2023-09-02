package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// IMPORTANT:
	// Please ensure that the path is correctly specified relative to this file.
	rootPath := "../../server"
	defaultFolder := "index" // Change the folder name here
	//src := "src"
	mux := http.NewServeMux()

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != rootPath {
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}
			mux.Handle("/"+relPath+"/", http.StripPrefix("/"+relPath+"/", http.FileServer(http.Dir(path))))
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	mux.Handle("/", http.FileServer(http.Dir(filepath.Join(rootPath, defaultFolder))))

	fmt.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", mux))
}
