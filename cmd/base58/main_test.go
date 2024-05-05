package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestCliRun(t *testing.T) {
	testCases := []struct {
		name       string
		args       []string
		input      string
		expected   string
		expectedRe *regexp.Regexp
		err        string
		errRe      *regexp.Regexp
	}{
		{
			name: "encode",
			args: []string{},
			input: `
0
32
64
000
512
1024
16777216
00068719476736
18446744073709551616
79228162514264337593543950336
`,
			expected: `
1
y
27
111
9Q
iE
2tZhm
1112NGvhhq
JPwcyDCgEuq
5QchsBFApWPVxyp9C
`,
		},
		{
			name: "encode flickr",
			args: []string{"--encoding", "flickr"},
			input: `
0
32
64
000
512
1024
16777216
00068719476736
18446744073709551616
79228162514264337593543950336
`,
			expected: `
1
y
27
111
9Q
iE
2tZhm
1112NGvhhq
JPwcyDCgEuq
5QchsBFApWPVxyp9C
`,
		},
		{
			name: "encode ripple",
			args: []string{"--encoding=ripple"},
			input: `
0
32
64
000
512
1024
16777216
00068719476736
18446744073709551616
79228162514264337593543950336
`,
			expected: `
r
Z
pf
rrr
9q
JC
p7zHM
rrrpo6WHHR
jFXUZedGCVR
nqUHTcgbQAFvYZQ9d
`,
		},
		{
			name: "encode bitcoin",
			args: []string{"-e", "bitcoin"},
			input: `
0
32
64
000
512
1024
16777216
00068719476736
18446744073709551616
79228162514264337593543950336
`,
			expected: `
1
Z
27
111
9q
Jf
2UzHM
1112ohWHHR
jpXCZedGfVR
5qCHTcgbQwpvYZQ9d
`,
		},
		{
			name: "encode error",
			args: []string{},
			input: `foo
bar
`,
			err: `expecting a non-negative number but got "foo"
expecting a non-negative number but got "bar"
`,
		},
		{
			name: "encode multiple values in each line",
			args: []string{},
			input: `
0 32 64		  		000	512
   1024 16777216
`,
			expected: `
1 y 27		  		111	9Q
   iE 2tZhm
`,
		},
		{
			name: "decode",
			args: []string{"-D"},
			input: `
1
y
27
111
9Q
iE
2tZhm
1112NGvhhq
JPwcyDCgEuq
5QchsBFApWPVxyp9C
`,
			expected: `
0
32
64
000
512
1024
16777216
00068719476736
18446744073709551616
79228162514264337593543950336
`,
		},
		{
			name: "decode flickr",
			args: []string{"--decode", "--encoding", "flickr"},
			input: `
1
y
27
111
9Q
iE
2tZhm
1112NGvhhq
JPwcyDCgEuq
5QchsBFApWPVxyp9C
`,
			expected: `
0
32
64
000
512
1024
16777216
00068719476736
18446744073709551616
79228162514264337593543950336
`,
		},
		{
			name: "decode ripple",
			args: []string{"--encoding=ripple", "--decode"},
			input: `
r
Z
pf
rrr
9q
JC
p7zHM
rrrpo6WHHR
jFXUZedGCVR
nqUHTcgbQAFvYZQ9d
`,
			expected: `
0
32
64
000
512
1024
16777216
00068719476736
18446744073709551616
79228162514264337593543950336
`,
		},
		{
			name: "decode bitcoin",
			args: []string{"-D", "-e", "bitcoin"},
			input: `
1
Z
27
111
9q
Jf
2UzHM
1112ohWHHR
jpXCZedGfVR
5qCHTcgbQwpvYZQ9d
`,
			expected: `
0
32
64
000
512
1024
16777216
00068719476736
18446744073709551616
79228162514264337593543950336
`,
		},
		{
			name: "decode multiple values in each line",
			args: []string{"-D"},
			input: `
1 y 27		  		111	9Q
   iE 2tZhm
`,
			expected: `
0 32 64		  		000	512
   1024 16777216
`,
		},
		{
			name:     "short clumped options",
			args:     []string{"-De=ripple"},
			input:    "r\n",
			expected: "0\n",
		},
		{
			name: "decode error",
			args: []string{"--decode"},
			input: `FOO
Fal
`,
			err: `invalid character 'O' in decoding a base58 string "FOO"
invalid character 'l' in decoding a base58 string "Fal"
`,
		},
		{
			name:  "decode flag error",
			args:  []string{"--decode=foo"},
			errRe: regexp.MustCompile(name + ": boolean flag `--decode' cannot have an argument\n"),
		},
		{
			name: "encoding error",
			args: []string{"--encoding", "foo"},
			errRe: regexp.MustCompile(name + ": invalid argument for flag `--encoding': " +
				"expected one of \\[flickr, ripple, bitcoin\\] but got foo\n"),
		},
		{
			name: "encoding error",
			args: []string{"--encoding"},
			err:  name + ": expected argument for flag `--encoding'\n",
		},
		{
			name:  "negative number error",
			input: "-100000000000000000000",
			err: `expecting a non-negative number but got "-100000000000000000000"
`,
		},
		{
			name: "input error",
			args: []string{"--input"},
			err:  name + ": expected argument for flag `--input'\n",
		},
		{
			name: "input error file",
			args: []string{"--input", "xxx"},
			errRe: regexp.MustCompile(name + ": open xxx: (?:no such file or directory|" +
				"The system cannot find the file specified\\.)\n"),
		},
		{
			name: "input error file",
			args: []string{"--", "--input"},
			errRe: regexp.MustCompile(name + ": open --input: (?:no such file or directory|" +
				"The system cannot find the file specified\\.)\n"),
		},
		{
			name: "invalid flag",
			args: []string{"--foo"},
			err:  name + ": unknown flag `--foo'\n",
		},
		{
			name:       "version flag",
			args:       []string{"--version"},
			expectedRe: regexp.MustCompile("^" + name + " "),
		},
		{
			name:       "help flag",
			args:       []string{"--help"},
			expectedRe: regexp.MustCompile("-h, --help"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var outStream, errStream strings.Builder
			cli := cli{
				inStream:  strings.NewReader(tc.input),
				outStream: &outStream,
				errStream: &errStream,
			}
			got := cli.run(tc.args)
			if tc.err == "" && tc.errRe == nil {
				if expected := exitCodeOK; got != expected {
					t.Errorf("expected: %v\ngot: %v", expected, got)
				}
				if tc.expectedRe != nil {
					if got, expected := outStream.String(), tc.expectedRe; !expected.MatchString(got) {
						t.Errorf("expected pattern: %v\ngot: %v", expected, got)
					}
				} else {
					if got, expected := outStream.String(), tc.expected; got != expected {
						t.Errorf("expected: %v\ngot: %v", expected, got)
					}
				}
			} else {
				if expected := exitCodeErr; got != expected {
					t.Errorf("expected: %v\ngot: %v", expected, got)
				}
				if tc.errRe != nil {
					if got, expected := errStream.String(), tc.errRe; !expected.MatchString(got) {
						t.Errorf("expected pattern: %v\ngot: %v", expected, got)
					}
				} else {
					if got, expected := errStream.String(), tc.err; got != expected {
						t.Errorf("expected: %v\ngot: %v", expected, got)
					}
				}
			}
		})
	}
}
