package cluster

import (
	"fmt"
	"io"

	"github.com/neticdk-k8s/ic/internal/render"
	"github.com/neticdk-k8s/ic/internal/ui"
	"sigs.k8s.io/yaml"
)

const (
	FormatJson  = "json"
	FormatTable = "table"
	FormatPlain = "plain"
)

type Renderer interface {
	// Render renders the cluster
	Render(format string) error
}

type renderer struct {
	data   []byte
	writer io.Writer
}

type clusterRenderer struct {
	renderer
	cluster *clusterResponse
}

// NewClusterRenderer creates a new renderer of a single cluster
func NewClusterRenderer(cluster *clusterResponse, jsonData []byte, writer io.Writer) *clusterRenderer {
	cr := &clusterRenderer{
		renderer: renderer{
			data:   jsonData,
			writer: writer,
		},
		cluster: cluster,
	}
	return cr
}

// Render renders the cluster
func (r *clusterRenderer) Render(format string) error {
	switch format {
	case FormatJson:
		return r.renderJSON()
	case FormatPlain, FormatTable:
		return r.renderText()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *clusterRenderer) renderText() error {
	data := [][]string{
		{"ID:", fmt.Sprintf("%s.%s", r.cluster.Name, r.cluster.ProviderName)},
		{"Name:", r.cluster.Name},
		{"Provider:", r.cluster.ProviderName},
		{"NRN:", r.cluster.NRN},
		{"Description:", r.cluster.Description},
		{"Type:", r.cluster.ClusterType},
		{"Partition:", r.cluster.Partition},
		{"Region:", r.cluster.Region},
		{"Environment:", r.cluster.EnvironmentName},
		{"Resilience Zone:", r.cluster.ResilienceZone},
		{"Infrastructure Provider:", r.cluster.InfrastructureProvider},
		{"Kubernetes Provider:", r.cluster.KubernetesProvider},
		{"Kubernetes Version:", r.cluster.KubernetesVersion},
		{"Client Version:", r.cluster.ClientVersion},
	}
	ui.RenderKVTable(r.writer, "Base Information", data)

	if r.cluster.ControlPlaneCapacity != nil {
		allocMem, unit := render.BytesToBinarySI(r.cluster.ControlPlaneCapacity.MemoryBytes)
		data = [][]string{
			{"Nodes:", fmt.Sprintf("%d", r.cluster.ControlPlaneCapacity.NodeCount)},
			{"Allocatable CPU:", fmt.Sprintf("%dm", r.cluster.ControlPlaneCapacity.CoresMillis)},
			{"Allocatable Memory:", fmt.Sprintf("%.f%s", allocMem, unit)},
		}
		ui.RenderKVTable(r.writer, "Control Plane Capacity", data)
	}

	if r.cluster.WorkerNodesCapacity != nil {
		allocMem, unit := render.BytesToBinarySI(r.cluster.WorkerNodesCapacity.MemoryBytes)
		data = [][]string{
			{"Nodes:", fmt.Sprintf("%d", r.cluster.WorkerNodesCapacity.NodeCount)},
			{"Allocatable CPU:", fmt.Sprintf("%dm", r.cluster.WorkerNodesCapacity.CoresMillis)},
			{"Allocatable Memory:", fmt.Sprintf("%.f%s", allocMem, unit)},
		}
		ui.RenderKVTable(r.writer, "Worker Nodes Capacity", data)
	}

	return nil
}

func (r *clusterRenderer) renderJSON() error {
	return render.PrettyPrintJSON(r.data, r.writer)
}

type clustersRenderer struct {
	renderer
	noHeaders bool
	clusters  *clusterListResponse
}

// NewClustersRenderer creates a new renderer for a list of clusters
func NewClustersRenderer(clusters *clusterListResponse, jsonData []byte, writer io.Writer, noHeaders bool) *clustersRenderer {
	cr := &clustersRenderer{
		renderer: renderer{
			writer: writer,
			data:   jsonData,
		},
		noHeaders: noHeaders,
		clusters:  clusters,
	}
	return cr
}

// Render renders the cluster list
func (r *clustersRenderer) Render(format string) error {
	switch format {
	case FormatJson:
		return r.renderJSON()
	case FormatPlain, FormatTable:
		return r.renderTable()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *clustersRenderer) renderTable() error {
	var headers []string
	if !r.noHeaders {
		headers = []string{"provider", "id", "rz", "version"}
	}
	table := ui.NewTable(r.writer, headers)
	for _, c := range r.clusters.Clusters {
		_ = table.Append(
			[]string{
				c.ProviderName,
				fmt.Sprintf("%s.%s", c.Name, c.ProviderName),
				c.ResilienceZone,
				c.KubernetesVersion,
			},
		)
	}
	_ = table.Render()
	return nil
}

func (r *clustersRenderer) renderJSON() error {
	return render.PrettyPrintJSON(r.data, r.writer)
}

type clusterNodesRenderer struct {
	renderer
	noHeaders bool
	nodes     *clusterNodesListResponse
}

// NewClusterNodesRenderer creates a new renderer for a list of cluster nodes
func NewClusterNodesRenderer(nodes *clusterNodesListResponse, jsonData []byte, writer io.Writer, noHeaders bool) *clusterNodesRenderer {
	cnr := &clusterNodesRenderer{
		renderer: renderer{
			writer: writer,
			data:   jsonData,
		},
		noHeaders: noHeaders,
		nodes:     nodes,
	}
	return cnr
}

// Render renders the cluster node list
func (r *clusterNodesRenderer) Render(format string) error {
	switch format {
	case FormatJson:
		return r.renderJSON()
	case FormatPlain, FormatTable:
		return r.renderTable()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *clusterNodesRenderer) renderTable() error {
	var headers []string
	if !r.noHeaders {
		headers = []string{"name", "cp", "kubelet", "cpu (alloc)", "mem (alloc)", "cpu (cap)", "mem (cap)"}
	}
	table := ui.NewTable(r.writer, headers)
	for _, n := range r.nodes.Nodes {
		allocMem, allocMemUnit := render.BytesToBinarySI(int64(n.AllocatableMemoryBytes))
		capMem, capMemUnit := render.BytesToBinarySI(int64(n.CapacityMemoryBytes))
		_ = table.Append(
			[]string{
				n.Name,
				fmt.Sprintf("%t", n.IsControlPlane),
				n.KubeletVersion,
				fmt.Sprintf("%dm", int64(n.AllocatableCPUMillis)),
				fmt.Sprintf("%.f%s", allocMem, allocMemUnit),
				fmt.Sprintf("%dm", int64(n.CapacityCPUMillis)),
				fmt.Sprintf("%.f%s", capMem, capMemUnit),
			},
		)
	}
	_ = table.Render()
	return nil
}

func (r *clusterNodesRenderer) renderJSON() error {
	return render.PrettyPrintJSON(r.data, r.writer)
}

type clusterNodeRenderer struct {
	renderer
	node *clusterNodeResponse
}

// NewClusterNodeRenderer creates a new renderer of a single cluster node
func NewClusterNodeRenderer(node *clusterNodeResponse, jsonData []byte, writer io.Writer) *clusterNodeRenderer {
	cnr := &clusterNodeRenderer{
		renderer: renderer{
			data:   jsonData,
			writer: writer,
		},
		node: node,
	}
	return cnr
}

// Render renders the cluster node
func (r *clusterNodeRenderer) Render(format string) error {
	switch format {
	case FormatJson:
		return r.renderJSON()
	case FormatPlain, FormatTable:
		return r.renderText()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (r *clusterNodeRenderer) renderText() error {
	allocMem, allocMemUnit := render.BytesToBinarySI(int64(r.node.AllocatableMemoryBytes))
	capMem, capMemUnit := render.BytesToBinarySI(int64(r.node.CapacityMemoryBytes))
	data := [][]string{
		{"Name:", r.node.Name},
		{"Role:", r.node.Role},
		{"KubeProxyVersion:", r.node.KubeProxyVersion},
		{"KubeletVersion:", r.node.KubeletVersion},
		{"KernelVersion:", r.node.KernelVersion},
		{"CRI Name:", r.node.CRIName},
		{"CRI Version:", r.node.CRIVersion},
		{"Container Runtime:", r.node.ContainerRuntimeVersion},
		{"Control Panel", fmt.Sprintf("%t", r.node.IsControlPlane)},
		{"Provider", r.node.Provider},
		{"Topology Region", r.node.TopologyRegion},
		{"Topology Zone", r.node.TopologyZone},
		{"CPU (Alloc)", fmt.Sprintf("%dm", int64(r.node.AllocatableCPUMillis))},
		{"Memory (Alloc)", fmt.Sprintf("%.f%s", allocMem, allocMemUnit)},
		{"CPU (Cap)", fmt.Sprintf("%dm", int64(r.node.CapacityCPUMillis))},
		{"Memory (Cap)", fmt.Sprintf("%.f%s", capMem, capMemUnit)},
	}
	ui.RenderKVTable(r.writer, "Node Information", data)

	return nil
}

func (r *clusterNodeRenderer) renderJSON() error {
	return render.PrettyPrintJSON(r.data, r.writer)
}

type clusterKubeConfigRenderer struct {
	renderer
}

// NewClusterKubeConfigRenderer creates a new renderer of a cluster kubeconfig
func NewClusterKubeConfigRenderer(data []byte, writer io.Writer) *clusterKubeConfigRenderer {
	r := &clusterKubeConfigRenderer{
		renderer: renderer{
			data:   data,
			writer: writer,
		},
	}
	return r
}

// Render renders the cluster node
func (r *clusterKubeConfigRenderer) Render(format string) error {
	switch format {
	case FormatJson:
		return r.renderJSON()
	default:
		render.String(r.data, r.writer)
		return nil
	}
}

func (r *clusterKubeConfigRenderer) renderJSON() error {
	jsonData, err := yaml.YAMLToJSON(r.data)
	if err != nil {
		return err
	}
	return render.PrettyPrintJSON(jsonData, r.writer)
}
