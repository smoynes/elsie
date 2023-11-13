// Package encoding includes implementations of encoding.TextMarshaler and encoding.TextUnmarshaler for
// binary object code. It uses Intel Hex as a defacto standard.
package encoding // import github.com/smoynes/elsie/internal/encoding

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/smoynes/elsie/internal/vm"
)

// HexEncoding implements marshalling and unmarshalling of Intel Hex files.
type HexEncoding struct {
	code *vm.ObjectCode
}

func (h HexEncoding) Code() vm.ObjectCode {
	code := vm.ObjectCode{}
	if h.code != nil {
		code = *h.code // copy
	}

	return code
}

func (h *HexEncoding) MarshalText() ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (h *HexEncoding) UnmarshalText(bs []byte) error {
	buf := bufio.NewReader(bytes.NewReader(bs))
	code := []vm.ObjectCode(nil)

	for { // iterate over lines in buffer

		var (
			recLen  int8   // Number of bytes in data field; excludes address, type, checksum fields.
			recAddr uint16 // Record address.
		)

		// First, look for leading ':'
		if prefix, err := buf.ReadByte(); err == io.EOF {
			break
		} else if err != nil {
			return err
		} else if prefix == '\n' {
			continue
		} else if prefix != ':' {
			return fmt.Errorf("%w: line does not start with ':'",
				invalidEncodingErr)
		}

		// Next, read two hexadecimal bytes.
		var countBytes [2]byte
		if _, err := io.ReadFull(buf, countBytes[:]); err != nil {
			return err
		}

		// Decode hex bytes into a byte value.
		if _, err := hex.Decode(countBytes[:1], countBytes[:]); err != nil {
			return fmt.Errorf("%w: %s", invalidEncodingErr, err)
		}

		// Convert the byte into an integer containing the record length.
		err := binary.Read(bytes.NewReader(countBytes[:1]), binary.BigEndian, &recLen)
		if err != nil {
			return fmt.Errorf("%w: %s", invalidEncodingErr, err)
		}

		// Eat to end of line. (?)
		if _, err := buf.ReadBytes('\n'); err == io.EOF {
			break
		}
	}

	if len(code) == 0 {
		return emptyErr
	}

	return nil
}

// kind represents the type of encoded record. Only the subset of record types supported by the
// encoder are supported.
type kind byte

const (
	kindData        kind = 0
	kindEOF         kind = 1
	kindSegmentAddr kind = 2
	kindLinearAddr  kind = 4
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
	// DecodingErr is a wrapped error that is returned when decoding fails.
	DecodingErr = &decodingError{}

	emptyErr           = fmt.Errorf("%w: no data decoded", DecodingErr)
	invalidEncodingErr = fmt.Errorf("%w: invalid encoding", DecodingErr)
)
