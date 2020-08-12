package base58

import (
	"fmt"
	"math/big"
)

// An Encoding is a radix 58 encoding/decoding scheme.
type Encoding struct {
	alphabet  [58]byte
	decodeMap [256]int64
}

// New creates a new base58 encoding.
func New(alphabet []byte) *Encoding {
	enc := &Encoding{}
	copy(enc.alphabet[:], alphabet[:])
	for i := range enc.decodeMap {
		enc.decodeMap[i] = -1
	}
	for i, b := range enc.alphabet {
		enc.decodeMap[b] = int64(i)
	}
	return enc
}

// FlickrEncoding is the encoding scheme used for Flickr's short URLs.
var FlickrEncoding = New([]byte("123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"))

// RippleEncoding is the encoding scheme used for Ripple addresses.
var RippleEncoding = New([]byte("rpshnaf39wBUDNEGHJKLM4PQRST7VWXYZ2bcdeCg65jkm8oFqi1tuvAxyz"))

// BitcoinEncoding is the encoding scheme used for Bitcoin addresses.
var BitcoinEncoding = New([]byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"))

var (
	zero  = new(big.Int)
	radix = big.NewInt(58)
)

const radixInt = uint64(58)

func reverse(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

type encodeError []byte

func (err encodeError) Error() string {
	return fmt.Sprintf("expecting a non-negative number but got %q", []byte(err))
}

// Encode encodes the number represented in the byte array base 10.
func (enc *Encoding) Encode(src []byte) ([]byte, error) {
	if len(src) < 20 { // 10^19 < math.MaxUint64 < 10^20
		return enc.encodeSmall(src)
	}
	n, ok := new(big.Int).SetString(string(src), 10)
	if !ok {
		return nil, encodeError(src)
	}
	bytes := make([]byte, 0, len(src))
	for _, c := range src {
		if c == '0' {
			bytes = append(bytes, enc.alphabet[0])
		} else {
			break
		}
	}
	zerocnt := len(bytes)
	mod := new(big.Int)
	for {
		switch n.Cmp(zero) {
		case 1:
			n.DivMod(n, radix, mod)
			bytes = append(bytes, enc.alphabet[mod.Int64()])
		case 0:
			reverse(bytes[zerocnt:])
			return bytes, nil
		default:
			return nil, encodeError(src)
		}
	}
}

func (enc *Encoding) encodeSmall(src []byte) ([]byte, error) {
	bytes := make([]byte, 0, len(src))
	for _, c := range src {
		if c == '0' {
			bytes = append(bytes, enc.alphabet[0])
		} else {
			break
		}
	}
	if len(bytes) == len(src) {
		return bytes, nil
	}
	n, err := parseUint64(src[len(bytes):])
	if err != nil {
		return nil, err
	}
	return enc.appendEncodeUint64(bytes, n), nil
}

func (enc *Encoding) appendEncodeUint64(buf []byte, n uint64) []byte {
	zerocnt := len(buf)
	var mod uint64
	for n > 0 {
		n, mod = n/radixInt, n%radixInt
		buf = append(buf, enc.alphabet[mod])
	}
	reverse(buf[zerocnt:])
	return buf
}

// Decode decodes the base58 encoded bytes.
func (enc *Encoding) Decode(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}
	var zeros []byte
	for i, c := range src {
		if c == enc.alphabet[0] && i < len(src)-1 {
			zeros = append(zeros, '0')
		} else {
			break
		}
	}
	n := new(big.Int)
	var i int64
	for _, c := range src {
		if i = enc.decodeMap[c]; i < 0 {
			return nil, fmt.Errorf("invalid character '%c' in decoding a base58 string \"%s\"", c, src)
		}
		n.Add(n.Mul(n, radix), big.NewInt(i))
	}
	return n.Append(zeros, 10), nil
}

// UnmarshalFlag implements flags.Unmarshaler
func (enc *Encoding) UnmarshalFlag(value string) error {
	switch value {
	case "flickr":
		*enc = *FlickrEncoding
	case "ripple":
		*enc = *RippleEncoding
	case "bitcoin":
		*enc = *BitcoinEncoding
	default:
		return fmt.Errorf("unknown encoding: %s", value)
	}
	return nil
}

func parseUint64(src []byte) (uint64, error) {
	var n uint64
	for _, c := range src {
		if '0' <= c && c <= '9' {
			n = n*10 + uint64(c&0xF)
		} else {
			return 0, encodeError(src)
		}
	}
	return n, nil
}
