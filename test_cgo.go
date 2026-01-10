//go:build ignore

package main

/*
#cgo CFLAGS: -I./vendor/whisper.cpp/include -I./vendor/whisper.cpp/ggml/include
#cgo LDFLAGS: -L./vendor/whisper.cpp/build/src -L./vendor/whisper.cpp/build/ggml/src -L./vendor/whisper.cpp/build/ggml/src/ggml-blas -L./vendor/whisper.cpp/build/ggml/src/ggml-metal
#cgo LDFLAGS: -Wl,-rpath,./vendor/whisper.cpp/build/src -Wl,-rpath,./vendor/whisper.cpp/build/ggml/src -Wl,-rpath,./vendor/whisper.cpp/build/ggml/src/ggml-blas -Wl,-rpath,./vendor/whisper.cpp/build/ggml/src/ggml-metal
#cgo LDFLAGS: -lwhisper -lggml -lggml-base -lggml-cpu -lm -lstdc++
#cgo darwin LDFLAGS: -lggml-metal -lggml-blas
#cgo darwin LDFLAGS: -framework Accelerate -framework Metal -framework Foundation -framework CoreGraphics
#include <whisper.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("Testing Whisper.cpp CGO integration...")

	// Test basic whisper functions
	fmt.Printf("Whisper sample rate: %d\n", int(C.WHISPER_SAMPLE_RATE))
	fmt.Printf("System info: %s\n", C.GoString(C.whisper_print_system_info()))

	// Test with a dummy model path
	modelPath := C.CString("models/nonexistent.bin")
	defer C.free(unsafe.Pointer(modelPath))

	// This should fail gracefully since the model doesn't exist
	ctx := C.whisper_init_from_file_with_params(modelPath, C.whisper_context_default_params())
	if ctx == nil {
		fmt.Println("✓ CGO integration working - model loading failed as expected (model doesn't exist)")
	} else {
		fmt.Println("✓ CGO integration working - model loaded successfully")
		C.whisper_free(ctx)
	}

	fmt.Println("CGO compilation and linking successful!")
}