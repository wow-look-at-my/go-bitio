package bitio

import (
	"fmt"
)

// BitPos represents a position in a bit stream as total bits.
type BitPos uint64

// NewPosition creates a new BitPos from bytes and bits.
func NewPosition(bytes, bits uint64) BitPos {
	return BitPos(bytes*8 + bits)
}

// FromBits creates a BitPos from a total bit count.
func FromBits(bits uint64) BitPos {
	return BitPos(bits)
}

// FromBytes creates a BitPos from a byte count.
func FromBytes(bytes uint64) BitPos {
	return BitPos(bytes * 8)
}

// Bytes returns the byte component of the position.
func (p BitPos) Bytes() uint64 {
	return uint64(p) / 8
}

// Bits returns the bit component of the position (0-7).
func (p BitPos) Bits() uint8 {
	return uint8(p % 8)
}

// TotalBits returns the total number of bits represented by this position.
func (p BitPos) TotalBits() uint64 {
	return uint64(p)
}

// TotalBytes returns the minimum number of bytes needed to contain this position.
func (p BitPos) TotalBytes() uint64 {
	return (uint64(p) + 7) / 8
}

// IsByteAligned returns true if the position is byte-aligned (bits == 0).
func (p BitPos) IsByteAligned() bool {
	return p%8 == 0
}

// String returns a string representation of the position.
func (p BitPos) String() string {
	return fmt.Sprintf("%d:%d", p.Bytes(), p.Bits())
}

// Add adds a size to a position, returning a new position.
func (p BitPos) Add(s BitSize) BitPos {
	return p + BitPos(s)
}

// Sub subtracts a size from a position, returning a new position.
func (p BitPos) Sub(s BitSize) BitPos {
	return p - BitPos(s)
}

// Diff returns the size between two positions.
func (p BitPos) Diff(other BitPos) BitSize {
	return BitSize(p - other)
}
