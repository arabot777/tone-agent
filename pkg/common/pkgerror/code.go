package pkgerror

import (
	"fmt"
	"sync"
)

var unknownCoder = defaultCoder{100000, "An internal server error occurred", ""}

// Coder defines an interface for an error code detail information.
type Coder interface {
	// External (user) facing error text.
	String() string

	// Reference returns the detail documents for user.
	Reference() string

	// Code returns the code of the coder
	Code() int
}

type defaultCoder struct {
	// C refers to the integer code of the ErrCode.
	C int

	// External (user) facing error text.
	Ext string

	// Ref specify the reference document.
	Ref string
}

// Code returns the integer code of the coder.
func (coder defaultCoder) Code() int {
	return coder.C
}

// String implements stringer. String returns the external error message,
// if any.
func (coder defaultCoder) String() string {
	return coder.Ext
}

// Reference returns the reference document.
func (coder defaultCoder) Reference() string {
	return coder.Ref
}

// codes contains a map of error codes to metadata.
var (
	codes   = map[int]Coder{}
	codeMux = &sync.Mutex{}
)

// MustRegister register a user define error code.
// It will panic when the same Code already exist.
func MustRegister(coder Coder) {
	codeMux.Lock()
	defer codeMux.Unlock()

	if _, ok := codes[coder.Code()]; ok {
		panic(fmt.Sprintf("code: %d already exist", coder.Code()))
	}

	codes[coder.Code()] = coder
}

// ParseCoder parse any error into *withCode.
// nil error will return nil direct.
// None withStack error will be parsed as ErrUnknown.
func ParseCoder(err error) Coder {
	if err == nil {
		return nil
	}

	if v, ok := err.(*withCode); ok {
		if coder, ok := codes[v.code]; ok {
			return coder
		}
	}

	return unknownCoder
}

// ParseError parse err from withCode
func ParseError(err error) string {
	if err == nil {
		return ""
	}

	if v, ok := err.(*withCode); ok {
		return v.err.Error()
	}

	return ""
}

// IsCode reports whether any error in err's chain contains the given error code.
func IsCode(err error, code int) bool {
	if v, ok := err.(*withCode); ok {
		if v.code == code {
			return true
		}

		if v.cause != nil {
			return IsCode(v.cause, code)
		}

		return false
	}

	return false
}

func init() {
	codes[unknownCoder.Code()] = unknownCoder
}
