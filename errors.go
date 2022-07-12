package xvalidator

import (
	"fmt"

	"github.com/ccbhj/xvalidator/internal"
	"github.com/pkg/errors"
)

var ErrInvalidStruct = errors.New("invalid struct cannot be validated")
var ErrInvalidValidatorArgument = errors.New("invalid validator argument")
var ErrUnknownValidator = errors.New("unknown validator")
var ErrStructNotRegister = errors.New("struct not registered")
var ErrUnknownConst = errors.New("unknown const")
var ErrInvalidArgument = errors.New("invalid argument for valiator")
var ErrInvalidValidatorSyntax = internal.ErrInvalidValidatorSyntax

type ValidatorError struct {
	Reason    string
	FieldName string
}

func (e ValidatorError) Error() string {
	return fmt.Sprintf("validate fail for field %s: %s", e.FieldName, e.Reason)
}
