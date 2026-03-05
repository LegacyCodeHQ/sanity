package formatters

import "hash/fnv"

// EdgeLabel returns a deterministic 3-character alphanumeric label for the
// directed edge from → to. The same pair always produces the same label.
func EdgeLabel(from, to string) string {
	h := fnv.New64a()
	h.Write([]byte(from))
	h.Write([]byte{0})
	h.Write([]byte(to))
	return encodeAlpha(h.Sum64(), 3)
}

const alphaChars = "abcdefghijklmnopqrstuvwxyz"

func encodeAlpha(n uint64, length int) string {
	buf := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		buf[i] = alphaChars[n%26]
		n /= 26
	}
	return string(buf)
}
