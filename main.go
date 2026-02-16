/*
Copyright © 2025 Aris Tzermias
*/
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/tzermias/deskctl/cmd"
)

func main() {
	// Create context that cancels on interrupt signals
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	// Pass context to Execute
	cmd.Execute(ctx)
}
