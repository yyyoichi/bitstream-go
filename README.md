# bitstream-go

A Go library for efficient bit-level reading and writing operations on integer slices.

## Features

- **BitReader**: Read bits from integer slices with configurable padding
- **BitWriter**: Write bits to integer slices with automatic allocation
- Generic support for `uint8`, `uint16`, `uint32`, `uint64`, and `uint`
- Thread-safe `BitWriter` operations
- Configurable left and right padding for each element

## Installation

```bash
go get github.com/yyyoichi/bitstream-go
```

## Usage

### BitReader

Read bits from an integer slice:

```go
package main

import bitstream "github.com/yyyoichi/bitstream-go"

func main() {
    data := []uint8{0b10101100, 0b11100011}
    reader := bitstream.NewBitReader(data, 0, 0) // no padding
    
    // Read 12 bits from position 0
    value := reader.Read16R(12, 0) // returns 0b101011001110
    
    // Read 4 bits from position 2
    value2 := reader.Read8R(4, 2)  // returns 0b1110
}
```

### BitWriter

Write bits to create an integer slice:

```go
package main

import bitstream "github.com/yyyoichi/bitstream-go"

func main() {
    writer := bitstream.NewBitWriter[uint8](0, 0) // no padding
    
    // Write 12 bits from a uint16
    writer.Write16(0, 12, 0b101011001110)
    
    // Write a single boolean
    writer.Bool(true)
    
    // Get the result
    data, bits := writer.Data()
    // data = []uint8{0b10101100, 0b11100000}
    // bits = 13
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

- `NewBitReader[T](data []T, leftPadd, rightPadd int) *BitReader[T]`
- `SetBits(bits int)` - Limit readable range
- `Read8R(bits, n int) uint8` - Read up to 8 bits
- `Read16R(bits, n int) uint16` - Read up to 16 bits
- `Read32R(bits, n int) uint32` - Read up to 32 bits
- `Read64R(bits, n int) uint64` - Read up to 64 bits
- `Bit() int` - The total number of valid bits

### BitWriter

- `NewBitWriter[T](leftPadd, rightPadd int) *BitWriter[T]`
- `Write8(leftPadd, bits int, data uint8)` - Write up to 8 bits
- `Write16(leftPadd, bits int, data uint16)` - Write up to 16 bits
- `Write32(leftPadd, bits int, data uint32)` - Write up to 32 bits
- `Write64(leftPadd, bits int, data uint64)` - Write up to 64 bits
- `WriteBool(data bool)` - Write a single bit
- `Data() ([]T, int)` - Get data and bit count
- `AnyData() (any, int)` - Get data as 'any type' and bit count

## License

Apache 2.0
