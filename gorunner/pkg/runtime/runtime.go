package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Resut of running code
type RunResult struct {
	Stdout []byte `json:"stdout"`
	Stderr []byte `json:"stderr"`

	TimeTook time.Duration `json:"timeTook"`
	ExitCode int           `json:"exitCode"`
}

// Provides methods for managing a user-specific environment
//
// While it is ok to not switch users during debugging, executing
// arbitrary code in a production environment should be done with
// necessary restricions
type SafeEnvProvider interface {
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

	if err := r.InitEnvironment(ctx, code); err != nil {
		return nil, fmt.Errorf("creating environment: %w", err)
	}

	cmd, err := r.env.Login(ctx)
	if err != nil {
		return nil, fmt.Errorf("logging in: %w", err)
	}

	stdin := strings.NewReader("go run " + r.root + "/main.go")
	cmd.Stdin = stdin

	// Read all data from outputs, while the command is still running
	stdout, stderr, err := getOutPipes(cmd)
	if err != nil {
		return nil, err
	}

	var (
		// Final buffers to write output to
		finStdout bytes.Buffer
		// Final buffers to write output to
		finStderr bytes.Buffer

		// For parallel reading of outpus during execution
		readWg sync.WaitGroup
	)

	slog.Info("Started execution", slog.String("cmd", cmd.String()))

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

	// Command start time
	start := time.Now()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting shell: %w", err)
	}

	// Finish reading before comamnd completion, cannot be done other way round
	readWg.Wait()

	// An error other than exiterror indicates a system error
	if err := cmd.Wait(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			slog.Warn("Non-zero exitcode running user code", slog.Int("code", exitErr.ExitCode()))
		} else {
			return nil, fmt.Errorf("running cmd: %w", err)
		}
	}

	res := &RunResult{
		Stderr:   finStderr.Bytes(),
		Stdout:   finStdout.Bytes(),
		ExitCode: cmd.ProcessState.ExitCode(),
		TimeTook: time.Now().Sub(start),
	}

	slog.Debug("Finished running user code", slog.Any("result", res))
	slog.Info("Finished running user code", slog.Any("result", res), slog.Duration("timeTook", time.Now().Sub(start)))

	return res, nil
}

// Create a clean directory with go.mod and main.go files
func (r Runtime) InitEnvironment(ctx context.Context, code string) error {
	slog.Info("Started preparing runtime environment")

	start := time.Now()

	if err := clearDirectory(r.root); err != nil {
		return fmt.Errorf("preparing root dir at %s: %w", r.root, err)
	}

	if err := goModInit(ctx, r.root); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}

	// Prepare a main.go file for running
	if err := writeMain(r.root, code); err != nil {
		return fmt.Errorf("writing main.go: %w", err)
	}

	slog.Info("Finished preparing runtime environment", slog.Duration("timeTook", time.Now().Sub(start)))

	return nil
}

// Cleans a directory with all its contents and recreates it with 0777 perms
func clearDirectory(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("removing: %w", err)
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("creating dir: %w", err)
	}
	return nil
}

// Get output pipes fro a comand (stdout, stderr)
//
// Pipesdo usually do not need to be closed manually, as they are autmoatically closed
// when the comand exits
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

// Create go mod file in a directory
func goModInit(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "init", "gorunner")
	slog.Info("Creating go mod", slog.String("dir", dir))

	if out, err := cmd.CombinedOutput(); err != nil {
		// if a go.mod is already present in a directory, go mod init exits with a non-zero
		if !strings.Contains(string(out), "go.mod already exists") {
			slog.Error("Failed to create go mod", slog.String("dir", dir), slog.String("output", string(out)))
			return err
		}
	} else {
		slog.Debug("Created go mod", slog.String("output", string(out)))
	}

	return nil
}

func writeMain(root string, code string) error {
	f, err := os.Create(filepath.Join(root, "main.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write([]byte(code))

	slog.Debug("Wrote main.go", slog.String("content", code))

	return nil
}
