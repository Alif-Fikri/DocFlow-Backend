package converter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func (c *Converter) MergePDF(inputPaths []string, outPath string) error {
	if len(inputPaths) < 2 {
		return fmt.Errorf("merge requires at least 2 files")
	}
	if err := api.MergeCreateFile(inputPaths, outPath, false, nil); err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}
	return nil
}

func (c *Converter) SplitPDF(inputPath, outDir string) ([]string, error) {
	if err := api.SplitFile(inputPath, outDir, 1, nil); err != nil {
		return nil, fmt.Errorf("split failed: %w", err)
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return nil, fmt.Errorf("read split output: %w", err)
	}
	var parts []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".pdf" {
			parts = append(parts, filepath.Join(outDir, e.Name()))
		}
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("split produced no pages")
	}
	return parts, nil
}

func (c *Converter) CompressPDF(inputPath, outPath string) error {
	if err := api.OptimizeFile(inputPath, outPath, nil); err != nil {
		return fmt.Errorf("compress failed: %w", err)
	}
	return nil
}
