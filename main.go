package main

import "github.com/mdy/melody/cli"

// Populated from Makefile
var version = "<unknown>"

// Let's go!
func main() {
	cli.Main(version)
}
