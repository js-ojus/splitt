// Copyright 2017 JONNALAGADDA Srinivas
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// Flags holds the input specification for the job.
type Flags struct {
	mode         string
	input        string
	outDir       string
	outPrefix    string
	start        int64
	unit         int64
	outSuffixLen int
}

func parseArgs() (*Flags, error) {
	var flags Flags
	flag.StringVar(&flags.mode, "mode", "bytes", "'bytes' or 'lines' to determine the unit for splitting; defaults to 'bytes'")
	flag.StringVar(&flags.input, "in", "", "path to the input file")
	flag.StringVar(&flags.outDir, "dir", ".", "directory to which split files should be written; defaults to current directory")
	flag.StringVar(&flags.outPrefix, "pref", "", "prefix to be used for split files")
	flag.Int64Var(&flags.start, "start", 0, "byte or line offset at which to start reading input file; default to 0")
	flag.Int64Var(&flags.unit, "size", 0, "number of bytes or lines after which to split")
	flag.IntVar(&flags.outSuffixLen, "extlen", 3, "number of digits for file name suffix; defaults to 3")
	flag.Parse()

	if !(flags.mode == "bytes" || flags.mode == "lines") {
		return nil, errors.New("mode should be 'bytes' or 'lines'")
	}

	if flags.input == "" {
		return nil, errors.New("specify the input file to be split")
	}

	if flags.outDir == "" {
		return nil, errors.New("specify the output directory")
	}

	if flags.outPrefix == "" {
		return nil, errors.New("specify the output prefix to be used for split files")
	}

	if flags.unit == 0 {
		return nil, errors.New("specify a positive size for splitting at")
	}

	if flags.outSuffixLen == 0 {
		return nil, errors.New("specify the number of output suffix digits to be used for split files")
	}

	return &flags, nil
}

//

func splitBytes(in io.ReadSeeker, flags *Flags) error {
	count := 0
	var offset int64
	stat, _ := os.Stat(flags.input)

	// If a starting offset has been given, seek there.
	if flags.start >= stat.Size() {
		return errors.New("starting offset at or beyond input file size")
	}
	in.Seek(flags.start, io.SeekStart)
	offset = flags.start

	// Main loop of splitting.
	for offset < stat.Size() {
		count++
		ofname := flags.outDir + "/" + flags.outPrefix + "." + fmt.Sprintf("%0[1]*[2]d", flags.outSuffixLen, count)
		out, err := os.Create(ofname)
		if err != nil {
			return errors.New("unable to create output file : " + ofname)
		}
		defer out.Close()

		c, err := io.CopyN(out, in, flags.unit)
		offset += c
		if c < flags.unit {
			if offset < stat.Size() {
				return errors.New("error before EOF : " + err.Error())
			}
		}
	}

	return nil
}

//

func splitLines(in io.ReadSeeker, flags *Flags) error {
	// TODO(js): implement this
	return nil
}

//

func main() {
	flags, err := parseArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	in, err := os.Open(flags.input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not open input file : "+err.Error())
		return
	}
	defer in.Close()

	switch flags.mode {
	case "bytes":
		err = splitBytes(in, flags)

	case "lines":
		err = splitLines(in, flags)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
