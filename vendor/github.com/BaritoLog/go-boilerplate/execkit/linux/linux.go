package linux

import (
	"fmt"
	"os/exec"
)

func Download(url, output string) *exec.Cmd {
	return exec.Command("curl", url, "-o", output)
}

func ExtractGzip(path, directory string) *exec.Cmd {
	return exec.Command("tar", "xvzf", path, "-C", directory)
}

func Remove(file string) *exec.Cmd {
	return exec.Command("rm", file)
}

func Bash(format string, v ...interface{}) *exec.Cmd {
	return exec.Command("sh", "-c", fmt.Sprintf(format, v...))
}
