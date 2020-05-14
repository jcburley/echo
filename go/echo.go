package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/chzyer/readline"
	"io"
	"net"
	"os"
)

var (
	helpFlag   bool
	noReadline bool
	eval       string
	socket     string
	Stderr     io.Writer = os.Stderr
	Stdout     io.Writer = os.Stdout
	Stdin      io.Reader = os.Stdin
)

func ExitEcho(rc int) {
	os.Exit(rc)
}

func InitSocket(fn func(func() (string, error), *bufio.Writer)) {
	l, err := net.Listen("tcp", socket)
	if err != nil {
		fmt.Fprintf(Stderr, "Cannot start listening on %s: %s\n", socket, err.Error())
		ExitEcho(12)
	}
	defer l.Close()

	fmt.Printf("Listening at %s...\n", l.Addr())

	conn, err := l.Accept() // Wait for a single connection
	if err != nil {
		fmt.Fprintf(Stderr, "Cannot start accepting on %s: %s\n",
			l.Addr(), err.Error())
		ExitEcho(13)
	}

	defer func() {
		conn.Close()
	}()

	fmt.Printf("Accepting client at %s...\n", conn.RemoteAddr())

	fmt.Fprintf(conn, "Welcome to echo, client at %s. Close the connection to exit.\n", conn.RemoteAddr())

	input := bufio.NewReader(conn)
	fn(func() (string, error) {
		return input.ReadString('\n')
	},
		bufio.NewWriter(conn))
}

func echoInput(readFn func() (string, error), writer *bufio.Writer) {
	for {
		line, err := readFn()
		if line != "" {
			writer.WriteString(line)
			writer.Flush()
		}
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(Stderr, "error reading input: %v\n", err)
			}
			break
		}
	}
}

func main() {
	if eval != "" {
		fmt.Println(eval)
		ExitEcho(0)
	}

	if helpFlag {
		flag.Usage()
		ExitEcho(0)
	}

	if socket == "" {
		if noReadline {
			input := bufio.NewReader(Stdin)
			echoInput(func() (string, error) {
				return input.ReadString('\n')
			},
				bufio.NewWriter(Stdout))
		} else {
			rl, err := readline.New("")
			if err != nil {
				fmt.Fprintf(Stderr, "Cannot init readline: %v\n", err)
				ExitEcho(1)
			}
			defer rl.Close()
			for {
				line, err := rl.Readline()
				if err != nil {
					if err != io.EOF {
						fmt.Fprintf(Stderr, "error reading input: %v\n", err)
					}
					break
				}
				fmt.Println(line)
			}
		}
	} else {
		InitSocket(echoInput)
	}

	ExitEcho(0)
}

func init() {
	flag.BoolVar(&noReadline, "no-readline", false, "Use only canonical line input from stdin")
	flag.BoolVar(&helpFlag, "h", false, "Print usage info and exit")
	flag.BoolVar(&helpFlag, "help", false, "Print usage info and exit")
	flag.StringVar(&eval, "i", "", "input string (instead of reading from stdin)")
	flag.StringVar(&eval, "input", "", "input string (instead of reading from stdin)")
	flag.StringVar(&socket, "socket", "", "socket upon which to listen for connections and then echo their input")
	flag.Parse()
}
