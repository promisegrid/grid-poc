package main

import (
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	// Get the global document object.
	document := js.Global.Get("document")

	// Create a button element.
	button := document.Call("createElement", "button")
	button.Set("innerHTML", "Click me")

	// Append the button to the body.
	document.Get("body").Call("appendChild", button)

	// Add a click event listener that creates and appends a paragraph.
	button.Call("addEventListener", "click", func() {
		p := document.Call("createElement", "p")
		p.Set("innerHTML", "Hello, big world!")
		document.Get("body").Call("appendChild", p)
	})
}
