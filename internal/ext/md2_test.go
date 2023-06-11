package ext

import (
	"fmt"
	"os"
	"testing"

	testing2 "github.com/rokkerruslan/dnska/testing"
)

// md2adapter is a helper that incomsulate some calls of md2 api.
func md2adapter(in []byte) []byte {
	ctx := NewMD2()
	ctx.Update(in)

	out := [16]byte{}
	ctx.Final(out[:])

	return out[:]
}

func encodeStringAdapter(in []byte) []byte {
	return []byte(EncodeString(string(in)))
}

func TestEncode(t *testing.T) {
	impls := map[string]func([]byte) []byte{
		"rud":    Encode,
		"stream": md2adapter,
		"string": encodeStringAdapter,
	}

	cases := []struct {
		in   string
		want string
	}{
		{
			in:   "",
			want: "8350e5a3e24c153df2275c9f80692773",
		},
		{
			in:   "A",
			want: "08e2a3810d8426443ecacaf47aeedd17",
		},
		{
			in:   "Ehal Greka",
			want: "27049028cba7bb9a8d850c6ed7a96a4b",
		},
		{
			in:   "Ehal Greka Cher",
			want: "c8f36ddc84aec549cf3a32650f33dbde",
		},
		{
			in:   "Ehal Greka Chere",
			want: "2b66d579cb41f02c13f008f624b87d2f",
		},
		{
			in:   "Ehal Greka Cherez Reku Vidit Greka v Reke Rak, Sunul Greka Ruku v Reku, Rak za Ruku Greku Zap",
			want: "e2cfccd7913c4cf8038af35294128cb0",
		},
		{
			in:   "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789",
			want: "da33def2a42df13975352846c30338cd",
		},
		{
			in:   "Всё хорошо что хорошо кончается",
			want: "fddf7dfd01e4710e01e81120cd72382e",
		},
	}

	for name, impl := range impls {
		t.Run(name, func(t *testing.T) {
			for _, c := range cases {
				got := impl([]byte(c.in))

				testing2.Assert(t, fmt.Sprintf("%x", got), c.want)
			}
		})
	}
}

func TestEncodeWap(t *testing.T) {
	buf, err := os.ReadFile("./wap.txt")
	if err != nil {
		t.Fatal(err)
	}

	testing2.Assert(t, fmt.Sprintf("%x", Encode(buf)), "d4351dd2fcf66e8d3e7729d0b243a4cd")
	testing2.Assert(t, fmt.Sprintf("%x", md2adapter(buf)), "d4351dd2fcf66e8d3e7729d0b243a4cd")
}

// Iteration 1.
//
// go test -bench=. -benchmem -cpuprofile cpu.out -memprofile mem.out -benchtime=10s ./internal/ext
// goos: darwin
// goarch: arm64
// pkg: github.com/rokkerruslan/dnska/internal/ext
// BenchmarkEncode-8   	      18	 639737715 ns/op	 6668548 B/op	       2 allocs/op
// PASS
// ok  	github.com/rokkerruslan/dnska/internal/ext	13.010s

func BenchmarkEncode(b *testing.B) {
	buf, err := os.ReadFile("./wap.txt")
	if err != nil {
		b.Fatal(err)
	}

	impls := map[string]func([]byte) []byte{
		"rud":    Encode,
		"stream": md2adapter,
	}

	for name, impl := range impls {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				impl(buf)
			}
		})
	}
}
