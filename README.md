# timeit

The `timeit` utility measures the time of command execution.

It has some features inspired by the FreeBSD `/usr/bin/time`:

1. Human friendly output (for example: `1h32m43s` instead of `5563.00`)

## Examples

    $ timeit sleep 61
    timeit results:
    real: 1m1.007506918s

# Status

Version 0. Working and tested, but expect breaking changes.

## Supported platforms

Unix-like and macOS.

## Build and install

* Option 1.
  1. Install [task](https://taskfile.dev/).
  2. `$ task`

* Option 2.
  1. `$ go build ./cmd/timeit`
  2. `$ go test ./...`

Then, copy the executable to a directory in your `$PATH`, for example `~/bin`.

## License

This code is released under the MIT license, see file [LICENSE](LICENSE).

## Credits

* FreeBSD `/usr/bin/time` ([man page], [C source]).

[man page]: https://www.freebsd.org/cgi/man.cgi?query=time
[C source]: https://github.com/freebsd/freebsd/blob/master/usr.bin/time/time.c
