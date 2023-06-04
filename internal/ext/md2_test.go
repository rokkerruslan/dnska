package ext

import (
	"fmt"
	"testing"

	testing2 "github.com/rokkerruslan/dnska/testing"
)

func TestEncode(t *testing.T) {
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
			in: "Всё хорошо что хорошо кончается",
			want: "fddf7dfd01e4710e01e81120cd72382e",
		},
	}

	for _, c := range cases {
		got := Encode([]byte(c.in))

		testing2.Assert(t, fmt.Sprintf("%x", got), c.want)
	}
}
