package bitio

import (
	"bytes"
	"testing"
	"github.com/wow-look-at-my/testify/assert"
	"github.com/wow-look-at-my/testify/require"
)

func TestPosition(t *testing.T) {
	t.Run("NewPosition normalizes", func(t *testing.T) {
		p := NewPosition(1, 10) // 10 bits = 1 byte + 2 bits
		assert.Equal(t, uint64(2), p.Bytes())
		assert.Equal(t, uint8(2), p.Bits())
	})

	t.Run("FromBits", func(t *testing.T) {
		p := FromBits(19) // 2 bytes + 3 bits
		assert.Equal(t, uint64(2), p.Bytes())
		assert.Equal(t, uint8(3), p.Bits())
	})

	t.Run("TotalBits", func(t *testing.T) {
		p := NewPosition(3, 5)
		assert.Equal(t, uint64(29), p.TotalBits())

	})

	t.Run("Add", func(t *testing.T) {
		a := NewPosition(1, 5)
		b := NewPosition(2, 4)
		c := a.Add(b)
		assert.Equal(t, uint64(4), c.Bytes())
		assert.Equal(t, uint8(1), c.Bits())
	})

	t.Run("Sub", func(t *testing.T) {
		a := NewPosition(3, 2)
		b := NewPosition(1, 5)
		c := a.Sub(b)
		assert.Equal(t, uint64(1), c.Bytes())
		assert.Equal(t, uint8(5), c.Bits())
	})
}

func TestReaderBasic(t *testing.T) {
	data := []byte{0xAB, 0xCD, 0xEF, 0x12, 0x34}
	r := NewReader(data)

	t.Run("ReadUint8 full byte", func(t *testing.T) {
		val, err := r.ReadUint8(8)
		require.Nil(t, err)
		assert.Equal(t, uint8(0xAB), val)

	})

	t.Run("ReadUint8 partial", func(t *testing.T) {
		// Position is now at byte 1
		// 0xCD = 1100 1101
		val, err := r.ReadUint8(4) // Should read 1101 = 0x0D
		require.Nil(t, err)
		assert.Equal(t, uint8(0x0D), val)
	})

	t.Run("ReadUint8 crosses byte boundary", func(t *testing.T) {
		// Position is now at byte 1, bit 4
		// Remaining bits of 0xCD: 1100 (high nibble)
		// First bits of 0xEF: 1110 1111
		// Reading 8 bits: 1100 from 0xCD + 1111 from 0xEF = 1111 1100 = 0xFC
		val, err := r.ReadUint8(8)
		require.Nil(t, err)
		assert.Equal(t, uint8(0xFC), val)

	})
}

func TestReaderBitAligned(t *testing.T) {
	// Test reading at various bit offsets
	data := []byte{0b10110100, 0b11001010}

	t.Run("Read 3 bits", func(t *testing.T) {
		r := NewReader(data)
		val, _ := r.ReadUint8(3) // 100 = 4
		assert.Equal(t, uint8(4), val)
	})

	t.Run("Read 5 bits then 6 bits", func(t *testing.T) {
		r := NewReader(data)
		v1, _ := r.ReadUint8(5) // 10100 = 20
		assert.Equal(t, uint8(20), v1)

		// After reading 5 bits, we're at bit 5
		// Reading 6 more bits: bits 5-7 of byte 0 (101) + bits 0-2 of byte 1 (010)
		// Combined little-endian: 010 101 = 21
		v2, _ := r.ReadUint8(6)
		assert.Equal(t, uint8(21), v2)
	})
}

