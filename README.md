# ping-monitor: Summarize host connectivity

Ping a list of hosts and output connectivity statistics about each of them.

### `--help`

```
SYNOPSIS:
    ping-monitor [--count|-c <int>] [--help|-h|-?] [--icmp|-i]
                 [--output-csv|--csv] [<args>]

OPTIONS:
    --count|-c <int>      (default: 30)

    --help|-h|-?          (default: false)

    --icmp|-i             requires privileges (default: false)

    --output-csv|--csv    (default: false)
```

### Example

```
❯ ping-monitor -c 5 google.com github.com
            Packets   Round trip times
Host        Received  Minimum      Maximum      Mean         Std. Dev.
google.com    5/5     6.477ms      13.387ms     8.9992ms     2.859179ms
github.com    5/5     83.166ms     86.511ms     84.501ms     1.174088ms
```

## License

This project dual-licensed under the Apache 2 and MIT licenses. You may choose
to use either.

  * [Apache License, Version 2.0](LICENSE-APACHE)
  * [MIT license](LICENSE-MIT)

### Contributions

Unless you explicitly state otherwise, any contribution you submit as defined
in the Apache 2.0 license shall be dual licensed as above, without any
additional terms or conditions.
