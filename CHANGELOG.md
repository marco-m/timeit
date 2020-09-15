# timeit Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Fixed

### Changed, breaking

### Changed

### New

## [v0.4.0] - [2020-09-15]

### Fixed

### Changed, breaking

### Changed

- Print newline at the beginning of each ticker [GH-2](https://github.com/marco-m/timeit/issues/2).
- Use Go 1.15.
- Update Taskfile to task v3.

### New

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
