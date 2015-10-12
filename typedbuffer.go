/**
 * A set of utility methods that encodes the type of an object and the object value
 * in a byte slice. For small values the object value is encoded with the type.
 *
 * Note that the encoded buffer preserves the order of the native object,
 * so that sorting buffers according to their natural order should match
 * sorting objects of the same type according to their natural order
 *
 * Sorting objects of different types is not supported. Also, sorting byte slices of different length
 * doesn't return the expected results (longer slices win) but sorting buffer contains different types
 * in the same order (records) should work as expected.
 *
 * Long are "compressed" and only use enough bytes to represent the value
 * (i.e. 255 will use 1 byte, 65535 will use 2 bytes and so long). To do this
 * and also preserve order the type encodes a representation of the number "size"
 * so that larger number will have a higher type.
 *
 * Supported types:
 *
 * Nil
 *   byte 00 - nil first (nil comes before any other value)
 *   byte FF - nil last (nil comes after any other value)
 *
 * Boolean:
 *   byte 0E - bool false
 *   byte 0F - bool true
 *
 * Bytes:
 *   byte 10+size     : 0 to 15 bytes
 *   byte 20+size     : 16 to 31 bytes
 *   byte 30+size     : 32 to 47 bytes
 *   byte 40+size     : 48 to 60 bytes
 *   byte 4D XX       : 61+XX bytes
 *   byte 4E XXXX     : 316+XXXX bytes
 *   byte 4F XXXXXXXX : 65851+XXXXXXXXXX bytes
 *
 * Date:
 *   byte 50 [8 bytes] - Date from bytes as long
 *
 * Compact Date (see Long) :
 *   byte D4+size [n bytes] - Date as compacted long from 1/1/2015
 *   byte 54+size [n bytes] - Date as compacted long before 1/1/2015
 *
 * Long:
 *   byte E0 - Long 0L
 *   byte E1 - Long 1L
 *   ...
 *   byte E7 - Long 7L
 *   byte E8 [1 byte]  Long 8L to 255L
 *   byte E9 [2 bytes] Long 256L to 65535L
 *   ...
 *   byte EF [8 bytes] Long from bytes
 *   byte 6F - Long -1L
 *   byte 6E - Long -2L
 *   ...
 *   byte 68 - Long -8L
 *   byte 67 [1 byte] - Long -9L to -256
 *   byte 66 [2 bytes] - Long -256 to -65536L
 *   ...
 *   byte 60 [8 bytes] - Long from bytes (negative value)
 *
 * Unsigned Long:
 *   byte 80 - Unsigned long 0
 *   byte 81 - Unsigned long 1
 *   ...
 *   byte 8F - Unsigned long 15
 *   byte 90 - Unsigned long 16
 *   byte 91 [1 byte] 17 to 255
 *   byte 92 [2 bytes] 256 to 65535
 *   ...
 *   byte 98 [8 bytes] Unsigned long from bytes
 *
 * Double:
 *   byte F7 - Double.NaN
 *   byte F6 - Double.POSITIVE_INFINITY
 *   byte F5 + bytes[8] - Double from bytes (positive value)
 *   byte F4 - Double +0.0
 *   byte 76 - Double -0.0
 *   byte 75 + bytes[8] - Double from bytes (negative value)
 *   byte 74 - Double.NEGATIVE_INFINITY
 *
 */
package typedbuffer

import (
	"errors"
)

