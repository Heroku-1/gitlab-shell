package main

import (
	"fmt"
	"os"

	"gitlab.com/gitlab-org/gitlab-shell/internal/command"
	"gitlab.com/gitlab-org/gitlab-shell/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/internal/executable"
	"gitlab.com/gitlab-org/gitlab-shell/internal/logger"
	"gitlab.com/gitlab-org/gitlab-shell/internal/sshenv"
)

func main() {
	readWriter := &readwriter.ReadWriter{
		Out:    os.Stdout,
		In:     os.Stdin,
		ErrOut: os.Stderr,
	}

	executable, err := executable.New(executable.Healthcheck, false)
	if err != nil {
		fmt.Fprintln(readWriter.ErrOut, "Failed to determine executable, exiting")
		os.Exit(1)
	}

	config, err := config.NewFromDirExternal(executable.RootDir)
	if err != nil {
		fmt.Fprintln(readWriter.ErrOut, "Failed to read config, exiting")
		os.Exit(1)
	}

	logCloser := logger.Configure(config)
	defer logCloser.Close()

	cmd, err := command.New(executable, os.Args[1:], sshenv.Env{}, config, readWriter)
	if err != nil {
		fmt.Fprintf(readWriter.ErrOut, "%v\n", err)
		os.Exit(1)
	}

	ctx, finished := command.Setup(executable.Name, config)
	defer finished()

	if err = cmd.Execute(ctx); err != nil {
		fmt.Fprintf(readWriter.ErrOut, "%v\n", err)
		os.Exit(1)
	}
}
