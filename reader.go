package bitio

import (
	"encoding/binary"
	"errors"
	"io"
)

// Reader reads data at arbitrary bit positions from a byte slice.
type Reader struct {
	data     []byte
	start    BitPos
	pos      BitPos
	end      BitPos
	msbFirst bool // If true, read bits from MSB to LSB within each byte
}

// SetMSBFirst sets whether bits are read MSB first (true) or LSB first (false, default).
func (r *Reader) SetMSBFirst(msb bool) {
	r.msbFirst = msb
}

// MSBFirst returns true if the reader is in MSB-first mode.
func (r *Reader) MSBFirst() bool {
	return r.msbFirst
}

// NewReader creates a new Reader from data.
func NewReader(data []byte) *Reader {
	return &Reader{
		data:  data,
		start: 0,
		pos:   0,
		end:   FromBytes(uint64(len(data))),
	}
}

// NewReaderWithBounds creates a Reader with explicit start and end positions.
func NewReaderWithBounds(data []byte, start, end BitPos) *Reader {
	return &Reader{
		data:  data,
		start: start,
		pos:   start,
		end:   end,
	}
}

// Position returns the current position.
func (r *Reader) Position() BitPos {
	return r.pos
}

// LocalPosition returns the position relative to start.
func (r *Reader) LocalPosition() BitSize {
	return r.pos.Diff(r.start)
}

// StartPosition returns the start position.
func (r *Reader) StartPosition() BitPos {
	return r.start
}

// EndPosition returns the end position.
func (r *Reader) EndPosition() BitPos {
	return r.end
}

// Length returns the total length from start to end.
func (r *Reader) Length() BitSize {
	return r.end.Diff(r.start)
}

// Remaining returns the remaining length from current position to end.
func (r *Reader) Remaining() BitSize {
	return r.end.Diff(r.pos)
}

// IsAtEnd returns true if at the end position.
func (r *Reader) IsAtEnd() bool {
	return r.pos == r.end
}

// SeekMode defines the seek reference point and direction.
type SeekMode int

const (
	// SeekSet seeks to an absolute position from the start of data.
	SeekSet SeekMode = iota
	// SeekFwd seeks forward from current position.
	SeekFwd
	// SeekBack seeks backward from current position.
	SeekBack
	// SeekEnd seeks backward from end position.
	SeekEnd
)

// Seek moves the read position.
func (r *Reader) Seek(offset BitSize, mode SeekMode) error {
	var newPos BitPos
	switch mode {
	case SeekSet:
		newPos = BitPos(offset)
	case SeekFwd:
		newPos = r.pos.Add(offset)
	case SeekBack:
		if offset > r.pos.Diff(r.start) {
			return errors.New("seek before start")
		}
		newPos = r.pos.Sub(offset)
	case SeekEnd:
		if offset > BitSize(r.end) {
			return errors.New("seek before start")
		}
		newPos = r.end.Sub(offset)
	default:
		return errors.New("invalid seek mode")
	}

	if newPos < r.start {
		return errors.New("seek before start")
	}
	if newPos > r.end {
		return errors.New("seek past end")
	}
	r.pos = newPos
	return nil
}

// bitMask returns a mask with the lowest n bits set.
func bitMask(n uint8) uint64 {
	if n >= 64 {
		return ^uint64(0)
	}
	return (1 << n) - 1
}

// readUint8 reads up to 8 bits from the buffer at the given byte offset and bit offset (LSB first).
func readUint8(data []byte, byteOff uint64, bitOff, bits uint8) uint8 {
	if bits == 0 {
		return 0
	}
	// How many bytes do we need to read?
	bytesNeeded := (bitOff + bits + 7) / 8
	if bytesNeeded == 1 {
		return (data[byteOff] >> bitOff) & uint8(bitMask(bits))
	}
	// Need to read 2 bytes
	val := uint16(data[byteOff]) | uint16(data[byteOff+1])<<8
	return uint8((val >> bitOff) & uint16(bitMask(bits)))
}

// reverseByte reverses the bits in a byte.
func reverseByte(b uint8) uint8 {
	b = (b&0xF0)>>4 | (b&0x0F)<<4
	b = (b&0xCC)>>2 | (b&0x33)<<2
	b = (b&0xAA)>>1 | (b&0x55)<<1
	return b
}

