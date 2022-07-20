package xvalidator

import (
	"reflect"
	"regexp"

	"github.com/ccbhj/xvalidator/internal"
)

var namePat = regexp.MustCompile(nameRegex)
var constNamePat = regexp.MustCompile(nameRegex)

const (
	notEmptyValidatorName    string = "not_empty"
	maxValidatorName         string = "max"
	minValidatorName         string = "min"
	lenValidatorName         string = "len"
	iRangeValidatorName      string = "irange"
	stringRangeValidatorName string = "srange"
	structValidatorName      string = "strct"
	regexValidatorName       string = "regex"
)

func init() {
	RegisterValidator(maxValidatorName, MaxValidator)
	RegisterValidator(minValidatorName, MinValidator)
	RegisterValidator(iRangeValidatorName, IntRangeValidator)
	RegisterValidator(stringRangeValidatorName, StringRangeValidator)
	RegisterValidator(structValidatorName, StructValidator)
	RegisterValidator(regexValidatorName, RegexMatchValiator)
	RegisterValidator(notEmptyValidatorName, NotEmptyValidator)
	RegisterValidator(lenValidatorName, LenValidator)
}

// RegisterConstStr registers a string constant
// name must start with letter and consist of letters and numbers
func RegisterConstStr(name, val string) {
	if !constNamePat.MatchString(name) {
		panic("invalid constant name")
	}
	registeredConstStr[name] = val
}

// RegisterConstStr registers an integer constant
// name must start with letter and consist of letters and numbers
func RegisterConstInt(name string, val uint64) {
	if !constNamePat.MatchString(name) {
		panic("invalid constant name")
	}
	registeredConstInt[name] = val
}

// RegisterValidator registers a custom validator
// name must start with letter and consist of letters and numbers
func RegisterValidator(name string, factory func(args ValidatorArgs) Validator) {
	if !namePat.MatchString(string(name)) {
		panic("invalid constant name")
	}
	registeredValidator[name] = factory
}

// RegisterStruct generate a validator for a struct pointer or struct value
func RegisterStruct(strct interface{}) {
	typ := internal.TypeIndirect(reflect.TypeOf(strct))
	_, in := registeredStruct[typ]
	if in {
		return
	}
	registeredStruct[typ] = NewStructValidator(strct)
}

// ValidateStruct validates a struct pointer of struct value
// The struct must be registed before ValidateStruct is called
func ValidateStruct(strct interface{}) error {
	typ := internal.TypeIndirect(reflect.TypeOf(strct))
	vld, in := registeredStruct[typ]
	if !in {
		return ErrStructNotRegister
	}
	return vld(strct)
}
