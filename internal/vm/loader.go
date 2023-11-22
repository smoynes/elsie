package vm

// loader.go holds an object loader.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/smoynes/elsie/internal/log"
)

// Loader takes object code and loads it into the machine's memory.
type Loader struct {
	vm  *LC3
	log *log.Logger
}

// NewLoader creates a new object loader.
func NewLoader(vm *LC3) *Loader {
	logger := log.DefaultLogger()

	return &Loader{
		vm:  vm,
		log: logger,
	}
}

// Load loads the object code starting at its origin address.
func (l *Loader) Load(obj ObjectCode) (uint16, error) {
	if len(obj.Code) == 0 {
		return 0, fmt.Errorf("%w: object too small", ErrObjectLoader)
	}

	var (
		addr  = obj.Orig
		count = uint16(0)
	)

	for _, code := range obj.Code {
		err := l.vm.Mem.store(addr, code)

		if err != nil {
			return count, fmt.Errorf("%w: %w", ErrObjectLoader, err)
		}

		count++
		addr++
	}

	return count, nil
}

// LoadVector stores the object and sets the vector-table entry to the object's origin address.
func (l *Loader) LoadVector(vector Word, obj ObjectCode) (uint16, error) {
	l.log.Debug("Loading vector", "vec", vector, "obj", obj)

	if count, err := l.Load(obj); err != nil {
		return count, err
	} else if err = l.vm.Mem.store(vector, obj.Orig); err != nil {
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
func (obj *ObjectCode) read(b []byte) (int, error) {
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