func TestWriterBasic(t *testing.T) {
	t.Run("Write full bytes", func(t *testing.T) {
		w := NewWriterAutoGrow()
		w.WriteUint8(0xAB, 8)
		w.WriteUint8(0xCD, 8)

		data := w.Data()
		assert.True(t, bytes.Equal(data, []byte{0xAB, 0xCD}))

	})

	t.Run("Write partial bits", func(t *testing.T) {
		w := NewWriterAutoGrow()
		w.WriteUint8(0x05, 4)	// 0101
		w.WriteUint8(0x0A, 4)	// 1010
		// Result: 1010 0101 = 0xA5

		data := w.Data()
		assert.False(t, len(data) != 1 || data[0] != 0xA5)

	})

	t.Run("Write crosses byte boundary", func(t *testing.T) {
		w := NewWriterAutoGrow()
		w.WriteUint8(0x07, 3)	// 111
		w.WriteUint16(0x1FF, 9)	// 1 1111 1111
		// Byte 0: 111 + 11111 = 1111 1111 = 0xFF
		// Byte 1: 0000 0001 = 0x01
		// Wait, let me recalculate...
		// Write 3 bits: xxx00111
		// Write 9 bits: 0x1FF = 0b1_1111_1111
		// After 3 bits, we write at bit 3
		// bits 0-4 of 0x1FF (11111) go to bits 3-7 of byte 0
		// bits 5-8 of 0x1FF (1111) go to bits 0-3 of byte 1
		// Byte 0: 11111 111 = 0xFF
		// Byte 1: 0000 1111 = 0x0F

		data := w.Data()
		assert.False(t, len(data) != 2 || data[0] != 0xFF || data[1] != 0x0F)

	})
}

func TestRoundTrip(t *testing.T) {
	t.Run("Various bit widths", func(t *testing.T) {
		w := NewWriterAutoGrow()
		w.WriteUint8(5, 3)
		w.WriteUint16(1000, 12)
		w.WriteUint32(0xDEADBEEF, 32)
		w.WriteUint8(7, 4)

		r := w.ToReader()
		v1, _ := r.ReadUint8(3)
		v2, _ := r.ReadUint16(12)
		v3, _ := r.ReadUint32(32)
		v4, _ := r.ReadUint8(4)
		assert.Equal(t, uint8(5), v1)
		assert.Equal(t, uint16(1000), v2)
		assert.Equal(t, uint32(0xDEADBEEF), v3)
		assert.Equal(t, uint8(7), v4)

	})

	t.Run("String round trip", func(t *testing.T) {
		w := NewWriterAutoGrow()
		w.WriteString("hello")

		r := w.ToReader()
		s, err := r.ReadString()
		require.Nil(t, err)
		assert.Equal(t, "hello", s)

	})

	t.Run("Varint round trip", func(t *testing.T) {
		testCases := []uint32{0, 1, 127, 128, 16383, 16384, 0x7FFFFFFF}
		for _, tc := range testCases {
			w := NewWriterAutoGrow()
			w.WriteUvarint(tc)

			r := w.ToReader()
			got, err := r.ReadUvarint()
			require.Nil(t, err)
			assert.Equal(t, tc, got)

		}
	})
}

func TestSeek(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78}
	r := NewReader(data)

	r.ReadUint8(8)	// Read first byte
	r.SeekBits(-8)	// Seek back

	val, _ := r.ReadUint8(8)
	assert.Equal(t, uint8(0x12), val)

	r.Seek(FromBits(16), SeekSet)
	val, _ = r.ReadUint8(8)
	assert.Equal(t, uint8(0x56), val)

}

