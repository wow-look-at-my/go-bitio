package bitio

import (
	"fmt"
)

// Position represents a position in a bit stream as total bits.
type Position uint64

// Zero is the zero position.
const Zero Position = 0

// NewPosition creates a new Position from bytes and bits.
func NewPosition(bytes uint64, bits uint8) Position {
	return Position(bytes*8 + uint64(bits))
}

// FromBits creates a Position from a total bit count.
func FromBits(bits uint64) Position {
	return Position(bits)
}

// FromBytes creates a Position from a byte count.
func FromBytes(bytes uint64) Position {
	return Position(bytes * 8)
}

// Bytes returns the byte component of the position.
func (p Position) Bytes() uint64 {
	return uint64(p) / 8
}

// Bits returns the bit component of the position (0-7).
func (p Position) Bits() uint8 {
	return uint8(p % 8)
}

// TotalBits returns the total number of bits represented by this position.
func (p Position) TotalBits() uint64 {
	return uint64(p)
}

// TotalBytes returns the minimum number of bytes needed to contain this position.
func (p Position) TotalBytes() uint64 {
	return (uint64(p) + 7) / 8
}

// IsZero returns true if the position is zero.
func (p Position) IsZero() bool {
	return p == 0
}

// IsByteAligned returns true if the position is byte-aligned (bits == 0).
func (p Position) IsByteAligned() bool {
	return p%8 == 0
}

// Equal returns true if two positions are equal.
func (p Position) Equal(other Position) bool {
	return p == other
}

// Less returns true if p < other.
func (p Position) Less(other Position) bool {
	return p < other
}

// LessOrEqual returns true if p <= other.
func (p Position) LessOrEqual(other Position) bool {
	return p <= other
}

// Greater returns true if p > other.
func (p Position) Greater(other Position) bool {
	return p > other
}

// GreaterOrEqual returns true if p >= other.
func (p Position) GreaterOrEqual(other Position) bool {
	return p >= other
}

// Add returns p + other.
func (p Position) Add(other Position) Position {
	return p + other
}

// Sub returns p - other. Panics if other > p.
func (p Position) Sub(other Position) Position {
	if other > p {
		panic("negative BitPosition not supported")
	}
	return p - other
}

// Mul returns p * scalar.
func (p Position) Mul(scalar uint64) Position {
	return Position(uint64(p) * scalar)
}

// String returns a string representation of the position.
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Bytes(), p.Bits())
}
