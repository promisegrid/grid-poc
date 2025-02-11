package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {
	// Read slides.md file.
	mdBytes, err := ioutil.ReadFile("slides.md")
	if err != nil {
		log.Fatalf("Error reading slides.md: %v", err)
	}
	mdContent := string(mdBytes)

	// Extract the title from slides.md.
	// The title is assumed to be the content of the first line starting with "#"
	scanner := bufio.NewScanner(strings.NewReader(mdContent))
	var title string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			// Remove the leading '#' and any surrounding whitespace.
			title = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			break
		}
	}
	if title == "" {
		title = "Slides"
	}

	// Read the slides template (slides.thtml)
	tmplBytes, err := ioutil.ReadFile("slides.thtml")
	if err != nil {
		log.Fatalf("Error reading slides.thtml: %v", err)
	}
	tmplContent := string(tmplBytes)

	// Parse the template.
	tmpl, err := template.New("slides").Parse(tmplContent)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// Prepare a template data struct.
	data := struct {
		Title  string
		Slides string
	}{
		Title:  title,
		Slides: mdContent,
	}

	// Execute the template with the provided data.
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	// Write the executed template to slides.html.
	if err := ioutil.WriteFile("slides.html", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Error writing slides.html: %v", err)
	}

	// Inform the user of the generated file.
	fmt.Println("Generated slides.html successfully.")

	// Set up HTTP server to serve slides.html on localhost:8192.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "slides.html")
	})

	fmt.Println("Serving slides.html on http://localhost:8192")
	if err := http.ListenAndServe(":8192", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
