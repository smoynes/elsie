package encoding_test

import (
	"encoding"
	"testing"

	. "github.com/smoynes/elsie/internal/encoding"
)

// Assert interface implemented.
var (
	_ encoding.TextMarshaler   = (*HexEncoder)(nil)
	_ encoding.TextUnmarshaler = (*HexEncoder)(nil)
)

func TestHexEncoder_MarshalText(t *testing.T) {

}
