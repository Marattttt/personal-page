package runtime

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"time"

	"github.com/Marattttt/personal-page-libs/userenv"
	"github.com/stretchr/testify/assert"
)

func TestHelloWorld(t *testing.T) {
	const code = `package main
import "fmt"

func main() {
	fmt.Println("Hello world")
}`

	var (
		env = userenv.SameUserEnv{}
		lck = &sync.Mutex{}
		dir = "/tmp/gorunner/test/"

		r = NewRuntime(lck, dir, env)

		expect = RunResult{Stdout: []byte("Hello world\n"), Stderr: nil, ExitCode: 0}
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	res, err := r.Run(ctx, code)

	if assert.NoError(t, err, "A system error happened") {
		assert.Equal(t, res.Stdout, expect.Stdout, "Should produce same stdout")
		assert.Equal(t, res.Stderr, expect.Stderr, "Should produce same stderr")
		assert.Equal(t, res.ExitCode, expect.ExitCode, "Should produce same exit code")
	}
}

func TestCouldNotCompile(t *testing.T) {
	const code = `invalid code`

	var (
		env = userenv.SameUserEnv{}
		lck = &sync.Mutex{}
		dir = "/tmp/gorunner/test/"

		r = NewRuntime(lck, dir, env)
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	res, err := r.Run(ctx, code)

	if assert.NoError(t, err, "A system error happened") {
		assert.Equal(t, string(res.Stdout), "", "Nothing in stdout")
		// go compiler always exits with 1 when invalid code is passed
		assert.Equal(t, res.ExitCode, 1, "Should exit with 1")

		// Strip the path prefix that is separated by a single space
		// The +1 removes that space from the output
		//
		// example: invalid.go:1:1: expected 'package', found invalid
		errMsg := string(res.Stderr)[bytes.IndexRune(res.Stderr, ' ')+1:]

		assert.Equal(t, errMsg, "expected 'package', found invalid\n", "Error message from compiler")
	}
}
