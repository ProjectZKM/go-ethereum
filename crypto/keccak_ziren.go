// Copyright 2025 The go-ethereum Authors
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

package crypto

import (
	"unsafe"

	"github.com/ProjectZKM/Ziren/crates/go-runtime/zkvm_runtime"
	"github.com/ethereum/go-ethereum/common"
)

// zirenKeccak256 hashes the concatenation of chunks with one KECCAK_SPONGE
// syscall. The syscall reads blocks of 36 words, 34 words of rate (136 bytes)
// followed by two zero words; on little-endian mipsle a block's rate bytes are
// its words, so the message is copied into place rather than packed word by
// word, and the digest words are its bytes.
func zirenKeccak256(chunks ...[]byte) (out [32]byte) {
	n := 0
	for _, c := range chunks {
		n += len(c)
	}
	blocks := n/136 + 1
	// one allocation: the sponge input followed by the 17-word result, whose
	// address would otherwise escape to the heap on its own
	words := make([]uint32, blocks*36+17)
	res := words[blocks*36:]
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&words[0])), blocks*36*4)
	pos := 0
	for _, c := range chunks {
		for len(c) > 0 {
			blk, off := pos/136, pos%136
			k := copy(buf[blk*144+off:blk*144+136], c)
			c = c[k:]
			pos += k
		}
	}
	last := (blocks - 1) * 144
	buf[last+n%136] ^= 0x01
	buf[last+135] ^= 0x80
	res[16] = uint32(blocks * 36)
	zkvm_runtime.SyscallKeccakSponge(unsafe.Pointer(&words[0]), unsafe.Pointer(&res[0]))
	*(*[8]uint32)(unsafe.Pointer(&out)) = *(*[8]uint32)(unsafe.Pointer(&res[0]))
	return out
}

// zirenKeccakState implements the KeccakState interface using the Ziren zkvm_runtime.
// It accumulates data written to it and uses the zkvm's Keccak256 system call for hashing.
type zirenKeccakState struct {
	buf    []byte   // accumulated data
	sum    [32]byte // cached result, valid when hashed
	hashed bool     // whether sum holds the hash of buf
}

func newZirenKeccakState() KeccakState {
	return &zirenKeccakState{
		buf: make([]byte, 0, 512), // pre-allocate reasonable capacity
	}
}

func (s *zirenKeccakState) Write(p []byte) (n int, err error) {
	s.buf = append(s.buf, p...)
	s.hashed = false
	return len(p), nil
}

func (s *zirenKeccakState) Sum(b []byte) []byte {
	s.computeHashIfNeeded()
	return append(b, s.sum[:]...)
}

func (s *zirenKeccakState) Reset() {
	s.buf = s.buf[:0]
	s.hashed = false
}

func (s *zirenKeccakState) Size() int {
	return 32
}

func (s *zirenKeccakState) BlockSize() int {
	return 136 // Keccak256 rate
}

func (s *zirenKeccakState) Read(p []byte) (n int, err error) {
	s.computeHashIfNeeded()

	if len(p) == 0 {
		return 0, nil
	}

	n = copy(p, s.sum[:])
	return n, nil
}

func (s *zirenKeccakState) computeHashIfNeeded() {
	if !s.hashed {
		s.sum = zirenKeccak256(s.buf)
		s.hashed = true
	}
}

// NewKeccakState creates a new KeccakState
// This uses a Ziren-optimized implementation that leverages the zkvm_runtime.Keccak256 system call.
func NewKeccakState() KeccakState {
	return newZirenKeccakState()
}

// Keccak256 calculates and returns the Keccak256 hash using the Ziren zkvm_runtime implementation.
func Keccak256(data ...[]byte) []byte {
	result := zirenKeccak256(data...)
	return result[:]
}

// Keccak256Hash calculates and returns the Keccak256 hash as a Hash using the Ziren zkvm_runtime implementation.
func Keccak256Hash(data ...[]byte) common.Hash {
	return common.Hash(zirenKeccak256(data...))
}
