package orchestrator

import (
	"fmt"
	"strings"

	"cloudcanvas-backend/internal/aws/validation"
	"cloudcanvas-backend/internal/diagram/graph"
	"cloudcanvas-backend/internal/diagram/parser"
	mapper "cloudcanvas-backend/internal/mapper"
)

// ────────────────────────────────────────────────────────────
// Result types
// ────────────────────────────────────────────────────────────

// ResourceResult holds the mapped domain resource and any validation error for it.
type ResourceResult struct {
	Resource        *mapper.Resource
	ValidationError error // nil means valid
}

// OrchestrationResult is the final output produced by the pipeline.
type OrchestrationResult struct {
	// Graph is the normalized, in-memory representation of the diagram.
	Graph *graph.DiagramGraph

	// Resources contains one entry per non-visual-only diagram node.
	Resources []ResourceResult

	// Errors is a flat list of all accumulated warnings or hard errors.
	Errors []string
}

// IsValid returns true only when no validation errors were found across all resources.
func (r *OrchestrationResult) IsValid() bool {
	for _, res := range r.Resources {
		if res.ValidationError != nil {
			return false
		}
	}
	return true
}

// ────────────────────────────────────────────────────────────
// Resource-type normalizer
// Maps frontend canvas resource type strings (e.g. "route-table", "security-group")
// to the canonical keys used in the AWS validator registry (e.g. "route_table").
// ────────────────────────────────────────────────────────────

var resourceTypeNormalizer = map[string]string{
	// Networking
	"vpc":              "vpc",
	"subnet":           "subnet",
	"security-group":   "security_group",
	"security_group":   "security_group",
	"internet-gateway": "internet_gateway",
	"internet_gateway": "internet_gateway",
	"nat-gateway":      "nat_gateway",
	"nat_gateway":      "nat_gateway",
	"elastic-ip":       "elastic_ip",
	"elastic_ip":       "elastic_ip",
	"route-table":      "route_table",
	"route_table":      "route_table",
	"network-acl":      "network_acl",
	"network_acl":      "network_acl",
	"network-interface": "network_interface",
	"network_interface": "network_interface",
	"vpc-endpoint":     "vpc_endpoint",
	"vpc_endpoint":     "vpc_endpoint",

	// Storage
	"s3":           "s3_bucket",
	"s3-bucket":    "s3_bucket",
	"s3_bucket":    "s3_bucket",
	"ebs":          "ebs_volume",
	"ebs-volume":   "ebs_volume",
	"ebs_volume":   "ebs_volume",

	// IAM
	"iam-role":             "iam_role",
	"iam_role":             "iam_role",
	"iam-policy":           "iam_policy",
	"iam_policy":           "iam_policy",
	"iam-user":             "iam_user",
	"iam_user":             "iam_user",
	"iam-group":            "iam_group",
	"iam_group":            "iam_group",
	"iam-instance-profile": "iam_instance_profile",
	"iam_instance_profile": "iam_instance_profile",

	// Compute
	"ec2":               "ec2_instance",
	"ec2-instance":      "ec2_instance",
	"ec2_instance":      "ec2_instance",
	"lambda":            "lambda_function",
	"lambda-function":   "lambda_function",
	"lambda_function":   "lambda_function",
	"launch-template":   "launch_template",
	"launch_template":   "launch_template",
	"load-balancer":     "load_balancer",
	"load_balancer":     "load_balancer",
	"lb-listener":       "lb_listener",
	"lb_listener":       "lb_listener",
	"lb-target-group":   "lb_target_group",
	"lb_target_group":   "lb_target_group",
	"autoscaling-group": "autoscaling_group",
	"autoscaling_group": "autoscaling_group",
	"autoscaling-policy": "autoscaling_policy",
	"autoscaling_policy": "autoscaling_policy",
}

// normalizeResourceType converts a frontend canvas type key to its AWS validator key.
// If no mapping is found it falls back to replacing hyphens with underscores and lowercasing.
func normalizeResourceType(rawType string) string {
	lower := strings.ToLower(rawType)
	if canonical, ok := resourceTypeNormalizer[lower]; ok {
		return canonical
	}
	// Graceful fallback
	return strings.ReplaceAll(lower, "-", "_")
}

// ────────────────────────────────────────────────────────────
// nodeToResource converts a single graph.Node to a mapper.Resource
// ────────────────────────────────────────────────────────────

