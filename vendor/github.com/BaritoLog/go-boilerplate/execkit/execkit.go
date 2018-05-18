package execkit

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"

	"github.com/BaritoLog/go-boilerplate/execkit/linux"
)

// Print all command and out
func Print(cmdWriter, stdOutErr io.Writer, cmds ...*exec.Cmd) (err error) {
	for _, cmd := range cmds {
		cmd.Stdout = stdOutErr
		cmd.Stderr = stdOutErr

		if cmdWriter != nil {
			cmdWriter.Write([]byte("\n> "))
			cmdWriter.Write(Bytes(cmd))
			cmdWriter.Write([]byte("\n"))
		}

		err = cmd.Run()
		if err != nil {
			return
		}
	}

	return
}

// Run all commands
func Run(cmdWriter io.Writer, cmds ...*exec.Cmd) (err error) {
	for _, cmd := range cmds {
		buf := bytes.Buffer{}

		cmd.Stdout = &buf
		cmd.Stderr = &buf

		if cmdWriter != nil {
			cmdWriter.Write([]byte("> "))
			cmdWriter.Write(Bytes(cmd))
			cmdWriter.Write([]byte("\n\n"))
		}

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("%s: %s\n", err.Error(), buf.String())
		}
	}

	return

}

// Bytes
func Bytes(cmd *exec.Cmd) []byte {
	buf := bytes.Buffer{}
	for i, arg := range cmd.Args {
		if i > 0 {
			buf.Write([]byte(" "))
		}
		buf.Write([]byte(arg))
	}

	return buf.Bytes()
}

func Pid(keywords ...string) (pid []byte, err error) {
	buf := bytes.Buffer{}
	buf.WriteString("ps ax")
	for _, keyword := range keywords {
		buf.WriteString(" | grep " + keyword)
	}
	buf.WriteString(" | grep -v grep")
	buf.WriteString(" | awk '{print $1}'")
	pid, err = linux.Bash(buf.String()).Output()
	return
}
