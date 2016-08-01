package base58

import (
	"fmt"
	"math/big"
)

// An Encoding is a radix 58 encoding/decoding scheme.
type Encoding struct {
	alphabet  []byte
	decodeMap [256]byte
}

func encoding(alphabet []byte) *Encoding {
	enc := &Encoding{
		alphabet: alphabet,
	}
	for i := range enc.decodeMap {
		enc.decodeMap[i] = 0xFF
	}
	for i, b := range alphabet {
		enc.decodeMap[b] = byte(i)
	}
	return enc
}

// FlickrEncoding is the encoding scheme used in Flickr's short URLs.
var FlickrEncoding = encoding([]byte("123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"))

// RippleEncoding is the encoding scheme used Ripple addresses.
var RippleEncoding = encoding([]byte("rpshnaf39wBUDNEGHJKLM4PQRST7VWXYZ2bcdeCg65jkm8oFqi1tuvAxyz"))

// BitcoinEncoding is the encoding scheme used Bitcoin addresses.
var BitcoinEncoding = encoding([]byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"))

func (encoding *Encoding) index(c byte) int64 {
	b := encoding.decodeMap[c]
	if b == 0xFF {
		return -1
	}
	return int64(b)
}

func (encoding *Encoding) at(idx int64) byte {
	return encoding.alphabet[idx]
}

var radix = big.NewInt(58)

func reverse(data []byte) []byte {
	for i, j := 0, len(data)-1; i < j; i++ {
		data[i], data[j] = data[j], data[i]
		j--
	}
	return data
}

// Encode encodes the number represented in the byte array base 10.
func (encoding *Encoding) Encode(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}
	n, ok := new(big.Int).SetString(string(src), 10)
	if !ok {
		return nil, fmt.Errorf("Expecting a number but got \"%s\".", string(src))
	}
	bytes := make([]byte, 0, len(src))
	for _, c := range src {
		if c == '0' {
			bytes = append(bytes, encoding.at(0))
		} else {
			break
		}
	}
	nprefix := len(bytes)
	mod := new(big.Int)
	zero := big.NewInt(0)
	for {
		switch n.Cmp(zero) {
		case 1:
			n.DivMod(n, radix, mod)
			bytes = append(bytes, encoding.at(mod.Int64()))
		case 0:
			reverse(bytes[nprefix:])
			return bytes, nil
		default:
			return nil, fmt.Errorf("Expecting a positive number in base58 encoding but got \"%s\".", n)
		}
	}
}

// Decode decodes the base58 encoded bytes.
func (encoding *Encoding) Decode(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}
	var zeros []byte
	for i, c := range src {
		if c == encoding.at(0) && i < len(src)-1 {
			zeros = append(zeros, '0')
		} else {
			break
		}
	}
	n := new(big.Int)
	var i int64
	for _, c := range src {
		if i = encoding.index(c); i < 0 {
			return nil, fmt.Errorf("Invalid character '%c' in decoding a base58 string \"%s\".", c, src)
		}
		n.Add(n.Mul(n, radix), big.NewInt(i))
	}
	return n.Append(zeros, 10), nil
}
