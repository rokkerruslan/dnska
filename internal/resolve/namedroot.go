package resolve

import (
	"bufio"
	_ "embed"
	"log"
	"strings"
)

var namedRootIndex []string

func init() {
	lines, err := ReadNamedRootFile()
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range lines {
		if len(line) != 4 {
			continue
		}

		qType := line[2]
		addr := line[3]

		if qType != "A" {
			continue
		}

		namedRootIndex = append(namedRootIndex, addr)
	}
}

//go:embed files/named.root
var namedRootContent string

func ReadNamedRootFile() ([][]string, error) {
	var records [][]string

	scanner := bufio.NewScanner(strings.NewReader(namedRootContent))
	for scanner.Scan() {
		line := scanner.Text()

		if i := strings.IndexByte(line, ';'); i != -1 {
			line = line[:i]
		}

		var parts []string

		for _, el := range strings.Split(line, " ") {
			if part := strings.TrimSpace(el); part != "" {
				parts = append(parts, part)
			}
		}

		if len(parts) == 0 {
			continue
		}

		records = append(records, parts)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return records, nil
}
