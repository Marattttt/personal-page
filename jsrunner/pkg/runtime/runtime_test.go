package runtime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Marattttt/personal-page-libs/userenv"
	"github.com/stretchr/testify/assert"
)

func TestHelloWorld(t *testing.T) {
	const code = `console.log('Hello world')`
	var (
		env = userenv.SameUserEnv{}
		lck = &sync.Mutex{}
		dir = "/tmp/jsrunner/test/"

		r = NewRuntime(lck, dir, env)

		expect = RunResult{Stdout: []byte("Hello world\n"), Stderr: nil, ExitCode: 0}
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	res, err := r.Run(ctx, code)

	if assert.NoError(t, err, "A system error happened") {
		assert.Equal(t, string(expect.Stdout), string(res.Stdout), "Should produce same stdout")
		assert.Equal(t, string(expect.Stderr), string(res.Stderr), "Should produce same stderr")
		assert.Equal(t, expect.ExitCode, res.ExitCode, "Should produce same exit code")
	}
}
