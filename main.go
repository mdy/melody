package main

import "github.com/melodysh/melody/cli"

// Populated from Makefile
var version = "<unknown>"

// Let's go!
func main() {
	cli.Main(version)
}
