package xvalidator

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/ccbhj/xvalidator/internal"
	"github.com/pkg/errors"
)

var registeredStruct = make(map[reflect.Type]Validator)
var registeredConstInt = make(map[string]uint64)
var registeredConstStr = make(map[string]string)
var registeredValidator = make(map[string]func(args ValidatorArgs) Validator)

const DefaultTagName = "xvldt"

type ValidatorArgs struct {
	Strs []string
	Ints []uint64
	Typ  reflect.Type
}

type Validator func(interface{}) error

func dummyValidator(interface{}) error {
	return nil
}

// And concat two Validator into a new Validator
// Validator v will be executed first, if it return a not-nil error, return
// immediately, otherwise execute other.
func (v Validator) And(other Validator) Validator {
	if other == nil && v == nil {
		return dummyValidator
	} else if v == nil {
		return other
	} else if other == nil {
		return v
	}
	return func(i interface{}) error {
		err := v(i)
		if err == nil {
			return other(i)
		}
		return err
	}
}

// WithName return a Validator that will fill the field name into ValidatorError
// if any error occurs
func (v Validator) WithName(s string) Validator {
	if v == nil {
		v = dummyValidator
	}
	return func(i interface{}) error {
		var e ValidatorError
		if err := v(i); err != nil && errors.As(err, &e) {
			e.FieldName = s
			return e
		} else if err != nil {
			return err
		}
		return nil
	}
}

// NewStructValidator parse the 'xvldt' tag in struct's fields and return a new
// Validator.
// All validator can be seperated with ',', and must carry '()' even if the
// validator need no arguments.
func NewStructValidator(args interface{}) Validator {
	val := reflect.Indirect(reflect.ValueOf(args))
	if val.Kind() != reflect.Struct {
		panic(errors.WithMessage(ErrInvalidValidatorArgument, "must be struct or struct pointer"))
	}

	typ := val.Type()
	vlds := make([]Validator, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag, has := field.Tag.Lookup(DefaultTagName)
		if !has {
			vlds[i] = nil
			continue
		}
		// parse all the validator name and arguments
		var vld Validator
		for _, match := range internal.ParseAllValidatorName(tag) {
			fn := strings.TrimSpace(match[1])
			argStr := strings.TrimSpace(match[2])
			v, in := registeredValidator[fn]
			if !in {
				panic(ErrUnknownValidator)
			}
			arg, err := internal.ParseArguments(argStr)
			if err != nil {
				panic(err)
			}

			// replace the variable with the registered value
			for _, v := range arg.Vars {
				ic, in := registeredConstInt[v]
				if in {
					arg.Ints = append(arg.Ints, ic)
					break
				}
				sc, in := registeredConstStr[v]
				if !in {
					panic(errors.WithMessage(ErrUnknownConst, v))
				}
				arg.Strs = append(arg.Strs, sc)
			}

			typ := field.Type
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			va := ValidatorArgs{
				Strs: arg.Strs,
				Ints: arg.Ints,
				Typ:  typ,
			}

			if vld == nil {
				vld = v(va)
			} else {
				vld = vld.And(v(va))
			}
		}
		vlds[i] = vld.WithName(field.Name)
	}

	return func(arg interface{}) error {
		val := reflect.Indirect(reflect.ValueOf(arg))
		if !val.IsValid() {
			return ErrInvalidStruct
		}
		for i, vld := range vlds {
			if vld == nil {
				continue
			}
			field := reflect.Indirect(val.Field(i))
			if !field.CanInterface() {
				continue
			}
			if err := vld(field.Interface()); err != nil {
				return err
			}
		}
		return nil
	}
}

// StringRangeValidator return a Validator that check whether a string value in a
// string list or not
func StringRangeValidator(arg ValidatorArgs) Validator {
	v := make(map[string]struct{}, len(arg.Strs))
	for _, x := range arg.Strs {
		v[x] = struct{}{}
	}
	return func(arg interface{}) error {
		s, ok := arg.(string)
		if !ok {
			return errors.New("not a string")
		}
		_, in := v[s]
		if !in {
			return ValidatorError{
				Reason: "invalid value",
			}
		}
		return nil
	}
}