func TestReadUint64(t *testing.T) {
	// Create data with a known 64-bit value
	data := []byte{0xEF, 0xBE, 0xAD, 0xDE, 0xBE, 0xBA, 0xFE, 0xCA, 0xFF}

	t.Run("Aligned 64 bits", func(t *testing.T) {
		r := NewReader(data)
		val, err := r.ReadUint64(64)
		require.Nil(t, err)
		assert.Equal(t, uint64(0xCAFEBABEDEADBEEF), val)
	})

	t.Run("Unaligned 64 bits", func(t *testing.T) {
		r := NewReader(data)
		r.ReadUint8(4) // Offset by 4 bits
		val, err := r.ReadUint64(64)
		require.Nil(t, err)
		// After shifting right 4 bits and reading across 9 bytes
		expected := uint64(0xFCAFEBABEDEADBEE)
		assert.Equal(t, expected, val)
	})

	t.Run("Various bit widths", func(t *testing.T) {
		r := NewReader(data)
		// Read progressively larger values
		v1, _ := r.ReadUint64(8)
		assert.Equal(t, uint64(0xEF), v1)

		r.pos = Zero
		v2, _ := r.ReadUint64(16)
		assert.Equal(t, uint64(0xBEEF), v2)

		r.pos = Zero
		v3, _ := r.ReadUint64(32)
		assert.Equal(t, uint64(0xDEADBEEF), v3)

		r.pos = Zero
		v4, _ := r.ReadUint64(48)
		assert.Equal(t, uint64(0xBABEDEADBEEF), v4)
	})
}

func TestReadStringN(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteString("hello world")

	r := w.ToReader()
	s, err := r.ReadStringN(5)
	require.Nil(t, err)
	assert.Equal(t, "hell", s) // Reads 4 chars + null
}

func TestReadBytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	r := NewReader(data)

	result, err := r.ReadBytes(3)
	require.Nil(t, err)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, result)
}

func TestTakeSpan(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78}
	r := NewReader(data)

	span, err := r.TakeSpan(FromBits(16))
	require.Nil(t, err)

	// Original reader should have advanced
	assert.Equal(t, FromBits(16), r.Position())

	// Span should read the first 2 bytes
	val, _ := span.ReadUint16(16)
	assert.Equal(t, uint16(0x3412), val)
}

func TestSpan(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78}
	r := NewReader(data)

	span, err := r.Span(FromBits(8), FromBits(24))
	require.Nil(t, err)

	val, _ := span.ReadUint16(16)
	assert.Equal(t, uint16(0x5634), val)

	// Error cases
	_, err = r.Span(FromBits(24), FromBits(8))
	assert.NotNil(t, err)
}

func TestWriteUint64(t *testing.T) {
	w := NewWriterAutoGrow()
	err := w.WriteUint64(0xDEADBEEFCAFEBABE, 64)
	require.Nil(t, err)

	r := w.ToReader()
	val, _ := r.ReadUint64(64)
	assert.Equal(t, uint64(0xDEADBEEFCAFEBABE), val)
}

func TestWriteBytesAligned(t *testing.T) {
	w := NewWriterAutoGrow()
	err := w.WriteBytesAligned([]byte{0xAB, 0xCD, 0xEF})
	require.Nil(t, err)

	assert.Equal(t, []byte{0xAB, 0xCD, 0xEF}, w.Data())
}

func TestWriteFromReader(t *testing.T) {
	// Create source data
	src := NewReader([]byte{0x12, 0x34, 0x56})

	// Copy to writer
	w := NewWriterAutoGrow()
	err := w.WriteFromReaderN(src, FromBits(24))
	require.Nil(t, err)

	assert.Equal(t, []byte{0x12, 0x34, 0x56}, w.Data())
}

func TestPadToByte(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteUint8(0x07, 3) // Write 3 bits
	w.PadToByte()         // Should write 5 zero bits

	assert.Equal(t, FromBits(8), w.Length())
	assert.Equal(t, []byte{0x07}, w.Data())
}

func TestWriterSeek(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteUint8(0xAB, 8)
	w.WriteUint8(0xCD, 8)

	// Seek back and overwrite
	err := w.Seek(Zero)
	require.Nil(t, err)
	w.WriteUint8(0xFF, 8)

	assert.Equal(t, []byte{0xFF, 0xCD}, w.Data())

	// Seek past end should fail
	err = w.Seek(FromBits(100))
	assert.NotNil(t, err)
}

func TestWriterGrow(t *testing.T) {
	w := NewWriterAutoGrow()

	// Write enough to trigger growth
	for i := 0; i < 100; i++ {
		w.WriteUint32(0xDEADBEEF, 32)
	}

	assert.Equal(t, FromBits(3200), w.Length())
}

