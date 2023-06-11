package ext

func Encode(in []byte) []byte {
	ctx := NewMD2()

	ctx.Update(in)

	out := [16]byte{}

	ctx.Final(out[:])

	return out[:]
}

func EncodeString(in string) string {
	ctx := NewMD2()

	ctx.Update([]byte(in))

	out := [16]byte{}

	ctx.Final(out[:])

	return string(out[:])
}

type MD2 struct {
	state    [16]byte
	checksum [16]byte
	count    uint64   // number of bytes, modulo 16
	buf      [16]byte // non processed data
}

func NewMD2() *MD2 {
	return &MD2{}
}

func (md2 *MD2) Update(in []byte) {
	md2.update(in)
}

func (md2 *MD2) Final(out []byte) {
	index := md2.count
	padLen := 16 - index

	md2.update(padding[padLen])
	md2.update(md2.checksum[:])

	copy(out[:16], md2.state[:])
}

func (md2 *MD2) update(in []byte) {
	index := md2.count
	md2.count = (index + uint64(len(in))) & 15
	remain := 16 - index

	var i uint64
	if uint64(len(in)) >= remain {
		copy(md2.buf[index:], in[:remain])
		transform(md2.state[:], md2.checksum[:], md2.buf[:])

		for i = remain; (i + 15) < uint64(len(in)); i += 16 {
			transform(md2.state[:], md2.checksum[:], in[i:])
		}
		index = 0
	} else {
		i = 0
	}

	// Copy remaining input
	copy(md2.buf[index:], in[i:])
}

func transform(state, checksum, block []byte) {
	x := [48]uint8{}

	for i := 0; i < 16; i++ {
		x[i] = state[i]
		x[i+16] = block[i]
		x[i+16*2] = state[i] ^ block[i]
	}

	t := uint8(0)
	for i := uint8(0); i < 18; i++ {
		for j := range x {
			x[j] ^= pi[t]
			t = x[j]
		}

		t = t + i // wrap around
	}

	for i := range state {
		state[i] = x[i]
	}

	t = checksum[15]
	for i := range checksum {
		checksum[i] = checksum[i] ^ pi[block[i]^t]
		t = checksum[i]
	}
}

var padding = [][]byte{
	nil,
	[]byte("\001"),
	[]byte("\002\002"),
	[]byte("\003\003\003"),
	[]byte("\004\004\004\004"),
	[]byte("\005\005\005\005\005"),
	[]byte("\006\006\006\006\006\006"),
	[]byte("\007\007\007\007\007\007\007"),
	[]byte("\010\010\010\010\010\010\010\010"),
	[]byte("\011\011\011\011\011\011\011\011\011"),
	[]byte("\012\012\012\012\012\012\012\012\012\012"),
	[]byte("\013\013\013\013\013\013\013\013\013\013\013"),
	[]byte("\014\014\014\014\014\014\014\014\014\014\014\014"),
	[]byte("\015\015\015\015\015\015\015\015\015\015\015\015\015"),
	[]byte("\016\016\016\016\016\016\016\016\016\016\016\016\016\016"),
	[]byte("\017\017\017\017\017\017\017\017\017\017\017\017\017\017\017"),
	[]byte("\020\020\020\020\020\020\020\020\020\020\020\020\020\020\020\020"),
}

// pi is Permutation of 0..255 constructed from the digits of pi. It
// gives a "random" nonlinear byte substitution operation.
var pi = [256]byte{
	41, 46, 67, 201, 162, 216, 124, 1, 61, 54, 84, 161, 236, 240, 6,
	19, 98, 167, 5, 243, 192, 199, 115, 140, 152, 147, 43, 217, 188,
	76, 130, 202, 30, 155, 87, 60, 253, 212, 224, 22, 103, 66, 111, 24,
	138, 23, 229, 18, 190, 78, 196, 214, 218, 158, 222, 73, 160, 251,
	245, 142, 187, 47, 238, 122, 169, 104, 121, 145, 21, 178, 7, 63,
	148, 194, 16, 137, 11, 34, 95, 33, 128, 127, 93, 154, 90, 144, 50,
	39, 53, 62, 204, 231, 191, 247, 151, 3, 255, 25, 48, 179, 72, 165,
	181, 209, 215, 94, 146, 42, 172, 86, 170, 198, 79, 184, 56, 210,
	150, 164, 125, 182, 118, 252, 107, 226, 156, 116, 4, 241, 69, 157,
	112, 89, 100, 113, 135, 32, 134, 91, 207, 101, 230, 45, 168, 2, 27,
	96, 37, 173, 174, 176, 185, 246, 28, 70, 97, 105, 52, 64, 126, 15,
	85, 71, 163, 35, 221, 81, 175, 58, 195, 92, 249, 206, 186, 197,
	234, 38, 44, 83, 13, 110, 133, 40, 132, 9, 211, 223, 205, 244, 65,
	129, 77, 82, 106, 220, 55, 200, 108, 193, 171, 250, 36, 225, 123,
	8, 12, 189, 177, 74, 120, 136, 149, 139, 227, 99, 232, 109, 233,
	203, 213, 254, 59, 0, 29, 57, 242, 239, 183, 14, 102, 88, 208, 228,
	166, 119, 114, 248, 235, 117, 75, 10, 49, 68, 80, 180, 143, 237,
	31, 26, 219, 153, 141, 51, 159, 17, 131, 20,
}
