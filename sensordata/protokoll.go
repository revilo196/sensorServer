package sensordata

import (
	"bytes"
	"encoding/binary"
	"math"
)

func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

func uint64frombytes(bytes []byte) uint64 {
	bits := binary.LittleEndian.Uint64(bytes)
	return bits
}

func DecodeParsePackage(pack []byte) (int, []SensorWert) {
	n := len(pack) / 16
	values := make([]SensorWert, n)

	for i := 0; i < n; i++ {
		val := pack[i*16 : (i+1)*16]
		buf1 := bytes.NewReader(val)
		var f1, f2 float32
		var in uint64

		binary.Read(buf1, binary.LittleEndian, &f1)
		binary.Read(buf1, binary.LittleEndian, &f2)
		binary.Read(buf1, binary.LittleEndian, &in)

		values[i] = SensorWert{f1, f2, in}
	}

	return n, values
}