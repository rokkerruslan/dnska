## DNSka - toy implementation of Domain Name System (DNS).

### Quickstart

Ensure that you have installed Go https://go.dev/doc/manage-install

```text
$ ./scripts/app.sh install
$ dnska version
```

Try to use stub resolver:

```text
$ dnska lookup --only-answer example.com
([]dnska.ResourceRecord) (len=1 cap=1) {
 (dnska.ResourceRecord) {
  Name: (string) (len=11) "example.com",
  Type: (dnska.QType) QTypeA,
  Class: (dnska.QClass) ClassIN,
  TTL: (uint32) 71124,
  RDLength: (uint16) 4,
  RData: (string) (len=13) "93.184.216.34"
 }
}
```

Try to run proxy name server:

```text
$ sudo dnska app

# And into another terminal:
$ dnska lookup --addr :2053 --type 28 --only-answer example.com
([]dnska.ResourceRecord) (len=1 cap=1) {
 (dnska.ResourceRecord) {
  Name: (string) (len=11) "example.com",
  Type: (dnska.QType) QTypeAAAA,
  Class: (dnska.QClass) ClassIN,
  TTL: (uint32) 14522,
  RDLength: (uint16) 16,
  RData: (string) (len=39) "2606:2800:0220:0001:0248:1893:25c8:1946"
 }
}
```

Encoding and decoding DNS packets:

```text
$ dnska encode [FILENAME]
$ dnska decode [FILENAME]
```

### Todo

- Caching name server.
- Recursive resolving.
- Loading zones and working as authoritative name server.

### RFCs

The DNS has a lot of RFC for describing its own features. Below
is a list of RFCs that I leant upon.

- DOMAIN NAMES - CONCEPTS AND FACILITIES [RFC1035](https://datatracker.ietf.org/doc/html/rfc1034)
- DOMAIN NAMES - IMPLEMENTATION AND SPECIFICATION [RFC1035](https://datatracker.ietf.org/doc/html/rfc1035)

  Base documents about DNS design. Not all record formats are supported. Currently, only:
  + A
  + CNAME

- DNS Extensions to Support IP Version 6 [RFC2396](https://datatracker.ietf.org/doc/html/rfc3596)

  Introduce the AAAA record type.