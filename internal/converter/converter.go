package converter

import (
	"fmt"
	"strings"

	"github.com/Alif-Fikri/DocFlow-Backend/internal/config"
)

type Converter struct {
	cfg config.Config
}

func New(cfg config.Config) *Converter {
	return &Converter{cfg: cfg}
}

var SupportedTargets = map[string]bool{
	"pdf":  true,
	"docx": true,
	"xlsx": true,
	"pptx": true,
	"txt":  true,
	"odt":  true,
	"ods":  true,
	"odp":  true,
	"rtf":  true,
	"csv":  true,
	"html": true,
}

func ValidateTarget(target string) error {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return fmt.Errorf("target format is required")
	}
	if !SupportedTargets[target] {
		return fmt.Errorf("unsupported target format %q", target)
	}
	return nil
}
