package main

import (
	"bufio"
	"io"
	"os"
)

type sid int

const (
	start sid = iota
	inString
	afterQuote
	afterCR
	accept
)

type state struct {
	sid    sid
	acc    []byte
	quoted bool
}

// Do some cleanup on a CSV file. Specifically:
//   - Replace double-quoted empty strings with unquoted empty strings
//   - Convert CRLF line terminations to LF
//   - Preserve existing embedded double quotes, e.g. "50"" flat-screen"
func main() {
	reader := bufio.NewReader(os.Stdin)
	state := makeState()
	for {
		c, err := reader.ReadByte()
		if err == io.EOF {
			return
		}
		if has, out := state.feed(c); has {
			os.Stdout.Write(out)
		}
	}
}

func makeState() state {
	return state{
		start,
		make([]byte, 0, 50),
		false,
	}
}

func (st *state) feed(c byte) (has bool, out []byte) {
	switch st.sid {
	case start:
		switch c {
		case '\r':
			st.sid = afterCR
		case ',':
			fallthrough
		case '\n':
			st.acc = append(st.acc, c)
			st.sid = accept
		case '"':
			st.quoted = true
			fallthrough
		default:
			st.acc = append(st.acc, c)
			st.sid = inString
		}
	case inString:
		if st.quoted {
			switch c {
			case '"':
				st.sid = afterQuote
			default:
				st.acc = append(st.acc, c)
			}
		} else {
			switch c {
			case ',':
				fallthrough
			case '\n':
				st.acc = append(st.acc, c)
				st.sid = accept
			case '\r':
				st.sid = afterCR
			default:
				st.acc = append(st.acc, c)
			}
		}
	case afterQuote:
		switch c {
		case '"':
			st.acc = append(st.acc, '"', '"')
			st.sid = inString
		case ',':
			fallthrough
		case '\n':
			st.acc = append(st.acc, '"', c)
			st.sid = accept
		case '\r':
			st.acc = append(st.acc, '"')
			st.sid = afterCR
		default:
			panic("unpaired double quote in string")
		}
	case afterCR:
		switch c {
		case '\n':
			st.acc = append(st.acc, c)
			st.sid = accept
		default:
			panic("stray \\r in input")
		}
	case accept:
		panic("shouldn't process another character when in accept state")
	}
	if st.sid == accept {
		if st.quoted && len(st.acc) == 3 {
			st.acc = st.acc[2:3]
		}
		out = make([]byte, len(st.acc))
		copy(out, st.acc)
		st.sid = start
		st.acc = st.acc[:0]
		st.quoted = false
		return true, out
	}
	return false, []byte{}
}