const (
	/** All positive values have this bit set */
	BB_POSITIVE byte = 0x80
	BB_NEGATIVE byte = 0x00

	BB_TYPE_MASK byte = 0xF0 // this is actually type + cardinality

	/** Nil values */
	BB_NIL_FIRST = 0x00
	BB_NIL_LAST  = 0xFF

	/** Boolean values */
	BB_BOOLEAN       = 0x0E
	BB_BOOLEAN_FALSE = BB_BOOLEAN | 0
	BB_BOOLEAN_TRUE  = BB_BOOLEAN | 1

	/** Bytes values */
	BB_BYTES       = 0x10
	BB_BYTES_LEN_1 = 0x4D
	BB_BYTES_LEN_2 = 0x4E
	BB_BYTES_LEN_4 = 0x4E

	/** Date values */
	BB_DATE = 0x50

	/** Compact date values */
	BB_COMPACT_DATE  = 0x54
	BB_POSITIVE_DATE = BB_COMPACT_DATE | BB_POSITIVE
	BB_NEGATIVE_DATE = BB_COMPACT_DATE | BB_NEGATIVE

	/** Integer values */
	BB_INT                = 0x60
	BB_INT_POSITIVE_VALUE = BB_INT | BB_POSITIVE | 0x08
	BB_INT_NEGATIVE_VALUE = BB_INT | BB_NEGATIVE
	BB_SMALL_POSITIVE     = BB_INT | BB_POSITIVE
	BB_SMALL_NEGATIVE     = BB_INT | BB_NEGATIVE | 0x08

	BB_INT_MASK = BB_INT_POSITIVE_VALUE

	SMALL_NEGATIVE_INT = -8
	SMALL_POSITIVE_INT = +7
	SMALL_INT_MASK     = 0x07
	SMALL_NEG_MASK     = ^SMALL_INT_MASK

	MIN_SMALL_POSITIVE = BB_SMALL_POSITIVE | (0 & SMALL_INT_MASK)
	MAX_SMALL_POSITIVE = BB_SMALL_POSITIVE | (SMALL_POSITIVE_INT & SMALL_INT_MASK)

	MIN_SMALL_NEGATIVE = BB_SMALL_NEGATIVE | (SMALL_NEGATIVE_INT & SMALL_INT_MASK)
	MAX_SMALL_NEGATIVE = BB_SMALL_NEGATIVE | (-1 & SMALL_INT_MASK)

	/** Double values */
	BB_DOUBLE                   = 0x70
	BB_DOUBLE_NAN               = (BB_DOUBLE | BB_POSITIVE) + 0x03
	BB_DOUBLE_POSITIVE_INFINITY = (BB_DOUBLE | BB_POSITIVE) + 0x02
	BB_DOUBLE_POSITIVE_ZERO     = (BB_DOUBLE | BB_POSITIVE) + 0x00
	BB_DOUBLE_POSITIVE_VALUE    = (BB_DOUBLE | BB_POSITIVE) + 0x01
	BB_DOUBLE_NEGATIVE_ZERO     = (BB_DOUBLE | BB_NEGATIVE) + 0x02
	BB_DOUBLE_NEGATIVE_VALUE    = (BB_DOUBLE | BB_NEGATIVE) + 0x01
	BB_DOUBLE_NEGATIVE_INFINITY = (BB_DOUBLE | BB_NEGATIVE) + 0x00

	/** Unsigned values */
	BB_UINT     = 0x80
	BB_UINT_VAR = 0x90

	BB_UINT_MASK = 0xE0

	SMALL_UINT      = 16
	SMALL_UINT_MASK = 0x1F

	MIN_SMALL_UINT = BB_UINT
	MAX_SMALL_UINT = BB_UINT + SMALL_UINT
)

var (
	NilFirst = []byte{BB_NIL_FIRST}
	NilLast  = []byte{BB_NIL_LAST}

	True  = []byte{BB_BOOLEAN_TRUE}
	False = []byte{BB_BOOLEAN_FALSE}

	Zero          = []byte{BB_SMALL_POSITIVE | 0}
	One           = []byte{BB_SMALL_POSITIVE | 1}
	MinusOne      = []byte{BB_SMALL_NEGATIVE | (-1 & SMALL_INT_MASK)}
	SmallNegative = []byte{BB_SMALL_NEGATIVE | (SMALL_NEGATIVE_INT & SMALL_INT_MASK)}

	Bytes1 = []byte{BB_BYTES_LEN_1}
	Bytes2 = []byte{BB_BYTES_LEN_2}
	Bytes4 = []byte{BB_BYTES_LEN_4}

	NoEncoding           = errors.New("no encoding")
	EmptyBufferError     = errors.New("empty buffer")
	CorruptedBufferError = errors.New("corrupted buffer")
)

