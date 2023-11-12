// Package encoding includes implementations of encoding.TextMarshaler and encoding.TextUnmarshaler for
// binary object code. It uses Intel Hex as a defacto standard.
package encoding // import github.com/smoynes/elsie/internal/encoding

import (
	"errors"
)
w
type HexEncoder struct{}

func (h *HexEncoder) MarshalText() ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (h *HexEncoder) UnmarshalText(bs []byte) error {
	return errors.New("not implemented")
}
