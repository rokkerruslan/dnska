package ext

func longestSuffix(a, b string) string {
	ai := len(a) - 1
	bi := len(b) - 1

	l := 0

	for ai >= 0 && bi >= 0 {
		if a[ai] != b[bi] {
			break
		}

		l++
		ai--
		bi--
	}

	return a[len(a)-l:]
}
