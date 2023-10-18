package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/smoynes/elsie/internal/log"
)

// loader.go holds an object loader.

// Loader takes object code and loads it into the machine's memory.
type Loader struct {
	log *log.Logger
}

// NewLoader creates a new object loader.
func NewLoader() *Loader {
	logger := log.DefaultLogger()
	return &Loader{log: logger}
}

// Load loads the object code starting at its origin address.
func (l *Loader) Load(vm *LC3, obj ObjectCode) (uint16, error) {
	var count uint16

	if len(obj.Code) == 0 {
		return count, ErrObjectLoader
	}

	addr := obj.Orig

	for _, code := range obj.Code {
		err := vm.Mem.store(addr, Word(code))

		if err != nil {
			return count, fmt.Errorf("%w: %w", ErrObjectLoader, err)
		}

		count++
		addr++
	}

	return count, nil
}

func (l *Loader) LoadVector(vm *LC3, vector Word, handler ObjectCode) (uint16, error) {
	if count, err := l.Load(vm, handler); err != nil {
		return count, err
	} else if err = vm.Mem.store(vector, handler.Orig); err != nil {
		return count, fmt.Errorf("%w: %w", ErrObjectLoader, err)
	} else {
		return count, nil
	}
}

// ObjectCode is a data structure that holds code and its origin offset in memory. Code may be
// comprised of either instructions or data.
type ObjectCode struct {
	Orig Word
	Code []Word
}

// Read loads an object from bytes.
func (obj *ObjectCode) Read(b []byte) (int, error) {
	var count int

	if len(b) < 2 {
		return 0, fmt.Errorf("%w: object code too small", ErrObjectLoader)
	}

	in := bytes.NewReader(b)
	err := binary.Read(in, binary.BigEndian, &obj.Orig)

	if err != nil {
		return count, fmt.Errorf("%w: %w", ErrObjectLoader, err)
	}

	count += 2

	obj.Code = make([]Word, len(b)/2-1)
	err = binary.Read(in, binary.BigEndian, obj.Code)

	if err != nil {
		return count, fmt.Errorf("%w: %w", ErrObjectLoader, err)
	}

	count += len(obj.Code) * 2

	return count, nil
}

var ErrObjectLoader = errors.New("loader error")
