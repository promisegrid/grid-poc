# WASM Component Model Hello Demo

This project demonstrates a first draft of a "Hello" demo that leverages the WebAssembly Component Model. The demo defines a well‐structured component interface using WebAssembly Interface Types (WIT) and a component manifest to drive the transformation of a traditional Go WASM binary into a component module.

Below is the list of all files needed to build and run the demo:

- **hello.go**: Go source code that implements the demo logic using syscall/js and exposes a component function (runDemo) to initialize the UI and IndexedDB workflow.
- **hello.wit**: WIT file defining the interface for the component. It declares the exported function that the host can call.
- **hello.yaml**: YAML manifest providing configuration for componentization; used by the WASM component toolchain to transform the WASM binary.
- **main.wasm**: The resulting componentized WebAssembly binary generated from hello.go via the component toolchain.
- **hello.html**: HTML file that loads the WASM component via the JavaScript bootstrap and provides the user interface.
- **main.js**: JavaScript bootstrap script that instantiates the componentized WebAssembly module and invokes the exported function.
- **Makefile**: Build script that compiles the Go source to a WASM binary and then uses a component model tool (e.g. wasm-tools) to produce the componentized main.wasm.

## Build and Run Workflow

1. **Compile the Go Source:**  
   The Makefile compiles `hello.go` into a standard WASM binary:
   ```
   GOOS=js GOARCH=wasm go build -o hello.wasm hello.go
   ```

2. **Componentize the WASM Binary:**  
   Using the provided `hello.wit` and `hello.yaml`, the toolchain converts `hello.wasm` into a component model compliant module (`main.wasm`):
   ```
   wasm-tools component new hello.wasm -c hello.yaml -o main.wasm
   ```

3. **Prepare the Web Assets:**  
   The HTML and JavaScript files are set up to load and instantiate the WASM component. The JS bootstrap calls the exported `runDemo` function to kick off the demo.

4. **Run the Demo:**  
   Serve the project directory (e.g. via a local HTTP server) and navigate to `hello.html` in your browser. The demo will display a button that, when clicked, interacts with IndexedDB and displays a message on the page.

Happy coding with WASM components!
