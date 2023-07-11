package proto

// Reference - https://datatracker.ietf.org/doc/html/rfc1035

// Message Format
//
// All communications inside the domain protocol are carried in a single
// format called a message. The top level format of message is divided
// into 5 sections (some of which are empty in certain cases) shown below:
//
// +---------------------+
// |        Header       |
// +---------------------+
// |       Question      | the question for the name server
// +---------------------+
// |        Answer       | RRs answering the question
// +---------------------+
// |      Authority      | RRs pointing toward an authority
// +---------------------+
// |      Additional     | RRs holding additional information
// +---------------------+

type Message struct {
	Header     Header
	Question   []Question
	Answer     []ResourceRecord
	Authority  []ResourceRecord
	Additional []ResourceRecord
}

// Header Section Format
//
//	                                1  1  1  1  1  1
//	  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                      ID                       |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    QDCOUNT                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    ANCOUNT                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    NSCOUNT                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    ARCOUNT                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
type Header struct {
	// ID is a 16 bit identifier assigned by the program that
	// generates any kind of query. This identifier is copied
	// the corresponding reply and can be used by the requester
	// to match up replies to outstanding queries.
	ID uint16

	// Response (QR) is a one bit field that specifies whether
	// this message is a query (0/false), or a response (1/true).
	Response bool

	// Opcode is a four bit field that specifies kind of
	// query in this message. This value is set by the
	// ordinator of a query and copied into response. The
	// values are:
	//   0    a standard query (QUERY)
	//   1    an inverse query (IQUERY)
	//   2    a server status request (STATUS)
	//   3-15 reserved for future use
	Opcode Opcode

	// AuthoritativeAnswer (AA) is valid in responses, and specifies
	// that the responding name server  is an authority for the domain
	// name in question section.
	//
	// Note that the contents of the answer section may have multiple
	// owner names because of aliases. The AuthoritativeAnswer bit
	// corresponds to the name which matches the query name, or the
	// first owner name in the answer section.
	AuthoritativeAnswer bool

	// TruncateCation (TC) bit specifies that this message was truncated
	// due to length greater than that permitted on the transmission
	// channel.
	TruncateCation bool

	// RecursionDesired bit may be set in a query and is copied
	// into the response. If RD is set, it directs the name server
	// to pursue the query recursively. Recursive query support is
	// optional.
	RecursionDesired bool

	// RecursionAvailable is set or cleared in a response, and denotes whether
	// recursive query support is available in the name server.
	RecursionAvailable bool

	Z byte

	// RCode is a 4 bit field is set as part of responses. The
	// values have the following interpretation:
	//
	//  0 No error condition
	//
	//  1 Format error - The name server was
	//    unable to interpret the query.
	//
	//  2 Server failure - The name server was
	//    unable to process this query due to a
	//    problem with the name server.
	//
	//  3 Name Error - Meaningful only for
	//    responses from an authoritative name
	//    server, this code signifies that the
	//    domain name referenced in the query does
	//    not exist.
	//
	//  4 Not Implemented - The name server does
	//    not support the requested kind of query.
	//
	//  5 Refused - The name server refuses to
	//    perform the specified operation for
	//    policy reasons.  For example, a name
	//    server may not wish to provide the
	//    information to the particular requester,
	//    or a name server may not wish to perform
	//    a particular operation (e.g., zone)
	RCode RCode

	// QDCount is an integer specifying the number of
	// entries in the question section.
	QDCount uint16

	// ANCount is an integer specifying the number of
	// response records in the answer section.
	ANCount uint16

	// NSCount is an integer specifying the number of
	// name server resource records in the authority
	// records.
	NSCount uint16

	// ARCount is an integer specifying the number of
	// resource records in the additional records section.
	ARCount uint16
}

// Question Section Format
//
// The question section is used to carry the "question" in most queries,
// i.e., the parameters that define what is being asked. The section
// contains QDCOUNT (usually 1) entries, each of the following format:
//
//	                                1  1  1  1  1  1
//	  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                                               |
//	/                     QNAME                     /
//	/                                               /
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     QTYPE                     |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     QCLASS                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
type Question struct {
	Name  string
	Type  QType
	Class QClass
}

