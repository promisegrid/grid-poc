const go = new Go();

// Instantiate the componentized WASM module (main.wasm) and then invoke the exported runDemo function.
if (WebAssembly.instantiateStreaming) {
  WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
    // Call the exported function to set up the demo.
    if (typeof runDemo === "function") {
      runDemo();
    }
  });
} else {
  fetch("main.wasm")
    .then(response => response.arrayBuffer())
    .then(bytes => WebAssembly.instantiate(bytes, go.importObject))
    .then((result) => {
      go.run(result.instance);
      if (typeof runDemo === "function") {
        runDemo();
      }
    });
}
