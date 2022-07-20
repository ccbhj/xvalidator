package internal

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")

var typeOfUint64 = reflect.TypeOf(uint64(0))

// TypeIndirect return the value type of the pointer if typ's kind is pointer
// otherwise return typ directly
func TypeIndirect(typ reflect.Type) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

// ToUint64 convert all integer type to ToUint64,
// string can also be parsed into uint64
func ToUint64(args interface{}) (uint64, error) {
	switch v := args.(type) {
	case int:
		return uint64(v), nil
	case int8:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		val := reflect.ValueOf(args)
		if val.Type().ConvertibleTo(typeOfUint64) {
			return val.Convert(typeOfUint64).Uint(), nil
		}
	}
	return 0, ErrInvalidValidatorSyntax
}

var vldPat = regexp.MustCompile(FnRegex)

// ParseAllValidatorName return all validator name and its arguments in s
// Note that validator name must start with letter
// example:
//     "range(1, 2, 3)" => ["range", "1, 2, 3"]
func ParseAllValidatorName(s string) [][]string {
	return vldPat.FindAllStringSubmatch(s, -1)
}

type ArgsInfos struct {
	Strs []string
	Ints []uint64
	Vars []string
}

//         [,space]                         [,space]
//         +-----+                  +-----------------------+ +-digit-+
//         |     |                  |                       | |       |
//         |     |                  |                       | |       |
//         +->******** <------------+                     **********  |
//      +-----* init * ------------digit----------------> * number *<-+
//      |  +->******** <------------+                     **********
//      |  |     |                  |
//      |  |     |                  +---------------quote------------------+
//      |  |     +------------------------quote-------------------------+  |
// [A-Z |  |                                                            |  |
// 0-9_]|  |  [,space]                                                          |  |
//      |  |                                      [./]                  |  |
//      |  +-*********              ********** ----------> ********** <-+  |
//      +--->* label *<-+           * escape * <--slash--- * string *------+
//           *********  |           **********             **********<-+
//               |      |                                       |      |
//               |      |                                       |      |
//               +------+                                       +------+
//    					  [A-Z0-9_]                                       [.]

// ParseArguments parse the arguments list into 3 kind of lists - strings, integers and variables
// Only valid unicode char can be supported
// All string must quote with '', you can use '/' to escape characters.
// All variables must all be capital letters or '_'.
// All arguments should be seperated with ','
func ParseArguments(s string) (*ArgsInfos, error) {
	var (
		i     uint64
		sb    = strings.Builder{}
		state = initS

		consts []string
		strs   []string
		ints   []uint64
		err    error
	)

	for _, r := range s {
		switch state {
		case initS:
			err = handleInitS(&state, r, &i, &sb)
		case constS:
			err = handleConstS(&state, r, &sb, &consts)
		case intS:
			err = handleIntS(&state, r, &i, &ints)
		case strS:
			err = handleStrS(&state, r, &sb, &strs)
		case escapeS:
			err = handleEscapeS(&state, r, &sb)
		}

		if err != nil {
			return nil, err
		}
	}
	// handle the final state so that state can transit to initS
	switch state {
	case constS:
		consts = append(consts, sb.String())
	case intS:
		ints = append(ints, i)
	case strS, escapeS:
		// unquoted string
		return nil, ErrInvalidValidatorSyntax
	}
	return &ArgsInfos{
		Strs: strs,
		Ints: ints,
		Vars: consts,
	}, nil
}

const (
	initS = iota
	strS
	intS
	constS
	escapeS
)

const escapeRune = '\\'
const quoteRune = '\''
const sepRune = ','
const constsSepRune = '_'

func handleInitS(state *int, r rune, i *uint64, sb *strings.Builder) error {
	if unicode.IsDigit(r) {
		*i = uint64(r - '0')
		*state = intS
	} else if unicode.IsUpper(r) {
		sb.WriteRune(r)
		*state = constS
	} else if r == quoteRune {
		*state = strS
	} else if unicode.IsSpace(r) || r == sepRune {
		return nil
	} else {
		return ErrInvalidValidatorSyntax
	}
	return nil
}

func handleIntS(state *int, r rune, i *uint64, ints *[]uint64) error {
	if r == sepRune || unicode.IsSpace(r) {
		*ints = append(*ints, *i)
		*state = initS
	} else if unicode.IsDigit(r) {
		*i = *i*10 + uint64(r-'0')
	} else {
		return ErrInvalidValidatorSyntax
	}
	return nil
}

func handleConstS(state *int, r rune, sb *strings.Builder, consts *[]string) error {
	if r == sepRune || unicode.IsSpace(r) {
		*consts = append(*consts, sb.String())
		sb.Reset()
		*state = initS
	} else if unicode.IsUpper(r) || unicode.IsDigit(r) || r == constsSepRune {
		_, err := sb.WriteRune(r)
		if err != nil {
			return errors.WithMessage(err, "fail to write rune in string")
		}
	} else {
		return ErrInvalidValidatorSyntax
	}
	return nil
}

func handleStrS(state *int, r rune, sb *strings.Builder, strs *[]string) error {
	if r == quoteRune {
		*strs = append(*strs, sb.String())
		sb.Reset()
		*state = initS
	} else if r == escapeRune {
		*state = escapeS
	} else if unicode.IsPrint(r) {
		sb.WriteRune(r)
	} else {
		return ErrInvalidValidatorSyntax
	}
	return nil
}

func handleEscapeS(state *int, r rune, sb *strings.Builder) error {
	if unicode.IsPrint(r) {
		sb.WriteRune(r)
		*state = strS
		return nil
	}
	return ErrInvalidValidatorSyntax
}
