package template

import (
	"errors"
)

var (
	ErrTemplateNotFound             = errors.New("template with the given ID is not found")
	ErrFailedToParseTemplate        = errors.New("failed to parse the template")
	ErrFailedToRenderTemplate       = errors.New("failed to render the template")
	ErrFailedToCopyTemplateFile     = errors.New("failed to copy the template file")
	ErrFailedToDownloadTemplateFile = errors.New("failed to download the template file")
	ErrTemplateFileOrUrlIsRequired  = errors.New("template has no source template or source url")
)
