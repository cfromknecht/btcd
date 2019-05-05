package work

import (
	"encoding/binary"
	"math/big"
)

const (
	width  int = 8
	uwidth     = uint(width)
)

type UInt256 [width]uint32

func (u *UInt256) NBits() uint {
	for w := uwidth - 1; w < uwidth; w-- {
		if u[w] > 0 {
			for b := uint(31); b > 0; b-- {
				if u[w]&(1<<b) > 0 {
					return 32*w + b + 1
				}
			}
			return 32*w + 1
		}
	}
	return 0
}

func (u *UInt256) Int() *big.Int {
	words := make([]big.Word, width)
	for i, word := range u[:] {
		words[i] = big.Word(word)
	}
	return new(big.Int).SetBits(words)
}

func (u *UInt256) NBytes() uint {
	bits := u.NBits()
	return (bits + 7) / 8
}

func (u *UInt256) Bytes() []byte {
	b := make([]byte, 4*width)
	for i := width - 1; i >= 0; i-- {
		offset := 28 - 4*i
		binary.BigEndian.PutUint32(b[offset:offset+4], u[i])
	}
	return b
}

func (u *UInt256) Write(b []byte) {
	bb := make([]byte, 4*width)
	for i := 0; i < width; i++ {
		offset := 4 * i
		binary.LittleEndian.PutUint32(bb[offset:offset+4], u[i])
	}
	copy(b, bb)
}

func (u *UInt256) Read(b []byte) *UInt256 {
	bb := make([]byte, 4*width)
	copy(bb, b)
	for i := 0; i < width; i++ {
		offset := 4 * i
		u[i] = binary.LittleEndian.Uint32(bb[offset : offset+4])
	}
	return u
}

func NewUInt256FromUint64(val uint64) *UInt256 {
	u := UInt256FromUint64(val)
	return &u
}

func UInt256FromUint64(val uint64) UInt256 {
	var u UInt256
	u[0] = uint32(val)
	u[1] = uint32(val >> 32)
	return u
}

func (u *UInt256) Lsh(shift uint) *UInt256 {
	a := *u
	*u = UInt256{}

	k := shift / 32
	shift = shift % 32
	for i := uint(0); i < uwidth; i++ {
		if i+k+1 < uwidth && shift != 0 {
			u[i+k+1] |= (a[i] >> (32 - shift))
		}
		if i+k < uwidth {
			u[i+k] |= (a[i] << shift)
		}
	}

	return u
}

func (u *UInt256) Rsh(shift uint) *UInt256 {
	a := *u
	*u = UInt256{}

	k := int(shift / 32)
	shift = shift % 32
	for i := 0; i < width; i++ {
		if i-k-1 >= 0 && shift != 0 {
			u[i-k-1] |= (a[i] << (32 - shift))
		}
		if i-k >= 0 {
			u[i-k] |= (a[i] >> shift)
		}
	}

	return u
}

func (u *UInt256) Cmp(o *UInt256) int {
	for i := width - 1; i >= 0; i-- {
		if u[i] < o[i] {
			return -1
		}
		if u[i] > o[i] {
			return 1
		}
	}

	return 0
}

func (u *UInt256) Neg() *UInt256 {
	for i := 0; i < width; i++ {
		u[i] = ^u[i]
	}
	return u.Inc()
}

func (u *UInt256) Add(o *UInt256) *UInt256 {
	var carry uint64
	for i := 0; i < width; i++ {
		n := carry + uint64(u[i]) + uint64(o[i])
		u[i] = uint32(n & 0xffffffff)
		carry = n >> 32
	}
	return u
}

func (u *UInt256) Int64() int64 {
	ret := int64(u[0]) | (int64(u[1]<<32) & 0x7fffffff)
	if u[width-1]&0x80000000 != 0 {
		ret = -ret
	}
	return ret
}

func (u *UInt256) Mul(o *UInt256) *UInt256 {
	var ret UInt256
	for j := 0; j < width; j++ {
		var carry uint64
		for i := 0; i+j < width; i++ {
			n := carry + uint64(ret[i+j]) + uint64(u[j]) + uint64(o[i])
			ret[i+j] = uint32(n & 0xffffffff)
			carry = n >> 32
		}
	}

	*u = ret
	return u
}

func (u *UInt256) Set(o *UInt256) *UInt256 {
	*u = *o
	return u
}

func (u *UInt256) Sub(o *UInt256) *UInt256 {
	negO := *o
	negO.Neg()
	return u.Add(&negO)
}

func (u *UInt256) Div(o *UInt256) *UInt256 {
	num := *u
	div := *o
	*u = UInt256{}

	numBits := num.NBits()
	divBits := div.NBits()

	if divBits == 0 {
		panic("divide by zero")
	}
	if divBits > numBits {
		return u
	}

	shift := int(numBits - divBits)
	div.Lsh(uint(shift))
	for shift >= 0 {
		if num.Cmp(&div) > 0 {
			num.Sub(&div)
			u[shift/32] |= 1 << uint(shift&31)
		}
		div.Rsh(1)
		shift--
	}

	return u
}

func (u *UInt256) Inc() *UInt256 {
	for i := 0; i < width; i++ {
		if u[i]++; u[i] != 0 {
			return u
		}
	}
	return u
}

func (u *UInt256) Sign() int {
	for i := 0; i < width; i++ {
		if u[i] > 0 {
			if u[width-1]&0x80000000 != 0 {
				return -1
			} else {
				return 1
			}
		}
	}
	return 0
}