// readUint8MSB reads up to 8 bits from the buffer (MSB first).
func readUint8MSB(data []byte, byteOff uint64, bitOff, bits uint8) uint8 {
	if bits == 0 {
		return 0
	}
	// MSB-first: bit 0 is the most significant bit of the byte
	// We read from the high bits down
	bytesNeeded := (bitOff + bits + 7) / 8
	if bytesNeeded == 1 {
		// Shift right to get the bits we want at the bottom, then mask
		shift := 8 - bitOff - bits
		return (data[byteOff] >> shift) & uint8(bitMask(bits))
	}
	// Need to read 2 bytes - combine them MSB first
	val := uint16(data[byteOff])<<8 | uint16(data[byteOff+1])
	shift := 16 - bitOff - bits
	return uint8((val >> shift) & uint16(bitMask(bits)))
}

// readUint16 reads up to 16 bits from the buffer (LSB first).
func readUint16(data []byte, byteOff uint64, bitOff, bits uint8) uint16 {
	if bits == 0 {
		return 0
	}
	bytesNeeded := (bitOff + bits + 7) / 8
	switch bytesNeeded {
	case 1:
		return uint16(data[byteOff]>>bitOff) & uint16(bitMask(bits))
	case 2:
		val := binary.LittleEndian.Uint16(data[byteOff:])
		return (val >> bitOff) & uint16(bitMask(bits))
	default: // 3 bytes
		val := uint32(data[byteOff]) | uint32(data[byteOff+1])<<8 | uint32(data[byteOff+2])<<16
		return uint16((val >> bitOff) & uint32(bitMask(bits)))
	}
}

// readUint16MSB reads up to 16 bits from the buffer (MSB first).
func readUint16MSB(data []byte, byteOff uint64, bitOff, bits uint8) uint16 {
	if bits == 0 {
		return 0
	}
	bytesNeeded := (bitOff + bits + 7) / 8
	switch bytesNeeded {
	case 1:
		shift := 8 - bitOff - bits
		return uint16(data[byteOff]>>shift) & uint16(bitMask(bits))
	case 2:
		val := binary.BigEndian.Uint16(data[byteOff:])
		shift := 16 - bitOff - bits
		return (val >> shift) & uint16(bitMask(bits))
	default: // 3 bytes
		val := uint32(data[byteOff])<<16 | uint32(data[byteOff+1])<<8 | uint32(data[byteOff+2])
		shift := 24 - bitOff - bits
		return uint16((val >> shift) & uint32(bitMask(bits)))
	}
}

// readUint32 reads up to 32 bits from the buffer (LSB first).
func readUint32(data []byte, byteOff uint64, bitOff, bits uint8) uint32 {
	if bits == 0 {
		return 0
	}
	bytesNeeded := (bitOff + bits + 7) / 8
	switch bytesNeeded {
	case 1:
		return uint32(data[byteOff]>>bitOff) & uint32(bitMask(bits))
	case 2:
		val := binary.LittleEndian.Uint16(data[byteOff:])
		return uint32(val>>bitOff) & uint32(bitMask(bits))
	case 3:
		val := uint32(data[byteOff]) | uint32(data[byteOff+1])<<8 | uint32(data[byteOff+2])<<16
		return (val >> bitOff) & uint32(bitMask(bits))
	case 4:
		val := binary.LittleEndian.Uint32(data[byteOff:])
		return (val >> bitOff) & uint32(bitMask(bits))
	default: // 5 bytes
		val := uint64(binary.LittleEndian.Uint32(data[byteOff:])) | uint64(data[byteOff+4])<<32
		return uint32((val >> bitOff) & uint64(bitMask(bits)))
	}
}

