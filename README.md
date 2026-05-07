# virgo

Turn web pages into Markdown using the Chrome DevTools Protocol

![Demo GIF](./virgo.gif)

## Usage

```
virgo
NAME:
   virgo - A tool for converting webpages to Markdown and plaintext.

USAGE:
   virgo [global options] [command [command options]]

VERSION:
   0.0.0

COMMANDS:
   screenshot  Screenshot one or more URLs
   markdown    Get the markdown content of a URL
   plaintext   Get the plantext content of a URL
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### Markdown subcommand

```
virgo markdown --help
NAME:
   virgo markdown - Get the markdown content of a URL

USAGE:
   virgo markdown [options] url

OPTIONS:
   --include-images, -i       include images in markdown output (default: false)
   --debug, -d                enable debug logging (default: false)
   --headfull, -H             run browser in headfull mode (default: false)
   --concurrency int, -c int  number of concurrent workers (default: 0)
   --remote-port int          remote DevTools port (default: 0)
   --remote-host string       remote DevTools host
   --device-type string       device type (desktop/mobile/tablet) (default: "desktop")
   --device-size string       device size preset (default: "large")
   --user-agent string        browser user-agent preset (default: "chrome")
   --help, -h                 show help
```

## Development

You can build the CLI tool with the following:

```
make build.cli
make test
```
