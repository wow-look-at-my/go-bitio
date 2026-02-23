package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/wow-look-at-my/go-bitio"
)

const (
	TypeUint8   = 0
	TypeUint16  = 1
	TypeUint32  = 2
	TypeUint64  = 3
	TypeString  = 4
	TypeFloat32 = 5
	TypeFloat64 = 6
	TypeVaruint = 7
)

func main() {
	const numOps = 100000

	rng := rand.New(rand.NewSource(rand.Int63()))
	w := bitio.NewWriterAutoGrow()

	for i := 0; i < numOps; i++ {
		typeTag := uint8(rng.Intn(8))
		w.WriteUint8(typeTag, 3)

		switch typeTag {
		case TypeUint8:
			bitCount := uint8(1 + rng.Intn(8))
			w.WriteUint8(bitCount, 4)
			w.WriteUint8(uint8(rng.Uint32()), bitCount)

		case TypeUint16:
			bitCount := uint8(1 + rng.Intn(16))
			w.WriteUint8(bitCount, 5)
			w.WriteUint16(uint16(rng.Uint32()), bitCount)

		case TypeUint32:
			bitCount := uint8(1 + rng.Intn(32))
			w.WriteUint8(bitCount, 6)
			w.WriteUint32(rng.Uint32(), bitCount)

		case TypeUint64:
			bitCount := uint8(1 + rng.Intn(64))
			w.WriteUint8(bitCount, 7)
			w.WriteUint64(rng.Uint64(), bitCount)

		case TypeString:
			strLen := 1 + rng.Intn(32)
			str := make([]byte, strLen)
			for j := 0; j < strLen; j++ {
				str[j] = byte(33 + rng.Intn(94))
			}
			w.WriteString(string(str))

		case TypeFloat32:
			w.WriteFloat32(rng.Float32())

		case TypeFloat64:
			w.WriteFloat64(rng.Float64())

		case TypeVaruint:
			w.WriteUvarint(rng.Uint32())
		}
	}

	if err := os.WriteFile("testdata/bench_stream.bin", w.Bytes(), 0644); err != nil {
		panic(err)
	}

	fmt.Printf("Written %d bytes\n", len(w.Bytes()))
}
