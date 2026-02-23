package bitio

import (
	"fmt"
)

// BitSize represents a size/length in bits.
type BitSize uint64

// NewSize creates a new BitSize from bytes and bits.
func NewSize(bytes, bits uint64) BitSize {
	return BitSize(bytes*8 + bits)
}

// Bytes returns the byte component of the size.
func (s BitSize) Bytes() uint64 {
	return uint64(s) / 8
}

// Bits returns the bit component of the size (0-7).
func (s BitSize) Bits() uint8 {
	return uint8(s % 8)
}

// TotalBits returns the total number of bits.
func (s BitSize) TotalBits() uint64 {
	return uint64(s)
}

// TotalBytes returns the minimum number of bytes needed to contain this size.
func (s BitSize) TotalBytes() uint64 {
	return (uint64(s) + 7) / 8
}

// IsByteAligned returns true if the size is byte-aligned.
func (s BitSize) IsByteAligned() bool {
	return s%8 == 0
}

// String returns a string representation of the size.
func (s BitSize) String() string {
	return fmt.Sprintf("%d:%d", s.Bytes(), s.Bits())
}
