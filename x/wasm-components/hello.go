//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
)

// runDemo is the exported function that initializes the demo.
// It creates a button in the DOM and sets up an IndexedDB demonstration.
func runDemo(this js.Value, args []js.Value) interface{} {
	document := js.Global().Get("document")

	// Create a button element.
	button := document.Call("createElement", "button")
	button.Set("innerHTML", "Store & Retrieve from IndexedDB (Component Model)")

	// Append the button to the body.
	document.Get("body").Call("appendChild", button)

	// Define the click event handler.
	clickHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		indexedDB := js.Global().Get("indexedDB")
		if indexedDB.IsUndefined() {
			p := document.Call("createElement", "p")
			p.Set("innerHTML", "IndexedDB is not supported in this browser.")
			document.Get("body").Call("appendChild", p)
			return nil
		}

		// Open (or create) a database named "MyDatabase" with version 1.
		request := indexedDB.Call("open", "MyDatabase", 1)

		// onupgradeneeded event handler to create an object store.
		var onUpgrade js.Func
		onUpgrade = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			db := request.Get("result")
			db.Call("createObjectStore", "greetings")
			request.Set("onupgradeneeded", js.Undefined())
			onUpgrade.Release()
			return nil
		})
		request.Set("onupgradeneeded", onUpgrade)

		// onsuccess event handler for database open.
		var onSuccess js.Func
		onSuccess = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			db := request.Get("result")
			tx := db.Call("transaction", "greetings", "readwrite")
			store := tx.Call("objectStore", "greetings")

			// Store a value with key "hello".
			putReq := store.Call("put", "Hello, IndexedDB (Component Model)!", "hello")

			// on success for put request.
			var onPutSuccess js.Func
			onPutSuccess = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				getReq := store.Call("get", "hello")
				var onGetSuccess js.Func
				onGetSuccess = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					val := getReq.Get("result").String()
					p := document.Call("createElement", "p")
					p.Set("innerHTML", "IndexedDB Value: "+val)
					document.Get("body").Call("appendChild", p)
					getReq.Set("onsuccess", js.Undefined())
					onGetSuccess.Release()
					return nil
				})
				getReq.Set("onsuccess", onGetSuccess)
				putReq.Set("onsuccess", js.Undefined())
				onPutSuccess.Release()
				return nil
			})
			putReq.Set("onsuccess", onPutSuccess)
			request.Set("onsuccess", js.Undefined())
			onSuccess.Release()
			return nil
		})
		request.Set("onsuccess", onSuccess)

		return nil
	})
	button.Call("addEventListener", "click", clickHandler)

	return nil
}

func main() {
	// Expose the runDemo function to be callable from JavaScript.
	js.Global().Set("runDemo", js.FuncOf(runDemo))
	// Keep the Go program running.
	select {}
}
