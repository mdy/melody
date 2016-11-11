# Melody - Dependency Manager for Go

[Get started](https://melody.sh/docs/howto) &nbsp;&middot;&nbsp;
[Install Melody](https://melody.sh/docs/howto/install) &nbsp;&middot;&nbsp;
[Command reference](https://melody.sh/docs/commands) &nbsp;&middot;&nbsp;
[Repo cache][melody-api]

[![Version Badge](https://badge.fury.io/mdy/github.com%2Fmdy%2Fmelody.svg)](https://melody.sh/github.com/mdy/melody)
[![Build Status](https://travis-ci.org/mdy/melody.svg?branch=master)](https://travis-ci.org/mdy/melody)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdy/melody)](https://goreportcard.com/report/github.com/mdy/melody)

Melody is a dependency manager with a focus on speed and reliability.  In some instances, it is **up to 3x faster** than `go get -d` and other dependency managers.

Sponsored by [Gemfury](https://gemfury.com), Melody has been used for production builds since July 2016.

## About

Melody is a tool that enables Go developers to manage project's dependencies and ensure fast, consistent, and repeatable builds.  We've adopted the [following principles](#credits-and-inspiration) to make this happen:

**All dependencies are vendored** to prevent multiple projects from clobbering shared repositories in GOPATH.

**Human-friendly config file** explicitly specifies project details and dependencies with corresponding version restrictions.

**Human-readable lock file** to record and track exact revision of each installed repository.  This file is used to deterministically recreate the `vendor` directory.

**Cloud-assisted repository indexing and caching** allows for much faster and more-reliable builds.  [melodyAPI][melody-api] integration makes sure your build is fast, and that a deleted repository or tag does not break future builds.

> Please note that the [melodyAPI][melody-api] cloud cache may be cold during the beta period due to a low traffic.  This may cause slowness during your installs, but it will get faster as our userbase grows.

Melody requires Go 1.6+. Although it may work with GO15VENDOREXPERIMENT flag, Go 1.5 is not supported.

## Documentation

[Melody documentation](https://melody.sh/docs/) is now part of the [Melody website](https://melody.sh/).  Below are a few quick links to get you started:

- [Installing Melody](https://melody.sh/docs/howto/install)
- [Start a new Go project](https://melody.sh/docs/howto/usage)
- [Command reference](https://melody.sh/docs/commands)
- [Go project layout](https://melody.sh/docs/reference/layout/)
- [Specifying dependencies](https://melody.sh/docs/reference/dependencies/)

## Contribution and Improvements

We encourage you to contribute to Melody! The current iteration of Melody is just a preview of what it could be.  Check [GitHub issues][issues] to see how you can help.

### Building from source

Although you can use `go get` to install Melody, we, of course, recommend using Melody to prepare the project: 

```bash
$ git clone https://github.com/mdy/melody.git
$ cd melody; melody install
$ make build
```

### Running tests

Once you have all the requirements to build Melody, you can run the tests after populating the test data:

```bash
$ make install
$ make test
```

### Submitting updates

If you would like to contribute to this project, just do the following:

1. Fork the repo on Github.
2. Add your features and make commits to your forked repo.
3. Make a pull request to this repo.
4. Review will be done and changes will be requested.
5. Once changes are done or no changes are required, pull request will be merged.
6. The next release will have your changes in it.

Please take a look at the issues page if you want to get started.

If you think it would be nice to have a particular feature that is presently not implemented, we would love to hear your ideas and consider working on it.  Just open an issue in Github.

## Credits and inspiration

Aside from the dependencies specified in the `Melody.toml` file that make Melody possible, we drew ideas, inspiration, and sometimes ported code directly from:

- [Bundler](http://bundler.io) - Ruby dependency manager
- [Cargo](http://doc.crates.io) - Rust dependency manager
- [Composer](https://getcomposer.org) - PHP dependency manager

Melody was started as an internal project at [Gemfury](https://gemfury.com), and the company will continue to sponsor its maintenance and future development.

## Questions

Please use the [tag "melody" on StackOverflow][questions] or [file a Github Issue][issues] if you have any other questions or problems.

## License

Melody is Copyright © 2016 Michael Rykov. See LICENSE file for terms of use and redistribution.

[questions]: http://stackoverflow.com/questions/ask?tags=melody
[issues]: https://github.com/mdy/melody/issues
[melody-api]: https://melody.sh/api/
