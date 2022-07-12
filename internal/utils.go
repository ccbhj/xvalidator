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

func TypeIndirect(typ reflect.Type) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

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

func ParseAllValidatorName(s string) [][]string {
	return vldPat.FindAllStringSubmatch(s, -1)
}

func parseValidatorName(s string) (fn string, args []string) {
	matches := vldPat.FindStringSubmatch(s)
	if len(matches) != 2 {
		return "", nil
	}
	return matches[0], strings.Split(matches[1], ",")
}

const (
	initS = iota
	strS
	intS
	constS
	escapeS
)

const escapeRune = '\\'
const labelRune = '$'
const quoteRune = '\''
const sepRune = ','
const constsSepRune = '_'

type ArgsInfos struct {
	Strs []string
	Ints []uint64
	Vars []string
}

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

// parseArguments parse the arguments list into 3 kind of lists - strings,
// integers and variables
//
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
	switch state {
	case constS:
		consts = append(consts, sb.String())
	case intS:
		ints = append(ints, i)
	case strS, escapeS:
		// unquote string
		return nil, ErrInvalidValidatorSyntax
	}
	return &ArgsInfos{
		Strs: strs,
		Ints: ints,
		Vars: consts,
	}, nil
}
