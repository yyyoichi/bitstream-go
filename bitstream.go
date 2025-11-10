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
//
// Panics if leftPadd + rightPadd >= element bit size, as this would leave no valid bits to read.
func NewBitReader[T IntegerType](data []T, leftPadd, rightPadd int) *BitReader[T] {
	var zero T
	size := int(unsafe.Sizeof(zero)) * 8
	if leftPadd+rightPadd >= size {
		panic("bitstream: padding sum must be less than element bit size")
	}
	s := size - leftPadd - rightPadd
	return &BitReader[T]{
		data: data,
		bits: len(data) * s,
		s:    s,
		msb:  T(1) << (size - leftPadd - 1),
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
