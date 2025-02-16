const go = new Go();
if (WebAssembly.instantiateStreaming) {
  WebAssembly.instantiateStreaming(fetch("hello.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
  });
} else {
  fetch("hello.wasm")
    .then(response => response.arrayBuffer())
    .then(bytes => WebAssembly.instantiate(bytes, go.importObject))
    .then((result) => {
      go.run(result.instance);
    });
}
