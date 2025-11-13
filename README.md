# bitstream-go

A Go library for efficient bit-level reading and writing operations on integer slices.

## Features

- **BitReader**: Read bits from integer slices with configurable padding
  - Block-based reading (`Read8R`, `Read16R`, etc.)
  - Cursor-based reading (`ReadBit`, `ReadBitAt`, `Pos`, `Seek`)
  - **Not thread-safe**: Use separate instances for concurrent access
- **BitWriter**: Write bits to integer slices with automatic allocation
  - Block-based writing (`Write8`, `Write16`, etc.)
  - Cursor-based writing (`WriteBit`, `WriteBitAt`, `Pos`, `Seek`)
  - **Thread-safe**: All operations protected by mutex
- Generic support for `uint8`, `uint16`, `uint32`, `uint64`, and `uint`
- Configurable left and right padding for each element
- Error handling following Go standard library conventions (`io.EOF`, `ErrNegativePosition`)

## Installation

```bash
go get github.com/yyyoichi/bitstream-go
```

## Usage

### BitReader

Read bits from an integer slice:

```go
package main

import (
    "fmt"
    "io"
    "github.com/yyyoichi/bitstream-go"
)

func main() {
    data := []uint8{0b10101100, 0b11100011}
    reader := bitstream.NewBitReader(data, 0, 0) // no padding
    
    // Block-based reading
    value := reader.Read16R(12, 0) // returns 0b101011001110
    value2 := reader.Read8R(4, 2)  // returns 0b1110
    
    // Cursor-based sequential reading
    bit, err := reader.ReadBit() // reads first bit (true), advances cursor
    if err == io.EOF {
        fmt.Println("end of data")
    }
    
    // Random access reading (doesn't move cursor)
    bit2, err := reader.ReadBitAt(5) // reads bit at position 5
    
    // Cursor positioning
    reader.Seek(8) // move cursor to position 8
    pos := reader.Pos() // get current position
}
```

### BitWriter

Write bits to create an integer slice:

```go
package main

import "github.com/yyyoichi/bitstream-go"

func main() {
    writer := bitstream.NewBitWriter[uint8](0, 0) // no padding
    
    // Block-based writing
    writer.Write16(0, 12, 0b101011001110)
    writer.WriteBool(true)
    
    // Cursor-based sequential writing
    writer.WriteBit(true)  // writes at current position, advances cursor
    writer.WriteBit(false)
    
    // Random access writing (doesn't move cursor)
    writer.WriteBitAt(5, true) // writes bit at position 5
    
    // Cursor positioning
    writer.Seek(0)  // move cursor to start
    pos := writer.Pos() // get current position
    
    // Get the result
    data := writer.Data()
    bits := writer.Bits()
    // data = []uint8{...}
    // bits = total number of bits written
}
```

### Padding

Configure padding to skip bits at the start or end of each element:

```go
// Skip 2 bits from the left and 1 bit from the right in each uint8
reader := bitstream.NewBitReader(data, 2, 1) // 5 valid bits per element

writer := bitstream.NewBitWriter[uint16](4, 0) // skip 4 top bits, 12 valid bits per element
```

## API

### BitReader

**Constructor:**
- `NewBitReader[T](data []T, leftPadd, rightPadd int) *BitReader[T]` - Create a new reader

**Block-based reading:**
- `Read8R(bits, n int) uint8` - Read up to 8 bits from n-th block
- `Read16R(bits, n int) uint16` - Read up to 16 bits from n-th block
- `Read32R(bits, n int) uint32` - Read up to 32 bits from n-th block
- `Read64R(bits, n int) uint64` - Read up to 64 bits from n-th block

**Cursor-based reading:**
- `ReadBit() (bool, error)` - Read one bit at cursor and advance (returns `io.EOF` if out of bounds)
- `ReadBitAt(pos int) (bool, error)` - Read one bit at position without moving cursor (returns `io.EOF` if out of bounds, `ErrNegativePosition` for negative positions)
- `Pos() int` - Get current cursor position
- `Seek(pos int) error` - Set cursor position (returns `ErrNegativePosition` for negative positions)

**Other:**
- `Bits() int` - Get total number of valid bits
- `SetBits(bits int)` - Limit readable range
- `Data() []T` - Get source data slice
- `AnyData() any` - Get source data as 'any' type

### BitWriter

**Constructor:**
- `NewBitWriter[T](leftPadd, rightPadd int) *BitWriter[T]` - Create a new writer

**Block-based writing:**
- `Write8(leftPadd, bits int, data uint8)` - Write up to 8 bits
- `Write16(leftPadd, bits int, data uint16)` - Write up to 16 bits
- `Write32(leftPadd, bits int, data uint32)` - Write up to 32 bits
- `Write64(leftPadd, bits int, data uint64)` - Write up to 64 bits
- `WriteBool(data bool)` - Write a single bit

**Cursor-based writing:**
- `WriteBit(bit bool) error` - Write one bit at cursor and advance (auto-extends data slice)
- `WriteBitAt(pos int, bit bool) error` - Write one bit at position without moving cursor (supports overwriting, returns `ErrNegativePosition` for negative positions)
- `Pos() int` - Get current cursor position (thread-safe)
- `Seek(pos int) error` - Set cursor position (returns `ErrNegativePosition` for negative positions)

**Other:**
- `Data() []T` - Get accumulated data slice
- `AnyData() any` - Get data as 'any' type
- `Bits() int` - Get total number of bits written

## License

Apache 2.0
