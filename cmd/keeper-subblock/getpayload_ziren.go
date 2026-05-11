//go:build ziren

package main

import (
	zkruntime "github.com/ProjectZKM/Ziren/crates/go-runtime/zkvm_runtime"
)

func getInput() []byte {
	return zkruntime.Read[[]byte]()
}

