package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"routerdash/internal/routerdash"
)

//go:embed web/build/index.html web/build/robots.txt web/build/favicon.ico web/build/_app
var staticFiles embed.FS
var version = "dev"

func main() {
	port := env("ROUTERDASH_ADDR", ":8080")
	runner := routerdash.NewRunner(os.Getenv("ROUTERDASH_FAKE") == "1")
	collector := routerdash.NewCollector(runner, time.Now)
	server := routerdash.NewServer(collector, staticFS(), version)

	log.Printf("routerdash %s listening on %s", version, port)
	if err := http.ListenAndServe(port, server); err != nil {
		log.Fatal(err)
	}
}

func env(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func staticFS() fs.FS {
	sub, err := fs.Sub(staticFiles, "web/build")
	if err != nil {
		log.Fatal(err)
	}
	return sub
}
