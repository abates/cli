package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Value interface {
	String() string
	Set(string) error
}

type SliceValue interface {
	String() string
	Set([]string) error
}

type boolValue bool

func (b *boolValue) Get() interface{} { return bool(*b) }
func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errParse
	}
	*b = boolValue(v)
	return err
}

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

type intValue int

func (i *intValue) Get() interface{} { return int(*i) }

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = intValue(v)
	return err
}

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

type int64Value int64

func (i *int64Value) Get() interface{} { return int64(*i) }
func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = int64Value(v)
	return err
}

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }

type uintValue uint

func (i *uintValue) Get() interface{} { return uint(*i) }
func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = uintValue(v)
	return err
}

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

type uint64Value uint64

func (i *uint64Value) Get() interface{} { return uint64(*i) }
func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

type stringValue string

func (s *stringValue) Get() interface{} { return string(*s) }
func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) String() string { return string(*s) }

type float64Value float64

func (f *float64Value) Get() interface{} { return float64(*f) }
func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		err = numError(err)
	}
	*f = float64Value(v)
	return err
}

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

type durationValue time.Duration

func (d *durationValue) Get() interface{} { return time.Duration(*d) }
func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		err = errParse
	}
	*d = durationValue(v)
	return err
}

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

type Arguments struct {
	input []string
	args  []*argument
}

type argument struct {
	value interface{}
	desc  string
}

func (args *Arguments) Bool(desc string) *bool {
	p := new(bool)
	args.BoolVar(p, desc)
	return p
}

func (args *Arguments) BoolVar(p *bool, desc string) { args.Var((*boolValue)(p), desc) }

func (args *Arguments) Duration(desc string) *time.Duration {
	p := new(time.Duration)
	args.DurationVar(p, desc)
	return p
}

func (args *Arguments) DurationVar(p *time.Duration, desc string) {
	args.Var((*durationValue)(p), desc)
}

func (args *Arguments) Float64(desc string) *float64 {
	p := new(float64)
	args.Float64Var(p, desc)
	return p
}

func (args *Arguments) Float64Var(p *float64, desc string) { args.Var((*float64Value)(p), desc) }

func (args *Arguments) Int(desc string) *int {
	p := new(int)
	args.IntVar(p, desc)
	return p
}

func (args *Arguments) IntVar(p *int, desc string) { args.Var((*intValue)(p), desc) }

func (args *Arguments) Int64(desc string) *int64 {
	p := new(int64)
	args.Int64Var(p, desc)
	return p
}

func (args *Arguments) Int64Var(p *int64, desc string) { args.Var((*int64Value)(p), desc) }

func (args *Arguments) String(desc string) *string {
	p := new(string)
	args.StringVar(p, desc)
	return p
}

func (args *Arguments) StringVar(p *string, desc string) { args.Var((*stringValue)(p), desc) }

func (args *Arguments) Uint(desc string) *uint {
	p := new(uint)
	args.UintVar(p, desc)
	return p
}

func (args *Arguments) UintVar(p *uint, desc string) { args.Var((*uintValue)(p), desc) }

func (args *Arguments) Uint64(desc string) *uint64 {
	p := new(uint64)
	args.Uint64Var(p, desc)
	return p
}

func (args *Arguments) Uint64Var(p *uint64, desc string) { args.Var((*uint64Value)(p), desc) }

func (args *Arguments) Var(value Value, desc string) {
	args.args = append(args.args, &argument{value, desc})
}

func (args *Arguments) VarSlice(value SliceValue, desc string) {
	args.args = append(args.args, &argument{value, desc})
}

func (args *Arguments) Len() int { return len(args.args) }

func (args *Arguments) Args() []string {
	/*if len(args.input) > len(args.args) {
		return args.input[len(args.args):]
	}
	return []string{}*/
	return args.input
}

func (args *Arguments) Parse(input []string) error {
	if len(input) < len(args.args) {
		return errNumArguments
	}
	args.input = []string{}
	for i, arg := range args.args {
		if s, ok := arg.value.(SliceValue); ok {
			return s.Set(input[i:len(input)])
		} else if s, ok := arg.value.(Value); ok {
			err := s.Set(input[i])
			if err != nil {
				return err
			}
		} else {
			panic(fmt.Sprintf("huh? value should have been Value or SliceValue got %T", arg.value))
		}
	}
	args.input = input[len(args.args):]
	return nil
}

func (args *Arguments) Usage(writer io.Writer) {
	desc := []string{}
	for _, arg := range args.args {
		desc = append(desc, arg.desc)
	}
	writer.Write([]byte(strings.Join(desc, " ")))
}
