package tcp

import "strings"

const (
	// from include/net/tcp.h:
	headerFIN = 0x01
	headerSYN = 0x02
	headerRST = 0x04
	headerPSH = 0x08
	headerACK = 0x10
	headerURG = 0x20
	headerECE = 0x40
	headerCWR = 0x80
)

var (
	// States holds a map of  states with their associated int.
	States = map[uint8]string{
		1:  "ESTABLISHED",
		2:  "SYNSENT",
		3:  "SYNRECV",
		4:  "FINWAIT1",
		5:  "FINWAIT2",
		6:  "TIMEWAIT",
		7:  "CLOSE",
		8:  "CLOSEWAIT",
		9:  "LASTACK",
		10: "LISTEN",
		11: "CLOSING",
		12: "NEWSYNRECV",
	}
)

// FlagsToString returns a string representation of the tcp flags.
func FlagsToString(flags uint8) string {
	a := []string{}
	if (flags & headerFIN) != 0 {
		a = append(a, "FIN")
	}
	if (flags & headerSYN) != 0 {
		a = append(a, "SYN")
	}
	if (flags & headerRST) != 0 {
		a = append(a, "RST")
	}
	if (flags & headerPSH) != 0 {
		a = append(a, "PSH")
	}
	if (flags & headerACK) != 0 {
		a = append(a, "ACK")
	}
	if (flags & headerURG) != 0 {
		a = append(a, "URG")
	}
	if (flags & headerECE) != 0 {
		a = append(a, "ECE")
	}
	if (flags & headerCWR) != 0 {
		a = append(a, "CWR")
	}

	return strings.Join(a, " | ")
}