// readUint32MSB reads up to 32 bits from the buffer (MSB first).
func readUint32MSB(data []byte, byteOff uint64, bitOff, bits uint8) uint32 {
	if bits == 0 {
		return 0
	}
	bytesNeeded := (bitOff + bits + 7) / 8
	switch bytesNeeded {
	case 1:
		shift := 8 - bitOff - bits
		return uint32(data[byteOff]>>shift) & uint32(bitMask(bits))
	case 2:
		val := binary.BigEndian.Uint16(data[byteOff:])
		shift := 16 - bitOff - bits
		return uint32(val>>shift) & uint32(bitMask(bits))
	case 3:
		val := uint32(data[byteOff])<<16 | uint32(data[byteOff+1])<<8 | uint32(data[byteOff+2])
		shift := 24 - bitOff - bits
		return (val >> shift) & uint32(bitMask(bits))
	case 4:
		val := binary.BigEndian.Uint32(data[byteOff:])
		shift := 32 - bitOff - bits
		return (val >> shift) & uint32(bitMask(bits))
	default: // 5 bytes
		val := uint64(data[byteOff])<<32 | uint64(binary.BigEndian.Uint32(data[byteOff+1:]))
		shift := 40 - bitOff - bits
		return uint32((val >> shift) & uint64(bitMask(bits)))
	}
}

// readUint64 reads up to 64 bits from the buffer (LSB first).
func readUint64(data []byte, byteOff uint64, bitOff, bits uint8) uint64 {
	if bits == 0 {
		return 0
	}
	bytesNeeded := (bitOff + bits + 7) / 8
	switch bytesNeeded {
	case 1:
		return uint64(data[byteOff]>>bitOff) & bitMask(bits)
	case 2:
		val := binary.LittleEndian.Uint16(data[byteOff:])
		return uint64(val>>bitOff) & bitMask(bits)
	case 3:
		val := uint32(data[byteOff]) | uint32(data[byteOff+1])<<8 | uint32(data[byteOff+2])<<16
		return uint64(val>>bitOff) & bitMask(bits)
	case 4:
		val := binary.LittleEndian.Uint32(data[byteOff:])
		return uint64(val>>bitOff) & bitMask(bits)
	case 5:
		val := uint64(binary.LittleEndian.Uint32(data[byteOff:])) | uint64(data[byteOff+4])<<32
		return (val >> bitOff) & bitMask(bits)
	case 6:
		val := uint64(binary.LittleEndian.Uint32(data[byteOff:])) | uint64(binary.LittleEndian.Uint16(data[byteOff+4:]))<<32
		return (val >> bitOff) & bitMask(bits)
	case 7:
		val := uint64(binary.LittleEndian.Uint32(data[byteOff:])) |
			uint64(data[byteOff+4])<<32 | uint64(data[byteOff+5])<<40 | uint64(data[byteOff+6])<<48
		return (val >> bitOff) & bitMask(bits)
	case 8:
		val := binary.LittleEndian.Uint64(data[byteOff:])
		return (val >> bitOff) & bitMask(bits)
	default: // 9 bytes needed when bitOff > 0 and bits == 64
		part1 := binary.LittleEndian.Uint64(data[byteOff:]) >> bitOff
		part2 := uint64(data[byteOff+8]) << (64 - bitOff)
		return (part1 | part2) & bitMask(bits)
	}
}

// readUint64MSB reads up to 64 bits from the buffer (MSB first).
func readUint64MSB(data []byte, byteOff uint64, bitOff, bits uint8) uint64 {
	if bits == 0 {
		return 0
	}
	bytesNeeded := (bitOff + bits + 7) / 8
	switch bytesNeeded {
	case 1:
		shift := 8 - bitOff - bits
		return uint64(data[byteOff]>>shift) & bitMask(bits)
	case 2:
		val := binary.BigEndian.Uint16(data[byteOff:])
		shift := 16 - bitOff - bits
		return uint64(val>>shift) & bitMask(bits)
	case 3:
		val := uint32(data[byteOff])<<16 | uint32(data[byteOff+1])<<8 | uint32(data[byteOff+2])
		shift := 24 - bitOff - bits
		return uint64(val>>shift) & bitMask(bits)
	case 4:
		val := binary.BigEndian.Uint32(data[byteOff:])
		shift := 32 - bitOff - bits
		return uint64(val>>shift) & bitMask(bits)
	case 5:
		val := uint64(data[byteOff])<<32 | uint64(binary.BigEndian.Uint32(data[byteOff+1:]))
		shift := 40 - bitOff - bits
		return (val >> shift) & bitMask(bits)
	case 6:
		val := uint64(binary.BigEndian.Uint16(data[byteOff:]))<<32 | uint64(binary.BigEndian.Uint32(data[byteOff+2:]))
		shift := 48 - bitOff - bits
		return (val >> shift) & bitMask(bits)
	case 7:
		val := uint64(data[byteOff])<<48 | uint64(data[byteOff+1])<<40 | uint64(data[byteOff+2])<<32 |
			uint64(binary.BigEndian.Uint32(data[byteOff+3:]))
		shift := 56 - bitOff - bits
		return (val >> shift) & bitMask(bits)
	case 8:
		val := binary.BigEndian.Uint64(data[byteOff:])
		shift := 64 - bitOff - bits
		return (val >> shift) & bitMask(bits)
	default: // 9 bytes needed when bitOff > 0 and bits == 64
		part1 := binary.BigEndian.Uint64(data[byteOff:]) << bitOff
		part2 := uint64(data[byteOff+8]) >> (8 - bitOff)
		return (part1 | part2) & bitMask(bits)
	}
}

