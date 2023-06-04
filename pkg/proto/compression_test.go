package proto

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/rokkerruslan/dnska/internal/limits"
	"github.com/rokkerruslan/dnska/pkg/bv"
	testing2 "github.com/rokkerruslan/dnska/testing"
)

func TestLabelsIndex_EncodeName(t *testing.T) {
	s0 := "F.ISI.ARPA"         // a sequence of labels ending in a zero octet
	s1 := "FOO.BAR.F.ISI.ARPA" // a sequence of labels ending with a pointer
	s2 := "BAR.F.ISI.ARPA"     // a pointer

	index := labelsIndex{}

	buf := bv.NewByteView(make([]byte, limits.UDPPayloadSizeLimit))

	if err := index.EncodeName(buf, s0); err != nil {
		t.Fatal(err)
	}
	if err := index.EncodeName(buf, s1); err != nil {
		t.Fatal(err)
	}
	if err := index.EncodeName(buf, s2); err != nil {
		t.Fatal(err)
	}

	out := bv.NewByteView(buf.Bytes())

	decodeIndex := labelsIndex{}

	s0got, err := decodeIndex.DecodeName(out)
	if err != nil {
		t.Fatal(err)
	}

	testing2.Assert(t, s0got, "F.ISI.ARPA")

	s1got, err := decodeIndex.DecodeName(out)
	if err != nil {
		t.Fatal(err)
	}

	testing2.Assert(t, s1got, "FOO.BAR.F.ISI.ARPA")

	s2got, err := decodeIndex.DecodeName(out)
	if err != nil {
		t.Fatal(err)
	}

	testing2.Assert(t, s2got, "BAR.F.ISI.ARPA")
}

func TestExample(t *testing.T) {
	b := bv.NewByteView(make([]byte, 512))

	li := labelsIndex{}

	testing2.ThisIsFine(t, li.EncodeName(b, "Z"))
	testing2.ThisIsFine(t, li.EncodeName(b, "C.B.A"))
	testing2.ThisIsFine(t, li.EncodeName(b, "F.E.D.C.B.A"))
	//testing2.ThisIsFine(t, li.EncodeName(b, "D.C.B.A"))
	//testing2.ThisIsFine(t, li.EncodeName(b, "E.D.C.B.A"))

	//oldAlgorithm := 26
	//newAlgorithm := 24

	t.Log(spew.Sdump(b.Bytes()))
	//t.Log(oldAlgorithm, newAlgorithm)
}

func TestExample2(t *testing.T) {
	b := bv.NewByteView(make([]byte, 512))
	li := labelsIndex{}

	list := []string{
		"A.B.C.D.E.F.G.H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"Z.",
		"Y.Z.",
		"X.Y.Z.",
		"W.X.Y.Z.",
		"V.W.X.Y.Z.",
		"U.V.W.X.Y.Z.",
		"T.U.V.W.X.Y.Z.",
		"S.T.U.V.W.X.Y.Z.",
		"R.S.T.U.V.W.X.Y.Z.",
		"Q.R.S.T.U.V.W.X.Y.Z.",
		"P.Q.R.S.T.U.V.W.X.Y.Z.",
		"O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"G.H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"F.G.H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"E.F.G.H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"D.E.F.G.H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"C.D.E.F.G.H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"B.C.D.E.F.G.H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
		"A.B.C.D.E.F.G.H.I.J.K.L.M.N.O.P.Q.R.S.T.U.V.W.X.Y.Z.",
	}

	for _, line := range list {
		t.Run(line, func(t *testing.T) {
			testing2.ThisIsFine(t, li.EncodeName(b, line))
		})
	}

	total := 26     // alphabet
	total += 26     // size of every label
	total += 26 * 2 // links
	total += 1      // end byte

	t.Log(spew.Sdump(b.Bytes()))

	testing2.Assert(t, b.Bytes(), total)
}
