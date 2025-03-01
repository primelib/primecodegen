package openapigenerator

import (
	"os/exec"
)

func IsBinaryAvailable(binary string) bool {
	_, err := exec.LookPath(binary)
	if err != nil {
		return false
	}

	return true
}
