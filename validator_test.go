package xvalidator

import (
	"testing"

	"github.com/ccbhj/xvalidator/internal"
	"github.com/stretchr/testify/assert"
)

func TestMaxMinValidator(t *testing.T) {
	fn := MaxValidator(ValidatorArgs{Ints: []uint64{100}})
	assert.Nil(t, fn(99))
	assert.Nil(t, fn(1))
	assert.Nil(t, fn(100))
	assert.NotNil(t, fn(101))
	assert.NotNil(t, fn(1000))

	fn = MinValidator(ValidatorArgs{Ints: []uint64{100}})
	assert.NotNil(t, fn(99))
	assert.NotNil(t, fn(1))
	assert.Nil(t, fn(101))
	assert.Nil(t, fn(1000))
}

func TestRangeValidator(t *testing.T) {
	fn := IntRangeValidator(ValidatorArgs{Ints: []uint64{1, 2, 3}})
	assert.Nil(t, fn(1))
	assert.Nil(t, fn(2))
	assert.Nil(t, fn(3))
	assert.NotNil(t, fn(4))
	assert.NotNil(t, fn(5))

	fn = StringRangeValidator(ValidatorArgs{Strs: []string{"1", "2", "3"}})
	assert.Nil(t, fn("1"))
	assert.Nil(t, fn("2"))
	assert.Nil(t, fn("3"))
	assert.NotNil(t, fn("4"))
	assert.NotNil(t, fn("5"))
}

func TestNewStructValidator(t *testing.T) {
	type TestStruct struct {
		A          int `xvldt:"irange(1, 99, 100), max(99)"`
		B          int `xvldt:"max(100)"`
		unexported int `xvldt:"min(1)"`
	}
	RegisterStruct(TestStruct{})
	var a = TestStruct{A: 1, B: 99}
	var b = TestStruct{A: 100, B: 99}
	var c = TestStruct{A: 1000, B: 99}
	var d = TestStruct{A: 1, B: 101}

	assert.Nil(t, ValidateStruct(a))
	assert.NotNil(t, ValidateStruct(b))
	assert.NotNil(t, ValidateStruct(c))
	assert.NotNil(t, ValidateStruct(d))
}

func TestStructValidator(t *testing.T) {
	max := 100
	type A struct {
		I int    `xvldt:"max(MAX_INT)"`
		S string `xvldt:"regex('[a-zA-Z]+')"`
	}
	type B struct {
		J int
		A A `xvldt:"strct()"`
	}
	RegisterConstInt("MAX_INT", uint64(max))
	RegisterStruct(A{})
	RegisterStruct(B{})

	// B's validation should depend on A
	var a = A{I: 99, S: "abc"}
	var b = B{J: 99, A: a}

	assert.Nil(t, ValidateStruct(b))

	b.A.I = 1000
	assert.NotNil(t, ValidateStruct(b))
	b.A.I = 99
	b.A.S = "123"
	assert.NotNil(t, ValidateStruct(b))
}

func TestParseArguments(t *testing.T) {
	testCases := map[string]internal.ArgsInfos{
		`I9, I_1, I____10`: {
			Vars: []string{"I9", "I_1", "I____10"},
		},
		`123, 456, 789`: {
			Ints: []uint64{123, 456, 789},
		},
		`'hello', 'xvalidator', 'are  ', 's0 ', 'c007 !!!'`: {
			Strs: []string{"hello", "xvalidator", "are  ", "s0 ", "c007 !!!"},
		},
		`INT_MAX, 123, 'hello'`: {
			Strs: []string{"hello"},
			Ints: []uint64{123},
			Vars: []string{"INT_MAX"},
		},
		` 123 `: {
			Ints: []uint64{123},
		},
		` 'test', `: {
			Strs: []string{"test"},
		},
		` '\'test\'' `: {
			Strs: []string{"'test'"},
		},
	}

	// invalid cases
	errCases := []string{
		string('\u0000'),    // char not printable
		`'hello`,            // unclosed quote
		`CONST_not_capital`, // constant not in capital
		`10_23`,             // invalid number
		`|..`,               //  unknown token
	}

	for _, s := range errCases {
		act, err := internal.ParseArguments(s)
		if err == nil {
			t.Logf("result=%v", act)
			t.FailNow()
		}
	}

	for s, exp := range testCases {
		act, err := internal.ParseArguments(s)
		if !assert.Nil(t, err) {
			t.Logf("input=%q result=%v", s, act)
			t.FailNow()
		}
		assert.Equal(t, exp.Vars, act.Vars)
		assert.Equal(t, exp.Ints, act.Ints)
		assert.Equal(t, exp.Strs, act.Strs)
	}
}
