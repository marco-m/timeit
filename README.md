# timeit

[![CI status](https://github.com/marco-m/timeit/actions/workflows/ci.yml/badge.svg)](https://github.com/marco-m/timeit/actions)

The `timeit` utility measures the time of command execution.

It has some features inspired by the FreeBSD `/usr/bin/time`:

1. Human friendly output (for example: `1h32m43s` instead of `5563.00`)

## Examples

Time a command, with or without options:

    $ timeit sleep 61
    timeit results:
    real: 1m1.008s

Time a shell construct: you have to time the execution of a subshell, for
example:

    $ timeit fish -c 'for i in (seq 3); sleep 1; echo $i; end'
    1
    2
    3
    timeit results:
    real: 3.035s

Time a command and print intermediate timings (color output by default):

    $ timeit --ticker 30s sleep 60
    timeit ticker: running for 30s
    timeit ticker: running for 1m
    timeit results:
    real: 1m0.005s

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

## Install from source

1. Install [Task](https://taskfile.dev/).
2. Run it: `task`.

Then, copy the executable to a directory in your `$PATH`.

## Making a release

    $ env RELEASE_TAG=v0.1.0 summon task release

## License

This code is released under the MIT license, see file [LICENSE](LICENSE).

## Credits

* FreeBSD `/usr/bin/time` ([man page], [C source]).

[man page]: https://www.freebsd.org/cgi/man.cgi?query=time
[C source]: https://github.com/freebsd/freebsd/blob/master/usr.bin/time/time.c
