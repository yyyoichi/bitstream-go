package bitstream

import "testing"

func TestBitReader(t *testing.T) {
	t.Run("right", func(t *testing.T) {
		readers := []struct {
			name   string
			reader interface {
				right(bits int, n int) (b uint64)
			}
		}{
			{"uint8", NewBitReader([]uint8{
				0b10101100,
				0b11100011,
				0b11000011,
				0b11100000,
			}, 0, 0)},
			{"uint16", NewBitReader([]uint16{
				0b1010110011100011,
				0b1100001111100000,
			}, 0, 0)},
			{"uint32", NewBitReader([]uint32{
				0b10101100111000111100001111100000,
			}, 0, 0)},
			{"uint64", NewBitReader([]uint64{
				uint64(0b10101100111000111100001111100000) << 32,
			}, 0, 0)},
		}
		test := []struct {
			bits     int
			n        int
			expected uint64
		}{
			{0, 0, 0b0},
			{0, 20, 0b0},
			{1, 0, 0b1},
			{1, 63, 0b0},
			{3, 0, 0b101},
			{3, 2, 0b1},
			{4, 0, 0b1010},
			{8, 0, 0b10101100},
			{12, 0, 0b101011001110},
			{12, 1, 0b001111000011},
			{16, 0, 0b1010110011100011},
			{16, 1, 0b1100001111100000},
			{16, 2, 0b0},
		}
		for _, r := range readers {
			t.Run(r.name, func(t *testing.T) {
				for _, tt := range test {
					result := r.reader.right(tt.bits, tt.n)
					if result != tt.expected {
						t.Errorf("right(%d, %d) = %08b; want %08b", tt.bits, tt.n, result, tt.expected)
					}
				}
			})
		}
	})
	t.Run("right_BoundaryValue", func(t *testing.T) {
		reader := NewBitReader([]uint8{0b11111111}, 0, 0)
		test := []struct {
			bits     int
			n        int
			expected uint64
		}{
			{1, 8, 0b0},
			{2, 4, 0b00},
			{3, 2, 0b110},
			{5, 1, 0b11100},
			{6, 1, 0b110000},
			{7, 1, 0b1000000},
		}
		for _, tt := range test {
			result := reader.right(tt.bits, tt.n)
			if result != tt.expected {
				t.Errorf("right(%d, %d) = %08b; want %08b", tt.bits, tt.n, result, tt.expected)
			}
		}
	})
	t.Run("right_withPadding", func(t *testing.T) {
		var data = []uint8{
			0b10101100,
			0b11100011,
			0b11000011,
			0b11100000,
		}
		test := []struct {
			lp, rp   int
			bits     int
			n        int
			expected uint64
		}{
			{1, 0, 1, 0, 0b0},
			{1, 0, 1, 1, 0b1},
			{2, 0, 3, 0, 0b101},
			{2, 0, 3, 6, 0b100},
			{2, 0, 4, 1, 0b0010},
			{2, 0, 6, 3, 0b100000},
			{3, 0, 4, 1, 0b0},
			{4, 0, 4, 1, 0b11},
			{5, 0, 4, 1, 0b1101},
			{6, 0, 4, 1, 0b1100},
			{7, 0, 4, 1, 0b0},

			{0, 1, 1, 0, 0b1},
			{0, 1, 1, 7, 0b1},
			{0, 2, 3, 0, 0b101},
			{0, 2, 3, 2, 0b111},
			{0, 2, 4, 1, 0b1111},
			{0, 3, 4, 1, 0b1111},
			{0, 4, 4, 1, 0b1110},
			{0, 5, 4, 1, 0b1111},
			{0, 6, 4, 1, 0b1111},
			{0, 7, 4, 1, 0b0},

			{1, 1, 1, 0, 0b0},
			{1, 1, 1, 1, 0b1},
			{1, 1, 3, 1, 0b110},
			{1, 1, 3, 6, 0b110},
			{1, 1, 5, 3, 0b111},
		}
		for _, tt := range test {
			reader := NewBitReader(data, tt.lp, tt.rp)
			result := reader.right(tt.bits, tt.n)
			if result != tt.expected {
				t.Errorf("right(%d, %d) with lp %d and rp %d = %08b; want %08b", tt.bits, tt.n, tt.lp, tt.rp, result, tt.expected)
			}
		}
	})
	t.Run("U16R_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when bits > 16")
			}
		}()
		reader := NewBitReader([]uint8{0xFF, 0xFF, 0xFF}, 0, 0)
		reader.right(17, 0) // Should panic
	})

	t.Run("SetBits", func(t *testing.T) {
		data := []uint8{0b11111111}
		reader := NewBitReader(data, 1, 1)
		reader.SetBits(5)
		r := reader.right(8, 0)
		if r != 0b11111000 {
			t.Errorf("SetBits or right failed: got %08b; want %08b", r, 0b11111000)
		}
	})
}

