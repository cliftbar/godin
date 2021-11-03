---
title: "Go/Wasm <-> JS Memory"
date: 2021-11-02T20:56:45-07:00
draft: false
tags:
    - "30 Minute Post"
---

This post will be about how I accessed the Go/Wasm memory buffer from the Javascript side.  I wanted to get this one out, because it took me a very long time to work through it, and I had to piece together a ton of random internet resources.  All the code is pulled from the site, but I've simplified it the best I can.

# Go/Wasm instantiation in JS
When instantiating the Go/Wasm "application" in JS-land, its important to keep a reference to the WebAssembly object used in that process.  It gets used to access the memory buffer later

```html
<script>
    let wasmRef;  // Need to keep track of this
    const go = new Go();
    async function init() {
        wasmRef = await WebAssembly.instantiateStreaming(fetch("/wasm/lib.wasm"), go.importObject)
        await go.run(wasmRef.instance);
    }
    init();
</script>
```

# Go/Wasm store data into buffer, return pointer
Well, explicitly storing data into the buffer isn't required, since the buffer is the entire memory space available to the Go/Wasm code.  But, it needs to be encoded into a format that can be read on the JS side.  In this case, I encode the GeoJSON Object into a byte slice (though I'm going to use a simpler string below).

```go
package main

import (
	"encoding/json"
	"reflect"
	"syscall/js"
	"unsafe"
)

// Pin buffer to global, so it doesn't get GC'd
var wasmMemoryBuffer []byte

func GetString(this js.Value, i []js.Value) interface{} {

	emptyGeoJSON := map[string]interface{}{
		"type": "FeatureCollection",
		"features": make([]interface{}, 0),
	}

	wasmMemoryBuffer, _ = json.Marshal(emptyGeoJSON)
	buffHeader := (*reflect.SliceHeader)(unsafe.Pointer(&wasmMemoryBuffer))

	retMap := map[string]interface{}{
		"ptr": buffHeader.Data,
		"len":  len(wasmMemoryBuffer),
	}

	return retMap
}

func registerCallbacks() {
	js.Global().Set("GetString", js.FuncOf(GetString))
}

func main(){
	c := make(chan struct{}, 0)

	// register functions
	registerCallbacks()
	<-c
}
```

The major steps here are:
1. Keep the reference to the `wasmMemoryBuffer` byte slice global, so it doesn't get garbage collected unexpectedly.  This will require explicit functions to clear out `wasmMemoryBuffer` if you want to deallocate it (which is just a function that sets it to an empty slice)
2. The simple GeoJSON object is turned into a byte slice using the standard `json.Marshal` function
3. This is the odd bit of pointer magic, and it requires knowing what the `slice` type is under the hood ([first reasonable link](https://faun.pub/slices-in-golang-introduction-4b11ac451049) I found).  The upshot is this line, `buffHeader := (*reflect.SliceHeader)(unsafe.Pointer(&wasmMemoryBuffer))`, gets you the slice header for `wasmMemoryBuffer`, and the pointer to the backing array is accessible at `buffHeader.Data`
4. `retMap` is just a basic map to cleanly package up both the pointer to our data and the length of our data, we'll need both on the JS side.

# Read from buffer on JS side
On the JS side, we get the pointer information, and use them to pull the data out of the correct section of WebAssembly memory.

```js
let ptrMap = GetString();

let wasmBuffer = wasmRef.instance.exports.mem.buffer;
let dataSection = wasmBuffer.slice(ptrMap.ptr, ptrMap.ptr + ptrMap.len);
let decodedGeoJSON = new TextDecoder("utf-8").decode(wasmBuffer);

let GeoJSONObjects = JSON.parse(decodedGeoJSON)
```

Here, we:
1. Call the Go/Wasm function to get the pointer and data length
2. Get a reference to the Wasm memory area (some online posts have this as `wasmRef.instance.exports.memory.buffer`, note memory vs mem, I'm not sure if the API changed or something)
3. use the pointer and data length to get the correct section of bytes out of the memory.  The datatype here is an Array of Uint8 values.
4. Finally, use a TextDecode class to decode the Uint8 values, and run that through JSON.parse() to get the objects out.

# Wrap Up
That's the key parts! I didn't address getting the Go code compiled to Wasm, the general structure of the code, nor any of the mapping stuff.  I think the biggest thing I was tripped up on was the amount old information floating around in web resources, many were several years old which definitely mattered.  The other big gotcha was that, initially, I was returning the pointer to the slice `wasmMemoryBuffer` _itself_ (`&wasmMemoryBuffer`), rather than the pointer to the slice _backing array_ (`buffHeader.Data`), which was very confusing as you end up in an inscrutable memory area if you do that.

I'm not sure, at the data size I'm doing now, that this technique matters too much over just returning the string and eating the copy going from Go/Wasm to JS (if there's a copy?  I don't know that for sure, information is hard to find).  But, this method has some neat synergies with using ProtoBuff instead of a string to encode the data, since ProtoBuff is a binary format it can encode objects directly to and from the memory space, without needing an intermediary string decode or json parse.