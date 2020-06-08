# timeit

The `timeit` utility measures the time of command execution.

It has some features inspired by the FreeBSD `/usr/bin/time`:

1. Human friendly output (for example: `1h32m43s` instead of `5563.00`)

## Examples

Timing a command, with or without options:

    $ timeit sleep 61
    timeit results:
    real: 1m1.007506918s

Timing a shell construct: you have to time the execution of a subshell, for
example:

    $ timeit fish -c 'for i in (seq 3); sleep 1; echo $i; end'
    1
    2
    3
    timeit results:
    real: 3.035378818s

Time a command and print intermediate timings:

    $ timeit -ticker 30s sleep 60
    timeit ticker: running since 30.00119935s
    timeit ticker: running since 1m0.00466081s
    timeit results:
    real: 1m0.005122556s

# Status

Before 1.0.0. Working and tested, backwards incompatible changes possible.

## Supported platforms

Unix-like and macOS.

## Installation

1. Download the archive for your platform from the [releases
  page](https://github.com/marco-m/timeit/releases).
2. Unarchive and copy the `timeit` executable somewhere in your `$PATH`. I like
   to use `$HOME/bin/`.

### Installation for macOS

You have to cope with the macOS gatekeeper, that will put the executable in
quarantine, since it is not signed nor notarized. There are two options:

1. Download the archive with a command-line tool, like curl or wget.
2. Download the archive with a web browser, unarchive and run
   ```
   $ xattr -d com.apple.quarantine timeit
   ```

## Build and install

1. Install [task](https://taskfile.dev/).
2. `$ task`

Then, copy the executable to a directory in your `$PATH`.

## Making a release

    $ env RELEASE_TAG=v0.1.0 summon task release

## License

This code is released under the MIT license, see file [LICENSE](LICENSE).

## Credits

* FreeBSD `/usr/bin/time` ([man page], [C source]).

[man page]: https://www.freebsd.org/cgi/man.cgi?query=time
[C source]: https://github.com/freebsd/freebsd/blob/master/usr.bin/time/time.c
