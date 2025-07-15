package app

import "github.com/lwm-galactic/app-cli/cli"

// CliOptions configuration options for reading parameters from the command line.
type CliOptions interface {
	Flags() (fss cli.NamedFlagSets)
	Validate() []error
}

// CompletableOptions abstracts options which can be completed.
type CompletableOptions interface {
	Complete() error
}

// PrintableOptions abstracts options which can be printed.
type PrintableOptions interface {
	String() string
}