func TestFixedWriter(t *testing.T) {
	buf := make([]byte, 4)
	w := NewWriter(buf)

	w.WriteUint32(0x12345678, 32)
	assert.Equal(t, []byte{0x78, 0x56, 0x34, 0x12}, buf)

	// Writing more should fail
	err := w.WriteUint8(0xFF, 8)
	assert.NotNil(t, err)
}

func TestPositionComparisons(t *testing.T) {
	a := NewPosition(1, 5)
	b := NewPosition(2, 3)
	c := NewPosition(1, 5)

	assert.True(t, a.Less(b))
	assert.True(t, a.LessOrEqual(b))
	assert.True(t, a.LessOrEqual(c))
	assert.True(t, b.Greater(a))
	assert.True(t, b.GreaterOrEqual(a))
	assert.True(t, a.GreaterOrEqual(c))
	assert.True(t, a.Equal(c))

	// Test Mul
	d := NewPosition(1, 2).Mul(3)
	assert.Equal(t, uint64(30), d.TotalBits()) // (8+2)*3 = 30
}

func TestPositionHelpers(t *testing.T) {
	p := NewPosition(2, 3)
	assert.Equal(t, "2:3", p.String())
	assert.False(t, p.IsZero())
	assert.False(t, p.IsByteAligned())

	aligned := FromBytes(5)
	assert.True(t, aligned.IsByteAligned())
	assert.Equal(t, uint64(5), aligned.TotalBytes())
}

func TestReaderHelpers(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78}
	r := NewReader(data)

	assert.Equal(t, Zero, r.Position())
	assert.Equal(t, Zero, r.LocalPosition())
	assert.Equal(t, FromBytes(4), r.Length())
	assert.Equal(t, FromBytes(4), r.Remaining())
	assert.False(t, r.IsAtEnd())
	assert.Equal(t, data, r.Data())

	r.ReadUint32(32)
	assert.True(t, r.IsAtEnd())
}

func TestReaderSeekModes(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78}
	r := NewReader(data)

	// SeekEnd
	r.Seek(FromBits(8), SeekEnd)
	assert.Equal(t, FromBits(24), r.Position())

	// SeekStart
	r.Seek(FromBits(8), SeekStart)
	assert.Equal(t, FromBits(8), r.Position())

	// SeekBytes forward
	r.SeekBytes(1)
	assert.Equal(t, FromBits(16), r.Position())
}

func TestReaderClone(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78}
	r := NewReader(data)
	r.ReadUint8(8)

	clone := r.Clone()
	assert.Equal(t, r.Position(), clone.Position())

	// Advancing clone shouldn't affect original
	clone.ReadUint8(8)
	assert.Equal(t, FromBits(8), r.Position())
	assert.Equal(t, FromBits(16), clone.Position())
}

func TestReadBit(t *testing.T) {
	data := []byte{0b10101010}
	r := NewReader(data)

	b0, _ := r.ReadBit()
	b1, _ := r.ReadBit()
	b2, _ := r.ReadBit()
	b3, _ := r.ReadBit()

	assert.False(t, b0)
	assert.True(t, b1)
	assert.False(t, b2)
	assert.True(t, b3)
}

func TestReadSignedInts(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteInt8(-1, 8)
	w.WriteInt16(-1000, 16)
	w.WriteInt32(-100000, 32)
	w.WriteInt64(-1, 64)

	r := w.ToReader()
	v1, _ := r.ReadInt8(8)
	v2, _ := r.ReadInt16(16)
	v3, _ := r.ReadInt32(32)
	v4, _ := r.ReadInt64(64)

	assert.Equal(t, int8(-1), v1)
	assert.Equal(t, int16(-1000), v2)
	assert.Equal(t, int32(-100000), v3)
	assert.Equal(t, int64(-1), v4)
}

