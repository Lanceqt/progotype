package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	rootPath := "../../server"
	mux := http.NewServeMux()
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		mux.HandleFunc("/"+relPath, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path)
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", mux))
}
