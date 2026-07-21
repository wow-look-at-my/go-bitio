package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wow-look-at-my/go-bitio"
)

func TestGenerateTestStream(t *testing.T) {
	// Generate with fixed seed for determinism
	data := GenerateTestStream(42, 1000)

	// Should produce non-empty output
	assert.NotEmpty(t, data)

	// Verify the stream is readable
	r := bitio.NewReader(data)
	count := 0

	for r.Remaining().TotalBits() > 64 {
		typeTag, err := r.ReadUint8(3)
		assert.Nil(t, err)
		assert.LessOrEqual(t, typeTag, uint8(7))

		switch typeTag {
		case TypeUint8:
			bitCount, _ := r.ReadUint8(4)
			assert.LessOrEqual(t, bitCount, uint8(8))
			_, err := r.ReadUint8(bitCount)
			assert.Nil(t, err)

		case TypeUint16:
			bitCount, _ := r.ReadUint8(5)
			assert.LessOrEqual(t, bitCount, uint8(16))
			_, err := r.ReadUint16(bitCount)
			assert.Nil(t, err)

		case TypeUint32:
			bitCount, _ := r.ReadUint8(6)
			assert.LessOrEqual(t, bitCount, uint8(32))
			_, err := r.ReadUint32(bitCount)
			assert.Nil(t, err)

		case TypeUint64:
			bitCount, _ := r.ReadUint8(7)
			assert.LessOrEqual(t, bitCount, uint8(64))
			_, err := r.ReadUint64(bitCount)
			assert.Nil(t, err)

		case TypeString:
			s, err := r.ReadString()
			assert.Nil(t, err)
			assert.NotEmpty(t, s)

		case TypeFloat32:
			_, err := r.ReadFloat32()
			assert.Nil(t, err)

		case TypeFloat64:
			_, err := r.ReadFloat64()
			assert.Nil(t, err)

		case TypeVaruint:
			_, err := r.ReadVarint()
			assert.Nil(t, err)
		}
		count++
	}

	// Should have read approximately 1000 ops (minus any at the end that don't fit)
	assert.Greater(t, count, 900)
}

func TestGenerateTestStreamDeterministic(t *testing.T) {
	// Same seed should produce same output
	data1 := GenerateTestStream(123, 100)
	data2 := GenerateTestStream(123, 100)
	assert.Equal(t, data1, data2)

	// Different seed should produce different output
	data3 := GenerateTestStream(456, 100)
	assert.NotEqual(t, data1, data3)
}
