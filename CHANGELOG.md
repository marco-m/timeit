# timeit Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - [Unreleased]

First release.

### Fixes

### Breaking Changes

### Changes

### Additions

- Basic timing functionalities.
- Print timing results also if child exits with error.
- Return correct exit code if child is terminated by a signal (128 + sigNum).
- Ignore SIGINT; let child handle it (see commit comments for ac061824f).
- flag `-version` reports the git commit.


[v0.1.0]: https://github.com/marco-m/timeit/releases/tag/v0.1.0
