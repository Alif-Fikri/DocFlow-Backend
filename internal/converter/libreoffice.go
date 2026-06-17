package converter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (c *Converter) Convert(ctx context.Context, inputPath, target, outDir string) (string, error) {
	target = strings.ToLower(strings.TrimSpace(target))
	if err := ValidateTarget(target); err != nil {
		return "", err
	}

	if strings.EqualFold(filepath.Ext(inputPath), ".pdf") {
		switch target {
		case "docx":
			return c.pdfToDocx(ctx, inputPath, outDir)
		case "pptx":
			return c.pdfToPptx(ctx, inputPath, outDir)
		case "xlsx":
			return c.pdfToXlsx(ctx, inputPath, outDir)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.ConvertTimeout)
	defer cancel()

	profileDir, err := os.MkdirTemp(c.cfg.WorkDir, "lo-profile-")
	if err != nil {
		return "", fmt.Errorf("create profile dir: %w", err)
	}
	defer os.RemoveAll(profileDir)

	profileURI := "file://" + filepath.ToSlash(profileDir)

	args := []string{
		"--headless",
		"--norestore",
		"--nolockcheck",
		"-env:UserInstallation=" + profileURI,
		"--convert-to", target,
		"--outdir", outDir,
		inputPath,
	}

	cmd := exec.CommandContext(ctx, c.cfg.SofficeBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("conversion timed out after %s", c.cfg.ConvertTimeout)
		}
		return "", fmt.Errorf("libreoffice failed: %v: %s", err, strings.TrimSpace(string(out)))
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	produced := filepath.Join(outDir, base+"."+target)
	if _, err := os.Stat(produced); err != nil {
		return "", fmt.Errorf("expected output %q not found (soffice output: %s)", produced, strings.TrimSpace(string(out)))
	}
	return produced, nil
}
