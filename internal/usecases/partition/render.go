package partition

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/neticdk-k8s/ic/internal/render"
	"github.com/neticdk-k8s/ic/internal/ui"
)

type Renderer interface {
	// Render renders the partitions
	Render(format string) error
}

type renderer struct {
	writer io.Writer
}

type partitionsRenderer struct {
	renderer
	noHeaders  bool
	partitions []string
}

// NewPartitionsRenderer creates a new renderer of a list of partitions
func NewPartitionsRenderer(partitions []string, writer io.Writer, noHeaders bool) *partitionsRenderer {
	cr := &partitionsRenderer{
		renderer: renderer{
			writer: writer,
		},
		noHeaders:  noHeaders,
		partitions: partitions,
	}
	return cr
}

// Render renders the partitions
func (r *partitionsRenderer) Render(format string) error {
	switch format {
	case "json":
		return r.renderJSON()
	case "plain", "table":
		return r.renderText()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *partitionsRenderer) renderText() error {
	var headers []string
	if !r.noHeaders {
		headers = []string{"partitions"}
	}
	table := ui.NewTable(r.writer, headers)
	for _, r := range r.partitions {
		_ = table.Append([]string{r})
	}
	_ = table.Render()

	return nil
}

func (r *partitionsRenderer) renderJSON() error {
	body, err := json.Marshal(r.partitions)
	if err != nil {
		return err
	}
	return render.PrettyPrintJSON(body, r.writer)
}
