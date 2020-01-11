package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	helpFlag    bool
	versionFlag bool
	eval        string
	filename    string
	noReadline  bool
	Stderr      = os.Stderr
	Stdout      = os.Stdout
	Stdin       = os.Stdin
)

func ExitEcho(rc int) {
	os.Exit(rc)
}

func isNumber(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

func notOption(arg string) bool {
	return arg == "-" || !strings.HasPrefix(arg, "-") || isNumber(arg[1:])
}

func parseArgs(args []string) {
	length := len(args)
	stop := false
	missing := false
	noFileFlag := false
	var i int
	for i = 1; i < length; i++ { // shift
		switch args[i] {
		case "-": // denotes stdin
			stop = true
		case "--help", "-h":
			helpFlag = true
			return // don't bother parsing anything else
		case "--version", "-v":
			versionFlag = true
		case "-e", "--eval":
			if i < length-1 && notOption(args[i+1]) {
				i += 1 // shift
				eval = args[i]
			} else {
				missing = true
			}
		case "--no-readline":
			noReadline = true
		case "--file":
			if i < length-1 && notOption(args[i+1]) {
				i += 1 // shift
				filename = args[i]
			}
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(Stderr, "Error: Unrecognized option '%s'\n", args[i])
				ExitEcho(2)
			}
			stop = true
		}
		if stop || missing {
			break
		}
	}
	if missing {
		fmt.Fprintf(Stderr, "Error: Missing argument for '%s' option\n", args[i])
		ExitEcho(3)
	}
	if i < length && !noFileFlag && filename == "" {
		filename = args[i]
		i += 1 // shift
	}
	if i < length {
		fmt.Fprintf(Stderr, "Error: Excess command-line arguments: %s\n", args[i:])
		ExitEcho(4)
	}
}

func main() {
	parseArgs(os.Args)

	if eval != "" {
		fmt.Println(eval)
		ExitEcho(0)
	}

	input := bufio.NewReader(Stdin)
	line := ""
	for {
		r, _, err := input.ReadRune()
		if err != nil {
			break
		}
		if r == '\n' {
			fmt.Println(line)
			line = ""
		} else if r != '\r' {
			line += string(r)
		}
	}
	ExitEcho(0)
}
