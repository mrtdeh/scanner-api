package engines

import (
	"context"
	"os/exec"
)

type CommandExecutor interface {
	ExecCommand(ctx context.Context, name string, args ...string) ([]byte, error)
}

type DefaultCommandExcutor struct{}

func (d *DefaultCommandExcutor) ExecCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
