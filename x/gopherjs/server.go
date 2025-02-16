package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	// docroot as arg1
	docroot := os.Args[1]

	// Define the address flag to allow setting the listening address.
	addr := flag.String("addr", "127.0.0.1:8736", "HTTP network address")
	flag.Parse()

	// Create a file server handler to serve files from the current directory.
	// It will serve hello.html as the default page when accessing "/".
	fs := http.FileServer(http.Dir(docroot))
	http.Handle("/", fs)

	log.Printf("Serving GopherJS demo on http://%s/", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
