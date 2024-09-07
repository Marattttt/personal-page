package runtime

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

// Resut of running code
type RunResult struct {
	Stdout []byte
	Stderr []byte

	Duration time.Duration
	ExitCode int
}

// Provides methods for managing a user-specific environment
//
// While it is ok to not switch users during debugging, executing
// arbitrary code in a production environment should be done with
// necessary restricions
type SafeEnvProvider interface {
	// Prepare a root directory for code execution
	Prepare(ctx context.Context, root string) error
	// Provide a logged in cmd for code execution and compilation
	Login(ctx context.Context) (*exec.Cmd, error)
}

type Runtime struct {
	// Lock during execution to prevent process collisions
	lck  sync.Locker
	root string
	env  SafeEnvProvider
}

func NewRuntime(lck sync.Locker, runDir string, provider SafeEnvProvider) Runtime {
	return Runtime{
		lck:  lck,
		env:  provider,
		root: runDir,
	}
}
