// Package encoding includes implementations of encoding.TextMarshaler and encoding.TextUnmarshaler
// to encode and decode binary object code. It is based on Intel Hex file-encoding.
//
// Each file is composed of lines composed of a prefix, length, address, type, (optional data) and a
// checksum. In shorthand:
//
//	:LLAAAATT[DD...]CC
//	0123456789
//
// See [Grammar] for a formal grammar.
//
// # Bugs
//
// This is not a complete implementation Intel Hex encoding; it is for internal use, only. It
// supports minimal record types, specifically just the data and end-of-file record types.
package encoding

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/smoynes/elsie/internal/vm"
)

const Grammar = `
file  = { line } ;
line  = ':' len addr data check nl ;
len   = byte ;
addr  = byte byte ;
data  = { byte }
byte  = hex hex ;
hex   = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'
      | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' ;
nl    = '\n' ;
`

// HexEncoding implements marshalling and unmarshalling of ELSIE binaries as Intel Hex files.
type HexEncoding struct {
	code []vm.ObjectCode
}

// Code returns the collected object code.
func (h HexEncoding) Code() []vm.ObjectCode {
	return h.code
}

func (h *HexEncoding) MarshalText() ([]byte, error) {
	var (
		buf   bytes.Buffer // Buffered output.
		check byte         // Checksum accumulator.
		val   [2]byte      // Value to encode.

		hexEnc = hex.NewEncoder(&buf) // Translates output to hex.
	)

	for i := range h.code {
		code := h.code[i]

		_ = buf.WriteByte(':')

		l := len(code.Code)
		val[0] = byte(l * 2)
		_, _ = hexEnc.Write(val[:1])
		check += val[0]

		val[0] = byte(code.Orig >> 8)
		val[1] = byte(code.Orig & 0x00ff)
		_, _ = hexEnc.Write(val[:])
		check += val[0]
		check += val[1]

		buf.WriteByte('0')
		buf.WriteByte('0')

		for _, word := range code.Code {
			val[0] = byte(word & 0xff00 >> 8)
			val[1] = byte(word & 0x00ff)
			_, _ = hexEnc.Write(val[:])
			check += val[0]
			check += val[1]
		}

		val[0] = 1 + ^check
		_, _ = hexEnc.Write(val[:1])

		buf.WriteByte('\n')
	}

	buf.Write([]byte(":00000001ff\n"))

	return buf.Bytes(), nil
}

func (h *HexEncoding) UnmarshalText(bs []byte) error {
	line := bufio.NewScanner(bytes.NewReader(bs))

	for line.Scan() {
		var (
			rec []byte = line.Bytes() //nolint:stylecheck

			recLen   byte    // Number of bytes in data field; excludes address, type, checksum fields.
			recAddr  uint16  // Record address.
			recKind  kind    // Record type.
			recCheck byte    // Expected checksum.
			check    byte    // Accumulated checksum.
			dec      [4]byte // Decode buffer.
		)

		if len(rec) == 0 {
			break
		} else if token := rec[0]; token == '\n' {
			continue
		} else if token != ':' {
			return fmt.Errorf("%w: line does not start with ':'", errInvalidHex)
		}

		if _, err := hex.Decode(dec[:1], rec[1:3]); err != nil {
			return fmt.Errorf("%w: len:%s", errInvalidHex, err.Error())
		} else {
			recLen = dec[0]
		}

		check += dec[0]

		if _, err := hex.Decode(dec[:2], rec[3:7]); err != nil {
			return fmt.Errorf("%w: addr: %s", errInvalidHex, err.Error())
		} else {
			recAddr = binary.BigEndian.Uint16(dec[:2])
		}

		check += dec[0] + dec[1]

		if _, err := hex.Decode(dec[:1], rec[7:9]); err != nil {
			return fmt.Errorf("%w: type: %s", errInvalidHex, err.Error())
		} else {
			recKind = kind(dec[0])
		}

		check += dec[0]

		if _, err := hex.Decode(dec[:1], rec[len(rec)-2:]); err != nil {
			return fmt.Errorf("%w: check: %s", errInvalidHex, err.Error())
		} else {
			recCheck = dec[0]
		}

		if recLen%2 != 0 {
			return fmt.Errorf("%w: odd data length", errInvalidHex)
		} else if recKind == kindData && recLen > 0 {
			hexData := make([]byte, recLen)

			if _, err := hex.Decode(hexData, rec[9:9+recLen*2]); err != nil {
				return fmt.Errorf("%w: data: %s", errInvalidHex, err.Error())
			}

			code := make([]vm.Word, recLen/2)
			for i := byte(0); i < recLen/2; i++ {
				code[i] = vm.Word(hexData[2*i])<<8 | vm.Word(hexData[2*i+1])
				check += hexData[2*i]
				check += hexData[2*i+1]
			}

			check = 1 + ^check
			if check != recCheck {
				return fmt.Errorf("%w: checksum invalid: %02x != %02x",
					errInvalidHex, check, recCheck)
			}

			h.code = append(h.code, vm.ObjectCode{
				Orig: vm.Word(recAddr),
				Code: code,
			})
		} else if recKind == kindEOF {
			check = 1 + ^check
			if check != recCheck {
				return fmt.Errorf("%w: checksum invalid: %02x != %02x",
					errInvalidHex, check, recCheck)
			}
			break
		} else {
			return fmt.Errorf("%w: unexpected record type: %d", errInvalidHex, recKind)
		}
	}

	if len(h.code) == 0 {
		return errEmpty
	}

	return nil
}

// kind represents the type of encoded record. Only the subset of record types supported by the
// encoder are supported.
type kind byte

const (
	kindData kind = 0
	kindEOF  kind = 1
)

type decodingError struct{}

func (decodingError) Error() string {
	return "decoding error"
}

func (de *decodingError) Is(err error) bool {
	if de == err {
		return true
	} else if _, ok := err.(*decodingError); ok {
		return true
	} else {
		return false
	}
}

var (
	// ErrDecode is a wrapped error that is returned when decoding fails.
	ErrDecode = &decodingError{}

	errEmpty      = fmt.Errorf("%w: no data decoded", ErrDecode)
	errInvalidHex = fmt.Errorf("%w: invalid encoding", ErrDecode)
)
