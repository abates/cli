package cli

import (
	"fmt"
	"reflect"
	"time"
)

type callback struct {
	reflect.Value
	arguments Arguments
	variables []interface{}
	t         reflect.Type
	inputErr  error
}

func Callback(f interface{}, descriptions ...string) CommandFunc {
	cb := callback{Value: reflect.ValueOf(f), t: reflect.TypeOf(f)}
	cb.process(descriptions...)
	return cb.callback
}

func getError(values []reflect.Value) error {
	if len(values) > 0 {
		if values[len(values)-1].CanInterface() {
			i := values[len(values)-1].Interface()
			if err, ok := i.(error); ok {
				return err
			}
		}
	}
	return nil
}

func (cb *callback) callback(name string, args ...string) ([]string, error) {
	if cb.inputErr != nil {
		return args, cb.inputErr
	}

	err := cb.arguments.Parse(args)
	if err == nil {
		args = cb.arguments.Args()
		values := []reflect.Value{}
		for i, v := range cb.variables {
			if cb.t.In(i).Kind() == reflect.Ptr {
				values = append(values, reflect.ValueOf(v))
			} else {
				values = append(values, reflect.Indirect(reflect.ValueOf(v)))
			}
		}

		err = getError(cb.Call(values))
	}
	return args, err
}

func (cb *callback) addVar(v interface{}) {
	cb.variables = append(cb.variables, v)
}

func (cb *callback) tryValue(arg reflect.Type, description string, wantType reflect.Type, addTo reflect.Value) bool {
	if arg.Kind() != reflect.Ptr {
		return cb.tryValue(reflect.PtrTo(arg), description, wantType, addTo)
	}

	if arg.Implements(wantType) {
		v := reflect.New(arg.Elem())
		addTo.Call([]reflect.Value{v, reflect.ValueOf(description)})
		cb.addVar(v.Interface())
		return true
	}
	return false
}

func (cb *callback) process(descriptions ...string) {
	if cb.Kind() != reflect.Func {
		cb.inputErr = fmt.Errorf("Provided callback is not a function")
		return
	}

	for i := 0; i < cb.t.NumIn(); i++ {
		description := ""
		if i < len(descriptions) {
			description = descriptions[i]
		}
		inArg := cb.t.In(i)
		switch inArg {
		case reflect.TypeOf(false):
			cb.addVar(cb.arguments.Bool(description))
		case reflect.TypeOf(time.Duration(0)):
			cb.addVar(cb.arguments.Duration(description))
		case reflect.TypeOf(float64(0)):
			cb.addVar(cb.arguments.Float64(description))
		case reflect.TypeOf(int(0)):
			cb.addVar(cb.arguments.Int(description))
		case reflect.TypeOf(int64(0)):
			cb.addVar(cb.arguments.Int64(description))
		case reflect.TypeOf(""):
			cb.addVar(cb.arguments.String(description))
		case reflect.TypeOf(uint(0)):
			cb.addVar(cb.arguments.Uint(description))
		case reflect.TypeOf(uint64(0)):
			cb.addVar(cb.arguments.Uint64(description))
		default:
			if !cb.tryValue(inArg, description, reflect.TypeOf((*Value)(nil)).Elem(), reflect.ValueOf(cb.arguments.Var)) {
				if !cb.tryValue(inArg, description, reflect.TypeOf((*SliceValue)(nil)).Elem(), reflect.ValueOf(cb.arguments.VarSlice)) {
					cb.inputErr = fmt.Errorf("%v must implement either Value or ValueSlice interfaces", inArg)
				}
			}
		}
	}
}
