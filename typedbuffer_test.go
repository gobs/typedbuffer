package typedbuffer

import (
	"bytes"
	"testing"
)

func TestBool(t *testing.T) {
	t.Log("true:", EncodeBool(true))
	t.Log("false:", EncodeBool(false))
}

func TestInt(t *testing.T) {
	t.Log("Zero:", Zero)
	t.Log("One:", One)
	t.Log("MinusOne:", MinusOne)

	for i := SMALL_NEGATIVE_INT; i <= SMALL_POSITIVE_INT; i++ {
		t.Log(i, ":", EncodeInt64(int64(i)))
	}

	t.Log("10:", EncodeInt64(10))
	t.Log("100:", EncodeInt64(100))
	t.Log("1000:", EncodeInt64(1000))
	t.Log("-1000:", EncodeInt64(-1000))
}

func TestUint(t *testing.T) {
	t.Log("0", EncodeUint64(0))
	t.Log("1", EncodeUint64(1))
	t.Log("10", EncodeUint64(10))
	t.Log("16", EncodeUint64(16))
	t.Log("17", EncodeUint64(17))
	t.Log("20", EncodeUint64(20))
	t.Log("100", EncodeUint64(100))
	t.Log("300", EncodeUint64(300))
	t.Log("1000", EncodeUint64(1000))
	t.Log("1 byte", EncodeUint64(0xFF))
	t.Log("2 byte", EncodeUint64(0x1122))
	t.Log("3 byte", EncodeUint64(0x112233))
	t.Log("4 byte", EncodeUint64(0x11223344))
	t.Log("5 byte", EncodeUint64(0x1122334455))
	t.Log("6 byte", EncodeUint64(0x112233445566))
	t.Log("7 byte", EncodeUint64(0x11223344556677))
	t.Log("8 byte", EncodeUint64(0x1122334455667788))
}

func TestBytes(t *testing.T) {
	t.Log("Empty:", EncodeBytes([]byte{}))
	t.Log("Short:", EncodeBytes([]byte("hello, world!")))
	t.Log("Longer:", EncodeBytes([]byte("0123456789012345678901234567890123456789")))
	t.Log("Longer:", EncodeBytes([]byte("01234567890123456789012345678901234567890123456789012345678901234567890123456789")))
	t.Log("Longer:", EncodeBytes([]byte("0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")))

	var arr [100000]byte

	t.Log("Bytes1:", EncodeBytes(arr[:300]))
	t.Log("Bytes2:", EncodeBytes(arr[:1000]))
	t.Log("Bytes4:", EncodeBytes(arr[:70000]))
}

func TestDecode(t *testing.T) {
	v, b, err := Decode([]byte{})
	t.Log("Empty:", v, b, err)

	v, b, err = Decode(Zero)
	t.Log("Zero:", v, b, err)

	v, b, err = Decode(EncodeInt64(3))
	t.Log("3:", v, b, err)

	v, b, err = Decode(EncodeInt64(-3))
	t.Log("-3:", v, b, err)

	v, b, err = Decode(EncodeInt64(10))
	t.Log("10:", v, b, err)

	v, b, err = Decode(EncodeInt64(1000))
	t.Log("1000:", v, b, err)

	v, b, err = Decode(EncodeInt64(-1000))
	t.Log("-1000:", v, b, err)

	v, b, err = Decode(EncodeUint64(0))
	t.Log("0:", v, b, err)

	v, b, err = Decode(EncodeUint64(100))
	t.Log("100:", v, b, err)

	v, b, err = Decode(EncodeUint64(255))
	t.Log("255:", v, b, err)

	v, b, err = Decode(EncodeUint64(1000))
	t.Log("1000:", v, b, err)

	v, b, err = Decode(EncodeUint64(1000000))
	t.Log("1000000:", v, b, err)

	v, b, err = Decode(EncodeUint64(1000000000))
	t.Log("1000000000:", v, b, err)

	v, b, err = Decode(EncodeUint64(1000000000000))
	t.Log("1000000000000:", v, b, err)

	v, b, err = Decode(EncodeBool(true))
	t.Log("true:", v, b, err)

	v, b, err = Decode(EncodeBytes([]byte("jello, world!")))
	t.Log("bytes:", v, b, err)

	v, b, err = Decode(EncodeBytes([]byte("01234567890123456789012345678901234567890123456789012345678901234567890123456789")))
	t.Log("bytes:", v, b, err)

	v, b, err = Decode(EncodeBytes([]byte("012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")))
	t.Log("bytes:", v, b, err)

	v, b, err = Decode(EncodeBytes([]byte("012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")))
	t.Log("bytes:", v, b, err)
}

