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
	"strings"
	"sync"
	"time"
)

type RunResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	TimeTook time.Duration
}

// Provides methods for managing a user-specific environment
//
// While it is ok to not switch users during debugging, executing
// arbitrary code in a production environment should be done with
// necessary restricions
type EnvProvider interface {
	// Provide a logged in cmd for code execution and compilation
	Login(ctx context.Context) (*exec.Cmd, error)
}

type Runtime struct {
	// Lock during execution to prevent process collisions
	lck  sync.Locker
	root string
	env  EnvProvider
}

func NewRuntime(lck sync.Locker, runDir string, provider EnvProvider) Runtime {
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

	if err := prepare(r.root, code); err != nil {
		return nil, fmt.Errorf("preparing: %w", err)
	}

	cmd, err := r.env.Login(ctx)
	if err != nil {
		return nil, fmt.Errorf("logging in: %w", err)
	}

	nodePath, err := nodeAbsPath(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting node path: %w", err)
	}

	stdinStr := *nodePath + " " + path.Join(r.root, "index.js")
	stdin := strings.NewReader(stdinStr)
	cmd.Stdin = stdin
	slog.Info("Prepared stdin for shell", slog.String("in", stdinStr))

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
		Stdout:   finStdout.Bytes(),
		Stderr:   finStderr.Bytes(),
		ExitCode: cmd.ProcessState.ExitCode(),
		TimeTook: time.Now().Sub(start),
	}

	slog.Info("Finished running user code", slog.Any("result", res), slog.Duration("timeTook", time.Now().Sub(start)))

	return res, nil
}

func prepare(dir string, code string) error {
	if err := clearDirectory(dir); err != nil {
		return fmt.Errorf("clearing: %w", err)
	}

	if err := writeMain(dir, code); err != nil {
		return fmt.Errorf("creating main: %w", err)
	}

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

// Finds the absolute path to  go executable
func nodeAbsPath(ctx context.Context) (*string, error) {
	cmd := exec.CommandContext(ctx, "which", "node")
	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("Could not get path of the node executable")
		return nil, fmt.Errorf("running which node: %w", err)
	}

	s := strings.TrimSpace(string(out))
	return &s, nil
}

func writeMain(root string, code string) error {
	f, err := os.Create(path.Join(root, "index.js"))
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write([]byte(code))

	slog.Debug("Wrote main.go", slog.String("content", code))

	return nil
}
