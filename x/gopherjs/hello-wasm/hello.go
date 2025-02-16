package main

import (
	"syscall/js"
)

func main() {
	// Create a channel to prevent the function from exiting.
	done := make(chan struct{}, 0)

	// Get the global document object.
	document := js.Global().Get("document")

	// Create a button element.
	button := document.Call("createElement", "button")
	button.Set("innerHTML", "Click me")
	document.Get("body").Call("appendChild", button)

	// Define the callback function for the "click" event.
	clickCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Create a paragraph element.
		p := document.Call("createElement", "p")
		p.Set("innerHTML", "Hello, big world!")
		document.Get("body").Call("appendChild", p)
		return nil
	})
	// Ensure the callback is released when not needed.
	defer clickCallback.Release()

	// Add the click event listener to the button.
	button.Call("addEventListener", "click", clickCallback)

	// Block forever so that the WASM module stays alive.
	<-done
}
