package component

import (
	"fmt"
	"io"

	"github.com/neticdk-k8s/ic/internal/render"
	"github.com/neticdk-k8s/ic/internal/ui"
)

type Renderer interface {
	// Render renders the component
	Render(format string) error
}

type renderer struct {
	data   []byte
	writer io.Writer
}

type componentRenderer struct {
	renderer
	component *componentResponse
}

// NewComponentRenderer creates a new renderer of a single component
func NewComponentRenderer(component *componentResponse, jsonData []byte, writer io.Writer) *componentRenderer {
	cr := &componentRenderer{
		renderer: renderer{
			data:   jsonData,
			writer: writer,
		},
		component: component,
	}
	return cr
}

// Render renders the component
func (r *componentRenderer) Render(format string) error {
	switch format {
	case "json":
		return r.renderJSON()
	case "plain", "table":
		return r.renderText()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *componentRenderer) renderText() error {
	data := [][]string{
		{"Name:", r.component.Namespace},
		{"Namespace:", r.component.Name},
		{"Description:", r.component.Description},
		{"Type:", r.component.ComponentType},
		{"Source:", r.component.Source},
	}
	ui.RenderKVTable(r.writer, "Base Information", data)

	fmt.Fprintln(r.writer)
	fmt.Fprintln(r.writer, "Resilience Zones:")
	rzsHeaders := []string{"name", "version"}
	rzsTable := ui.NewTable(r.writer, rzsHeaders)
	for _, c := range r.component.ResilienceZones {
		_ = rzsTable.Append(
			[]string{
				c.Name,
				c.Version,
			},
		)
	}
	_ = rzsTable.Render()

	fmt.Fprintln(r.writer)
	fmt.Fprintln(r.writer, "Clusters:")
	clustersHeaders := []string{"name"}
	clustersTable := ui.NewTable(r.writer, clustersHeaders)
	for _, c := range r.component.Clusters {
		_ = clustersTable.Append([]string{c})
	}
	_ = clustersTable.Render()

	return nil
}

func (r *componentRenderer) renderJSON() error {
	return render.PrettyPrintJSON(r.data, r.writer)
}

type componentsRenderer struct {
	renderer
	noHeaders  bool
	components *componentListResponse
}

// NewComponentsRenderer creates a new renderer for a list of components
func NewComponentsRenderer(components *componentListResponse, jsonData []byte, writer io.Writer, noHeaders bool) *componentsRenderer {
	cr := &componentsRenderer{
		renderer: renderer{
			writer: writer,
			data:   jsonData,
		},
		noHeaders:  noHeaders,
		components: components,
	}
	return cr
}

// Render renders the component list
func (r *componentsRenderer) Render(format string) error {
	switch format {
	case "json":
		return r.renderJSON()
	case "plain", "table":
		return r.renderTable()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *componentsRenderer) renderTable() error {
	var headers []string
	if !r.noHeaders {
		headers = []string{"namespace", "name", "type"}
	}
	table := ui.NewTable(r.writer, headers)
	for _, c := range r.components.Components {
		_ = table.Append(
			[]string{
				c.Namespace,
				c.Name,
				c.ComponentType,
			},
		)
	}
	_ = table.Render()
	return nil
}

func (r *componentsRenderer) renderJSON() error {
	return render.PrettyPrintJSON(r.data, r.writer)
}
