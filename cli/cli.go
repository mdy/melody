package cli

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"path"
)

// Melody CLI entry point
func Main(version string) {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Package based dependency manager"
	app.Version = version
	app.Author = ""
	app.Email = ""

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug mode",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:  "log-level, l",
			Value: "fatal",
			Usage: "log level",
		},
	}

	// Configure based on global CLI flags
	app.Before = func(c *cli.Context) error {
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			log.Fatalf(err.Error())
		}

		log.SetOutput(os.Stderr)
		log.SetLevel(level)
		return nil
	}

	// See below...
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		if s, ok := err.(fmt.Stringer); ok {
			log.Error(err)
			fmt.Println(s.String())
		} else {
			log.Fatal(err)
		}
	}
}

var (
	commands = []cli.Command{
		{
			Name:   "init",
			Usage:  "Start a project",
			Action: initProject,
		}, {
			Name:      "install",
			ShortName: "i",
			Usage:     "Install dependencies",
			Action:    install,
		}, {
			Name:      "update",
			ShortName: "u",
			Usage:     "Update dependencies",
			Action:    update,
		}, {
			Name:      "outdated",
			ShortName: "o",
			Usage:     "Show outdated dependencies",
			Action:    outdated,
		}, {
			Name:   "lint",
			Usage:  "Validate configuration",
			Action: lint,
		}, {
			Name:   "list",
			Usage:  "List all dependencies",
			Action: list,
		}, {
			Name:   "get",
			Usage:  "Download and install a package",
			Action: getPackages,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "d",
					Usage: "Download only",
				},
				cli.BoolFlag{
					Name:  "u",
					Usage: "Force update",
				},
			},
		}, {
			Name:   "info",
			Usage:  "Show project info",
			Action: info,
		},
	}
)
