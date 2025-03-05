package specutil

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
)

var (
	ErrSpeakeasyNotInstalled = fmt.Errorf("speakeasy is not installed. Please follow the instructions at https://www.speakeasy.com/docs/speakeasy-reference/cli/getting-started")
)

func SpeakEasyFormat(file string) error {
	_, err := exec.LookPath("speakeasy")
	if err != nil {
		return ErrSpeakeasyNotInstalled
	}

	cmd := exec.Command("speakeasy",
		"openapi", "transform", "cleanup",
		"--schema", file,
		"--out", file,
		"--logLevel", "info",
	)
	cmd.Stderr = os.Stderr
	log.Trace().Str("cmd", cmd.String()).Msg("calling speakeasy to format openapi specification")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute primecodegen: %w", err)
	}

	err = os.WriteFile(file, output, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write updated OpenAPI spec: %w", err)
	}

	return nil
}

func SpeakEasyRemoveUnused(file string) error {
	_, err := exec.LookPath("speakeasy")
	if err != nil {
		return ErrSpeakeasyNotInstalled
	}

	cmd := exec.Command("speakeasy",
		"openapi", "transform", "remove-unused",
		"--schema", file,
		"--out", file,
		"--logLevel", "info",
	)
	cmd.Stderr = os.Stderr
	log.Trace().Str("cmd", cmd.String()).Msg("calling speakeasy to format openapi specification")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute primecodegen: %w", err)
	}

	err = os.WriteFile(file, output, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write updated OpenAPI spec: %w", err)
	}

	return nil
}
