package generator

import (
	"errors"
	"os"
	"os/exec"

	"github.com/primelib/primecodegen/pkg/logging"
)

// ErrPrintingPressNotInstalled is returned when the ppress binary cannot be found in the PATH
var ErrPrintingPressNotInstalled = errors.New("printing-press (ppress) cli is not installed")

type PrintingPressGenerator struct {
	OutputName string   `json:"-" yaml:"-"`
	APISpec    string   `json:"-" yaml:"-"`
	Args       []string `json:"-" yaml:"-"`
	Config     PrintingPressGeneratorConfig
}

type PrintingPressGeneratorConfig struct {
	// Skip toggles (Default to false, meaning artifacts ARE generated/shown by default)
	NoJson   bool `json:"noJson" yaml:"noJson"`
	NoHtml   bool `json:"noHtml" yaml:"noHtml"`
	NoLlm    bool `json:"noLlm" yaml:"noLlm"`
	NoFooter bool `json:"noFooter" yaml:"noFooter"`
	NoLogo   bool `json:"noLogo" yaml:"noLogo"`

	// Custom Layout/Footer attributes (Omitted if empty)
	Title           string `json:"title" yaml:"title"`
	FooterContent   string `json:"footerContent" yaml:"footerContent"`
	FooterLinkTitle string `json:"footerLinkTitle" yaml:"footerLinkTitle"`
	FooterUrl       string `json:"footerUrl" yaml:"footerUrl"`
}

// Name returns the name of the task
func (n *PrintingPressGenerator) Name() string {
	return "printingpress"
}

func (n *PrintingPressGenerator) GetOutputName() string {
	return n.OutputName
}

func (n *PrintingPressGenerator) Generate(opts GenerateOptions) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(opts.OutputDirectory, os.ModePerm); err != nil {
		return err
	}

	// Generate docs using the CLI
	return n.generateDocs(opts)
}

func (n *PrintingPressGenerator) generateDocs(opts GenerateOptions) error {
	_, err := exec.LookPath("ppress")
	if err != nil {
		return ErrPrintingPressNotInstalled
	}

	args := []string{
		"--output", opts.OutputDirectory,
	}

	if n.Config.NoJson {
		args = append(args, "--no-json")
	}
	if n.Config.NoHtml {
		args = append(args, "--no-html")
	}
	if n.Config.NoLlm {
		args = append(args, "--no-llm")
	}
	if n.Config.NoFooter {
		args = append(args, "--no-footer")
	}
	if n.Config.NoLogo {
		args = append(args, "--no-logo")
	}

	if n.Config.Title != "" {
		args = append(args, "--title", n.Config.Title)
	}
	if n.Config.FooterContent != "" {
		args = append(args, "--footer-content", n.Config.FooterContent)
	}
	if n.Config.FooterLinkTitle != "" {
		args = append(args, "--footer-link-title", n.Config.FooterLinkTitle)
	}
	if n.Config.FooterUrl != "" {
		args = append(args, "--footer-url", n.Config.FooterUrl)
	}

	if len(n.Args) > 0 {
		args = append(args, n.Args...)
	}
	args = append(args, n.APISpec)

	cmd := exec.Command("ppress", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logging.Trace("calling ppress to generate openapi docs", "cmd", cmd.String())

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