// PeekUint8 peeks up to 8 bits without advancing position.
func (r *Reader) PeekUint8(bits uint8) (uint8, error) {
	if bits > 8 {
		return 0, errors.New("bits must be <= 8")
	}
	if r.pos.Add(NewSize(0, uint64(bits))) > r.end {
		return 0, io.EOF
	}
	if r.msbFirst {
		return readUint8MSB(r.data, r.pos.Bytes(), r.pos.Bits(), bits), nil
	}
	return readUint8(r.data, r.pos.Bytes(), r.pos.Bits(), bits), nil
}

// PeekUint16 peeks up to 16 bits without advancing position.
func (r *Reader) PeekUint16(bits uint8) (uint16, error) {
	if bits > 16 {
		return 0, errors.New("bits must be <= 16")
	}
	if r.pos.Add(NewSize(0, uint64(bits))) > r.end {
		return 0, io.EOF
	}
	if r.msbFirst {
		return readUint16MSB(r.data, r.pos.Bytes(), r.pos.Bits(), bits), nil
	}
	return readUint16(r.data, r.pos.Bytes(), r.pos.Bits(), bits), nil
}

// PeekUint32 peeks up to 32 bits without advancing position.
func (r *Reader) PeekUint32(bits uint8) (uint32, error) {
	if bits > 32 {
		return 0, errors.New("bits must be <= 32")
	}
	if r.pos.Add(NewSize(0, uint64(bits))) > r.end {
		return 0, io.EOF
	}
	if r.msbFirst {
		return readUint32MSB(r.data, r.pos.Bytes(), r.pos.Bits(), bits), nil
	}
	return readUint32(r.data, r.pos.Bytes(), r.pos.Bits(), bits), nil
}

// PeekUint64 peeks up to 64 bits without advancing position.
func (r *Reader) PeekUint64(bits uint8) (uint64, error) {
	if bits > 64 {
		return 0, errors.New("bits must be <= 64")
	}
	if r.pos.Add(NewSize(0, uint64(bits))) > r.end {
		return 0, io.EOF
	}
	if r.msbFirst {
		return readUint64MSB(r.data, r.pos.Bytes(), r.pos.Bits(), bits), nil
	}
	return readUint64(r.data, r.pos.Bytes(), r.pos.Bits(), bits), nil
}

// ReadUint8 reads up to 8 bits and advances position.
func (r *Reader) ReadUint8(bits uint8) (uint8, error) {
	val, err := r.PeekUint8(bits)
	if err != nil {
		return 0, err
	}
	r.pos = r.pos.Add(NewSize(0, uint64(bits)))
	return val, nil
}

// ReadUint16 reads up to 16 bits and advances position.
func (r *Reader) ReadUint16(bits uint8) (uint16, error) {
	val, err := r.PeekUint16(bits)
	if err != nil {
		return 0, err
	}
	r.pos = r.pos.Add(NewSize(0, uint64(bits)))
	return val, nil
}

// ReadUint32 reads up to 32 bits and advances position.
func (r *Reader) ReadUint32(bits uint8) (uint32, error) {
	val, err := r.PeekUint32(bits)
	if err != nil {
		return 0, err
	}
	r.pos = r.pos.Add(NewSize(0, uint64(bits)))
	return val, nil
}

// ReadUint64 reads up to 64 bits and advances position.
func (r *Reader) ReadUint64(bits uint8) (uint64, error) {
	val, err := r.PeekUint64(bits)
	if err != nil {
		return 0, err
	}
	r.pos = r.pos.Add(NewSize(0, uint64(bits)))
	return val, nil
}

