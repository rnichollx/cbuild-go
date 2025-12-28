package ccommon

import (
	"os/exec"
	"strings"
)

func GCCIsRealGCC(path string) (bool, error) {
	cmd := exec.Command(path, "--version")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	output := string(out)
	if strings.Contains(strings.ToLower(output), "clang") {
		return false, nil
	}

	if strings.Contains(strings.ToLower(output), "gcc") {
		return true, nil
	}

	return false, nil
}
