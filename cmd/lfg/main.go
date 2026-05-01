// Command lfg is the TUI bootstrap CLI entrypoint.
//
// Thin shim around internal/cli — all subcommand wiring lives there
// so the package can be imported and exercised from tests.
package main

import "github.com/ptmaroct/lfg/internal/cli"

func main() { cli.Execute() }