func TestEncode(t *testing.T) {
	b, err := Encode(10, "hello", -100000, false)
	t.Log("encoded:", b, err)

	for len(b) > 0 {
		t.Log("decoding:", b)
		v, next, err := Decode(b)
		t.Log(v, next, err)

		b = next
	}
}

func MustEncode(values ...interface{}) []byte {
	b, err := Encode(values...)
	if err != nil {
		panic(err)
	}
	return b
}

type CompareItem struct {
	min, max []byte
}

func MustDecodeAll(b []byte) []interface{} {
	res, err := DecodeAll(b)
	if err != nil {
		panic("unexpected error")
	}

	return res
}

func TestCompare(t *testing.T) {
	tests := []CompareItem{
		CompareItem{EncodeInt64(10), EncodeInt64(10000000000)},
		CompareItem{EncodeInt64(10), EncodeInt64(100)},
		CompareItem{EncodeInt64(-1), EncodeInt64(10)},
		CompareItem{EncodeInt64(-2), EncodeInt64(-1)},
		CompareItem{EncodeInt64(-1000000000), EncodeInt64(-1)},
		CompareItem{MustEncode(-1245678, "dogs", true), MustEncode(100, "cats", false)},
		CompareItem{MustEncode(3, "dog", false), MustEncode(3, "dog", true)},
		CompareItem{MustEncode(3, "cat", false), MustEncode(3, "dog", false)},

		CompareItem{MustEncode(1, 50, 1000000), MustEncode(1, 300, 1)},
		CompareItem{MustEncode(1, 50, 1000000, 1), MustEncode(1, 300, 1)},
	}

	for _, tt := range tests {
		t.Log(MustDecodeAll(tt.min), "<", MustDecodeAll(tt.max), "[", tt.min, "<", tt.max, "]")
		if bytes.Compare(tt.min, tt.max) != -1 {
			t.Log(tt.min, "should be less than", tt.max)
			t.Fail()
		}
	}
}

func TestCompareNil(t *testing.T) {
	tests := []CompareItem{
		CompareItem{NilFirst, EncodeInt64(0)},
		CompareItem{NilFirst, EncodeInt64(-1)},
		CompareItem{NilFirst, EncodeInt64(1)},
		CompareItem{NilFirst, EncodeBool(true)},
		CompareItem{NilFirst, EncodeBool(false)},
		CompareItem{EncodeInt64(0), NilLast},
		CompareItem{EncodeInt64(-1), NilLast},
		CompareItem{EncodeInt64(1), NilLast},
		CompareItem{EncodeBool(true), NilLast},
		CompareItem{EncodeBool(false), NilLast},

		CompareItem{MustEncode(nil, 50, 1000000), MustEncode(0, nil, 1000000)},
		CompareItem{MustEncode(nil, 50, 1000000), MustEncode(0, nil, 1000000)},

		CompareItem{MustEncode(0, 50, nil), MustEncode(0, 50, 1)},
		CompareItem{MustEncode(0, 1, 1), MustEncode(0, 50, nil)},
	}

	for _, tt := range tests {
		t.Log(MustDecodeAll(tt.min), "<", MustDecodeAll(tt.max), "[", tt.min, "<", tt.max, "]")
		if bytes.Compare(tt.min, tt.max) != -1 {
			t.Log(tt.min, "should be less than", tt.max)
			t.Fail()
		}
	}
}
