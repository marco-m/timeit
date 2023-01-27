# timeit Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## UNRELEASED

### Fixed

- The displayed duration of each ticker is now correctly rounded to a unit proportional to the value of the `--ticker` flag.

### Breaking

- Due to the introduction of package kong to parse the command-line, now `sleepit` wants flags specified with two hyphens, for example `--sleep` instead of `-sleep`.

### Changed

- Timeit now always reports the command status. For example:
   ```
   $ timeit true
   timeit results:
       command succeeded
       real: 4ms

   $ timeit false
   timeit results:
       command failed: exit status 1
       real: 2ms

   $ timeit sleep 2
   ^Ctimeit: got signal name=interrupt count=1 disposition=ignore
   timeit results:
       command terminated abnormally: signal: interrupt
       real: 1.851s
   ```

## [v0.7.0] - [2023-01-21]

### Fixed

- Fix [#1](https://github.com/marco-m/timeit/issues/1) as non reproducible. In any case, added manual test configuration in testdata/terraform.
- Fix tests about signal handling. Code was already correct, but tests were not.

### Breaking

- Due to the introduction of package kong to parse the command-line, now `timeit` wants flags specified with two hyphens, for example `--ticker` instead of `-ticker`. 

### Changed

- Upgrade to Go 1.19.
- Upgrade dependencies.
- Integration tests are now driven by package [rogpeppe/go-internal/testscript](http://github.com/rogpeppe/go-internal).

### New

- The sleepit helper now performs optional signal handling and interruptible work and cleanup phases, to enable better tests:
  ```
  sleepit: sleep for the specified duration, optionally handling signals
  Usage: sleepit <command> [<args>]
  Commands
    default     Use default action: on reception of SIGINT terminate abruptly
    handle      Handle signals: on reception of SIGINT perform cleanup before exiting
    version     Show the sleepit version

  Usage of default:
    -sleep duration
          Sleep duration (default 5s)

  Usage of handle:
    -cleanup duration
          Cleanup duration (default 5s)
    -sleep duration
          Sleep duration (default 5s)
    -term-after N
          Terminate immediately after N signals.
          Default is to terminate only when the cleanup phase has completed.
  ```
- The sleepit helper now has also a good series of tests.
- Add CI: build and test with GitHub Actions, for platforms: Linux, macOS, Windows.
- Add basic and experimental support for Windows. Untested: signal handling.

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
[v0.7.0]: https://github.com/marco-m/timeit/releases/tag/v0.7.0
