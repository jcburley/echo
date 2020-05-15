package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/chzyer/readline"
	"io"
	"net"
	"os"
	"strings"
)

var (
	helpFlag   bool
	lineReader string
	eval       string
	connectTo  string
	socket     string
	prompt     string
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

	switch lineReader {
	case "":
		input := bufio.NewReader(conn)
		fn(func() (string, error) {
			fmt.Fprint(conn, prompt)
			return input.ReadString('\n')
		},
			bufio.NewWriter(conn))
	case "chzyer/readline":
		cfg := readline.Config{Prompt: prompt}
		rl, err := readline.HandleConn(cfg, conn)
		if err != nil {
			fmt.Fprintf(Stderr, "Cannot init readline: %v\n", err)
			ExitEcho(2)
		}
		echoInput(func() (string, error) {
			line, err := rl.Readline()
			if line != "" || err == nil {
				line += "\n"
			}
			return line, err
		},
			bufio.NewWriter(conn))
	}
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

var lineReaders = map[string]struct{}{
	"":                struct{}{},
	"chzyer/readline": struct{}{},
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

	if _, found := lineReaders[lineReader]; !found {
		fmt.Fprintf(Stderr, "Unsupported line reader %s; choose one of: %s\n", lineReader, keys(lineReaders))
		ExitEcho(4)
	}

	if connectTo != "" {
		n := "tcp"
		err := readline.DialRemote(n, connectTo)
		if err != nil {
			fmt.Fprintf(Stderr, "Cannot dial n=%s addr=%s: %v\n", n, connectTo, err)
			ExitEcho(3)
		}
		ExitEcho(0)
	}

	if socket == "" {
		switch lineReader {
		case "":
			input := bufio.NewReader(Stdin)
			echoInput(func() (string, error) {
				fmt.Print(prompt)
				return input.ReadString('\n')
			},
				bufio.NewWriter(Stdout))
		case "chzyer/readline":
			rl, err := readline.New(prompt)
			if err != nil {
				fmt.Fprintf(Stderr, "Cannot init readline: %v\n", err)
				ExitEcho(1)
			}
			defer rl.Close()
			echoInput(func() (string, error) {
				line, err := rl.Readline()
				if line != "" || err == nil {
					line += "\n"
				}
				return line, err
			},
				bufio.NewWriter(Stdout))
		}
	} else {
		InitSocket(echoInput)
	}

	ExitEcho(0)
}

func keys(m map[string]struct{}) string {
	var str []string
	for k, _ := range m {
		str = append(str, "\""+k+"\"")
	}
	return strings.Join(str, ", ")
}

func init() {
	flag.StringVar(&lineReader, "line-reader", "", fmt.Sprintf("Line reader to use, of: %s", keys(lineReaders)))
	flag.BoolVar(&helpFlag, "h", false, "Print usage info and exit")
	flag.BoolVar(&helpFlag, "help", false, "Print usage info and exit")
	flag.StringVar(&eval, "i", "", "input string (instead of reading from stdin)")
	flag.StringVar(&eval, "input", "", "input string (instead of reading from stdin)")
	flag.StringVar(&connectTo, "connect-to", "", "act as a client connecting to server at specified address")
	flag.StringVar(&socket, "socket", "", "socket upon which to listen for connections and then echo their input")
	flag.StringVar(&prompt, "prompt", "\000", "overrides the default prompt")
	flag.Parse()
	if prompt == "\000" {
		prompt = lineReader + "> "
	}
}
