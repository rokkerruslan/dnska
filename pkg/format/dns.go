package format

import (
	"fmt"
	"strings"

	"github.com/rokkerruslan/dnska/pkg/proto"
)

// FormatDNSMessage formats a DNS message in a readable zone file-like format.
//
// Example output:
//
//	; Query: google.com A IN
//	; Response: true, RCode: NoError (0 records)
//
//	; Answer Section:
//	google.com.             129     IN  A       142.250.185.46
func FormatDNSMessage(msg *proto.Message, questionName string) string {
	var sb strings.Builder

	// Query line
	if len(msg.Question) > 0 {
		q := msg.Question[0]
		fmt.Fprintf(&sb, "; Query: %s %s %s\n", q.Name, q.Type, q.Class)
	}

	// Response status
	fmt.Fprintf(&sb, "; Response: %v, RCode: %s (%d)\n",
		msg.Header.Response, msg.Header.RCode, msg.Header.RCode)

	// Answer Section
	if len(msg.Answer) > 0 {
		sb.WriteString(";\n; Answer Section:\n")
		for _, rr := range msg.Answer {
			sb.WriteString(formatResourceRecord(rr))
		}
	} else if msg.Header.ANCount > 0 {
		sb.WriteString(";\n; Answer Section: (empty)\n")
	}

	// Authority Section
	if len(msg.Authority) > 0 {
		sb.WriteString(";\n; Authority Section:\n")
		for _, rr := range msg.Authority {
			sb.WriteString(formatResourceRecord(rr))
		}
	}

	// Additional Section
	if len(msg.Additional) > 0 {
		sb.WriteString(";\n; Additional Section:\n")
		for _, rr := range msg.Additional {
			sb.WriteString(formatResourceRecord(rr))
		}
	}

	return sb.String()
}

// formatResourceRecord formats a single DNS resource record.
// Format: name TTL IN type rdata
func formatResourceRecord(rr proto.ResourceRecord) string {
	return fmt.Sprintf("%-30s %5d IN  %-6s %s\n",
		rr.Name, rr.TTL, rr.Type, rr.RData)
}

// FormatDNSAnswer formats only the answer records for a simple view.
//
// Example output:
//
//	google.com.        142.250.185.46
//	google.com.        2607:f8b0:4004:81b::200e
func FormatDNSAnswer(records []proto.ResourceRecord) string {
	if len(records) == 0 {
		return "(no records)\n"
	}

	var sb strings.Builder
	for _, rr := range records {
		fmt.Fprintf(&sb, "%-30s %s\n", rr.Name, rr.RData)
	}
	return sb.String()
}
