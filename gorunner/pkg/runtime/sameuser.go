package runtime

import (
	"context"
	"os/exec"
)

// Provides runtime environment for the same user the application is running as
type SameUserEnv struct{}

// Starts a bash session
func (SameUserEnv) Login(ctx context.Context) (*exec.Cmd, error) {
	// Bash can take code from stdin
	cmd := exec.CommandContext(ctx, "bash")
	return cmd, nil
}