func EncodeBool(b bool) []byte {
	if b {
		return True
	} else {
		return False
	}
}

func EncodeNil(first bool) []byte {
	if first {
		return NilFirst
	} else {
		return NilLast
	}
}

func EncodeInt(i int) []byte {
	return EncodeInt64(int64(i))
}

func EncodeUint(u uint) []byte {
	return EncodeUint64(uint64(u))
}

func EncodeInt64(i int64) []byte {
	switch {
	case i < SMALL_NEGATIVE_INT:
		return compactInt64(uint64(i), BB_INT_NEGATIVE_VALUE)

	case i > SMALL_POSITIVE_INT:
		return compactInt64(uint64(i), BB_INT_POSITIVE_VALUE)

	case i >= 0:
		// "small" positive value (0..+7)
		return []byte{BB_SMALL_POSITIVE + byte(i&SMALL_INT_MASK)}

	default:
		// "small" negative value (-8..-1)
		return []byte{BB_SMALL_NEGATIVE + byte(i&SMALL_INT_MASK)}
	}
}

func EncodeUint64(u uint64) []byte {
	if u <= SMALL_UINT {
		return []byte{BB_UINT + byte(u)}
	} else {
		return compactUint64(u)
	}
}

func compactInt64(v uint64, typ byte) []byte {
	bb := make([]byte, 0, 8)
	bits := 64 /* size of int64 */ - 8

	if (typ & BB_TYPE_MASK) == BB_INT_NEGATIVE_VALUE { // negative value
		for ; bits > 0; bits -= 8 {
			if ((v >> uint(bits)) & 0xff) != 0xff {
				break
			}
		}

		bb = append(bb, byte(typ+(7-(byte(bits/8)))))
	} else { // positive value
		for ; bits > 0; bits -= 8 {
			if (v >> uint(bits)) != 0 {
				break
			}
		}

		bb = append(bb, byte(typ+byte(bits/8)))
	}

	for ; bits >= 0; bits -= 8 {
		bb = append(bb, byte(v>>uint(bits)))
	}

	return bb
}

func uncompactInt64(bb []byte, positive bool) int64 {
	var l int64

	if !positive {
		l = -1
	}

	for _, b := range bb {
		l = (l << 8) | int64(b)
	}

	return l
}

func compactUint64(u uint64) []byte {
	bb := make([]byte, 0, 8)
	bits := 64 /* size of uint64 */ - 8

	for ; bits > 0; bits -= 8 {
		if (u >> uint(bits)) != 0 {
			break
		}
	}

	bb = append(bb, byte(BB_UINT_VAR+byte(bits/8)+1))

	for ; bits >= 0; bits -= 8 {
		bb = append(bb, byte(u>>uint(bits)))
	}

	return bb
}

func uncompactUint64(bb []byte) uint64 {
	var l uint64

	for _, b := range bb {
		l = (l << 8) | uint64(b)
	}

	return l
}

func EncodeBytes(bb []byte) []byte {
	l := len(bb)

	switch {
	case l <= 60:
		b := []byte{BB_BYTES | byte(l)}
		return append(b, bb...)

	case l <= (61 + 0xff):
		l -= 61
		b := append(Bytes1, byte(l))
		return append(b, bb...)

	case l <= (317 + 0xffff):
		l -= 317

		b := append(Bytes2, byte(l>>8), byte(l>>0))
		return append(b, bb...)

	case l <= (65851 + 0xffffffff):
		l -= 65851

		b := append(Bytes4, byte(l>>24), byte(l>>16), byte(l>>8), byte(l>>0))
		return append(b, bb...)

	default:
		panic("slice too long")
	}
}

func Encode(values ...interface{}) ([]byte, error) {
    return EncodeNils(true, values...)
}

