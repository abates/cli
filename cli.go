package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func Query(reader io.Reader, writer io.Writer, message string, acceptable ...string) (resp string) {
	accept := make(map[string]bool, len(acceptable))
	for _, a := range acceptable {
		accept[strings.ToLower(strings.TrimSpace(a))] = true
	}

	buf := bufio.NewReader(reader)
	for {
		fmt.Fprint(writer, message)
		resp, _ = buf.ReadString('\n')
		resp = strings.ToLower(strings.TrimSpace(resp))
		if accept[resp] {
			break
		}
		fmt.Fprintf(writer, "Invalid input\n")
	}
	return resp
}
