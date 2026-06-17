package converter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const pdf2docxScript = `import sys
from pdf2docx import Converter

src, dst = sys.argv[1], sys.argv[2]
cv = Converter(src)
try:
    cv.convert(dst)
finally:
    cv.close()
`

const pdf2pptxScript = `import sys, os, tempfile
import fitz
from pptx import Presentation
from pptx.util import Emu

src, dst = sys.argv[1], sys.argv[2]
doc = fitz.open(src)
prs = Presentation()
blank = prs.slide_layouts[6]
if doc.page_count:
    r = doc[0].rect
    prs.slide_width = Emu(int(r.width * 12700))
    prs.slide_height = Emu(int(r.height * 12700))
tmp = tempfile.mkdtemp()
for i in range(doc.page_count):
    pix = doc[i].get_pixmap(dpi=150)
    img = os.path.join(tmp, "p%d.png" % i)
    pix.save(img)
    slide = prs.slides.add_slide(blank)
    slide.shapes.add_picture(img, 0, 0, width=prs.slide_width, height=prs.slide_height)
prs.save(dst)
`

const pdf2xlsxScript = `import sys
import pdfplumber
from openpyxl import Workbook

src, dst = sys.argv[1], sys.argv[2]
wb = Workbook()
wb.remove(wb.active)
with pdfplumber.open(src) as pdf:
    for pi, page in enumerate(pdf.pages, start=1):
        ws = wb.create_sheet(title=("Page %d" % pi)[:31])
        tables = page.extract_tables()
        if tables:
            for table in tables:
                for row in table:
                    ws.append(["" if c is None else c for c in row])
                ws.append([])
        else:
            for line in (page.extract_text() or "").splitlines():
                ws.append([line])
if not wb.sheetnames:
    wb.create_sheet(title="Sheet1")
wb.save(dst)
`

func (c *Converter) pdfToDocx(ctx context.Context, inputPath, outDir string) (string, error) {
	return c.runPythonConversion(ctx, "_pdf2docx.py", pdf2docxScript, inputPath, outDir, "docx")
}

func (c *Converter) pdfToPptx(ctx context.Context, inputPath, outDir string) (string, error) {
	return c.runPythonConversion(ctx, "_pdf2pptx.py", pdf2pptxScript, inputPath, outDir, "pptx")
}

func (c *Converter) pdfToXlsx(ctx context.Context, inputPath, outDir string) (string, error) {
	return c.runPythonConversion(ctx, "_pdf2xlsx.py", pdf2xlsxScript, inputPath, outDir, "xlsx")
}

func (c *Converter) runPythonConversion(
	ctx context.Context,
	scriptName, scriptBody, inputPath, outDir, ext string,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.ConvertTimeout)
	defer cancel()

	scriptPath := filepath.Join(outDir, scriptName)
	if err := os.WriteFile(scriptPath, []byte(scriptBody), 0o600); err != nil {
		return "", fmt.Errorf("write %s: %w", scriptName, err)
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outPath := filepath.Join(outDir, base+"."+ext)

	cmd := exec.CommandContext(ctx, c.cfg.PythonBin, scriptPath, inputPath, outPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("conversion timed out after %s", c.cfg.ConvertTimeout)
		}
		return "", fmt.Errorf("%s failed: %v: %s", scriptName, err, strings.TrimSpace(string(out)))
	}
	if _, err := os.Stat(outPath); err != nil {
		return "", fmt.Errorf("%s produced no output: %s", scriptName, strings.TrimSpace(string(out)))
	}
	return outPath, nil
}
