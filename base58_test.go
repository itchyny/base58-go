package base58

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
)

type testcase struct {
	encoding  *Encoding
	testpairs []testpair
}

type testpair struct {
	decoded string
	encoded string
}

var testcases = []testcase{
	{FlickrEncoding, []testpair{
		{"", ""},
		{"0", "1"},
		{"32", "y"},
		{"64", "27"},
		{"000", "111"},
		{"512", "9Q"},
		{"1024", "iE"},
		{"16777216", "2tZhm"},
		{"00000000000", "11111111111"},
		{"00068719476736", "1112NGvhhq"},
		{"430804206899405823", "ZZZZZZZZZZ"},
		{"430804206899405824", "21111111111"},
		{"9999999999999999999", "pdjvYZfL3PR"},
		{"10000000000000000000", "pdjvYZfL3PS"},
		{"18446744073709551615", "JPwcyDCgEup"},
		{"18446744073709551616", "JPwcyDCgEuq"},
		{"00000000000000000000", "11111111111111111111"},
		{"00000000000000000001", "11111111111111111112"},
		{"79228162514264337593543950336", "5QchsBFApWPVxyp9C"},
		{"00000000018446744073709551616", "111111111JPwcyDCgEuq"},
	}},
	{RippleEncoding, []testpair{
		{"", ""},
		{"0", "r"},
		{"32", "Z"},
		{"64", "pf"},
		{"000", "rrr"},
		{"512", "9q"},
		{"1024", "JC"},
		{"16777216", "p7zHM"},
		{"00000000000", "rrrrrrrrrrr"},
		{"00068719476736", "rrrpo6WHHR"},
		{"430804206899405823", "zzzzzzzzzz"},
		{"430804206899405824", "prrrrrrrrrr"},
		{"9999999999999999999", "QDKWyzEmsFi"},
		{"10000000000000000000", "QDKWyzEmsF1"},
		{"18446744073709551615", "jFXUZedGCVQ"},
		{"18446744073709551616", "jFXUZedGCVR"},
		{"00000000000000000000", "rrrrrrrrrrrrrrrrrrrr"},
		{"00000000000000000001", "rrrrrrrrrrrrrrrrrrrp"},
		{"79228162514264337593543950336", "nqUHTcgbQAFvYZQ9d"},
		{"00000000018446744073709551616", "rrrrrrrrrjFXUZedGCVR"},
	}},
	{BitcoinEncoding, []testpair{
		{"", ""},
		{"0", "1"},
		{"32", "Z"},
		{"64", "27"},
		{"000", "111"},
		{"512", "9q"},
		{"1024", "Jf"},
		{"16777216", "2UzHM"},
		{"00000000000", "11111111111"},
		{"00068719476736", "1112ohWHHR"},
		{"430804206899405823", "zzzzzzzzzz"},
		{"430804206899405824", "21111111111"},
		{"9999999999999999999", "QDKWyzFm3pr"},
		{"10000000000000000000", "QDKWyzFm3ps"},
		{"18446744073709551615", "jpXCZedGfVQ"},
		{"18446744073709551616", "jpXCZedGfVR"},
		{"00000000000000000000", "11111111111111111111"},
		{"00000000000000000001", "11111111111111111112"},
		{"79228162514264337593543950336", "5qCHTcgbQwpvYZQ9d"},
		{"00000000018446744073709551616", "111111111jpXCZedGfVR"},
	}},
}

func TestEncode(t *testing.T) {
	for _, testcase := range testcases {
		for _, pair := range testcase.testpairs {
			got, err := testcase.encoding.Encode([]byte(pair.decoded))
			if err != nil {
				t.Fatalf("Error occurred while encoding %s (%s).", pair.decoded, err)
			}
			if string(got) != pair.encoded {
				t.Errorf("Encode(%s) = %s, want %s", pair.decoded, string(got), pair.encoded)
			}
		}
	}
}

func TestEncodeUint64(t *testing.T) {
	for _, testcase := range testcases {
		for i := range 100 {
			n := rand.Uint64() % uint64(math.Pow10(int(i/5)))
			got := testcase.encoding.EncodeUint64(n)
			expected, err := testcase.encoding.Encode([]byte(strconv.FormatUint(n, 10)))
			if err != nil {
				t.Fatalf("Error occurred while encoding %d (%s).", n, err)
			}
			if string(got) != string(expected) {
				t.Errorf("EncodeUint64(%d) = %s, want %s", n, got, expected)
			}
		}
	}
}

func TestDecode(t *testing.T) {
	for _, testcase := range testcases {
		for _, pair := range testcase.testpairs {
			got, err := testcase.encoding.Decode([]byte(pair.encoded))
			if err != nil {
				t.Fatalf("Error occurred while decoding %s (%s).", pair.encoded, err)
			}
			if string(got) != pair.decoded {
				t.Errorf("Decode(%s) = %s, want %s", pair.encoded, string(got), pair.decoded)
			}
		}
	}
}

func TestDecodeUint64(t *testing.T) {
	for _, testcase := range testcases {
		for i := range 100 {
			n := rand.Uint64() % uint64(math.Pow10(int(i/5)))
			src := testcase.encoding.EncodeUint64(n)
			got, err := testcase.encoding.DecodeUint64(src)
			if err != nil {
				t.Fatalf("Error occurred while decoding %s (%s).", src, err)
			}
			if got != n {
				t.Errorf("DecodeUint64(%s) = %d, want %d", src, got, n)
			}
		}
	}
}

func TestDecodeUint64_Overflow(t *testing.T) {
	for _, testcase := range testcases {
		src := testcase.encoding.EncodeUint64(math.MaxUint64)
		got, err := testcase.encoding.DecodeUint64([]byte(src))
		if err != nil {
			t.Fatalf("Error occurred while decoding %s (%s).", src, err)
		}
		if got != math.MaxUint64 {
			t.Errorf("DecodeUint64(%s) = %d, want %d", src, got, uint64(math.MaxUint64))
		}
		bs := []byte(src)
		bs[len(bs)-1]++
		got, err = testcase.encoding.DecodeUint64(bs)
		if err == nil {
			t.Errorf("Overflow error should occur while decoding %s but got %d.", bs, got)
		}
		src = []byte("aaaaaaaaaaaaaa")
		got, err = testcase.encoding.DecodeUint64(src)
		if err == nil {
			t.Errorf("Overflow error should occur while decoding %s but got %d.", src, got)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	for range b.N {
		for _, testcase := range testcases {
			for _, pair := range testcase.testpairs {
				_, _ = testcase.encoding.Encode([]byte(pair.decoded))
			}
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	for range b.N {
		for _, testcase := range testcases {
			for _, pair := range testcase.testpairs {
				_, _ = testcase.encoding.Decode([]byte(pair.encoded))
			}
		}
	}
}
