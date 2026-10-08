// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

//go:build ziren

package trie

import (
	"fmt"
	"unsafe"

	"github.com/ethereum/go-ethereum/rlp"
)

// hashNodeTab is the type word of a node interface holding a hashNode, taken
// from a value Go boxed itself.
var hashNodeTab = func() unsafe.Pointer {
	var n node = hashNode(nil)
	return (*[2]unsafe.Pointer)(unsafe.Pointer(&n))[0]
}()

// boxHashNode returns *h as a node whose data word points at h, as Go's own
// conversion would point at a heap copy; *h must not be modified afterwards.
func boxHashNode(h *hashNode) node {
	var n node
	w := (*[2]unsafe.Pointer)(unsafe.Pointer(&n))
	w[0], w[1] = hashNodeTab, unsafe.Pointer(h)
	return n
}

// decodeFull is the decodeFull of node_full.go with the hash children boxed
// into one array per node. Converting each hash child to the node interface
// allocates a slice header on its own, up to 16 per node and the largest
// source of allocations in a stateless block on 32-bit MIPS. A child that is
// not a 32-byte hash goes through decodeRef as before.
func decodeFull(hash, elems []byte) (*fullNode, error) {
	n := &fullNode{flags: nodeFlag{hash: hash}}
	var refs *[16]hashNode
	for i := 0; i < 16; i++ {
		if kind, val, rest, err := rlp.Split(elems); err == nil && kind == rlp.String && len(val) == hashLen {
			if refs == nil {
				refs = new([16]hashNode)
			}
			refs[i] = hashNode(val)
			n.Children[i], elems = boxHashNode(&refs[i]), rest
			continue
		}
		cld, rest, err := decodeRef(elems)
		if err != nil {
			return n, wrapError(err, fmt.Sprintf("[%d]", i))
		}
		n.Children[i], elems = cld, rest
	}
	val, _, err := rlp.SplitString(elems)
	if err != nil {
		return n, err
	}
	if len(val) > 0 {
		n.Children[16] = valueNode(val)
	}
	return n, nil
}
