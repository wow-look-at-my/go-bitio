package bitio

import (
	"errors"
	"math"
)

// Writer writes data at arbitrary bit positions to a byte slice.
type Writer struct {
	data		[]byte
	pos		Position
	end		Position
	autoGrow	bool
}

// NewWriter creates a new Writer with a fixed-size buffer.
func NewWriter(data []byte) *Writer {
	return &Writer{
		data:		data,
		pos:		zero,
		end:		zero,
		autoGrow:	false,
	}
}

// NewWriterSize creates a new Writer with an initial buffer size.
func NewWriterSize(size int) *Writer {
	return &Writer{
		data:		make([]byte, size),
		pos:		zero,
		end:		zero,
		autoGrow:	false,
	}
}

// NewWriterAutoGrow creates a new Writer that automatically grows its buffer.
func NewWriterAutoGrow() *Writer {
	return &Writer{
		data:		make([]byte, 64),
		pos:		zero,
		end:		zero,
		autoGrow:	true,
	}
}

// Position returns the current write position.
func (w *Writer) Position() Position {
	return w.pos
}

// Length returns the total length written so far.
func (w *Writer) Length() Position {
	return w.end
}

// Data returns the underlying data slice up to the written length.
func (w *Writer) Data() []byte {
	return w.data[:w.end.TotalBytes()]
}

// Bytes returns the written data as a byte slice.
// This is the same as Data() for byte-aligned writes.
func (w *Writer) Bytes() []byte {
	return w.Data()
}

// ensureCapacity ensures the buffer can hold at least newEnd bits.
func (w *Writer) ensureCapacity(newEnd Position) error {
	bytesNeeded := newEnd.TotalBytes()
	if bytesNeeded <= uint64(len(w.data)) {
		return nil
	}
	if !w.autoGrow {
		return errors.New("write would overflow buffer")
	}
	// Grow the buffer
	newCap := uint64(len(w.data)) * 2
	if newCap < bytesNeeded {
		newCap = bytesNeeded * 2
	}
	newData := make([]byte, newCap)
	copy(newData, w.data)
	w.data = newData
	return nil
}

// writeBits writes bits to the buffer at the given position.
func writeBits(data []byte, byteOff uint64, bitOff, bits uint8, value uint64) {
	if bits == 0 {
		return
	}

	// Mask the value to the number of bits we're writing
	value &= bitMask(bits)

	for bits > 0 {
		// How many bits can we write to the current byte?
		bitsInByte := 8 - bitOff
		if bitsInByte > bits {
			bitsInByte = bits
		}

		// Create mask for the bits we're writing
		mask := uint8(bitMask(bitsInByte)) << bitOff

		// Clear the bits we're about to write
		data[byteOff] &= ^mask

		// Write the bits
		data[byteOff] |= uint8(value<<bitOff) & mask

		// Move to next byte
		value >>= bitsInByte
		bits -= bitsInByte
		bitOff = 0
		byteOff++
	}
}

// WriteUint8 writes up to 8 bits.
func (w *Writer) WriteUint8(value uint8, bits uint8) error {
	if bits > 8 {
		return errors.New("bits must be <= 8")
	}
	newEnd := w.pos.Add(FromBits(uint64(bits)))
	if err := w.ensureCapacity(newEnd); err != nil {
		return err
	}
	writeBits(w.data, w.pos.Bytes(), w.pos.Bits(), bits, uint64(value))
	w.pos = newEnd
	if w.pos.Greater(w.end) {
		w.end = w.pos
	}
	return nil
}

// WriteUint16 writes up to 16 bits.
func (w *Writer) WriteUint16(value uint16, bits uint8) error {
	if bits > 16 {
		return errors.New("bits must be <= 16")
	}
	newEnd := w.pos.Add(FromBits(uint64(bits)))
	if err := w.ensureCapacity(newEnd); err != nil {
		return err
	}
	writeBits(w.data, w.pos.Bytes(), w.pos.Bits(), bits, uint64(value))
	w.pos = newEnd
	if w.pos.Greater(w.end) {
		w.end = w.pos
	}
	return nil
}

// WriteUint32 writes up to 32 bits.
func (w *Writer) WriteUint32(value uint32, bits uint8) error {
	if bits > 32 {
		return errors.New("bits must be <= 32")
	}
	newEnd := w.pos.Add(FromBits(uint64(bits)))
	if err := w.ensureCapacity(newEnd); err != nil {
		return err
	}
	writeBits(w.data, w.pos.Bytes(), w.pos.Bits(), bits, uint64(value))
	w.pos = newEnd
	if w.pos.Greater(w.end) {
		w.end = w.pos
	}
	return nil
}

// WriteUint64 writes up to 64 bits.
func (w *Writer) WriteUint64(value uint64, bits uint8) error {
	if bits > 64 {
		return errors.New("bits must be <= 64")
	}
	newEnd := w.pos.Add(FromBits(uint64(bits)))
	if err := w.ensureCapacity(newEnd); err != nil {
		return err
	}
	writeBits(w.data, w.pos.Bytes(), w.pos.Bits(), bits, value)
	w.pos = newEnd
	if w.pos.Greater(w.end) {
		w.end = w.pos
	}
	return nil
}

