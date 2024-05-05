package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
	"unicode"
	"unicode/utf8"

	"github.com/itchyny/base58-go"
	"github.com/jessevdk/go-flags"
)

const name = "base58"

const version = "0.2.1"

var revision = "HEAD"

func main() {
	os.Exit((&cli{
		inStream:  os.Stdin,
		outStream: os.Stdout,
		errStream: os.Stderr,
	}).run(os.Args[1:]))
}

const (
	exitCodeOK = iota
	exitCodeErr
)

type cli struct {
	inStream  io.Reader
	outStream io.Writer
	errStream io.Writer
}

type flagopts struct {
	Decode   bool             `short:"D" long:"decode" description:"decode input"`
	Encoding *base58.Encoding `short:"e" long:"encoding" default:"flickr" choice:"flickr" choice:"ripple" choice:"bitcoin" description:"encoding name"`
	Input    []string         `short:"i" long:"input" default:"-" description:"input file"`
	Output   string           `short:"o" long:"output" default:"-" description:"output file"`
	Version  bool             `short:"v" long:"version" description:"print version"`
}

func (cli *cli) run(args []string) int {
	var opts flagopts
	args, err := flags.NewParser(
		&opts, flags.HelpFlag|flags.PassDoubleDash,
	).ParseArgs(args)
	if err != nil {
		if err, ok := err.(*flags.Error); ok && err.Type == flags.ErrHelp {
			fmt.Fprintln(cli.outStream, err.Error())
			return exitCodeOK
		}
		fmt.Fprintln(cli.errStream, err.Error())
		return exitCodeErr
	}
	if opts.Version {
		fmt.Fprintf(cli.outStream, "%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return exitCodeOK
	}
	if opts.Output != "-" {
		file, err := os.Create(opts.Output)
		if err != nil {
			fmt.Fprintln(cli.errStream, err.Error())
			return exitCodeErr
		}
		defer file.Close()
		cli.outStream = file
	}
	var f func([]byte) ([]byte, error)
	if opts.Decode {
		f = opts.Encoding.Decode
	} else {
		f = opts.Encoding.Encode
	}
	if len(args) == 0 {
		args = append(args, opts.Input...)
	}
	status := exitCodeOK
	for _, name := range args {
		status = max(cli.runInternal(name, f), status)
	}
	return status
}

func (cli *cli) runInternal(name string, f func([]byte) ([]byte, error)) int {
	var in io.Reader
	if name == "-" {
		in = cli.inStream
	} else {
		file, err := os.Open(name)
		if err != nil {
			fmt.Fprintln(cli.errStream, err.Error())
			return exitCodeErr
		}
		defer file.Close()
		in = file
	}
	scanner := bufio.NewScanner(in)
	status := exitCodeOK
	for scanner.Scan() {
		result, err := processLine(scanner.Bytes(), f)
		if err != nil {
			fmt.Fprintln(cli.errStream, err.Error()) // should print error each line
			status = exitCodeErr
			continue
		}
		cli.outStream.Write(result)
		cli.outStream.Write([]byte{'\n'})
	}
	return status
}

func processLine(src []byte, f func([]byte) ([]byte, error)) ([]byte, error) {
	var results [][]byte
	for i := 0; len(src) > 0; src = src[i:] {
		if i = bytes.IndexFunc(src, unicode.IsSpace); i == 0 {
			_, width := utf8.DecodeRune(src)
			results = append(results, src[:width])
			src = src[width:]
			continue
		} else if i < 0 {
			i = len(src)
		}
		result, err := f(src[:i])
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if len(results) == 1 {
		return results[0], nil
	}
	return slices.Concat(results...), nil
}
