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
	RegisteredValidator(maxValidatorName, MaxValidator)
	RegisteredValidator(minValidatorName, MinValidator)
	RegisteredValidator(iRangeValidatorName, IntRangeValidator)
	RegisteredValidator(stringRangeValidatorName, StringRangeValidator)
	RegisteredValidator(structValidatorName, StructValidator)
	RegisteredValidator(regexValidatorName, RegexMatchValiator)
	RegisteredValidator(notEmptyValidatorName, NotEmptyValidator)
	RegisteredValidator(lenValidatorName, LenValidator)
}

func RegisteredConstStr(name, val string) {
	if !constNamePat.MatchString(name) {
		panic("invalid constant name")
	}
	registeredConstStr[name] = val
}

func RegisteredConstInt(name string, val uint64) {
	if !constNamePat.MatchString(name) {
		panic("invalid constant name")
	}
	registeredConstInt[name] = val
}

func RegisteredValidator(name string, factory func(args ValidatorArgs) Validator) {
	if !namePat.MatchString(string(name)) {
		panic("invalid constant name")
	}
	registeredValidator[name] = factory
}

func RegisteredStruct(strct interface{}) {
	typ := internal.TypeIndirect(reflect.TypeOf(strct))
	_, in := registeredStruct[typ]
	if in {
		return
	}
	registeredStruct[typ] = NewStructValidator(strct)
}

func ValidateStruct(strct interface{}) error {
	typ := internal.TypeIndirect(reflect.TypeOf(strct))
	vld, in := registeredStruct[typ]
	if !in {
		return ErrStructNotRegister
	}
	return vld(strct)
}
