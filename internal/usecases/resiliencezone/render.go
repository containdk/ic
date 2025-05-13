package resiliencezone

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/neticdk-k8s/ic/internal/render"
	"github.com/neticdk-k8s/ic/internal/ui"
)

type Renderer interface {
	// Render renders the resilience zones
	Render(format string) error
}

type renderer struct {
	writer io.Writer
}

type resilienceZonesRenderer struct {
	renderer
	noHeaders       bool
	resilienceZones []string
}

// NewResilienceZonesRenderer creates a new renderer of a list of resilienceZones
func NewResilienceZonesRenderer(resilienceZones []string, writer io.Writer, noHeaders bool) *resilienceZonesRenderer {
	cr := &resilienceZonesRenderer{
		renderer: renderer{
			writer: writer,
		},
		noHeaders:       noHeaders,
		resilienceZones: resilienceZones,
	}
	return cr
}

// Render renders the resilienceZones
func (r *resilienceZonesRenderer) Render(format string) error {
	switch format {
	case "json":
		return r.renderJSON()
	case "plain", "table":
		return r.renderText()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *resilienceZonesRenderer) renderText() error {
	var headers []string
	if !r.noHeaders {
		headers = []string{"resilience zones"}
	}
	table := ui.NewTable(r.writer, headers)
	for _, r := range r.resilienceZones {
		_ = table.Append([]string{r})
	}
	_ = table.Render()

	return nil
}

func (r *resilienceZonesRenderer) renderJSON() error {
	body, err := json.Marshal(r.resilienceZones)
	if err != nil {
		return err
	}
	return render.PrettyPrintJSON(body, r.writer)
}
