package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/itchyny/base58-go"
	"github.com/jessevdk/go-flags"
)

const name = "base58"

const version = "0.0.1"

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
	Decode   bool     `short:"D" long:"decode" description:"decodes input"`
	Encoding string   `short:"e" long:"encoding" default:"flickr" description:"encoding (flickr, ripple or bitcoin)"`
	Input    []string `short:"i" long:"input" default:"-" description:"input file"`
	Output   string   `short:"o" long:"output" default:"-" description:"output file"`
	Version  bool     `short:"v" long:"version" description:"print version"`
}

func (opt *flagopts) getEncoding() (*base58.Encoding, error) {
	var encoding *base58.Encoding
	switch opt.Encoding {
	case "flickr":
		encoding = base58.FlickrEncoding
	case "ripple":
		encoding = base58.RippleEncoding
	case "bitcoin":
		encoding = base58.BitcoinEncoding
	default:
		return nil, fmt.Errorf("unknown encoding: %s", opt.Encoding)
	}
	return encoding, nil
}

func (cli *cli) run(args []string) int {
	var opts flagopts
	args, err := flags.NewParser(
		&opts, flags.HelpFlag|flags.PassDoubleDash,
	).ParseArgs(args)
	if err != nil {
		fmt.Fprintln(cli.errStream, err.Error())
		return exitCodeErr
	}
	if opts.Version {
		fmt.Fprintf(cli.outStream, "%s %s\n", name, version)
		return exitCodeOK
	}
	var inputFiles []string
	for _, name := range append(opts.Input, args...) {
		if name != "" && name != "-" {
			inputFiles = append(inputFiles, name)
		}
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
	status := exitCodeOK
	var encoding *base58.Encoding
	if encoding, err = opts.getEncoding(); err != nil {
		fmt.Fprintln(cli.errStream, err.Error())
		return exitCodeErr
	}
	if len(inputFiles) == 0 {
		if s := cli.runInternal(opts.Decode, encoding, cli.inStream); s != exitCodeOK {
			status = s
		}
	}
	for _, name := range inputFiles {
		if s := cli.runFile(opts.Decode, encoding, name); s != exitCodeOK {
			status = s
		}
	}
	return status
}

func (cli *cli) runFile(decode bool, encoding *base58.Encoding, name string) int {
	file, err := os.Open(name)
	if err != nil {
		fmt.Fprintln(cli.errStream, err.Error())
		return exitCodeErr
	}
	defer file.Close()
	return cli.runInternal(decode, encoding, file)
}

func (cli *cli) runInternal(decode bool, encoding *base58.Encoding, in io.Reader) int {
	scanner := bufio.NewScanner(in)
	status := exitCodeOK
	var result []byte
	var err error
	for scanner.Scan() {
		src := scanner.Bytes()
		if decode {
			result, err = encoding.Decode(src)
		} else {
			result, err = encoding.Encode(src)
		}
		if err != nil {
			fmt.Fprintln(cli.errStream, err.Error()) // should print error each line
			status = exitCodeErr
			continue
		}
		cli.outStream.Write(result)
		cli.outStream.Write([]byte{0x0a})
	}
	return status
}
