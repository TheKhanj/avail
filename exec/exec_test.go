package exec

import (
	"log"
	"os"
	"testing"
)

func TestExec(t *testing.T) {
	shlex := `sh -c 'for i in $(seq 3); do echo "hello $i"; sleep 1; done'`
	e, err := New(
		WithLogger(log.New(os.Stderr, "log-prefix", 0)),
		WithShlex(shlex),
	)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = e.Run()
	if err != nil {
		t.Fatal(err)
	}
}
