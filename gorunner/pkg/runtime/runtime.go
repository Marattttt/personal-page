package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Resut of running code
type RunResult struct {
	Stdout []byte
	Stderr []byte

	TimeTook time.Duration
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

// TODO: add support for extra files, e.g. through variable arguments
func (r Runtime) Run(ctx context.Context, code string) (*RunResult, error) {
	r.lck.Lock()
	defer r.lck.Unlock()

	slog.Info("Started preparing runtime environment")

	start := time.Now()

	if err := r.env.Prepare(ctx, r.root); err != nil {
		return nil, fmt.Errorf("preparing: %w", err)
	}

	cmd, err := r.env.Login(ctx)
	if err != nil {
		return nil, fmt.Errorf("logging in: %w", err)
	}

	slog.Info("Finished preparing runtime environment", slog.Duration("timeTook", time.Now().Sub(start)))

	// Read all data from outputs, while the command is running
	stdout, stderr, err := getOutPipes(cmd)
	if err != nil {
		return nil, err
	}
	defer stdout.Close()
	defer stderr.Close()

	// Prepare a main.go file for running
	if err := writeMain(r.root, code); err != nil {
		return nil, fmt.Errorf("writing main.go: %w", err)
	}

	mainPath, err := filepath.Abs(path.Join(r.root, "main.go"))
	if err != nil {
		return nil, fmt.Errorf("getting abs path for just written main.go: %w", err)
	}

	stdin := strings.NewReader("go run " + mainPath)
	cmd.Stdin = stdin

	// Command start time
	startedAt := time.Now()

	var (
		// Final buffers to write output to
		finStdout bytes.Buffer
		// Final buffers to write output to
		finStderr bytes.Buffer

		// For parallel reading of outpus during execution
		readWg sync.WaitGroup
	)

	cmd.Start()

	readWg.Add(1)
	go func() {
		defer readWg.Done()
		b, err := io.ReadAll(stdout)
		if err != nil {
			slog.Error("Error reading stdout", slog.String("err", err.Error()))
		} else {
			finStdout.Write(b)
		}
	}()

	readWg.Add(1)
	go func() {
		defer readWg.Done()
		b, err := io.ReadAll(stderr)
		if err != nil {
			slog.Error("Error reading stderr", slog.String("err", err.Error()))
		} else {
			finStderr.Write(b)
		}
	}()

	// An error other than exiterror indicates a system error
	if err := cmd.Wait(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			slog.Warn("Non-zero exitcode running user code", slog.Int("code", exitErr.ExitCode()))
		} else {
			return nil, fmt.Errorf("running cmd: %w", err)
		}
	}

	// Finish reading
	readWg.Wait()

	return &RunResult{
			Stderr:   finStderr.Bytes(),
			Stdout:   finStdout.Bytes(),
			ExitCode: cmd.ProcessState.ExitCode(),
			TimeTook: time.Now().Sub(startedAt)},
		nil

}

func getOutPipes(cmd *exec.Cmd) (io.ReadCloser, io.ReadCloser, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("getting stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("getting stderr: %w", err)
	}

	return stdout, stderr, nil
}

func writeMain(root string, code string) error {
	f, err := os.Create(filepath.Join(root, "main.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write([]byte(code))

	return nil
}
