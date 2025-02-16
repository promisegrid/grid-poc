//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
)

func main() {
	// Get the global document object.
	document := js.Global().Get("document")

	// Create a button element.
	button := document.Call("createElement", "button")
	button.Set("innerHTML", "Click me")

	// Append the button to the body.
	document.Get("body").Call("appendChild", button)

	// Define the click event handler.
	clickHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		p := document.Call("createElement", "p")
		p.Set("innerHTML", "Hello, WASM world!")
		document.Get("body").Call("appendChild", p)
		return nil
	})
	// It's important to release the function when not used.
	defer clickHandler.Release()

	// Add the click event listener to the button.
	button.Call("addEventListener", "click", clickHandler)

	// Prevent the Go program from exiting.
	select {}
}