// IntRangeValidator return a Validator that check whether a integer value in a
// integer list or not. All integer and string can support.
func IntRangeValidator(arg ValidatorArgs) Validator {
	v := make(map[uint64]struct{}, len(arg.Ints))
	for _, x := range arg.Ints {
		v[x] = struct{}{}
	}
	return func(arg interface{}) error {
		i, err := internal.ToUint64(arg)
		if err != nil {
			return errors.WithMessage(err, "not an integer")
		}
		_, in := v[i]
		if !in {
			return ValidatorError{
				Reason: "invalid value",
			}
		}
		return nil
	}
}

// MaxValidator return a Validator that check whether an integer is less than the
// first arguments of the validator
func MaxValidator(arg ValidatorArgs) Validator {
	if len(arg.Ints) < 1 {
		panic(errors.New("MaxValidator required one integer"))
	}
	max := arg.Ints[0]
	return func(arg interface{}) error {
		i, err := internal.ToUint64(arg)
		if err != nil {
			return errors.WithMessage(err, "not an integer")
		}
		if i > max {
			return ValidatorError{
				Reason: "out of range",
			}
		}
		return nil
	}
}

// MaxValidator return a Validator that check whether an integer is larger than the
// first arguments of the validator
func MinValidator(arg ValidatorArgs) Validator {
	if len(arg.Ints) < 1 {
		panic(errors.New("MaxValidator required one integer"))
	}
	min := arg.Ints[0]
	return func(arg interface{}) error {
		i, err := internal.ToUint64(arg)
		if err != nil {
			return errors.WithMessage(err, "not an integer")
		}
		if i < min {
			return ValidatorError{
				Reason: "out of range",
			}
		}
		return nil
	}
}

// EmptyValidator return a Validator that check whether a string is not empty
func NotEmptyValidator(_ ValidatorArgs) Validator {
	return func(arg interface{}) error {
		s, ok := arg.(string)
		if !ok {
			return errors.New("not a string")
		}
		if strings.TrimSpace(s) == "" {
			return ValidatorError{
				Reason: "empty string",
			}
		}
		return nil
	}
}

// RegexMatchValiator return a Validator that check whether a string match the
// fisrt argument of the validator
func RegexMatchValiator(v ValidatorArgs) Validator {
	if v.Typ.Kind() != reflect.String {
		panic("invalid type for regex validator")
	}
	if len(v.Strs) < 1 {
		panic(errors.New("RegexValidator required one string"))
	}
	pat, err := regexp.Compile(v.Strs[0])
	if err != nil {
		panic(errors.WithMessage(err, "invalid regex pattern"))
	}
	return func(arg interface{}) error {
		s, ok := arg.(string)
		if !ok {
			return errors.New("not a string")
		}
		if !pat.MatchString(s) {
			return ValidatorError{
				Reason: "string not match pattern",
			}
		}
		return nil

	}
}

// LenMatchValiator return a Validator that check whether a string's length is
// the same as the first argument of the validator
func LenValidator(v ValidatorArgs) Validator {
	if v.Typ.Kind() != reflect.String {
		panic("invalid type for len validator")
	}
	if len(v.Ints) < 1 {
		panic("need an integer for len validator")
	}
	l := v.Ints[0]
	return func(arg interface{}) error {
		s, ok := arg.(string)
		if !ok {
			return errors.New("not a string")
		}

		if uint64(len(s)) != l {
			return ValidatorError{
				Reason: "invalid string length",
			}
		}
		return nil
	}
}

// StructValiator return a Validator that check whether a struct pointer of
// struct value can pass the validation.
func StructValidator(v ValidatorArgs) Validator {
	vld, in := registeredStruct[v.Typ]
	if !in {
		panic(errors.WithMessage(ErrStructNotRegister, v.Typ.Name()))
	}
	return vld
}
