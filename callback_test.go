package cli

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestCallbackDesc(t *testing.T) {
	f := func(a, b, c int) {}
	cb := callback{Value: reflect.ValueOf(f), t: reflect.TypeOf(f)}
	cb.process("<a>", "<b>")

	for i, want := range []string{"<a>", "<b>", ""} {
		if cb.arguments.args[i].desc != want {
			t.Errorf("Wanted description %q got %q", want, cb.arguments.args[i].desc)
		}
	}
}

func TestCallback(t *testing.T) {
	tests := []struct {
		desc    string
		cb      interface{}
		input   []string
		wantErr string
	}{
		{"bool", func(b bool) error { return fmt.Errorf("%v", b) }, []string{"true"}, "true"},
		{"bool (parse error)", func(b bool) error { return fmt.Errorf("%v", b) }, []string{"yo"}, "parse error"},
		{"duration", func(d time.Duration) error { return fmt.Errorf("%v", d) }, []string{"1s"}, "1s"},
		{"float64", func(f float64) error { return fmt.Errorf("%v", f) }, []string{"1.234"}, "1.234"},
		{"int", func(i int) error { return fmt.Errorf("%v", i) }, []string{"4234"}, "4234"},
		{"int64", func(i int64) error { return fmt.Errorf("%v", i) }, []string{"934"}, "934"},
		{"string", func(s string) error { return fmt.Errorf("%v", s) }, []string{"foo"}, "foo"},
		{"uint", func(u uint) error { return fmt.Errorf("%v", u) }, []string{"1234"}, "1234"},
		{"uint64", func(u uint64) error { return fmt.Errorf("%v", u) }, []string{"1234"}, "1234"},
		{"value", func(b *boolValue) error { return fmt.Errorf("%v", b.String()) }, []string{"true"}, "true"},
		{"bool/no pointer", func(b boolValue) error { return fmt.Errorf("%v", b.String()) }, []string{"true"}, "true"},
		{"no func", "hello world", []string{"true"}, "Provided callback is not a function"},
		{"int slice", func(i *intSlice) error { return fmt.Errorf("%v", i.String()) }, []string{"1", "2", "3", "4", "5"}, "1,2,3,4,5"},
		{"two values", func(a, b int) error { return fmt.Errorf("%d %d", a, b) }, []string{"1", "2"}, "1 2"},
		{"two expected one received", func(a, b int) error { return fmt.Errorf("%d %d", a, b) }, []string{"1"}, "Invalid Usage not enough arguments given"},
		{"non-value argument", func(a time.Time) error { return nil }, []string{"1"}, "time.Time must implement either Value or ValueSlice interfaces"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cmd := New("", ErrorHandlingOption(ContinueOnError))
			cmd.Callback = Callback(test.cb)

			_, err := cmd.Run(test.input)
			if err == nil {
				t.Errorf("Expected error")
			} else if err.Error() != test.wantErr {
				t.Errorf("Wanted error %s got %s", test.wantErr, err.Error())
			}
		})
	}
}
