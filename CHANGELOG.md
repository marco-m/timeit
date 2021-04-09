# timeit Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Fixed

- Fix [#1](https://github.com/marco-m/timeit/issues/1) as non reproducible. In any case, added manual test configuration in testdata/terraform.
- Fix tests about signal handling. Code was already correct, but tests were not.

### Changed, breaking

### Changed

- Upgrade to Go 1.16
- Upgrade dependencies

### New

- The sleepit helper now performs signal handling, to enable better tests.

## [v0.6.0] - [2021-01-17]

### New

- Add color output, enabled by default. Use flag `-no-color` to disable. If stderr is not a TTY (eg: redirected to a file), then coloring will be disabled.

## [v0.5.0] - [2020-09-16]

### Changed

- Round elapsed time to milliseconds [GH-4](https://github.com/marco-m/timeit/issues/4)

## [v0.4.0] - [2020-09-15]

### Changed

- Print newline at the beginning of each ticker [GH-2](https://github.com/marco-m/timeit/issues/2).
- Use Go 1.15.
- Update Taskfile to task v3.

## [v0.3.0] - [2020-06-08]

### New

- Add flag `-check-version`: check online if there is a more recent release
  available, courtesy of [taschino](https://github.com/marco-m/taschino).
- Add flag `-ticker <duration>`: print a status line each <duration>.

## [v0.2.0] - [2020-06-05]

### New

- Binary releases available from GitHub [releases
  page](https://github.com/marco-m/timeit/releases).
- Taskfile machinery to make GitHub releases, using
  [github-release](https://github.com/github-release/github-release)

## [v0.1.0] - [2020-06-02]

First release.

### New

- Basic timing functionalities.
- Print timing results also if child exits with error.
- Return correct exit code if child is terminated by a signal (128 + sigNum).
- Ignore SIGINT as /usr/bin/time does; let child handle it (see commit comments
  for ac061824f).
- flag `-version` reports the git commit.


[v0.1.0]: https://github.com/marco-m/timeit/releases/tag/v0.1.0
[v0.2.0]: https://github.com/marco-m/timeit/releases/tag/v0.2.0
[v0.3.0]: https://github.com/marco-m/timeit/releases/tag/v0.3.0
[v0.4.0]: https://github.com/marco-m/timeit/releases/tag/v0.4.0
[v0.5.0]: https://github.com/marco-m/timeit/releases/tag/v0.5.0
[v0.6.0]: https://github.com/marco-m/timeit/releases/tag/v0.6.0