// ResourceRecord Format
//
//	                                1  1  1  1  1  1
//	  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                                               |
//	/                                               /
//	/                      NAME                     /
//	|                                               |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                      TYPE                     |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     CLASS                     |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                      TTL                      |
//	|                                               |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                   RDLENGTH                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--|
//	/                     RDATA                     /
//	/                                               /
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
type ResourceRecord struct {
	Name  string
	Type  QType
	Class QClass

	// TTL is a field which is the time to live of the RR. This
	// field is a 32 bit integer in units of seconds, an is
	// primarily used by resolvers when they cache RRs. The TTL
	// describes how long a RR can be cached before it should be
	// discarded.
	//
	// The meaning of the TTL field is a time limit on how long an RR can be
	// kept in a cache. This limit does not apply to authoritative data in
	// zones; it is also timed out, but by the refreshing policies for the
	// zone. The TTL is assigned by the administrator for the zone where the
	// data originates. While short TTLs can be used to minimize caching, and
	// a zero TTL prohibits caching, the realities of Internet performance
	// suggest that these times should be on the order of days for the typical
	// host. If a change can be anticipated, the TTL can be reduced prior to
	// the change to minimize inconsistency during the change, and then
	// increased back to its former value following the change.
	TTL uint32

	RDLength uint16
	RData    string
}

//go:generate stringer -type=QClass
type QClass uint16

const (
	ClassUnknown QClass = 0
	ClassIN      QClass = 1
	ClassCS      QClass = 2
	ClassCH      QClass = 3
	ClassHS      QClass = 4
	ClassAny     QClass = 255
)

//go:generate stringer -type=QType -trimprefix=QType
type QType uint16

const (
	QTypeUnknown QType = 0
	QTypeA       QType = 1
	QTypeNS      QType = 2
	QTypeMD      QType = 3
	QTypeMF      QType = 4
	QTypeCName   QType = 5
	QTypeSOA     QType = 6
	QTypeMB      QType = 7
	QTypeMG      QType = 8
	QTypeMR      QType = 9
	QTypeNULL    QType = 10
	QTypeWKS     QType = 11
	QTypePTR     QType = 12
	QTypeHINFO   QType = 13
	QTypeMINFO   QType = 14
	QTypeMX      QType = 15
	QTypeTXT     QType = 16

	// QTypeAAAA (RFC 3596) resource record type is a record specific
	// to the Internet class that stores a single IPv6 address.
	QTypeAAAA QType = 28

	QTypeAXFR  QType = 252
	QTypeMAILB QType = 253
	QTypeMAILA QType = 254
	QTypeALL   QType = 255
)

func ParseQType(s string) QType {
	switch s {
	case "A":
		return QTypeA
	case "NS":
		return QTypeNS
	case "MD":
		return QTypeMD
	case "MF":
		return QTypeMF
	case "CNAME":
		return QTypeCName
	case "SOA":
		return QTypeSOA
	case "MB":
		return QTypeMB
	case "MG":
		return QTypeMG
	case "MR":
		return QTypeMR
	case "NULL":
		return QTypeNULL
	case "WKS":
		return QTypeWKS
	case "HINFO":
		return QTypeHINFO
	case "MINFO":
		return QTypeMINFO
	case "TXT":
		return QTypeTXT
	case "AAAA":
		return QTypeAAAA
	case "AXFR":
		return QTypeAXFR
	case "MAILB":
		return QTypeMAILB
	case "MAILA":
		return QTypeMAILA
	case "ALL":
		return QTypeALL
	default:
		return QTypeUnknown
	}
}

//go:generate stringer -type RCode
type RCode uint8

const (
	RCodeNoErrorCondition = RCode(0)
	RCodeFormatError      = RCode(1)
	RCodeServerFailure    = RCode(2)
	RCodeNameError        = RCode(3)
	RCodeNotImplemented   = RCode(4)

	//	RCodeRefused describes that the name server refuses to
	// perform the specified operation for policy reasons. For
	// example, a name server may not wish to provide the
	// information to the particular requester, or a name
	// server may not wish to perform a particular operation
	// (e.g., zone transfer) for particular data.
	RCodeRefused = RCode(5)
)

//go:generate stringer -type Opcode
type Opcode uint8

const (
	OpcodeQuery Opcode = iota
	OpcodeIQuery
	OpcodeStatus
)

type QR bool

// todo: add custom types for header values

const (
	QRQuery    QR = false
	QRResponse QR = true
)

func (v QR) String() string {
	if v {
		return "RESPONSE"
	}

	return "QUERY"
}
