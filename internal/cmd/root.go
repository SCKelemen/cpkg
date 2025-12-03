package cmd

import (
	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/clix/ext/help"
	"github.com/SCKelemen/clix/ext/version"
)

// Version is set at build time via ldflags.
// Example: go build -ldflags "-X github.com/SCKelemen/cpkg/internal/cmd.Version=1.1.0"
var Version = "dev"

// Commit is set at build time via ldflags.
// Example: go build -ldflags "-X github.com/SCKelemen/cpkg/internal/cmd.Commit=$(git rev-parse --short HEAD)"
var Commit = ""

// Date is set at build time via ldflags.
// Example: go build -ldflags "-X github.com/SCKelemen/cpkg/internal/cmd.Date=$(date -u +%Y-%m-%d)"
var Date = ""

func NewApp() *clix.App {
	app := clix.NewApp("cpkg",
		clix.WithAppDescription("Source-only package manager for C"),
	)

	app.Root = clix.NewGroup("cpkg", "Source-only package manager for C",
		initCmd,
		addCmd,
		tidyCmd,
		syncCmd,
		upgradeCmd,
		vendorCmd,
		statusCmd,
		listCmd,
		explainCmd,
		outdatedCmd,
		buildCmd,
		testCmd,
		graphCmd,
	)

	// Add help extension for "cpkg help [command]"
	app.AddExtension(help.Extension{})

	// Add version extension for "cpkg version" and "cpkg --version"
	app.AddExtension(version.Extension{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	})

	return app
}