func EncodeNils(nilFirst bool, values ...interface{}) ([]byte, error) {
	b := []byte{}

	for _, v := range values {
		if v == nil {
			b = append(b, EncodeNil(nilFirst)...)
			continue
		}

		switch t := v.(type) {
		case bool:
			b = append(b, EncodeBool(t)...)

		case int:
			b = append(b, EncodeInt(t)...)

		case int64:
			b = append(b, EncodeInt64(t)...)

		case uint64:
			b = append(b, EncodeUint64(t)...)

		case []uint64:
			for _, u := range t {
				b = append(b, EncodeUint64(u)...)
			}

		case []byte:
			b = append(b, EncodeBytes(t)...)

		case string:
			b = append(b, EncodeBytes([]byte(t))...)

		default:
			return nil, NoEncoding
		}
	}

	return b, nil
}

func Decode(b []byte) (interface{}, []byte, error) {
	if len(b) == 0 {
		return nil, nil, EmptyBufferError
	}

	k := b[0]
	next := b[1:]

	switch {
	case k == BB_NIL_FIRST || k == BB_NIL_LAST:
		return nil, next, nil

	case k == BB_BOOLEAN_FALSE:
		return false, next, nil

	case k == BB_BOOLEAN_TRUE:
		return true, next, nil

	case k >= BB_BYTES && k < BB_BYTES_LEN_1:
		k -= BB_BYTES
		if len(next) < int(k) {
			return nil, nil, CorruptedBufferError
		}
		return next[0:k], next[k:], nil

	case k == BB_BYTES_LEN_1:
		if len(next) < 1 {
			return nil, nil, CorruptedBufferError
		}
		k, next = next[0], next[1:]
		n := int(k) + 62
		if len(next) < int(n) {
			return nil, nil, CorruptedBufferError
		}
		return next[0:n], next[n:], nil

	case k == BB_BYTES_LEN_2:
		if len(next) < 2 {
			return nil, nil, CorruptedBufferError
		}

		k1, k2, next := next[0], next[1], next[2:]
		n := int(k1)*256 + int(k2) + 318
		if len(next) < int(n) {
			return nil, nil, CorruptedBufferError
		}
		return next[0:n], next[n:], nil

	case k >= MIN_SMALL_POSITIVE && k <= MAX_SMALL_POSITIVE:
		return int64(k & SMALL_INT_MASK), next, nil

	case k >= MIN_SMALL_NEGATIVE && k <= MAX_SMALL_NEGATIVE:
		return int64(k&SMALL_INT_MASK) | SMALL_NEG_MASK, next, nil

	case (k & BB_INT_MASK) == BB_INT_POSITIVE_VALUE:
		n := int(k&7) + 1
		if len(next) < n {
			return nil, nil, CorruptedBufferError
		}
		return uncompactInt64(next[0:n], true), next[n:], nil

	case (k & BB_INT_MASK) == BB_INT_NEGATIVE_VALUE:
		n := 8 - int(k&7)
		if len(next) < n {
			return nil, nil, CorruptedBufferError
		}
		return uncompactInt64(next[0:n], false), next[n:], nil

	case k >= MIN_SMALL_UINT && k <= MAX_SMALL_UINT:
		return uint64(k & SMALL_UINT_MASK), next, nil

	case (k & BB_UINT_MASK) == BB_UINT:
		n := int(k & 15)
		if n == 0 || n > 8 || len(next) < n {
			return nil, nil, CorruptedBufferError
		}
		return uncompactUint64(next[0:n]), next[n:], nil

	default:
		return nil, nil, CorruptedBufferError
	}
}

func DecodeAll(strings bool, b []byte) ([]interface{}, error) {
	res := make([]interface{}, 0)

	for {
		v, next, err := Decode(b)
		if err == EmptyBufferError {
			return res, nil
		}

		if err != nil {
			return nil, err
		}

                if strings {
                    if sb, ok := v.([]byte); ok {
		        v = string(sb)
                    }
                } 

                res = append(res, v)
		b = next
	}
}

func DecodeUintArray(b []byte) ([]uint64, error) {
	res := []uint64{}

	for {
		v, next, err := Decode(b)
		if err == EmptyBufferError {
			return res, nil
		}

		if err != nil {
			return nil, err
		}

		res = append(res, v.(uint64))
		b = next
	}
}