func TestReadFloats(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteFloat32(3.14159)
	w.WriteFloat64(2.71828)

	r := w.ToReader()
	f1, _ := r.ReadFloat32()
	f2, _ := r.ReadFloat64()

	assert.InDelta(t, 3.14159, f1, 0.00001)
	assert.InDelta(t, 2.71828, f2, 0.00001)
}

func TestWriteBit(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteBit(false)
	w.WriteBit(true)
	w.WriteBit(false)
	w.WriteBit(true)
	w.WriteBit(false)
	w.WriteBit(true)
	w.WriteBit(false)
	w.WriteBit(true)

	assert.Equal(t, []byte{0b10101010}, w.Data())
}

func TestWriteVarint(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteVarint(-100)

	r := w.ToReader()
	val, _ := r.ReadVarint()
	assert.Equal(t, int32(-100), val)
}

func TestWriteStringN(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteStringN("hello world", 6) // Should write "hello\0"

	r := w.ToReader()
	s, _ := r.ReadString()
	assert.Equal(t, "hello", s)
}

func TestWriteBytes(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteUint8(0x0F, 4) // Misalign
	w.WriteBytes([]byte{0xAB, 0xCD})

	// Should still work despite misalignment
	r := w.ToReader()
	r.ReadUint8(4) // Skip the first 4 bits
	b1, _ := r.ReadUint8(8)
	b2, _ := r.ReadUint8(8)
	assert.Equal(t, uint8(0xAB), b1)
	assert.Equal(t, uint8(0xCD), b2)
}

func TestWriterHelpers(t *testing.T) {
	w := NewWriterAutoGrow()
	w.WriteUint32(0xDEADBEEF, 32)

	assert.Equal(t, FromBits(32), w.Position())
	assert.Equal(t, FromBits(32), w.Length())
	assert.Equal(t, 4, len(w.Bytes()))
}

func TestErrorCases(t *testing.T) {
	data := []byte{0x12}
	r := NewReader(data)

	// Read past end
	_, err := r.ReadUint16(16)
	assert.NotNil(t, err)

	// Peek past end
	_, err = r.PeekUint16(16)
	assert.NotNil(t, err)

	// Invalid bit count
	_, err = r.ReadUint8(9)
	assert.NotNil(t, err)

	// Seek before start
	err = r.SeekBits(-100)
	assert.NotNil(t, err)

	// Invalid seek mode (using a bogus value)
	err = r.Seek(Zero, SeekMode(99))
	assert.NotNil(t, err)
}

func TestWriterErrorCases(t *testing.T) {
	buf := make([]byte, 1)
	w := NewWriter(buf)

	// Invalid bit count
	err := w.WriteUint16(0, 17)
	assert.NotNil(t, err)

	err = w.WriteUint32(0, 33)
	assert.NotNil(t, err)

	err = w.WriteUint64(0, 65)
	assert.NotNil(t, err)
}

func BenchmarkReadUint32(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}

	b.Run("Aligned", func(b *testing.B) {
		r := NewReader(data)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.pos = Zero
			for j := 0; j < 256; j++ {
				r.ReadUint32(32)
			}
		}
	})

	b.Run("Unaligned", func(b *testing.B) {
		r := NewReader(data)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.pos = FromBits(3)	// Start at bit 3
			for j := 0; j < 200; j++ {
				r.ReadUint32(32)
			}
		}
	})
}

func BenchmarkWriteUint32(b *testing.B) {
	b.Run("Aligned", func(b *testing.B) {
		w := NewWriterSize(4096)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w.Reset()
			for j := 0; j < 256; j++ {
				w.WriteUint32(0xDEADBEEF, 32)
			}
		}
	})

	b.Run("Unaligned", func(b *testing.B) {
		w := NewWriterSize(4096)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w.Reset()
			w.WriteUint8(0, 3)	// Misalign by 3 bits
			for j := 0; j < 200; j++ {
				w.WriteUint32(0xDEADBEEF, 32)
			}
		}
	})
}
