# WASM Component Model Hello Demo

This project demonstrates a version of the traditional "Hello" demo that leverages the WASM component model. It uses WebAssembly Interface Types (WIT) to explicitly define the component interfaces and a componentization workflow to transform the standard WASM binary into a component model module.

Below is a list of all files needed for the build and run process, along with one-line descriptions:

- **hello.go**: Go source code implementing the demo logic using syscall/js and structured for component model integration.
- **component.wit**: WIT file defining the component interface (functions, types, and contracts) exposed by the demo.
- **component.yaml**: Manifest file for component customization and configuration; used by the toolchain to produce a WASM component.
- **main.wasm**: The resulting WebAssembly binary after compiling hello.go and applying component model transformations.
- **hello.html**: HTML file that loads the WASM component and provides the UI for triggering the demo.
- **main.js**: JavaScript bootstrap script that instantiates and runs the WASM component in the browser.
- **Makefile**: Build script with targets to compile the Go code, transform the WASM binary into a component using the component model toolchain (e.g., wit-bindgen or wasm-tools), and copy necessary files for running the demo.

## Build and Run Workflow

1. **Compile the Go Source:**  
   Set the environment for WebAssembly and compile `hello.go` into `hello.wasm`:
   ```
   GOOS=js GOARCH=wasm go build -o hello.wasm hello.go
   ```

2. **Componentize the WASM Binary:**  
   Use the WASM component model toolchain in conjunction with `component.wit` and `component.yaml` to generate a componentized version of the module. This process transforms `hello.wasm` into a compliant component according to your project’s interface definitions.

3. **Prepare the Web Assets:**  
   Ensure that `hello.html` and `main.js` are set up to load and instantiate the componentized `main.wasm`.

4. **Run the Demo:**  
   Serve the directory (e.g., via a local HTTP server) and open `hello.html` in a modern web browser that supports WebAssembly and the component model.

Happy coding with WASM components!
