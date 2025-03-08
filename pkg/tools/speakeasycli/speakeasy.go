package speakeasycli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
)

var (
	ErrSpeakeasyNotInstalled    = fmt.Errorf("speakeasy is not installed. Please follow the instructions at https://www.speakeasy.com/docs/speakeasy-reference/cli/getting-started")
	ErrSpeakeasyTransformFailed = fmt.Errorf("failed to execute speakeasy transform command")
	ErrorSpeakeasyConvertFailed = fmt.Errorf("failed to execute speakeasy convert swagger to openapi command")
)

func SpeakEasyTransformCommand(file string, transformMethod string) ([]byte, error) {
	_, err := exec.LookPath("speakeasy")
	if err != nil {
		return nil, ErrSpeakeasyNotInstalled
	}

	cmd := exec.Command("speakeasy",
		"openapi", "transform", transformMethod,
		"--schema", file,
		"--logLevel", "info",
	)
	cmd.Stderr = os.Stderr
	log.Trace().Str("cmd", cmd.String()).Msg("calling speakeasy cli to transform openapi specification")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Join(ErrSpeakeasyTransformFailed, err)
	}

	return output, nil
}

func SpeakEasySwaggerConvertCommand(file string) ([]byte, error) {
	_, err := exec.LookPath("speakeasy")
	if err != nil {
		return nil, ErrSpeakeasyNotInstalled
	}

	cmd := exec.Command("speakeasy",
		"openapi", "transform", "convert-swagger",
		"--schema", file,
		"--logLevel", "info",
	)
	cmd.Stderr = os.Stderr
	log.Trace().Str("cmd", cmd.String()).Msg("calling speakeasy cli to convert openapi specification to swagger")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Join(ErrorSpeakeasyConvertFailed, err)
	}

	return output, nil
}
