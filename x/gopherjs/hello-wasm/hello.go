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
	button.Set("innerHTML", "Store & Retrieve from IndexedDB")

	// Append the button to the body.
	document.Get("body").Call("appendChild", button)

	// Define the click event handler.
	clickHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Get IndexedDB from the global object.
		indexedDB := js.Global().Get("indexedDB")
		if indexedDB.IsUndefined() {
			p := document.Call("createElement", "p")
			p.Set("innerHTML", "IndexedDB is not supported in this browser.")
			document.Get("body").Call("appendChild", p)
			return nil
		}

		// Open (or create) a database called "MyDatabase" with version 1.
		request := indexedDB.Call("open", "MyDatabase", 1)

		// onupgradeneeded event handler to create the object store.
		var onUpgrade js.Func
		onUpgrade = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// Get the database from the request.
			db := request.Get("result")
			// Create an object store named "greetings".
			db.Call("createObjectStore", "greetings")
			// Clean up: remove the upgrade handler.
			request.Set("onupgradeneeded", js.Undefined())
			onUpgrade.Release()
			return nil
		})
		request.Set("onupgradeneeded", onUpgrade)

		// onsuccess event handler when the database is opened.
		var onSuccess js.Func
		onSuccess = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			db := request.Get("result")
			// Start a read-write transaction on the "greetings" object store.
			tx := db.Call("transaction", "greetings", "readwrite")
			store := tx.Call("objectStore", "greetings")

			// Store a value "Hello, IndexedDB" with key "hello".
			putReq := store.Call("put", "Hello, IndexedDB", "hello")

			// Handler to be called when the put request succeeds.
			var onPutSuccess js.Func
			onPutSuccess = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				// Now get the stored value back.
				getReq := store.Call("get", "hello")
				var onGetSuccess js.Func
				onGetSuccess = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					val := getReq.Get("result").String()
					// Append the retrieved value to the document as a paragraph.
					p := document.Call("createElement", "p")
					p.Set("innerHTML", "IndexedDB Value: "+val)
					document.Get("body").Call("appendChild", p)
					// Cleanup the get-success handler.
					getReq.Set("onsuccess", js.Undefined())
					onGetSuccess.Release()
					return nil
				})
				getReq.Set("onsuccess", onGetSuccess)
				// Cleanup the put-success handler.
				putReq.Set("onsuccess", js.Undefined())
				onPutSuccess.Release()
				return nil
			})
			putReq.Set("onsuccess", onPutSuccess)
			// Cleanup the onsuccess handler for the open request.
			request.Set("onsuccess", js.Undefined())
			onSuccess.Release()
			return nil
		})
		request.Set("onsuccess", onSuccess)

		return nil
	})
	// Add the click event listener to the button.
	button.Call("addEventListener", "click", clickHandler)
	// Keep the Go program running.
	select {}
}
