package bitstream

import "testing"

func TestBitReader(t *testing.T) {
	t.Run("U16R", func(t *testing.T) {
		readers := []struct {
			name   string
			reader interface {
				U16R(bits int, n int) (b uint16)
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
			expected uint16
		}{
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
					result := r.reader.U16R(tt.bits, tt.n)
					if result != tt.expected {
						t.Errorf("U16R(%d, %d) = %08b; want %08b", tt.bits, tt.n, result, tt.expected)
					}
				}
			})
		}
	})
	t.Run("U16R_withPadding", func(t *testing.T) {
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
			expected uint16
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
			result := reader.U16R(tt.bits, tt.n)
			if result != tt.expected {
				t.Errorf("U16R(%d, %d) with lp %d and rp %d = %08b; want %08b", tt.bits, tt.n, tt.lp, tt.rp, result, tt.expected)
			}
		}
	})
}

func TestNewBitReader_Panic(t *testing.T) {
	tests := []struct {
		name      string
		leftPadd  int
		rightPadd int
	}{
		{"padding sum equals element size", 4, 4},
		{"padding sum exceeds element size", 5, 5},
		{"left padding only equals size", 8, 0},
		{"right padding only equals size", 0, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("NewBitReader(leftPadd=%d, rightPadd=%d) did not panic", tt.leftPadd, tt.rightPadd)
				}
			}()
			// Should panic
			NewBitReader([]uint8{0xFF}, tt.leftPadd, tt.rightPadd)
		})
	}
}
