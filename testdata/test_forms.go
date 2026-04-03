//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/form"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func main() {
	files, _ := filepath.Glob("testdata/*.pdf")
	for _, f := range files {
		fmt.Printf("=== %s ===\n", filepath.Base(f))
		data, _ := os.ReadFile(f)

		conf := model.NewDefaultConfiguration()

		// Step 1: List fields
		fields, err := api.FormFields(bytes.NewReader(data), conf)
		if err != nil {
			fmt.Printf("  LIST ERROR: %v\n", err)
			continue
		}
		fmt.Printf("  LIST OK: %d fields\n", len(fields))

		if len(fields) == 0 {
			continue
		}

		// Step 2: Build fill data — fill first text field found
		fg := form.FormGroup{Forms: []form.Form{{}}}
		filled := false
		for _, fld := range fields {
			if fld.Typ == form.FTText {
				fg.Forms[0].TextFields = append(fg.Forms[0].TextFields, &form.TextField{
					Name:  fld.Name,
					Value: "Test Value",
				})
				filled = true
				break
			}
		}
		if !filled {
			fmt.Printf("  SKIP FILL: no text fields\n")
			continue
		}

		fgJSON, _ := json.Marshal(fg)

		// Step 3: Fill form
		fillConf := model.NewDefaultConfiguration()
		fillConf.Cmd = model.FILLFORMFIELDS
		var out bytes.Buffer
		err = api.FillForm(bytes.NewReader(data), bytes.NewReader(fgJSON), &out, fillConf)
		if err != nil {
			fmt.Printf("  FILL ERROR: %v\n", err)
			continue
		}
		fmt.Printf("  FILL OK: %d bytes output\n", out.Len())

		// Step 4: Verify the filled PDF can be read back
		exportConf := model.NewDefaultConfiguration()
		var exportBuf bytes.Buffer
		err = api.ExportFormJSON(bytes.NewReader(out.Bytes()), &exportBuf, "", exportConf)
		if err != nil {
			fmt.Printf("  EXPORT ERROR: %v\n", err)
			continue
		}
		fmt.Printf("  EXPORT OK: verified\n")
	}
}
