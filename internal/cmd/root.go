package cmd

import (
	"github.com/SCKelemen/clix"
)

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

	return app
}
