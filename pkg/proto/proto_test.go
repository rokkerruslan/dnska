package proto

import (
	"net"
	"os"
	"testing"

	testing2 "github.com/rokkerruslan/dnska/testing"
)

func TestDecodeEncodeMessage(t *testing.T) {
	t.Run("standard-query.query.A.google.com", func(t *testing.T) {
		buf, err := os.ReadFile("testdata/standard-query.query.A.google.com")
		testing2.FailIfError(t, err)

		dec := NewDecoder()
		got, err := dec.Decode(buf)

		testing2.FailIfError(t, err)

		want := Message{
			Header: Header{
				ID:                  5140,
				Response:            false,
				Opcode:              0,
				AuthoritativeAnswer: false,
				TruncateCation:      false,
				RecursionDesired:    true,
				RecursionAvailable:  false,
				Z:                   0,
				RCode:               0,
				QDCount:             1,
				ANCount:             0,
				NSCount:             0,
				ARCount:             0,
			},
			Question: []Question{
				{
					Name:  "google.com",
					Type:  QTypeA,
					Class: ClassIN,
				},
			},
			Answer:     []ResourceRecord{},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		}

		testing2.Assert(t, got, want)
		//testing2.Assert(t, uint(len(buf)), nb.Pos())

		enc := NewEncoder(make([]byte, 512))
		outBuf, err := enc.Encode(got)

		testing2.ThisIsFine(t, err)

		testing2.Assert(t, outBuf, buf)
	})

	t.Run("standard-query.response.A.google.com", func(t *testing.T) {
		buf, err := os.ReadFile("testdata/standard-query.response.A.google.com")
		testing2.FailIfError(t, err)

		dec := NewDecoder()
		got, err := dec.Decode(buf)
		testing2.FailIfError(t, err)

		want := Message{
			Header: Header{
				ID:                  5140,
				Response:            true,
				Opcode:              0,
				AuthoritativeAnswer: false,
				TruncateCation:      false,
				RecursionDesired:    true,
				RecursionAvailable:  true,
				Z:                   0,
				RCode:               0,
				QDCount:             1,
				ANCount:             6,
				NSCount:             0,
				ARCount:             0,
			},
			Question: []Question{
				{
					Name:  "google.com",
					Type:  QTypeA,
					Class: ClassIN,
				},
			},
			Answer: []ResourceRecord{
				{
					Name:     "google.com",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      63,
					RDLength: 4,
					RData:    "142.250.150.100",
				},
				{
					Name:     "google.com",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      63,
					RDLength: 4,
					RData:    "142.250.150.138",
				},
				{
					Name:     "google.com",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      63,
					RDLength: 4,
					RData:    "142.250.150.139",
				},
				{
					Name:     "google.com",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      63,
					RDLength: 4,
					RData:    "142.250.150.102",
				},
				{
					Name:     "google.com",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      63,
					RDLength: 4,
					RData:    "142.250.150.101",
				},
				{
					Name:     "google.com",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      63,
					RDLength: 4,
					RData:    "142.250.150.113",
				},
			},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		}

		testing2.Assert(t, got, want)

		enc := NewEncoder(make([]byte, 512))
		outBuf, err := enc.Encode(got)

		testing2.ThisIsFine(t, err)

		testing2.Assert(t, outBuf, buf)
	})

	t.Run("standard-query.query.A.yahoo.com.opt.cookie", func(t *testing.T) {
		t.Skipf("fail without EDNS implementation")

		buf, err := os.ReadFile("testdata/standard-query.query.A.yahoo.com.opt.cookie")
		testing2.FailIfError(t, err)

		dec := NewDecoder()
		got, err := dec.Decode(buf)
		testing2.FailIfError(t, err)

		want := Message{
			Header: Header{
				ID:                  53786,
				Response:            false,
				Opcode:              OpcodeQuery,
				AuthoritativeAnswer: false,
				TruncateCation:      false,
				RecursionDesired:    true,
				RecursionAvailable:  false,
				RCode:               RCodeNoErrorCondition,
				QDCount:             1,
				ARCount:             1,
			},
			Question: []Question{
				{
					Name:  "yahoo.com",
					Type:  QTypeA,
					Class: ClassIN,
				},
			},
			Answer:    []ResourceRecord{},
			Authority: []ResourceRecord{},
			Additional: []ResourceRecord{
				{
					Type: 41, // todo: EDNS
				},
			},
		}

		testing2.Assert(t, got, want)
	})

	t.Run("standard-query.response.A.yahoo.com.opt.cookie", func(t *testing.T) {
		t.Skipf("fail without ENDS implementation")

		buf, err := os.ReadFile("testdata/standard-query.response.A.yahoo.com.opt.cookie")
		testing2.FailIfError(t, err)

		dec := NewDecoder()
		got, err := dec.Decode(buf)
		testing2.FailIfError(t, err)

		want := Message{
			Header: Header{
				ID:                  53786,
				Response:            true,
				Opcode:              OpcodeQuery,
				AuthoritativeAnswer: false,
				TruncateCation:      false,
				RecursionDesired:    true,
				RecursionAvailable:  true,
				Z:                   0,
				RCode:               RCodeNoErrorCondition,
				QDCount:             1,
				ANCount:             6,
				NSCount:             0,
				ARCount:             1,
			},
			Question: []Question{
				{
					Name:  "yahoo.com",
					Type:  QTypeA,
					Class: ClassIN,
				},
			},
			Answer: []ResourceRecord{
				{Name: "yahoo.com", Type: QTypeA, Class: ClassIN, TTL: 1214, RDLength: 4, RData: "74.6.143.25"},
				{Name: "yahoo.com", Type: QTypeA, Class: ClassIN, TTL: 1214, RDLength: 4, RData: "98.137.11.164"},
				{Name: "yahoo.com", Type: QTypeA, Class: ClassIN, TTL: 1214, RDLength: 4, RData: "98.137.11.163"},
				{Name: "yahoo.com", Type: QTypeA, Class: ClassIN, TTL: 1214, RDLength: 4, RData: "74.6.231.21"},
				{Name: "yahoo.com", Type: QTypeA, Class: ClassIN, TTL: 1214, RDLength: 4, RData: "74.6.231.20"},
				{Name: "yahoo.com", Type: QTypeA, Class: ClassIN, TTL: 1214, RDLength: 4, RData: "74.6.143.26"},
			},
			Authority: []ResourceRecord{},
			Additional: []ResourceRecord{
				{
					Type: 41, // todo: EDNS
				},
			},
		}

		testing2.Assert(t, got, want)
	})

	t.Run("standard-query.response.soa.com", func(t *testing.T) {
		buf, err := os.ReadFile("testdata/standard-query.response.soa.com")
		testing2.FailIfError(t, err)

		dec := NewDecoder()
		got, err := dec.Decode(buf)
		testing2.FailIfError(t, err)

		want := Message{
			Header: Header{
				ID:                  1,
				Response:            true,
				Opcode:              OpcodeQuery,
				AuthoritativeAnswer: false,
				TruncateCation:      false,
				RecursionDesired:    false,
				RecursionAvailable:  false,
				Z:                   0,
				RCode:               RCodeNoErrorCondition,
				QDCount:             1,
				ANCount:             0,
				NSCount:             13,
				ARCount:             15,
			},
			Question: []Question{
				{
					Name:  "com",
					Type:  QTypeSOA,
					Class: ClassIN,
				},
			},
			Answer: []ResourceRecord{},
			Authority: []ResourceRecord{
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 20,
					RData:    "a.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "b.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "c.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "d.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "e.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "f.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "g.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "h.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "i.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "j.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "k.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "l.gtld-servers.net",
				},
				{
					Name:     "com",
					Type:     QTypeNS,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "m.gtld-servers.net",
				},
			},
			Additional: []ResourceRecord{
				{
					Name:     "a.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.5.6.30",
				},
				{
					Name:     "b.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.33.14.30",
				},
				{
					Name:     "c.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.26.92.30",
				},
				{
					Name:     "d.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.31.80.30",
				},
				{
					Name:     "e.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.12.94.30",
				},
				{
					Name:     "f.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.35.51.30",
				},
				{
					Name:     "g.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.42.93.30",
				},
				{
					Name:     "h.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.54.112.30",
				},
				{
					Name:     "i.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.43.172.30",
				},
				{
					Name:     "j.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.48.79.30",
				},
				{
					Name:     "k.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.52.178.30",
				},
				{
					Name:     "l.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.41.162.30",
				},
				{
					Name:     "m.gtld-servers.net",
					Type:     QTypeA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 4,
					RData:    "192.55.83.30",
				},
				{
					Name:     "a.gtld-servers.net",
					Type:     QTypeAAAA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 16,
					RData:    "2001:0503:a83e:0000:0000:0000:0002:0030",
				},
				{
					Name:     "b.gtld-servers.net",
					Type:     QTypeAAAA,
					Class:    ClassIN,
					TTL:      172800,
					RDLength: 16,
					RData:    "2001:0503:231d:0000:0000:0000:0002:0030",
				},
			},
		}

		testing2.Assert(t, got, want)

		enc := NewEncoder(make([]byte, 512))
		outBuf, err := enc.Encode(got)

		testing2.ThisIsFine(t, err)

		dec2 := NewDecoder()
		decoded, err := dec2.Decode(outBuf)

		testing2.ThisIsFine(t, err)
		testing2.Assert(t, decoded, want)
	})
}

func BenchmarkUDPResolveAddr_Name(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = net.ResolveUDPAddr("udp", "google.com:53")
	}
}

func BenchmarkUDPResolveAddr_IP(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = net.ResolveUDPAddr("udp", "1.1.1.1:53")
	}
}
