package ext

// We begin by supposing that we have a b-byte message as input, and
// that we wish to find its message digest. Here b is an arbitrary
// non-negative integer; b may be zero, and it may be arbitrarily large.
// We imagine the bytes of the message written down as follows:
// m_0 m_1 ... m_{b-1}
// The following five steps are performed to compute the message digest
// of the message.

func Encode(in []byte) []byte {

	// Step 1. Append Padding Bytes
	//
	// The message is "padded" (extended) so that its length (in bytes) is
	// congruent to 0, modulo 16. That is, the message is extended so that
	// it is a multiple of 16 bytes long. Padding is always performed, even
	// if the length of the message is already congruent to 0, modulo 16.

	rem := len(in) % 16
	in = append(in, padding[16-rem]...)

	// Step 2. Append Checksum
	//
	//A 16-byte checksum of the message is appended to the result of the
	//previous step.
	//
	//	This step uses a 256-byte "random" permutation constructed from the
	// digits of pi. Let S[i] denote the i-th element of this table. The
	// table is given in the appendix.

	C := [16]uint8{}

	l := uint8(0)

	for i := 0; i < (len(in) / 16); i++ {
		for j := 0; j < 16; j++ {
			c := in[i*16+j]
			C[j] = C[j] ^ pi[c^l]
			l = C[j]
		}
	}

	in = append(in, C[:]...)

	// Step 3. Initialize MD Buffer

	buf := [48]uint8{}

	// Step 4. Process Message in 16-Byte Blocks

	// Process each 16-word block.
	for i := 0; i < (len(in) / 16); i++ {

		// Copy block i into X.
		for j := 0; j < 16; j++ {
			buf[16+j] = in[i*16+j]
			buf[32+j] = buf[16+j] ^ buf[j]
		}

		t := uint8(0)

		// Do 18 rounds.
		for j := uint8(0); j < 18; j++ {

			// Round j.
			for k := 0; k < 48; k++ {
				buf[k] ^= pi[t]
				t = buf[k]
			}

			t = t+j // % 256
		}
	}

	// Step 5. Output

	return buf[0:16]
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
