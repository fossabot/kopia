package cli

import (
	"context"

	"github.com/kopia/kopia/internal/serverapi"
)

var (
	serverResumeCommand = serverCommands.Command("resume", "Resume the scheduled snapshots for one or more sources")
)

func init() {
	serverResumeCommand.Action(serverAction(runServerResume))
}

func runServerResume(ctx context.Context, cli *serverapi.Client) error {
	return cli.Post("sources/resume", &serverapi.Empty{}, &serverapi.Empty{})
}
