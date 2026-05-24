# virgo

Turn web pages into Markdown using the Chrome DevTools Protocol

![Demo GIF](./virgo.gif)

## Install

You can install the primary CLI utility using `go install`:

```
go install github.com/mjc-gh/virgo/cmd/virgo@latest
```

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
    plaintext   Get the plaintext content of a URL
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

### Headless Shell

You can use the [chromedp headless
shell](https://hub.docker.com/r/chromedp/headless-shell/) container if
you do not have Chrome installed locally. This works well on Linux
servers:

```
docker pull chromedp/headless-shell:latest
docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell
virgo markdown --remote-port 9222 --remote-host 127.0.0.1 cnn.com
```

## Development

You need Go and `golangci-lint`

```
mise use -g go@1.26.3
mise use -g golangci-lint@v2.12.2
```

You can build and test the CLI tool with the following:

```
make build.cli
make test
```
