package cli

import (
	"fmt"
	"reflect"
	"time"
)

func Callback(cb interface{}, descriptions ...string) CommandFunc {
	var inputErr error
	arguments := Arguments{}
	variables := []interface{}{}

	v := reflect.ValueOf(cb)
	t := reflect.TypeOf(cb)

	if v.Kind() == reflect.Func {
		for i := 0; i < t.NumIn(); i++ {
			description := ""
			if i < len(descriptions) {
				description = descriptions[i]
			}
			inArg := t.In(i)
			switch inArg {
			case reflect.TypeOf(false):
				var b bool
				arguments.Bool(&b, description)
				variables = append(variables, &b)
			case reflect.TypeOf(time.Duration(0)):
				var d time.Duration
				arguments.Duration(&d, description)
				variables = append(variables, &d)
			case reflect.TypeOf(float64(0)):
				var f float64
				arguments.Float64(&f, description)
				variables = append(variables, &f)
			case reflect.TypeOf(int(0)):
				var i int
				arguments.Int(&i, description)
				variables = append(variables, &i)
			case reflect.TypeOf(int64(0)):
				var i int64
				arguments.Int64(&i, description)
				variables = append(variables, &i)
			case reflect.TypeOf(""):
				var s string
				arguments.String(&s, description)
				variables = append(variables, &s)
			case reflect.TypeOf(uint(0)):
				var u uint
				arguments.Uint(&u, description)
				variables = append(variables, &u)
			case reflect.TypeOf(uint64(0)):
				var u uint64
				arguments.Uint64(&u, description)
				variables = append(variables, &u)
			default:
				if inArg.Implements(reflect.TypeOf((*Value)(nil)).Elem()) {
					if inArg.Kind() == reflect.Ptr {
						u := reflect.New(inArg.Elem()).Interface().(Value)
						arguments.Var(u, description)
						variables = append(variables, u)
					} else {
						inputErr = fmt.Errorf("%v argument must be a pointer to a type implementing Value", inArg)
					}
				} else if inArg.Implements(reflect.TypeOf((*SliceValue)(nil)).Elem()) {
					if inArg.Kind() == reflect.Ptr {
						u := reflect.New(inArg.Elem()).Interface().(SliceValue)
						arguments.VarSlice(u, description)
						variables = append(variables, u)
					} else {
						inputErr = fmt.Errorf("%v argument must be a pointer to a type implementing Value", inArg)
					}
				} else {
					inputErr = fmt.Errorf("Type %s does not implement Value interface", inArg)
				}
			}
		}
	} else {
		inputErr = fmt.Errorf("Provided callback is not a function")
	}

	return func(name string, args ...string) ([]string, error) {
		if inputErr != nil {
			return args, inputErr
		}

		err := arguments.Parse(args)
		if err == nil {
			args = arguments.Args()
			values := []reflect.Value{}
			for i, v := range variables {
				if t.In(i).Kind() == reflect.Ptr {
					values = append(values, reflect.ValueOf(v))
				} else {
					values = append(values, reflect.Indirect(reflect.ValueOf(v)))
				}
			}

			ret := v.Call(values)
			if len(ret) > 0 {
				if ret[len(ret)-1].CanInterface() {
					i := ret[len(ret)-1].Interface()
					if err, ok := i.(error); ok {
						return args, err
					}
				}
			}
		}
		return args, err
	}
}
