package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/docopt/docopt-go"
	"github.com/erik/vxsv"
)

func main() {
	usage := fmt.Sprintf(`view [x] separated values

Usage:
  vxsv [-tpmxsd DELIMITER] [-n COUNT] ([-] | [<PATH>])
  vxsv -h | --help

Arguments:
  PATH  file to load [default: -]

Options:
  -h --help                   show this help message and exit.
  -s --stream                 handle streaming data.
  -d --delimiter=<DELIMITER>  separator for values [default: ,].
  -t --tabs                   use tabs as separator value.
  -p --psql                   parse output of psql cli (used as a pager)
  -m --mysql                  parse output of mysql cli
  -n --count=<COUNT>          only read <COUNT> records [default: all].

%s`, vxsv.HelpText)

	args, _ := docopt.Parse(usage, nil, true, "0.0.0", false)

	var count int64
	var data *vxsv.TabularData
	var err error

	// default to stdin if we don't have an explicit file passed in
	reader := io.Reader(os.Stdin)

	if args["<PATH>"] != nil {
		fileName, _ := args["<PATH>"].(string)
		file, err := os.Open(fileName)
		if err != nil {
			panic("Failed to open file")
		}

		reader = io.Reader(file)
	}

	if countStr, ok := args["--count"].(string); ok {
		if countStr == "all" {
			count = math.MaxInt64
		} else if count, err = strconv.ParseInt(countStr, 10, 64); err != nil {
			fmt.Printf("Invalid value given for count: %s\n", countStr)
			os.Exit(1)
		}
	}

	if args["--psql"] == true {
		if data, err = vxsv.ReadPSQLTable(reader, count); err != nil {
			fmt.Printf("Failed to read PSQL data: %v", err)
			os.Exit(1)
		}
	} else if args["--mysql"] == true {
		if data, err = vxsv.ReadMySQLTable(reader, count); err != nil {
			fmt.Printf("Failed to read MySQL data: %v", err)
			os.Exit(1)
		}
	} else {
		delimiter := ','
		if args["-t"] == true {
			delimiter = '\t'
		} else if args["--delimiter"] != nil {
			if delimiterStr, ok := args["--delimiter"].(string); !ok {
				panic("Couldn't grab delimiter")
			} else {
				delimiter = []rune(delimiterStr)[0]
			}
		}

		if data, err = vxsv.ReadCSVFile(reader, delimiter, count); err != nil {
			fmt.Printf("Failed to read CSV file (do you have the right delimiter?): %v\n", err)
			os.Exit(1)
		}
	}

	ui := vxsv.NewUI(data)
	if err := ui.Init(); err != nil {
		fmt.Printf("Failed to initialize terminal UI: %v\n", err)
		os.Exit(1)
	}

	ui.Loop()
}
