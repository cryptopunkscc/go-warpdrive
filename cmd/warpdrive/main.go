package main

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/go-warpdrive"
	"github.com/cryptopunkscc/go-warpdrive/jrpc"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	// setup connection
	pr, pw := io.Pipe()
	rw := &stdReadWrite{pr, os.Stdout}

	// serve
	cli := jrpc.Cli{Conn: rw}
	go func() {
		err := cli.Serve(ctx)
		if err != nil {
			log.Panicln(err)
		} else {
			os.Exit(0)
		}
	}()

	// handler args
	switch len(os.Args) > 1 {
	case true:
		// format application arguments and pass to cli
		args := strings.Join(os.Args[1:], " ")
		_, err := fmt.Fprint(pw, "prompt-off", "\n", args, "\n", "exit", "\n")
		if err != nil {
			log.Panicln(warpdrive.Error(err, "cannot write args"))
		}
	case false:
		// switch to interactive mode, pass std in to cli
		go func() {
			_, err := io.Copy(pw, os.Stdin)
			if err != nil {
				log.Panicln(warpdrive.Error(err, "cannot copy std in"))
			}
		}()
	}

	// Trap ctrl+c
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		for {
			<-sigCh
			println()
			log.Println("shutting down...")
			shutdown()

			<-sigCh
			println()
			log.Println("forcing shutdown...")
			os.Exit(0)
		}
	}()

	<-ctx.Done()

	_ = pw.Close()

	time.Sleep(50 * time.Millisecond)

	os.Exit(0)
}

type stdReadWrite struct {
	io.Reader
	io.Writer
}

func (s stdReadWrite) Close() error {
	return nil
}
