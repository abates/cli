package cli

import (
	"strings"
	"testing"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		desc       string
		input      string
		accept     []string
		wantOutput string
		wantResp   string
	}{
		{"good input", "y\n", []string{"Y"}, "", "y"},
		{"bad input", "n\ny\n", []string{"Y"}, "Invalid input\n", "y"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			reader := strings.NewReader(test.input)
			writer := &strings.Builder{}
			gotResp := Query(reader, writer, "", test.accept...)
			gotOutput := writer.String()

			if test.wantOutput != gotOutput {
				t.Errorf("want output %q got %q", test.wantOutput, gotOutput)
			}

			if test.wantResp != gotResp {
				t.Errorf("want resp %q got %q", test.wantResp, gotResp)
			}
		})
	}
}
