package cli

import (
	"flag"
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

func (il *intSlice) Set(str []string) error {
	for _, s := range str {
		v, err := strconv.Atoi(s)
		if err == nil {
			*il = append(*il, v)
		} else {
			return err
		}
	}
	return nil
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

func TestStringers(t *testing.T) {
	tests := []struct {
		desc    string
		input   interface{}
		want    string
		wantGet interface{}
	}{
		{"bool", boolValue(true), "true", true},
		{"int", intValue(42), "42", 42},
		{"int64", int64Value(43), "43", int64(43)},
		{"uint", uintValue(44), "44", uint(44)},
		{"uint64", uint64Value(45), "45", uint64(45)},
		{"string", stringValue("12345"), "12345", "12345"},
		{"float64", float64Value(46), "46", float64(46)},
		{"duration", durationValue(time.Second * 64), "1m4s", time.Second * 64},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			g := reflect.New(reflect.TypeOf(test.input))
			g.Elem().Set(reflect.ValueOf(test.input))
			if value, ok := g.Interface().(flag.Value); ok {
				got := value.String()
				if test.want != got {
					t.Errorf("Wanted string %q got %q", test.want, got)
				}
			} else {
				t.Errorf("Expected %T to implement flag.Value", test.input)
			}

			if getter, ok := g.Interface().(flag.Getter); ok {
				got := getter.Get()
				if test.wantGet != got {
					t.Errorf("Wanted value %v got %v", test.wantGet, got)
				}
			} else {
				t.Errorf("Expected %T to implement flag.Getter", test.input)
			}
		})
	}
}

func TestArguments(t *testing.T) {
	tests := []struct {
		desc    string
		input   []string
		cb      func(*Arguments) interface{}
		want    interface{}
		wantErr error
	}{
		{"bool", []string{"true"}, func(args *Arguments) interface{} { return args.Bool("bool") }, true, nil},
		{"bool", []string{"foobar"}, func(args *Arguments) interface{} { return args.Bool("bool") }, false, errParse},
		{"duration", []string{"64s"}, func(args *Arguments) interface{} { return args.Duration("duration") }, time.Second * 64, nil},
		{"duration err", []string{"sixty-four seconds"}, func(args *Arguments) interface{} { return args.Duration("duration") }, time.Duration(0), errParse},
		{"int", []string{"1"}, func(args *Arguments) interface{} { return args.Int("int") }, 1, nil},
		{"int errParse", []string{"one"}, func(args *Arguments) interface{} { return args.Int("int") }, 0, errParse},
		{"int errRange", []string{"18446744073709551615"}, func(args *Arguments) interface{} { return args.Int("int") }, 0, errRange},
		{"int64", []string{"2"}, func(args *Arguments) interface{} { return args.Int64("int64") }, int64(2), nil},
		{"int64 err", []string{"two"}, func(args *Arguments) interface{} { return args.Int64("int64") }, 0, errParse},
		{"float64", []string{"2.001"}, func(args *Arguments) interface{} { return args.Float64("float64") }, float64(2.001), nil},
		{"float64 err", []string{"two point zero zero one"}, func(args *Arguments) interface{} { return args.Float64("float64") }, 0, errParse},
		{"string", []string{"foobar"}, func(args *Arguments) interface{} { return args.String("string") }, "foobar", nil},
		{"uint", []string{"5"}, func(args *Arguments) interface{} { return args.Uint("uint") }, uint(5), nil},
		{"uint", []string{"five"}, func(args *Arguments) interface{} { return args.Uint("uint") }, 0, errParse},
		{"uint64", []string{"6"}, func(args *Arguments) interface{} { return args.Uint64("uint64") }, uint64(6), nil},
		{"uint64 err", []string{"six"}, func(args *Arguments) interface{} { return args.Uint64("uint64") }, 0, errParse},
		{"varslice", []string{"1", "2", "3"}, func(args *Arguments) interface{} {
			varSlice := []int{}
			args.VarSlice((*intSlice)(&varSlice), "n n n n...")
			return &varSlice
		}, []int{1, 2, 3}, nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			args := &Arguments{}
			g := test.cb(args)
			want := test.want
			gotErr := args.Parse(test.input)
			if test.wantErr != gotErr {
				t.Errorf("want err %v got %v", test.wantErr, gotErr)
			} else if gotErr == nil {
				got := reflect.ValueOf(g).Elem().Interface()
				if !reflect.DeepEqual(want, got) {
					t.Errorf("want %T got %T", want, got)
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
