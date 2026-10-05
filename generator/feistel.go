package generator

const (
	Max48Bits      = (uint64(1) << 48) - 1
	Mask24         = (uint64(1) << 24) - 1
	Base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Round keys for 24-bit half-mixing
var roundKeys = [4]uint64{
	0xA35B21,
	0xF4E19C,
	0x5C8A73,
	0x91D0E4,
}

// FeistelEncrypt ensures a strict 1-to-1 bijective mapping: Input != Input2 => Output != Output2
func FeistelEncrypt(input uint64) uint64 {
	input = input & Max48Bits

	left := (input >> 24) & Mask24
	right := input & Mask24

	for i := 0; i < 4; i++ {
		nextLeft := right

		// Non-linear round function F(right, key)
		x := (right ^ roundKeys[i]) & Mask24
		x = (x ^ (x >> 13)) * 0x5deece66d
		fOut := (x ^ (x >> 16)) & Mask24

		nextRight := (left ^ fOut) & Mask24
		left = nextLeft
		right = nextRight
	}

	return (left << 24) | right
}

// ToBase62 converts a 48-bit scrambled integer into a fixed 8-character string
func ToBase62(val uint64) string {
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = Base62Alphabet[val%62]
		val /= 62
	}
	return string(buf)
}
