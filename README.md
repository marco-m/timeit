# timeit

[![CI status](https://github.com/marco-m/timeit/actions/workflows/ci.yml/badge.svg)](https://github.com/marco-m/timeit/actions)

The `timeit` utility measures the time of command execution.

It has the following features:

- Color output.
- Human friendly output (for example: `1h32m43s` instead of `5563.00`), inspired by the FreeBSD `/usr/bin/time`.
- Periodic ticker.

## Examples

Time a command, with or without options:

    $ timeit sleep 61
    timeit results:
        command succeeded
        real: 1m1.008s

Time a shell construct: you have to time the execution of a subshell, for
example:

    $ timeit fish -c 'for i in (seq 3); sleep 1; echo $i; end'
    1
    2
    3
    timeit results:
        command succeeded
        real: 3.035s

Time a command and print intermediate timings (color output by default):

    $ timeit --ticker 30s sleep 60
    timeit ticker: running for 30s
    timeit ticker: running for 1m
    timeit results:
        command succeeded
        real: 1m0.005s

The termination status of the command is always clearly reported:

    $ timeit false
    timeit results:
        command failed: exit status 1
        real: 2ms

    $ timeit sleep 2
    ^Ctimeit: got signal name=interrupt count=1 disposition=ignore
    timeit results:
        command terminated abnormally: signal: interrupt
        real: 1.851s

Time a command, observe its output and summarize in-flight operations (example: pytest --verbose):

    $ timeit --ticker=30s --observe=pytest pytest --verbose testdata/pytest
    test_fruits.py::test_apple
    test_fruits.py::test_banana
    ...
    [gw3] [4%] PASSED test_fruits.py::test_banana
    test_herbs.py::test_coriander
    [gw2] [9%] PASSED test_herbs.py::test_basil
    ...
    timeit ticker: running for 2m0s
    in-flight:                                <== sorted by age (oldest first)
        test_fruits.py::test_appple    50s
        test_fruits.py::test_banana    48s
        test_herbs.py::test_coriander   3s
    ...
    finished:                                 <== optional, only if verbose
        test_fruits.py::test_banana    1h3m
        test_herbs.py::test_basil        3s
    ...
    timeit results:
        command succeeded
        real: 2h34m21s
        slowest flights:
            test_fruits.py::test_banana      1h3m
            test_herbs.py::test_coriander   48m3s

Check online if there is a more recent version:

    $ timeit --check-version
    installed version v0.2.1 is older than the latest version v0.3.0
    To upgrade visit https://github.com/marco-m/timeit

## Status

Pre 1.0.0. Working and tested, backwards incompatible changes possible.

## Supported platforms

Unix-like and macOS.

## Signal handling and exit status

`timeit`, like its ancestor `/usr/bin/time`, will ignore SIGINT (CTRL-C) and will transparently let the timed command decide how to handle the signal. This allows for example the timed command to react to SIGINT by entering a cleanup phase before exiting.

In any case, `timeit` will exit with the same exit status of the timed command.

## Install from binary package

1. Download the archive for your platform from the [releases
  page](https://github.com/marco-m/timeit/releases).
2. Unarchive and copy the `timeit` executable somewhere in your `$PATH`.

### Install for macOS

You have to cope with the macOS gatekeeper, that will put the executable in
quarantine, since it is not signed nor notarized. There are two options:

1. Download the archive with a command-line tool, like curl or wget.
2. Download the archive with a web browser, unarchive and run
   ```
   $ xattr -d com.apple.quarantine timeit
   ```

## Using the source

### Getting familiar with the build tool

    ./vis -h

### Installing from source

1. Run: `./vis build`.
2. Copy the `timeit` executable to a directory in your `$PATH`.

### Making a release

    $ env RELEASE_TAG=v0.1.0 summon task release

## License

This code is released under the MIT license, see file [LICENSE](LICENSE).

## Credits

- FreeBSD `/usr/bin/time` ([man page], [C source]).

[man page]: https://www.freebsd.org/cgi/man.cgi?query=time
[C source]: https://github.com/freebsd/freebsd/blob/master/usr.bin/time/time.c
