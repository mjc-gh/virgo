package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mjc-gh/virgo"
	"github.com/mjc-gh/virgo/engine"
	"github.com/mjc-gh/virgo/internal/browser"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

const URL = "url"

var ErrInvalidDeviceProperties = errors.New("invalid device properties")
var ErrScreenShotFailed = errors.New("screenshot result error")

var (
	logger  *zerolog.Logger
	version string
)

type taskCallbackFn = func(*cli.Command, *engine.Engine) error

func main() {
	baseFlags := []cli.Flag{
		&cli.BoolFlag{Name: "debug", Aliases: []string{"d"}, Usage: "enable debug logging"},
		&cli.BoolFlag{Name: "headfull", Aliases: []string{"H"}, Usage: "run browser in headfull mode"},
		&cli.IntFlag{Name: "concurrency", Aliases: []string{"c"}, Usage: "number of concurrent workers"},
		&cli.IntFlag{Name: "remote-port", Usage: "remote DevTools port"},
		&cli.StringFlag{Name: "remote-host", Usage: "remote DevTools host"},
		&cli.StringFlag{
			Name: "device-type", Value: "desktop", Usage: "device type (desktop/mobile/tablet)", Action: validDeviceType,
		},
		&cli.StringFlag{Name: "device-size", Value: "large", Usage: "device size preset", Action: validDeviceSize},
		&cli.StringFlag{Name: "user-agent", Value: "chrome", Usage: "browser user-agent preset"},
	}

	ver := version
	if ver == "" {
		ver = "0.0.0"
	}

	cmd := &cli.Command{
		Name:    "virgo",
		Version: ver,
		Usage:   "A tool for converting webpages to Markdown and plaintext.",
		Commands: []*cli.Command{
			{
				Name:  "screenshot",
				Usage: "Screenshot one or more URLs",
				Arguments: []cli.Argument{
					&cli.StringArgs{Name: URL, Min: 1, Max: -1},
				},
				Flags: append([]cli.Flag{
					&cli.StringFlag{Name: "output-dir", Value: "tmp/", Aliases: []string{"o"}, Usage: "directory for screenshots"},
				}, baseFlags...),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runTask(ctx, cmd, "screenshot", map[string]any{}, screenshotCallback)
				},
			}, {
				Name:      "markdown",
				Usage:     "Get the markdown content of a URL",
				Arguments: []cli.Argument{&cli.StringArg{Name: URL}},
				Flags: append([]cli.Flag{
					&cli.BoolFlag{Name: "include-images", Aliases: []string{"i"}, Usage: "include images in markdown output"},
				}, baseFlags...),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					params := map[string]any{
						"include-images": cmd.Bool("include-images"),
					}

					return runTask(ctx, cmd, "markdown", params, stdOutCallback)
				},
			}, {
				Name:      "plaintext",
				Usage:     "Get the plantext content of a URL",
				Arguments: []cli.Argument{&cli.StringArg{Name: URL}},
				Flags:     baseFlags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runTask(ctx, cmd, "plaintext", map[string]any{}, stdOutCallback)
				},
			}, {
				Name:      "links",
				Usage:     "Search for links on a page using fuzzy matching",
				Arguments: []cli.Argument{&cli.StringArg{Name: URL}},
				Flags: append([]cli.Flag{
					&cli.StringFlag{Name: "search", Aliases: []string{"s"}, Usage: "search term for fuzzy matching"},
					&cli.IntFlag{
						Name: "threshold", Aliases: []string{"t"}, Value: 3,
						Usage: "fuzzy match threshold (0=exact, higher=fuzzier)",
					},
				}, baseFlags...),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					params := map[string]any{
						"search":    cmd.String("search"),
						"threshold": cmd.Int("threshold"),
					}

					return runTask(ctx, cmd, "links", params, stdOutCallback)
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runTask(ctx context.Context, cmd *cli.Command, name string, params map[string]any, callback taskCallbackFn) error {
	logger = virgo.SetupLogger(cmd.Bool("debug"))

	deviceSize := cmd.StringArg("device-size")
	deviceType := cmd.StringArg("device-type")
	remoteHost := cmd.String("remote-host")
	remotePort := cmd.Int("remote-port")
	urls := cmd.StringArgs(URL)

	opts := []engine.Option{engine.WithLogger(virgo.Logger())}

	if len(urls) == 0 {
		url := cmd.StringArg(URL)
		urls = []string{url}
	}

	if remoteHost != "" && remotePort != 0 {
		opts = append(opts, engine.WithRemoteAllocator(remoteHost, remotePort))
	} else if cmd.Bool("headfull") {
		opts = append(opts, engine.WithHeadfullLocalAllocator())
	}

	e := engine.New(cmd.Int("concurrency"), opts...)
	e.Start(ctx)

	for _, url := range urls {
		t := engine.NewTask(
			name, url,
			engine.WithParams(params),
			engine.WithDeviceProperties(deviceType, deviceSize),
			engine.WithUserAgent(deviceType, cmd.StringArg("user-agent")),
		)

		e.Add(t)
	}

	go e.Shutdown()

	// TODO handle interrupt signal and wait for shutdown

	return callback(cmd, e)
}

func stdOutCallback(cmd *cli.Command, e *engine.Engine) error {
	for r := range e.Results() {
		var out string

		if r.Error != nil {
			logger.Warn().Msgf("result error: %v", r.Error)

			continue
		}

		logger.Debug().
			Str(URL, r.URL).
			Str("duration", r.Elapsed.String()).
			Msg("plaintext result")

		switch v := r.Result.(type) {
		case *engine.MarkdownResult:
			out = v.Content
		case *engine.PlaintextResult:
			out = v.Content
		case *engine.LinksResult:
			out = v.Content
		default:
			logger.Warn().Msg("plaintext result type assertion failed")

			continue
		}

		_, err := fmt.Fprint(os.Stdout, out)
		if err != nil {
			return err
		}
	}

	return nil
}

func screenshotCallback(cmd *cli.Command, e *engine.Engine) error {
	outputDir := cmd.String("output-dir")
	err := os.MkdirAll(outputDir, 0o750)
	if err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	for r := range e.Results() {
		if r.Error != nil {
			logger.Warn().Msgf("result error: %v", r.Error)

			continue
		}

		logger.Debug().
			Str(URL, r.URL).
			Str("duration", r.Elapsed.String()).
			Msg("screenshot result")

		sr, ok := r.Result.(*engine.ScreenshotResult)
		if !ok {
			return ErrScreenShotFailed
		}

		fileName, err := urlToFilename(r.URL)
		if err != nil {
			return fmt.Errorf("build screenshot filename: %w", err)
		}

		outPath := filepath.Join(outputDir, fileName+".png")
		out, err := os.Create(filepath.Clean(outPath))
		if err != nil {
			return fmt.Errorf("create screenshot file: %w", err)
		}

		if _, err = out.Write(*sr.Buffer); err != nil {
			// Try to close file, but ignore any errors
			_ = out.Close()

			return fmt.Errorf("write screenshot file: %w", err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("close screenshot file: %w", err)
		}
	}

	return nil
}

func validDeviceType(ctx context.Context, cmd *cli.Command, v string) error {
	if !browser.IsValidDeviceType(v) {
		return fmt.Errorf("%w: %v", ErrInvalidDeviceProperties, v)
	}

	return nil
}

func validDeviceSize(ctx context.Context, cmd *cli.Command, v string) error {
	if !browser.IsValidDeviceSize(v) {
		return fmt.Errorf("%w: %v", ErrInvalidDeviceProperties, v)
	}

	return nil
}

func urlToFilename(taskURL string) (string, error) {
	u, err := url.Parse(taskURL)
	if err != nil {
		return "", err
	}

	domain := u.Host
	path := u.Path

	combined := strings.Trim(domain+path, "/")

	safe := strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '.':
			return '_'
		default:
			return r
		}
	}, combined)

	if safe == "" {
		safe = "index"
	}

	return safe, nil
}
