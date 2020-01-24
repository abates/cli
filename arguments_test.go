package cli

import (
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type getter interface {
	Get() interface{}
}

type intSlice []int

func (il *intSlice) Set(str string) error {
	v, err := strconv.Atoi(str)
	if err == nil {
		*il = append(*il, v)
	}
	return err
}

func (il *intSlice) Get() interface{} {
	return ([]int)(*il)
}

func (il *intSlice) String() string {
	list := make([]string, len(*il))
	for i, v := range *il {
		list[i] = strconv.Itoa(v)
	}
	return strings.Join(list, ",")
}

func TestArguments(t *testing.T) {
	var b bool
	var d time.Duration
	var i int
	var i64 int64
	var f64 float64
	var s string
	var ui uint
	var ui64 uint64
	var varSlice []int

	tests := []struct {
		desc       string
		input      []string
		cb         func(*Arguments)
		want       interface{}
		wantString string
		wantErr    error
	}{
		{"bool", []string{"true"}, func(args *Arguments) { args.Bool(&b, "bool") }, true, "true", nil},
		{"bool", []string{"foobar"}, func(args *Arguments) { args.Bool(&b, "bool") }, false, "false", errParse},
		{"duration", []string{"64s"}, func(args *Arguments) { args.Duration(&d, "duration") }, time.Second * 64, "1m4s", nil},
		{"duration err", []string{"sixty-four seconds"}, func(args *Arguments) { args.Duration(&d, "duration") }, time.Duration(0), "", errParse},
		{"int", []string{"1"}, func(args *Arguments) { args.Int(&i, "int") }, 1, "1", nil},
		{"int errParse", []string{"one"}, func(args *Arguments) { args.Int(&i, "int") }, 0, "", errParse},
		{"int errRange", []string{"18446744073709551615"}, func(args *Arguments) { args.Int(&i, "int") }, 0, "", errRange},
		{"int64", []string{"2"}, func(args *Arguments) { args.Int64(&i64, "int64") }, int64(2), "2", nil},
		{"int64 err", []string{"two"}, func(args *Arguments) { args.Int64(&i64, "int64") }, 0, "", errParse},
		{"float64", []string{"2.001"}, func(args *Arguments) { args.Float64(&f64, "float64") }, float64(2.001), "2.001", nil},
		{"float64 err", []string{"two point zero zero one"}, func(args *Arguments) { args.Float64(&f64, "float64") }, 0, "", errParse},
		{"string", []string{"foobar"}, func(args *Arguments) { args.String(&s, "string") }, "foobar", "foobar", nil},
		{"uint", []string{"5"}, func(args *Arguments) { args.Uint(&ui, "uint") }, uint(5), "5", nil},
		{"uint", []string{"five"}, func(args *Arguments) { args.Uint(&ui, "uint") }, 0, "", errParse},
		{"uint64", []string{"6"}, func(args *Arguments) { args.Uint64(&ui64, "uint64") }, uint64(6), "6", nil},
		{"uint64 err", []string{"six"}, func(args *Arguments) { args.Uint64(&ui64, "uint64") }, 0, "", errParse},
		{"varslice", []string{"1", "2", "3"}, func(args *Arguments) { args.VarSlice((*intSlice)(&varSlice), "n n n n...") }, []int{1, 2, 3}, "1,2,3", nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			args := &Arguments{}
			test.cb(args)
			gotErr := args.Parse(test.input)
			if test.wantErr != gotErr {
				t.Errorf("want err %v got %v", test.wantErr, gotErr)
			} else if gotErr == nil {
				if g, ok := args.args[0].value.(getter); ok {
					got := g.Get()
					if !reflect.DeepEqual(test.want, got) {
						t.Errorf("want %v got %v", test.want, got)
					}

					gotString := args.args[0].value.String()
					if test.wantString != gotString {
						t.Errorf("want string %q got %q", test.wantString, gotString)
					}
				} else {
					t.Errorf("value %T does not implement getter", args.args[0].value)
				}
			}
		})
	}
}

func TestNumError(t *testing.T) {
	tests := []struct {
		desc  string
		input error
		want  error
	}{
		{"errParse", &strconv.NumError{Err: strconv.ErrSyntax}, errParse},
		{"errRange", &strconv.NumError{Err: strconv.ErrRange}, errRange},
		{"EOF", io.EOF, io.EOF},
		{"Other", &strconv.NumError{Err: io.EOF}, io.EOF},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := numError(test.input)
			if got != test.want {
				t.Errorf("want error %v got %v", test.want, got)
			}
		})
	}
}

type testValue struct{}

func (testValue) String() string     { return "" }
func (testValue) Set(s string) error { return nil }

func TestArgumentsParse(t *testing.T) {
	tests := []struct {
		desc     string
		input    []string
		values   []Value
		wantLen  int
		wantArgs []string
		wantErr  error
	}{
		{"test 1", []string{"true"}, []Value{testValue{}}, 1, []string{}, nil},
		{"test 2", []string{}, []Value{testValue{}}, 1, []string{}, errNumArguments},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			args := []*argument{}
			for _, value := range test.values {
				args = append(args, &argument{value: value})
			}

			arguments := &Arguments{args: args}
			gotErr := arguments.Parse(test.input)
			if test.wantErr != gotErr {
				t.Errorf("want error %v got %v", test.wantErr, gotErr)
			} else if gotErr == nil {
				gotLen := arguments.Len()
				if test.wantLen != gotLen {
					t.Errorf("want len %d got %d", test.wantLen, gotLen)
				}

				gotArgs := arguments.Args()
				if !reflect.DeepEqual(test.wantArgs, gotArgs) {
					t.Errorf("want args %v got %v", test.wantArgs, gotArgs)
				}
			}
		})
	}
}

func TestArgumentsUsage(t *testing.T) {
	tests := []struct {
		desc  string
		input []*argument
		want  string
	}{
		{"no args", nil, ""},
		{"one arg", []*argument{{desc: "<foo>"}}, "<foo>"},
		{"two args", []*argument{{desc: "<foo>"}, {desc: "<bar>"}}, "<foo> <bar>"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			builder := &strings.Builder{}
			args := &Arguments{args: test.input}
			args.Usage(builder)
			got := builder.String()
			if test.want != got {
				t.Errorf("want usage %q got %q", test.want, got)
			}
		})
	}
}
