package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/itchyny/base58-go"
	"github.com/jessevdk/go-flags"
)

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
}

func (opt *flagopts) encoding() *base58.Encoding {
	var encoding *base58.Encoding
	switch opt.Encoding {
	case "flickr":
		encoding = base58.FlickrEncoding
	case "ripple":
		encoding = base58.RippleEncoding
	case "bitcoin":
		encoding = base58.BitcoinEncoding
	default:
		fmt.Fprintf(os.Stderr, "Unknown encoding: %s.\n", opt.Encoding)
		os.Exit(1)
	}
	return encoding
}

type option struct {
	decode   bool
	encoding *base58.Encoding
}

func (cli *cli) run(args []string) int {
	var opts flagopts
	args, err := flags.ParseArgs(&opts, args)
	if err != nil {
		return exitCodeErr
	}
	var inputFiles []string
	for _, name := range append(opts.Input, args...) {
		if name != "" && name != "-" {
			inputFiles = append(inputFiles, name)
		}
	}
	var outFile io.Writer
	if opts.Output == "-" {
		outFile = cli.outStream
	} else {
		file, err := os.Create(opts.Output)
		if err != nil {
			fmt.Fprintln(cli.errStream, err.Error())
			return exitCodeErr
		}
		defer file.Close()
		outFile = file
	}
	status := exitCodeOK
	opt := &option{decode: opts.Decode, encoding: opts.encoding()}
	if len(inputFiles) == 0 {
		status = cli.runInternal(opt, cli.inStream, outFile)
	}
	for _, name := range inputFiles {
		file, err := os.Open(name)
		if err != nil {
			fmt.Fprintln(cli.errStream, err.Error())
			continue
		}
		defer file.Close()
		if s := cli.runInternal(opt, file, outFile); status < s {
			status = s
		}
	}
	return status
}

func (cli *cli) runInternal(opt *option, in io.Reader, out io.Writer) int {
	scanner := bufio.NewScanner(in)
	status := exitCodeOK
	var result []byte
	var err error
	for scanner.Scan() {
		src := scanner.Bytes()
		if opt.decode {
			result, err = opt.encoding.Decode(src)
		} else {
			result, err = opt.encoding.Encode(src)
		}
		if err != nil {
			fmt.Fprintln(cli.errStream, err.Error())
			status = exitCodeErr
			continue
		}
		out.Write(result)
		out.Write([]byte{0x0a})
	}
	return status
}
