package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
)

type DiffUserEnv struct {
	user string

	// Whether the user was checked for existance
	userChecked bool
}

func NewDiffUserEnv(user string, pass *string) (*DiffUserEnv, error) {
	const unamereg = `^[a-z_][a-z0-9_]{0,30}$`
	reg := regexp.MustCompile(unamereg)

	if !reg.Match([]byte(user)) {
		return nil, fmt.Errorf("username %s is not supported", user)
	}

	if pass != nil {
		slog.Warn("Password based login is not supported")
	}

	return &DiffUserEnv{
		user: user,
	}, nil
}

func (d DiffUserEnv) Login(ctx context.Context) (*exec.Cmd, error) {
	if !d.userChecked {
		checkUser := exec.CommandContext(ctx, "id", d.user)
		if err := checkUser.Run(); err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				slog.Warn("Could not check user id", slog.String("stderr", string(err.Stderr)))
			}

			return nil, fmt.Errorf("checking id of user '%s': %w", d.user, err)
		}
	}

	d.userChecked = true

	// Assumes running as the root user
	return exec.Command("sudo", "-u", d.user, "sh"), nil
}
