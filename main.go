package main

import "github.com/melody-sh/melody/cli"

// Populated from Makefile
var version = "<unknown>"

// Let's go!
func main() {
	cli.Main(version)
}
