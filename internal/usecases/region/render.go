package region

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/neticdk-k8s/ic/internal/render"
	"github.com/neticdk-k8s/ic/internal/ui"
)

type Renderer interface {
	// Render renders the regions
	Render(format string) error
}

type renderer struct {
	writer io.Writer
}

type regionsRenderer struct {
	renderer
	noHeaders bool
	regions   []string
}

// NewRegionsRenderer creates a new renderer of a list of regions
func NewRegionsRenderer(regions []string, writer io.Writer, noHeaders bool) *regionsRenderer {
	cr := &regionsRenderer{
		renderer: renderer{
			writer: writer,
		},
		noHeaders: noHeaders,
		regions:   regions,
	}
	return cr
}

// Render renders the regions
func (r *regionsRenderer) Render(format string) error {
	switch format {
	case "json":
		return r.renderJSON()
	case "plain", "table":
		return r.renderText()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *regionsRenderer) renderText() error {
	var headers []string
	if !r.noHeaders {
		headers = []string{"regions"}
	}
	table := ui.NewTable(r.writer, headers)
	for _, r := range r.regions {
		_ = table.Append([]string{r})
	}
	_ = table.Render()

	return nil
}

func (r *regionsRenderer) renderJSON() error {
	body, err := json.Marshal(r.regions)
	if err != nil {
		return err
	}
	return render.PrettyPrintJSON(body, r.writer)
}
