package bitio

import (
	"bytes"
	"testing"
)

func TestPosition(t *testing.T) {
	t.Run("NewPosition normalizes", func(t *testing.T) {
		p := NewPosition(1, 10) // 10 bits = 1 byte + 2 bits
		if p.Bytes() != 2 || p.Bits() != 2 {
			t.Errorf("expected 2:2, got %s", p)
		}
	})

	t.Run("FromBits", func(t *testing.T) {
		p := FromBits(19) // 2 bytes + 3 bits
		if p.Bytes() != 2 || p.Bits() != 3 {
			t.Errorf("expected 2:3, got %s", p)
		}
	})

	t.Run("TotalBits", func(t *testing.T) {
		p := NewPosition(3, 5)
		if p.TotalBits() != 29 {
			t.Errorf("expected 29, got %d", p.TotalBits())
		}
	})

	t.Run("Add", func(t *testing.T) {
		a := NewPosition(1, 5)
		b := NewPosition(2, 4)
		c := a.Add(b)
		if c.Bytes() != 4 || c.Bits() != 1 {
			t.Errorf("expected 4:1, got %s", c)
		}
	})

	t.Run("Sub", func(t *testing.T) {
		a := NewPosition(3, 2)
		b := NewPosition(1, 5)
		c := a.Sub(b)
		if c.Bytes() != 1 || c.Bits() != 5 {
			t.Errorf("expected 1:5, got %s", c)
		}
	})
}

func TestReaderBasic(t *testing.T) {
	data := []byte{0xAB, 0xCD, 0xEF, 0x12, 0x34}
	r := NewReader(data)

	t.Run("ReadUint8 full byte", func(t *testing.T) {
		val, err := r.ReadUint8(8)
		if err != nil {
			t.Fatal(err)
		}
		if val != 0xAB {
			t.Errorf("expected 0xAB, got 0x%X", val)
		}
	})

	t.Run("ReadUint8 partial", func(t *testing.T) {
		// Position is now at byte 1
		// 0xCD = 1100 1101
		val, err := r.ReadUint8(4) // Should read 1101 = 0x0D
		if err != nil {
			t.Fatal(err)
		}
		if val != 0x0D {
			t.Errorf("expected 0x0D, got 0x%X", val)
		}
	})

	t.Run("ReadUint8 crosses byte boundary", func(t *testing.T) {
		// Position is now at byte 1, bit 4
		// Remaining bits of 0xCD: 1100 (high nibble)
		// First bits of 0xEF: 1110 1111
		// Reading 8 bits: 1100 from 0xCD + 1111 from 0xEF = 1111 1100 = 0xFC
		val, err := r.ReadUint8(8)
		if err != nil {
			t.Fatal(err)
		}
		if val != 0xFC {
			t.Errorf("expected 0xFC, got 0x%X", val)
		}
	})
}

func TestReaderBitAligned(t *testing.T) {
	// Test reading at various bit offsets
	data := []byte{0b10110100, 0b11001010}

	t.Run("Read 3 bits", func(t *testing.T) {
		r := NewReader(data)
		val, _ := r.ReadUint8(3) // 100 = 4
		if val != 4 {
			t.Errorf("expected 4, got %d", val)
		}
	})

	t.Run("Read 5 bits then 6 bits", func(t *testing.T) {
		r := NewReader(data)
		v1, _ := r.ReadUint8(5) // 10100 = 20
		if v1 != 20 {
			t.Errorf("expected 20, got %d", v1)
		}
		v2, _ := r.ReadUint8(6) // 011011 = 27
		if v2 != 27 {
			t.Errorf("expected 27, got %d (pos: %s)", v2, r.Position())
		}
	})
}

func TestWriterBasic(t *testing.T) {
	t.Run("Write full bytes", func(t *testing.T) {
		w := NewWriterAutoGrow()
		w.WriteUint8(0xAB, 8)
		w.WriteUint8(0xCD, 8)

		data := w.Data()
		if !bytes.Equal(data, []byte{0xAB, 0xCD}) {
			t.Errorf("expected [AB CD], got %X", data)
		}
	})

	t.Run("Write partial bits", func(t *testing.T) {
		w := NewWriterAutoGrow()
		w.WriteUint8(0x05, 4) // 0101
		w.WriteUint8(0x0A, 4) // 1010
		// Result: 1010 0101 = 0xA5

		data := w.Data()
		if len(data) != 1 || data[0] != 0xA5 {
			t.Errorf("expected [A5], got %X", data)
		}
	})

	t.Run("Write crosses byte boundary", func(t *testing.T) {
		w := NewWriterAutoGrow()
		w.WriteUint8(0x07, 3)  // 111
		w.WriteUint16(0x1FF, 9) // 1 1111 1111
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
		if len(data) != 2 || data[0] != 0xFF || data[1] != 0x0F {
			t.Errorf("expected [FF 0F], got %X", data)
		}
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

		if v1 != 5 {
			t.Errorf("v1: expected 5, got %d", v1)
		}
		if v2 != 1000 {
			t.Errorf("v2: expected 1000, got %d", v2)
		}
		if v3 != 0xDEADBEEF {
			t.Errorf("v3: expected 0xDEADBEEF, got 0x%X", v3)
		}
		if v4 != 7 {
			t.Errorf("v4: expected 7, got %d", v4)
		}
	})

	t.Run("String round trip", func(t *testing.T) {
		w := NewWriterAutoGrow()
		w.WriteString("hello")

		r := w.ToReader()
		s, err := r.ReadString()
		if err != nil {
			t.Fatal(err)
		}
		if s != "hello" {
			t.Errorf("expected 'hello', got '%s'", s)
		}
	})

	t.Run("Varint round trip", func(t *testing.T) {
		testCases := []uint32{0, 1, 127, 128, 16383, 16384, 0x7FFFFFFF}
		for _, tc := range testCases {
			w := NewWriterAutoGrow()
			w.WriteUvarint(tc)

			r := w.ToReader()
			got, err := r.ReadUvarint()
			if err != nil {
				t.Fatal(err)
			}
			if got != tc {
				t.Errorf("varint %d: expected %d, got %d", tc, tc, got)
			}
		}
	})
}

func TestSeek(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78}
	r := NewReader(data)

	r.ReadUint8(8) // Read first byte
	r.SeekBits(-8) // Seek back

	val, _ := r.ReadUint8(8)
	if val != 0x12 {
		t.Errorf("expected 0x12, got 0x%X", val)
	}

	r.Seek(FromBits(16), SeekSet)
	val, _ = r.ReadUint8(8)
	if val != 0x56 {
		t.Errorf("expected 0x56, got 0x%X", val)
	}
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
			r.pos = FromBits(3) // Start at bit 3
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
			w.WriteUint8(0, 3) // Misalign by 3 bits
			for j := 0; j < 200; j++ {
				w.WriteUint32(0xDEADBEEF, 32)
			}
		}
	})
}