// WriteBit writes a single bit.
func (w *Writer) WriteBit(value bool) error {
	var v uint8
	if value {
		v = 1
	}
	return w.WriteUint8(v, 1)
}

// WriteInt8 writes up to 8 bits as a signed integer.
func (w *Writer) WriteInt8(value int8, bits uint8) error {
	return w.WriteUint8(uint8(value), bits)
}

// WriteInt16 writes up to 16 bits as a signed integer.
func (w *Writer) WriteInt16(value int16, bits uint8) error {
	return w.WriteUint16(uint16(value), bits)
}

// WriteInt32 writes up to 32 bits as a signed integer.
func (w *Writer) WriteInt32(value int32, bits uint8) error {
	return w.WriteUint32(uint32(value), bits)
}

// WriteInt64 writes up to 64 bits as a signed integer.
func (w *Writer) WriteInt64(value int64, bits uint8) error {
	return w.WriteUint64(uint64(value), bits)
}

// WriteFloat32 writes a float32 (32 bits).
func (w *Writer) WriteFloat32(value float32) error {
	return w.WriteUint32(math.Float32bits(value), 32)
}

// WriteFloat64 writes a float64 (64 bits).
func (w *Writer) WriteFloat64(value float64) error {
	return w.WriteUint64(math.Float64bits(value), 64)
}

// WriteString writes a null-terminated string.
func (w *Writer) WriteString(s string) error {
	for i := 0; i < len(s); i++ {
		if err := w.WriteUint8(s[i], 8); err != nil {
			return err
		}
	}
	return w.WriteUint8(0, 8)
}

// WriteStringN writes a string with null terminator, up to maxLen bytes total.
func (w *Writer) WriteStringN(s string, maxLen int) error {
	toWrite := len(s)
	if toWrite > maxLen-1 {
		toWrite = maxLen - 1
	}
	for i := 0; i < toWrite; i++ {
		if err := w.WriteUint8(s[i], 8); err != nil {
			return err
		}
	}
	return w.WriteUint8(0, 8)
}

// WriteBytes writes a byte slice.
func (w *Writer) WriteBytes(data []byte) error {
	for _, b := range data {
		if err := w.WriteUint8(b, 8); err != nil {
			return err
		}
	}
	return nil
}

// WriteBytesAligned writes bytes at byte-aligned position (faster path).
func (w *Writer) WriteBytesAligned(data []byte) error {
	if !w.pos.IsByteAligned() {
		return w.WriteBytes(data)
	}
	newEnd := w.pos.Add(FromBytes(uint64(len(data))))
	if err := w.ensureCapacity(newEnd); err != nil {
		return err
	}
	copy(w.data[w.pos.Bytes():], data)
	w.pos = newEnd
	if w.pos.Greater(w.end) {
		w.end = w.pos
	}
	return nil
}

// WriteUvarint writes a variable-length unsigned integer.
func (w *Writer) WriteUvarint(value uint32) error {
	for i := 0; i < 5; i++ {
		b := uint8(value & 0x7F)
		value >>= 7
		if value != 0 {
			b |= 0x80
		}
		if err := w.WriteUint8(b, 8); err != nil {
			return err
		}
		if value == 0 {
			break
		}
	}
	return nil
}

// WriteVarint writes a variable-length signed integer.
func (w *Writer) WriteVarint(value int32) error {
	return w.WriteUvarint(uint32(value))
}

// WriteFromReader copies data from a Reader.
func (w *Writer) WriteFromReader(r *Reader) error {
	return w.WriteFromReaderN(r, r.Remaining())
}

// WriteFromReaderN copies a specific number of bits from a Reader.
func (w *Writer) WriteFromReaderN(r *Reader, length Position) error {
	remaining := length
	for !remaining.IsZero() {
		bits := uint8(8)
		if remaining.Bytes() == 0 {
			bits = remaining.Bits()
		}
		val, err := r.ReadUint8(bits)
		if err != nil {
			return err
		}
		if err := w.WriteUint8(val, bits); err != nil {
			return err
		}
		remaining = remaining.Sub(FromBits(uint64(bits)))
	}
	return nil
}

// PadToByte adds zero bits until the position is byte-aligned.
func (w *Writer) PadToByte() error {
	if w.pos.Bits() == 0 {
		return nil
	}
	return w.WriteUint8(0, 8-w.pos.Bits())
}

// Seek moves the write position.
func (w *Writer) Seek(pos Position) error {
	if pos.Greater(w.end) {
		return errors.New("seek past end of written data")
	}
	w.pos = pos
	return nil
}

// Reset resets the writer to the beginning.
func (w *Writer) Reset() {
	w.pos = zero
	w.end = zero
}

// ToReader creates a Reader from the written data.
func (w *Writer) ToReader() *Reader {
	return NewReaderWithBounds(w.data, zero, w.end)
}
