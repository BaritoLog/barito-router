package execkit

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/BaritoLog/go-boilerplate/execkit/linux"
	. "github.com/BaritoLog/go-boilerplate/testkit"
)

func TestPrint(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {

		cmdWriter := bytes.Buffer{}
		stdOutErr := bytes.Buffer{}

		err := Print(&cmdWriter, &stdOutErr, linux.Bash("echo stdout; echo 1>&2 stderr"))
		FatalIfError(t, err)

		get := cmdWriter.String()
		want := "\n> sh -c echo stdout; echo 1>&2 stderr\n"
		FatalIf(t, get != want, "command writer get wrong value: %s", get)

		get = stdOutErr.String()
		want = "stdout\nstderr\n"
		FatalIf(t, get != want, "standar output/error writer get wrong value: %s", get)

	}
}

func TestPrint_GetError(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {

		cmdWriter := bytes.Buffer{}
		stdOutErr := bytes.Buffer{}

		err := Print(&cmdWriter, &stdOutErr,
			linux.Bash("echo stdout; echo 1>&2 stderr"),
			linux.Bash("bad_command"),
			linux.Bash("echo na"),
		)
		FatalIfWrongError(t, err, "exit status 127")

		get := cmdWriter.String()
		want := `
> sh -c echo stdout; echo 1>&2 stderr

> sh -c bad_command
`
		FatalIf(t, get != want, "command writer got wrong: %s", get)

		get = stdOutErr.String()
		want = `stdout
stderr
sh: bad_command: command not found
`
		FatalIf(t, get != want, "standar output/error writer got wrong: %s", get)

	}
}

func TestRun_GetError(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {

		cmdWriter := bytes.Buffer{}

		err := Run(&cmdWriter,
			linux.Bash("echo stdout; echo 1>&2 stderr"),
			linux.Bash("bad_command"),
			linux.Bash("echo na"),
		)

		FatalIfWrongError(t, err, "exit status 127: sh: bad_command: command not found\n\n")

		get := cmdWriter.String()
		want := "> sh -c echo stdout; echo 1>&2 stderr\n\n> sh -c bad_command\n\n"
		FatalIf(t, get != want, "command writer get wrong: %s", get)
	}
}

func TestRun(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {

		cmdWriter := bytes.Buffer{}

		err := Run(&cmdWriter, linux.Bash("echo stdout; echo 1>&2 stderr"))
		FatalIfError(t, err)

		get := cmdWriter.String()
		want := "> sh -c echo stdout; echo 1>&2 stderr\n\n"
		FatalIf(t, get != want, "command writer get wrong: %s", get)
	}
}

func TestPid(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {

		pid, err := Pid("system")
		FatalIfError(t, err)
		FatalIf(t, len(pid) < 1, "pid is empty")
	}
}
