package bitio

import (
	"fmt"
)

// Position represents a position in a bit stream as bytes + bits offset.
// The bits component is always in the range [0, 7].
type Position struct {
	bytes uint64
	bits  uint8
}

// Zero is the zero position.
var Zero = Position{}

// NewPosition creates a new Position from bytes and bits.
// If bits >= 8, it normalizes by adding to bytes.
func NewPosition(bytes uint64, bits uint8) Position {
	return Position{
		bytes: bytes + uint64(bits/8),
		bits:  bits % 8,
	}
}

// FromBits creates a Position from a total bit count.
func FromBits(bits uint64) Position {
	return Position{
		bytes: bits / 8,
		bits:  uint8(bits % 8),
	}
}

// FromBytes creates a Position from a byte count.
func FromBytes(bytes uint64) Position {
	return Position{bytes: bytes}
}

// Bytes returns the byte component of the position.
func (p Position) Bytes() uint64 {
	return p.bytes
}

// Bits returns the bit component of the position (0-7).
func (p Position) Bits() uint8 {
	return p.bits
}

// TotalBits returns the total number of bits represented by this position.
func (p Position) TotalBits() uint64 {
	return p.bytes*8 + uint64(p.bits)
}

// TotalBytes returns the minimum number of bytes needed to contain this position.
func (p Position) TotalBytes() uint64 {
	return p.bytes + uint64((p.bits+7)/8)
}

// IsZero returns true if the position is zero.
func (p Position) IsZero() bool {
	return p.bytes == 0 && p.bits == 0
}

// IsByteAligned returns true if the position is byte-aligned (bits == 0).
func (p Position) IsByteAligned() bool {
	return p.bits == 0
}

// Equal returns true if two positions are equal.
func (p Position) Equal(other Position) bool {
	return p.bytes == other.bytes && p.bits == other.bits
}

// Less returns true if p < other.
func (p Position) Less(other Position) bool {
	return p.bytes < other.bytes || (p.bytes == other.bytes && p.bits < other.bits)
}

// LessOrEqual returns true if p <= other.
func (p Position) LessOrEqual(other Position) bool {
	return p.bytes < other.bytes || (p.bytes == other.bytes && p.bits <= other.bits)
}

// Greater returns true if p > other.
func (p Position) Greater(other Position) bool {
	return p.bytes > other.bytes || (p.bytes == other.bytes && p.bits > other.bits)
}

// GreaterOrEqual returns true if p >= other.
func (p Position) GreaterOrEqual(other Position) bool {
	return p.bytes > other.bytes || (p.bytes == other.bytes && p.bits >= other.bits)
}

// Add returns p + other.
func (p Position) Add(other Position) Position {
	totalBits := p.bits + other.bits
	return Position{
		bytes: p.bytes + other.bytes + uint64(totalBits/8),
		bits:  totalBits % 8,
	}
}

// Sub returns p - other. Panics if other > p.
func (p Position) Sub(other Position) Position {
	if other.Greater(p) {
		panic("negative BitPosition not supported")
	}
	// Add 8 to handle borrow from bytes
	adjustedBits := p.bits + 8 - other.bits
	return Position{
		bytes: p.bytes - other.bytes - 1 + uint64(adjustedBits/8),
		bits:  adjustedBits % 8,
	}
}

// Mul returns p * scalar.
func (p Position) Mul(scalar uint64) Position {
	totalBits := p.TotalBits() * scalar
	return FromBits(totalBits)
}

// String returns a string representation of the position.
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.bytes, p.bits)
}
