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

package vm

// codeBitmapInternal is the codeBitmapInternal of analysis_bitmap.go with a
// 32-bit program counter. On 32-bit MIPS every uint64 compare, increment,
// index and shift of the generic loop is a multi-instruction sequence, and the
// analysis runs once for every distinct contract code a block calls. Code is
// far shorter than 2^32 bytes, and the bit operations are those of the generic
// loop, so the bitmap is the same.
func codeBitmapInternal(code, bits BitVec) BitVec {
	for pc := uint32(0); pc < uint32(len(code)); {
		op := OpCode(code[pc])
		pc++
		if int8(op) < int8(PUSH1) { // If not PUSH (the int8(op) > int(PUSH32) is always false).
			continue
		}
		numbits := op - PUSH1 + 1
		if numbits >= 8 {
			for ; numbits >= 16; numbits -= 16 {
				a := byte(0xFF << (pc % 8))
				bits[pc/8] |= a
				bits[pc/8+1] = 0xFF
				bits[pc/8+2] = ^a
				pc += 16
			}
			for ; numbits >= 8; numbits -= 8 {
				a := byte(0xFF << (pc % 8))
				bits[pc/8] |= a
				bits[pc/8+1] = ^a
				pc += 8
			}
		}
		if numbits == 0 {
			continue
		}
		// set1 and setN with the masks set2BitsMask .. set7BitsMask
		a := (uint16(1)<<numbits - 1) << (pc % 8)
		bits[pc/8] |= byte(a)
		if b := byte(a >> 8); b != 0 {
			bits[pc/8+1] = b
		}
		pc += uint32(numbits)
	}
	return bits
}
