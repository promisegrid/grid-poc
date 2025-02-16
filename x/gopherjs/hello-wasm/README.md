# WebAssembly Hello Demo

This demo runs a simple "Hello" example in a web browser using Go’s native WebAssembly support. The demo creates a button that, when clicked, appends a paragraph to the document. Below is the list of files needed to build and run the demo:

- **hello.go**: Go source code that uses the `syscall/js` package to interact with the browser’s DOM.
- **wasm_exec.js**: JavaScript support file provided by Go (found in `$GOROOT/misc/wasm/`) to bootstrap the WebAssembly runtime.
- **hello.wasm**: Compiled WebAssembly binary generated from `hello.go`.
- **hello.html**: HTML file that loads `wasm_exec.js` and `hello.wasm` to execute the demo in the browser.
- **Makefile**: Build script that sets the necessary environment variables (e.g., `GOOS=js`, `GOARCH=wasm`) and compiles the Go source into the WebAssembly binary.

## Building and Running the Demo

1. **Prepare the Environment**:  
   Ensure you have Go installed with WebAssembly support. Copy `wasm_exec.js` from your Go installation (typically at `$GOROOT/misc/wasm/wasm_exec.js`) into this project directory.

2. **Build the Demo**:  
   Run `make` in the project directory. The Makefile uses the commands:
   ```
   GOOS=js GOARCH=wasm go build -o hello.wasm hello.go
   ```
   to compile the source code into `hello.wasm`.

3. **Run the Demo**:  
   Open `hello.html` in any modern web browser that supports WebAssembly. The HTML file loads `wasm_exec.js` and the WebAssembly module, and then executes the demo.

Happy coding!