// ReadBit reads a single bit and returns it as a bool.
func (r *Reader) ReadBit() (bool, error) {
	val, err := r.ReadUint8(1)
	return val != 0, err
}

// ReadInt8 reads up to 8 bits as a signed integer.
func (r *Reader) ReadInt8(bits uint8) (int8, error) {
	val, err := r.ReadUint8(bits)
	return int8(val), err
}

// ReadInt16 reads up to 16 bits as a signed integer.
func (r *Reader) ReadInt16(bits uint8) (int16, error) {
	val, err := r.ReadUint16(bits)
	return int16(val), err
}

// ReadInt32 reads up to 32 bits as a signed integer.
func (r *Reader) ReadInt32(bits uint8) (int32, error) {
	val, err := r.ReadUint32(bits)
	return int32(val), err
}

// ReadInt64 reads up to 64 bits as a signed integer.
func (r *Reader) ReadInt64(bits uint8) (int64, error) {
	val, err := r.ReadUint64(bits)
	return int64(val), err
}

// ReadFloat32 reads 32 bits as a float32.
func (r *Reader) ReadFloat32() (float32, error) {
	val, err := r.ReadUint32(32)
	if err != nil {
		return 0, err
	}
	return float32FromBits(val), nil
}

// ReadFloat64 reads 64 bits as a float64.
func (r *Reader) ReadFloat64() (float64, error) {
	val, err := r.ReadUint64(64)
	if err != nil {
		return 0, err
	}
	return float64FromBits(val), nil
}

// ReadString reads a null-terminated string.
func (r *Reader) ReadString() (string, error) {
	var buf []byte
	for {
		b, err := r.ReadUint8(8)
		if err != nil {
			return string(buf), err
		}
		if b == 0 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf), nil
}

// ReadStringN reads a null-terminated string with a maximum length.
// If the string is longer than maxLen, it reads maxLen bytes and does not
// look for a null terminator.
func (r *Reader) ReadStringN(maxLen int) (string, error) {
	var buf []byte
	for i := 0; i < maxLen; i++ {
		b, err := r.ReadUint8(8)
		if err != nil {
			return string(buf), err
		}
		if b == 0 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf), nil
}

// ReadBytes reads n bytes into a new slice.
func (r *Reader) ReadBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		b, err := r.ReadUint8(8)
		if err != nil {
			return buf[:i], err
		}
		buf[i] = b
	}
	return buf, nil
}

// ReadUvarint reads a variable-length unsigned integer (up to 32 bits).
// Uses the same encoding as Source engine: 7 bits per byte, MSB indicates continuation.
func (r *Reader) ReadUvarint() (uint32, error) {
	var result uint32
	for i := uint8(0); i < 5; i++ {
		b, err := r.ReadUint8(8)
		if err != nil {
			return result, err
		}
		result |= uint32(b&0x7F) << (i * 7)
		if b&0x80 == 0 {
			break
		}
	}
	return result, nil
}

// ReadVarint reads a variable-length signed integer.
func (r *Reader) ReadVarint() (int32, error) {
	val, err := r.ReadUvarint()
	return int32(val), err
}

// Clone returns a copy of the reader with the same position.
func (r *Reader) Clone() *Reader {
	return &Reader{
		data:  r.data,
		start: r.start,
		pos:   r.pos,
		end:   r.end,
	}
}

// Span returns a new reader that covers a sub-range.
func (r *Reader) Span(start, end BitPos) (*Reader, error) {
	if start > end {
		return nil, errors.New("start must be <= end")
	}
	if start < r.start {
		return nil, errors.New("start out of bounds")
	}
	if end > r.end {
		return nil, errors.New("end out of bounds")
	}
	return &Reader{
		data:  r.data,
		start: start,
		pos:   start,
		end:   end,
	}, nil
}

// TakeSpan returns a new reader covering length bits from current position
// and advances the position by length.
func (r *Reader) TakeSpan(length BitSize) (*Reader, error) {
	endPos := r.pos.Add(length)
	if endPos > r.end {
		return nil, io.EOF
	}
	span, err := r.Span(r.pos, endPos)
	if err != nil {
		return nil, err
	}
	r.pos = endPos
	return span, nil
}

// Data returns the underlying data slice.
func (r *Reader) Data() []byte {
	return r.data
}
