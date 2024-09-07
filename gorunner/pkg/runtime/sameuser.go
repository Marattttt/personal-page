package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// Provides runtime environment for the same user the application is running as
type SameUserEnv struct{}

func (SameUserEnv) Prepare(ctx context.Context, root string) error {
	if err := os.RemoveAll(root); err != nil {
		return fmt.Errorf("removing root path: %w", err)
	}
	if err := os.MkdirAll(root, 777); err != nil {
		return fmt.Errorf("creating root dir: %w", err)
	}

	cmd := exec.CommandContext(ctx, "go", "mod", "init", "usercode")
	cmd.Dir = root

	if err := cmd.Run(); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("executing go mod init: %w", err)
		}
	}

	slog.Info("Prepared runtime env for current user", slog.String("path", root))
	return nil
}

// Starts a bash session
func (SameUserEnv) Login(ctx context.Context) (*exec.Cmd, error) {
	// Bash can take code from stdin
	cmd := exec.CommandContext(ctx, "bash")
	return cmd, nil
}