func nodeToResource(node *graph.Node, g *graph.DiagramGraph) *mapper.Resource {
	// Resolve region from the node's config or by walking up to the closest ancestor that has one
	region := resolveRegion(node, g)

	// Collect dependency IDs from dependency-typed edges where this node is the source
	dependsOn := make([]string, 0)
	for _, edge := range g.Edges {
		if edge.IsDependency() && edge.Source == node.ID {
			dependsOn = append(dependsOn, edge.Target)
		}
	}

	// Copy node config into metadata (strip internal/visual bookkeeping keys)
	metadata := make(map[string]interface{})
	for k, v := range node.Config {
		if k == "_varRefs" {
			continue
		}
		metadata[k] = v
	}

	// Propagate region into metadata so AWS models can use it
	if region != "" {
		metadata["region"] = region
	}

	// Build ResourceType from the normalised type string
	canonicalType := normalizeResourceType(node.ResourceType)
	resType := mapper.ResourceType{
		ID:   node.ID,
		Name: canonicalType,
	}

	var parentID *string
	if node.ParentID != nil {
		parentID = node.ParentID
	}

	return &mapper.Resource{
		ID:        node.ID,
		Name:      node.Label,
		Type:      resType,
		Provider:  "aws",
		Region:    region,
		ParentID:  parentID,
		DependsOn: dependsOn,
		Metadata:  metadata,
	}
}

// resolveRegion walks the node's ancestors to find a region value.
func resolveRegion(node *graph.Node, g *graph.DiagramGraph) string {
	// 1. Region directly inside node config
	if r, ok := node.Config["region"].(string); ok && r != "" {
		return r
	}
	if r, ok := node.Config["regionId"].(string); ok && r != "" {
		return r
	}

	// 2. Walk up through parent nodes
	current := node
	for current.ParentID != nil {
		parent, exists := g.GetNode(*current.ParentID)
		if !exists {
			break
		}
		if parent.IsRegion() {
			if name, ok := parent.Config["name"].(string); ok && name != "" {
				return name
			}
		}
		if r, ok := parent.Config["region"].(string); ok && r != "" {
			return r
		}
		current = parent
	}
	return ""
}

// ────────────────────────────────────────────────────────────
// Pipeline steps
// ────────────────────────────────────────────────────────────

// Step1_Parse parses raw IR JSON into an IRDiagram.
func Step1_Parse(jsonData []byte) (*parser.IRDiagram, error) {
	ir, err := parser.ParseIRDiagram(jsonData)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return ir, nil
}

// Step2_Normalize converts an IRDiagram into a normalized DiagramGraph.
func Step2_Normalize(ir *parser.IRDiagram) (*graph.DiagramGraph, error) {
	g, err := parser.NormalizeToGraph(ir)
	if err != nil {
		return nil, fmt.Errorf("normalize: %w", err)
	}
	return g, nil
}

// Step3_ToResources maps each non-visual-only graph node to a domain mapper.Resource.
func Step3_ToResources(g *graph.DiagramGraph) []*mapper.Resource {
	resources := make([]*mapper.Resource, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		// Skip visual-only decorators (regions are kept since they provide region context)
		if node.IsVisualOnly {
			continue
		}
		// Skip the region container node itself – it's structural, not a real resource
		if node.IsRegion() {
			continue
		}
		resources = append(resources, nodeToResource(node, g))
	}
	return resources
}

// Step4_Validate runs AWS struct-driven validation on each resource's metadata.
// Returns a slice of ResourceResult, one per resource.
func Step4_Validate(resources []*mapper.Resource) []ResourceResult {
	results := make([]ResourceResult, 0, len(resources))
	for _, res := range resources {
		err := validation.Validate(res.Type.Name, res.Metadata)
		results = append(results, ResourceResult{
			Resource:        res,
			ValidationError: err,
		})
	}
	return results
}

// ────────────────────────────────────────────────────────────
// Orchestrate – full pipeline in one call
// ────────────────────────────────────────────────────────────

// Orchestrate runs the complete pipeline:
//  1. Parse IR JSON
//  2. Normalize to graph
//  3. Map graph nodes → domain Resources
//  4. Validate each resource against its AWS model's Validate() method
//
// It always returns an OrchestrationResult even when validation errors are present
// so callers can inspect partial results. A non-nil error is returned only for
// hard failures (parse / normalize failures).
func Orchestrate(jsonData []byte) (*OrchestrationResult, error) {
	result := &OrchestrationResult{
		Errors: make([]string, 0),
	}

	// Step 1 – Parse
	ir, err := Step1_Parse(jsonData)
	if err != nil {
		return nil, err
	}

	// Step 2 – Normalize to graph
	g, err := Step2_Normalize(ir)
	if err != nil {
		return nil, err
	}
	result.Graph = g

	// Step 3 – Map nodes to domain resources
	domainResources := Step3_ToResources(g)

	// Step 4 – Validate each resource
	result.Resources = Step4_Validate(domainResources)

	// Aggregate human-readable errors for callers that just want a summary
	for _, res := range result.Resources {
		if res.ValidationError != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("[%s] %s: %v",
					res.Resource.Type.Name,
					res.Resource.Name,
					res.ValidationError,
				),
			)
		}
	}

	return result, nil
}
