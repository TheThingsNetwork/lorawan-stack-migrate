package awsiot

import (
	"go.thethings.network/lorawan-stack-migrate/pkg/commands"
	_ "go.thethings.network/lorawan-stack-migrate/pkg/source/awsiot"
)

const sourceName = "awsiot"

// Command represents the awsiot source.
var Command = commands.Source(sourceName,
	"Export devices from AWS IoT",
	commands.WithSourceOptions(
		commands.WithAliases([]string{"aws-iot"}),
	),
)
