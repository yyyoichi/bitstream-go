package bitstream

import "unsafe"

type IntegerType interface {
	~uint64 | ~uint32 | ~uint16 | ~uint8 | ~uint
}

// BitReader provides bit-level reading operations on integer slice data.
// It treats the data as a continuous bit stream, allowing precise bit extraction.
type BitReader[T IntegerType] struct {
	data []T // Source data to read bits from
	bits int // Total number of valid bits in the data
	s    int // Number of valid bits per element (element size - left padding - right padding)
	msb  T   // MSB mask for the valid bit range
}

// NewBitReader creates a new BitReader for manipulating bits from integer slice data.
// leftPadd specifies how many upper bits in each element should be treated as padding.
// rightPadd specifies how many lower bits in each element should be treated as padding.
// For example, if each element has 2 bits of padding at the top and 1 bit at the bottom,
// set leftPadd=2 and rightPadd=1.
// The reader will only access bits from position leftPadd to (element size - rightPadd).
func NewBitReader[T IntegerType](data []T, leftPadd, rightPadd int) *BitReader[T] {
	var zero T
	var s = int(unsafe.Sizeof(zero))*8 - leftPadd - rightPadd
	return &BitReader[T]{
		data: data,
		bits: len(data) * s,
		s:    s,
		msb:  T(1) << (s - 1 + rightPadd),
	}
}

// U16R reads a specified number of bits from the n-th position in the data.
// bits specifies how many bits to read (up to 16 bits).
// n specifies which block to read (0-indexed).
// Returns the bits as a uint16 value, right-aligned (LSB-aligned).
func (r BitReader[T]) U16R(bits, n int) (b uint16) {
	s := min(n*bits, r.bits)
	e := min(s+bits, r.bits)
	for i := s; i < e; i++ {
		b <<= 1
		mask := r.msb >> (i % r.s)
		if r.data[i/r.s]&mask != 0 {
			b |= 1
		}
	}
	return
}