func TestBitWriter(t *testing.T) {
	t.Run("writeU8", func(t *testing.T) {
		writer := NewBitWriter[uint64](0, 0)
		writer.Write8(0, 8, 255)
		if writer.bits != 8 {
			t.Errorf("expected bits to be 8, got %d", writer.bits)
		}
		if writer.data[0] != uint64(255)<<56 {
			t.Errorf("expected data[0] to be %08b, got %08b", uint64(255)<<56, writer.data[0])
		}
		writer.Write8(1, 7, 255)
		if writer.bits != 15 {
			t.Errorf("expected bits to be 15, got %d", writer.bits)
		}
		if writer.data[0] != (uint64(255)<<56)|(uint64(127)<<49) {
			t.Errorf("expected data[0] to be %08b, got %08b", (uint64(255)<<56)|(uint64(127)<<49), writer.data[0])
		}
		writer.Write8(0, 8, 0)
		writer.Write8(6, 1, 255)
		if writer.bits != 24 {
			t.Errorf("expected bits to be 24, got %d", writer.bits)
		}
		if writer.data[0] != (uint64(255)<<56)|(uint64(127)<<49)|(uint64(0b0)<<41)|(uint64(0b1)<<40) {
			t.Errorf("expected data[0] to be %08b, got %08b", (uint64(255)<<56)|(uint64(127)<<49)|(uint64(0b0)<<41)|(uint64(0b1)<<40), writer.data[0])
		}
		if len(writer.data) != 1 {
			t.Errorf("expected 1 element in data, got %d", len(writer.data))
		}
	})
	t.Run("write16", func(t *testing.T) {
		writer := NewBitWriter[uint64](0, 0)
		writer.Write16(0, 16, 65535)
		if writer.bits != 16 {
			t.Errorf("expected bits to be 16, got %d", writer.bits)
		}
		if writer.data[0] != uint64(65535)<<48 {
			t.Errorf("expected data[0] to be %08b, got %08b", uint64(65535)<<48, writer.data[0])
		}
		writer.Write16(2, 14, 65535)
		if writer.bits != 30 {
			t.Errorf("expected bits to be 30, got %d", writer.bits)
		}
		if writer.data[0] != (uint64(65535)<<48)|(uint64(0x3FFF)<<34) {
			t.Errorf("expected data[0] to be %08b, got %08b", (uint64(65535)<<48)|(uint64(0x3FFF)<<34), writer.data[0])
		}
		writer.Write16(0, 16, 0)
		writer.Write16(5, 2, 65535)
		if writer.bits != 48 {
			t.Errorf("expected bits to be 48, got %d", writer.bits)
		}
		if writer.data[0] != (uint64(65535)<<48)|(uint64(0x3FFF)<<34)|(uint64(0b0)<<18)|(uint64(0b11)<<16) {
			t.Errorf("expected data[0] to be %08b, got %08b", (uint64(65535)<<48)|(uint64(0x3FFF)<<34)|(uint64(0b0)<<18)|(uint64(0b11)<<16), writer.data[0])
		}
		if len(writer.data) != 1 {
			t.Errorf("expected 1 element in data, got %d", len(writer.data))
		}
	})
	t.Run("write32", func(t *testing.T) {
		writer := NewBitWriter[uint64](0, 0)
		writer.Write32(0, 32, 0xFFFFFFFF)
		if writer.bits != 32 {
			t.Errorf("expected bits to be 32, got %d", writer.bits)
		}
		if writer.data[0] != uint64(0xFFFFFFFF)<<32 {
			t.Errorf("expected data[0] to be %08b, got %08b", uint64(0xFFFFFFFF)<<32, writer.data[0])
		}
		writer.Write32(4, 28, 0xFFFFFFFF)
		if writer.bits != 60 {
			t.Errorf("expected bits to be 60, got %d", writer.bits)
		}
		if writer.data[0] != (uint64(0xFFFFFFFF)<<32)|(uint64(0x0FFFFFFF)<<4) {
			t.Errorf("expected data[0] to be %08b, got %08b", (uint64(0xFFFFFFFF)<<32)|(uint64(0x0FFFFFFF)<<4), writer.data[0])
		}
		writer.Write32(0, 32, 0)
		writer.Write32(10, 4, 0xFFFFFFFF)
		if writer.bits != 96 {
			t.Errorf("expected bits to be 96, got %d", writer.bits)
		}
		if writer.data[0] != uint64(0xFFFFFFFF)<<32|(uint64(0xFFFFFFF)<<4) {
			t.Errorf("expected data[0] to be %064b, got %064b", uint64(0xFFFFFFFF)<<32|(uint64(0xFFFFFFF)<<4), writer.data[0])
		}
		if writer.data[1] != uint64(0xF)<<32 {
			t.Errorf("expected data[1] to be %016b, got %016b", uint64(0xF)<<32, writer.data[1])
		}
		if len(writer.data) != 2 {
			t.Errorf("expected 2 elements in data, got %d", len(writer.data))
		}
	})
	t.Run("write64", func(t *testing.T) {
		writer := NewBitWriter[uint64](0, 0)
		writer.Write64(0, 64, 0xFFFFFFFFFFFFFFFF)
		if writer.bits != 64 {
			t.Errorf("expected bits to be 64, got %d", writer.bits)
		}
		if writer.data[0] != 0xFFFFFFFFFFFFFFFF {
			t.Errorf("expected data[0] to be %016x, got %016x", uint64(0xFFFFFFFFFFFFFFFF), writer.data[0])
		}
		writer.Write64(8, 56, 0xFFFFFFFFFFFFFFFF)
		if writer.bits != 120 {
			t.Errorf("expected bits to be 120, got %d", writer.bits)
		}
		if len(writer.data) != 2 {
			t.Errorf("expected 2 elements in data, got %d", len(writer.data))
		}
		if writer.data[1] != uint64(0x00FFFFFFFFFFFFFF)<<8 {
			t.Errorf("expected data[1] to be %016x, got %016x", uint64(0x00FFFFFFFFFFFFFF)<<8, writer.data[1])
		}
		writer.Write64(0, 64, 0)
		writer.Write64(20, 8, 0xFFFFFFFFFFFFFFFF)
		if writer.bits != 192 {
			t.Errorf("expected bits to be 192, got %d", writer.bits)
		}
		if writer.data[1] != uint64(0x00FFFFFFFFFFFFFF)<<8 {
			t.Errorf("expected data[1] to be %016x, got %016x", uint64(0x00FFFFFFFFFFFFFF)<<8, writer.data[1])
		}
		if writer.data[2] != uint64(0xFF) {
			t.Errorf("expected data[2] to be %016x, got %016x", uint64(0xFF), writer.data[2])
		}
		if len(writer.data) != 3 {
			t.Errorf("expected 3 elements in data, got %d", len(writer.data))
		}
	})
	t.Run("writeBool", func(t *testing.T) {
		writer := NewBitWriter[uint64](0, 0)
		writer.WriteBool(true)
		if writer.bits != 1 {
			t.Errorf("expected bits to be 1, got %d", writer.bits)
		}
		if writer.data[0] != uint64(1)<<63 {
			t.Errorf("expected data[0] to be %08b, got %08b", uint64(1)<<63, writer.data[0])
		}
		writer.WriteBool(false)
		writer.WriteBool(true)
		if writer.bits != 3 {
			t.Errorf("expected bits to be 2, got %d", writer.bits)
		}
		if writer.data[0] != (uint64(1)<<63)|(uint64(01)<<61) {
			t.Errorf("expected data[0] to be %08b, got %08b", (uint64(1)<<63)|(uint64(0)<<61), writer.data[0])
		}
	})
	t.Run("writeWithPadding", func(t *testing.T) {
		writer := NewBitWriter[uint8](1, 0)
		writer.Write8(0, 8, 0xFF)
		if writer.bits != 8 {
			t.Errorf("expected bits to be 8, got %d", writer.bits)
		}
		if writer.data[0] != uint8(0b01111111) {
			t.Errorf("expected data[0] to be %08b, got %08b", uint8(0b01111111), writer.data[0])
		}
		if writer.data[1] != uint8(0b01000000) {
			t.Errorf("expected data[1] to be %08b, got %08b", uint8(0b01000000), writer.data[1])
		}
		writer = NewBitWriter[uint8](1, 2)
		writer.Write8(0, 8, 0xFF)
		if writer.bits != 8 {
			t.Errorf("expected bits to be 8, got %d", writer.bits)
		}
		if writer.data[0] != uint8(0b01111100) {
			t.Errorf("expected data[0] to be %08b, got %08b", uint8(0b01111100), writer.data[0])
		}
		if writer.data[1] != uint8(0b01110000) {
			t.Errorf("expected data[1] to be %08b, got %08b", uint8(0b01110000), writer.data[1])
		}
	})
	t.Run("Data_uint8", func(t *testing.T) {
		writer := NewBitWriter[uint8](0, 0)
		writer.Write16(0, 16, 0xFFFF)
		data, bits := writer.Data()
		if bits != 16 {
			t.Errorf("expected bits to be 16, got %d", bits)
		}
		if len(data) != 2 {
			t.Errorf("expected data length to be 2, got %d", len(data))
		}
		if data[0] != 0xFF {
			t.Errorf("expected data[0] to be %08b, got %08b", 0xFF, data[0])
		}
		if data[1] != 0xFF {
			t.Errorf("expected data[1] to be %08b, got %08b", 0xFF, data[1])
		}
		writer = NewBitWriter[uint8](1, 1)
		writer.Write16(0, 16, 0xFFFF)
		data, bits = writer.Data()
		if bits != 16+3+2 {
			t.Errorf("expected bits to be 21, got %d", bits)
		}
		if len(data) != 3 {
			t.Errorf("expected data length to be 3, got %d", len(data))
		}
		if data[0] != 0b01111110 {
			t.Errorf("expected data[0] to be %08b, got %08b", 0b01111110, data[0])
		}
		if data[1] != 0b01111110 {
			t.Errorf("expected data[1] to be %08b, got %08b", 0b01111110, data[1])
		}
		if data[2] != 0b01111000 {
			t.Errorf("expected data[2] to be %08b, got %08b", 0b01111000, data[2])
		}
	})
	t.Run("Data_uint16", func(t *testing.T) {
		writer := NewBitWriter[uint16](0, 0)
		writer.Write32(0, 32, 0xFFFFFFFF)
		data, bits := writer.Data()
		if bits != 32 {
			t.Errorf("expected bits to be 32, got %d", bits)
		}
		if len(data) != 2 {
			t.Errorf("expected data length to be 2, got %d", len(data))
		}
		if data[0] != 0xFFFF {
			t.Errorf("expected data[0] to be %016b, got %016b", 0xFFFF, data[0])
		}
		if data[1] != 0xFFFF {
			t.Errorf("expected data[1] to be %016b, got %016b", 0xFFFF, data[1])
		}
		writer = NewBitWriter[uint16](2, 2)
		writer.Write32(0, 32, 0xFFFFFFFF)
		data, bits = writer.Data()
		if bits != 32+6+4 {
			t.Errorf("expected bits to be 42, got %d", bits)
		}
		if len(data) != 3 {
			t.Errorf("expected data length to be 3, got %d", len(data))
		}
		if data[0] != 0x3FFC {
			t.Errorf("expected data[0] to be %016b, got %016b", 0x3FFC, data[0])
		}
	})
	t.Run("Data_uint32", func(t *testing.T) {
		writer := NewBitWriter[uint32](0, 0)
		writer.Write64(0, 64, 0xFFFFFFFFFFFFFFFF)
		data, bits := writer.Data()
		if bits != 64 {
			t.Errorf("expected bits to be 64, got %d", bits)
		}
		if len(data) != 2 {
			t.Errorf("expected data length to be 2, got %d", len(data))
		}
		if data[0] != 0xFFFFFFFF {
			t.Errorf("expected data[0] to be %032b, got %032b", 0xFFFFFFFF, data[0])
		}
		if data[1] != 0xFFFFFFFF {
			t.Errorf("expected data[1] to be %032b, got %032b", 0xFFFFFFFF, data[1])
		}
		writer = NewBitWriter[uint32](4, 4)
		writer.Write64(0, 64, 0xFFFFFFFFFFFFFFFF)
		data, bits = writer.Data()
		if bits != 64+12+8 {
			t.Errorf("expected bits to be 88, got %d", bits)
		}
		if len(data) != 3 {
			t.Errorf("expected data length to be 3, got %d", len(data))
		}
		if data[0] != 0x0FFFFFF0 {
			t.Errorf("expected data[0] to be %032b, got %032b", 0x0FFFFFF0, data[0])
		}
		if data[1] != 0x0FFFFFF0 {
			t.Errorf("expected data[1] to be %032b, got %032b", 0x0FFFFFF0, data[1])
		}
		if data[2] != 0x0FFFF000 {
			t.Errorf("expected data[2] to be %032b, got %032b", 0x0FFFF000, data[2])
		}
	})
	t.Run("Data_uint64", func(t *testing.T) {
		writer := NewBitWriter[uint64](0, 0)
		writer.Write64(0, 64, 0xFFFFFFFFFFFFFFFF)
		writer.Write64(0, 64, 0xFFFFFFFFFFFFFFFF)
		data, bits := writer.Data()
		if bits != 128 {
			t.Errorf("expected bits to be 128, got %d", bits)
		}
		if len(data) != 2 {
			t.Errorf("expected data length to be 2, got %d", len(data))
		}
		if data[0] != 0xFFFFFFFFFFFFFFFF {
			t.Errorf("expected data[0] to be %064b, got %064b", uint64(0xFFFFFFFFFFFFFFFF), data[0])
		}
		if data[1] != 0xFFFFFFFFFFFFFFFF {
			t.Errorf("expected data[1] to be %064b, got %064b", uint64(0xFFFFFFFFFFFFFFFF), data[1])
		}

		writer = NewBitWriter[uint64](8, 8)
		writer.Write64(0, 64, 0xFFFFFFFFFFFFFFFF)
		data, bits = writer.Data()
		if bits != 64+16+8 {
			t.Errorf("expected bits to be 88, got %d", bits)
		}
		if len(data) != 2 {
			t.Errorf("expected data length to be 2, got %d", len(data))
		}
		if data[0] != uint64(0x00FFFFFFFFFFFF00) {
			t.Errorf("expected data[0] to be %064b, got %064b", uint64(0x00FFFFFFFFFFFF00), data[0])
		}
		if data[1] != uint64(0x00FFFF)<<40 {
			t.Errorf("expected data[1] to be %064b, got %064b", uint64(0x00FFFF)<<40, data[1])
		}
	})
}
