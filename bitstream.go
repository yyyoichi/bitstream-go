package bitstream

import (
	"sync"
	"unsafe"
)

type Unsigned interface {
	~uint64 | ~uint32 | ~uint16 | ~uint8 | ~uint
}

// BitReader provides bit-level reading operations on integer slice data.
// It treats the data as a continuous bit stream, allowing precise bit extraction.
type BitReader[T Unsigned] struct {
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
func NewBitReader[T Unsigned](data []T, leftPadd, rightPadd int) *BitReader[T] {
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

// SetBits sets the total number of valid bits in the BitReader.
// Data beyond the specified bits position will be treated as zero,
// regardless of the actual padding configuration.
// This is useful for limiting the readable range within the data.
func (r *BitReader[T]) SetBits(bits int) {
	r.bits = bits
}

// Read8R reads a specified number of bits from the n-th position in the data.
// bits specifies how many bits to read (up to 8 bits).
// n specifies which block to read (0-indexed).
// Returns the bits as a uint16 value, right-aligned (LSB-aligned).
//
// Panics if bits > 8, as uint16 can only hold 8 bits.
func (r *BitReader[T]) Read8R(bits, n int) uint8 {
	if bits > 8 {
		panic("bitstream: cannot read more than 8 bits into uint16")
	}
	return uint8(r.right(bits, n))
}

// Read16R reads a specified number of bits from the n-th position in the data.
// bits specifies how many bits to read (up to 16 bits).
// n specifies which block to read (0-indexed).
// Returns the bits as a uint16 value, right-aligned (LSB-aligned).
//
// Panics if bits > 16, as uint16 can only hold 16 bits.
func (r *BitReader[T]) Read16R(bits, n int) (b uint16) {
	if bits > 16 {
		panic("bitstream: cannot read more than 16 bits into uint16")
	}
	return uint16(r.right(bits, n))
}

// Read32R reads a specified number of bits from the n-th position in the data.
// bits specifies how many bits to read (up to 32 bits).
// n specifies which block to read (0-indexed).
// Returns the bits as a uint16 value, right-aligned (LSB-aligned).
//
// Panics if bits > 32, as uint16 can only hold 32 bits.
func (r *BitReader[T]) Read32R(bits, n int) (b uint32) {
	if bits > 32 {
		panic("bitstream: cannot read more than 32 bits into uint16")
	}
	return uint32(r.right(bits, n))
}

// Read64R reads a specified number of bits from the n-th position in the data.
// bits specifies how many bits to read (up to 64 bits).
// n specifies which block to read (0-indexed).
// Returns the bits as a uint16 value, right-aligned (LSB-aligned).
//
// Panics if bits > 64, as uint16 can only hold 64 bits.
func (r *BitReader[T]) Read64R(bits, n int) (b uint64) {
	if bits > 64 {
		panic("bitstream: cannot read more than 64 bits into uint16")
	}
	return r.right(bits, n)
}

// Bits returns the total number of valid bits in the BitReader.
func (r *BitReader[T]) Bits() int {
	return r.bits
}

func (r *BitReader[T]) right(bits, n int) (b uint64) {
	s := min(n*bits, r.bits)
	e := min(s+bits, r.bits)
	for i := s; i < e; i++ {
		b <<= 1
		mask := r.msb >> (i % r.s)
		if r.data[i/r.s]&mask != 0 {
			b |= 1
		}
	}
	for range bits - (e - s) {
		b <<= 1
	}
	return
}

// BitWriter provides bit-level writing operations to integer slice data.
// It treats the destination as a continuous bit stream, allowing precise bit insertion.
// BitWriter is safe for concurrent use.
type BitWriter[T Unsigned] struct {
	mu   *sync.Mutex
	data []T // Destination data to write bits into
	bits int // Total number of bits written so far
	s    int // Number of valid bits per element (element size - left padding - right padding)
	msb  T   // MSB mask for the valid bit range
	lp   int // Left padding bits
	rp   int // Right padding bits
}

// NewBitWriter creates a new BitWriter for writing bits to integer slice data.
// leftPadd specifies how many upper bits in each element should be treated as padding.
// rightPadd specifies how many lower bits in each element should be treated as padding.
// The writer will only write bits to the valid range between paddings.
//
// Panics if leftPadd + rightPadd >= element bit size, as this would leave no valid bits to write.
func NewBitWriter[T Unsigned](leftPadd, rightPadd int) *BitWriter[T] {
	var zero T
	size := int(unsafe.Sizeof(zero)) * 8
	if leftPadd+rightPadd >= size {
		panic("bitstream: padding sum must be less than element bit size")
	}
	s := size - leftPadd - rightPadd
	return &BitWriter[T]{
		mu:   &sync.Mutex{},
		data: make([]T, 0),
		bits: 0,
		s:    s,
		msb:  T(1) << (size - leftPadd - 1),
		lp:   leftPadd,
		rp:   rightPadd,
	}
}

// Write8 writes the specified bits from a uint8 value to the stream.
// leftPadd specifies how many upper bits to skip in the source data.
// bits specifies how many bits to write after skipping leftPadd bits.
//
// Panics if leftPadd + bits > 8.
func (w *BitWriter[T]) Write8(leftPadd, bits int, data uint8) {
	if leftPadd+bits > 8 {
		panic("bitstream: padding and bits exceed uint8 size")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := leftPadd; i < leftPadd+bits; i++ {
		w.write(data&(1<<(7-i)) != 0)
	}
}

// Write16 writes the specified bits from a uint16 value to the stream.
// leftPadd specifies how many upper bits to skip in the source data.
// bits specifies how many bits to write after skipping leftPadd bits.
//
// Panics if leftPadd + bits > 16.
func (w *BitWriter[T]) Write16(leftPadd, bits int, data uint16) {
	if leftPadd+bits > 16 {
		panic("bitstream: padding and bits exceed uint16 size")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := leftPadd; i < leftPadd+bits; i++ {
		w.write(data&(1<<(15-i)) != 0)
	}
}

// Write32 writes the specified bits from a uint32 value to the stream.
// leftPadd specifies how many upper bits to skip in the source data.
// bits specifies how many bits to write after skipping leftPadd bits.
//
// Panics if leftPadd + bits > 32.
func (w *BitWriter[T]) Write32(leftPadd, bits int, data uint32) {
	if leftPadd+bits > 32 {
		panic("bitstream: padding and bits exceed uint32 size")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := leftPadd; i < leftPadd+bits; i++ {
		w.write(data&(1<<(31-i)) != 0)
	}
}

// Write64 writes the specified bits from a uint64 value to the stream.
// leftPadd specifies how many upper bits to skip in the source data.
// bits specifies how many bits to write after skipping leftPadd bits.
//
// Panics if leftPadd + bits > 64.
func (w *BitWriter[T]) Write64(leftPadd, bits int, data uint64) {
	if leftPadd+bits > 64 {
		panic("bitstream: padding and bits exceed uint64 size")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := leftPadd; i < leftPadd+bits; i++ {
		w.write(data&(1<<(63-i)) != 0)
	}
}

// WriteBool writes a single boolean value as one bit to the stream.
func (w *BitWriter[T]) WriteBool(data bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.write(data)
}

// Data returns the accumulated data slice.
// Use Bits() to get the total number of valid bits written.
func (w *BitWriter[T]) Data() []T {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.data
}

// AnyData returns the accumulated data as an any type.
// This is useful when the exact type of the underlying data slice is not known at compile time.
// Use Bits() to get the total number of valid bits written.
func (w *BitWriter[T]) AnyData() any {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.data
}

// Bits returns the total number of valid bits in the BitWriter.
func (r *BitWriter[T]) Bits() int {
	return r.bits
}

func (w *BitWriter[T]) write(b bool) {
	idx := w.bits / w.s
	if idx >= len(w.data) {
		w.data = append(w.data, 0)
	}
	if b {
		w.data[idx] |= w.msb >> (w.bits % w.s)
	}
	w.bits += 1
}
