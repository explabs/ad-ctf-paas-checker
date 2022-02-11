package runner

import (
	"bytes"
	"os/exec"
	"strings"
)

func RunScript(args ...string) (result string, stderr string, err error) {
	output, stderr, err := runProcess(exec.Command("python3", args...))
	if err != nil {
		return "", stderr, err
	}
	return strings.TrimSuffix(output, "\n"), stderr, nil
}

func runProcess(process *exec.Cmd) (stdout, stderr string, err error) {
	var stdOutBuf, stdErrBuf bytes.Buffer
	process.Stdout = &stdOutBuf
	process.Stderr = &stdErrBuf
	if err = process.Run(); err != nil {
		return "", "", err
	}
	return stdOutBuf.String(), stdErrBuf.String(), nil
}
