package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"support/handler"
)

func removePHP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, ".php") {
			originalQuery := r.URL.RawQuery
			newPath := strings.Replace(r.URL.Path, ".php", "", -1)
			if originalQuery != "" {
				newPath += "?" + originalQuery
			}
			http.Redirect(w, r, newPath, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.IndexHandler)
	mux.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/doc", handler.DocHandler)
	mux.HandleFunc("/search", handler.SearchHandler)

	fmt.Println("Server running on http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", removePHP(mux)))
}
