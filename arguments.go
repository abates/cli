package cli

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

var (
	errParse        = errors.New("parse error")
	errRange        = errors.New("value out of range")
	errNumArguments = errors.New("not enough arguments given")
)

func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}
	if ne.Err == strconv.ErrSyntax {
		return errParse
	}
	if ne.Err == strconv.ErrRange {
		return errRange
	}
	return ne.Err
}

type Value interface {
	String() string
	Set(string) error
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
	value Value
	desc  string
	slice bool
}

func (args *Arguments) Bool(p *bool, desc string)              { args.Var((*boolValue)(p), desc) }
func (args *Arguments) Duration(p *time.Duration, desc string) { args.Var((*durationValue)(p), desc) }
func (args *Arguments) Float64(p *float64, desc string)        { args.Var((*float64Value)(p), desc) }
func (args *Arguments) Int(p *int, desc string)                { args.Var((*intValue)(p), desc) }
func (args *Arguments) Int64(p *int64, desc string)            { args.Var((*int64Value)(p), desc) }
func (args *Arguments) String(p *string, desc string)          { args.Var((*stringValue)(p), desc) }
func (args *Arguments) Uint(p *uint, desc string)              { args.Var((*uintValue)(p), desc) }
func (args *Arguments) Uint64(p *uint64, desc string)          { args.Var((*uint64Value)(p), desc) }

func (args *Arguments) Var(value Value, desc string) {
	args.args = append(args.args, &argument{value, desc, false})
}

func (args *Arguments) VarSlice(value Value, desc string) {
	args.args = append(args.args, &argument{value, desc, true})
}

func (args *Arguments) Len() int       { return len(args.args) }
func (args *Arguments) Args() []string { return args.input[len(args.args):] }

func (args *Arguments) Parse(input []string) error {
	if len(input) < len(args.args) {
		return errNumArguments
	}
	args.input = input
	for i, arg := range args.args {
		if arg.slice {
			for ; i < len(args.input); i++ {
				err := arg.value.Set(input[i])
				if err != nil {
					return err
				}
			}
			return nil
		} else {
			err := arg.value.Set(input[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (args *Arguments) Usage(writer io.Writer) {
	desc := []string{}
	for _, arg := range args.args {
		desc = append(desc, arg.desc)
	}
	writer.Write([]byte(strings.Join(desc, " ")))
}
