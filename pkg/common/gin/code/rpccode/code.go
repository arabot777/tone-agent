package rpccode

import _errors "tone/agent/pkg/common/pkgerror"

// ErrCode implements `gitee.com/arabot777/arabot-go/pkg/errors`.Coder interface.
type ErrCode struct {
	// C refers to the code of the ErrCode.
	C int

	// External (user) facing error text.
	Ext string

	// Ref specify the reference document.
	Ref string
}

var _ _errors.Coder = &ErrCode{}

// Code returns the integer code of ErrCode.
func (coder ErrCode) Code() int {
	return coder.C
}

// String implements stringer. String returns the external error message,
// if any.
func (coder ErrCode) String() string {
	return coder.Ext
}

// Reference returns the reference document.
func (coder ErrCode) Reference() string {
	return coder.Ref
}

// nolint: unparam
func register(code int, message string, refs ...string) {
	var reference string
	if len(refs) > 0 {
		reference = refs[0]
	}

	coder := &ErrCode{
		C:   code,
		Ext: message,
		Ref: reference,
	}

	_errors.MustRegister(coder)
}
