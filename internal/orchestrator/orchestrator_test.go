package orchestrator

import (
	"os"
	"testing"
)

// pathToSampleIR points to the shared sample diagram shipped with the repo.
const pathToSampleIR = "../../json-ir-data/digram-01.json"

// ────────────────────────────────────────────────────────────
// Helper
// ────────────────────────────────────────────────────────────

func loadSampleIR(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(pathToSampleIR)
	if err != nil {
		t.Fatalf("Could not read sample IR file: %v", err)
	}
	return data
}

// ────────────────────────────────────────────────────────────
// Step-level tests
// ────────────────────────────────────────────────────────────

func TestStep1_Parse(t *testing.T) {
	data := loadSampleIR(t)

	ir, err := Step1_Parse(data)
	if err != nil {
		t.Fatalf("Step1_Parse failed: %v", err)
	}

	if len(ir.Nodes) == 0 {
		t.Error("Expected at least one node after parse")
	}
	t.Logf("Parsed %d nodes, %d edges, %d variables", len(ir.Nodes), len(ir.Edges), len(ir.Variables))
}

func TestStep2_Normalize(t *testing.T) {
	data := loadSampleIR(t)

	ir, err := Step1_Parse(data)
	if err != nil {
		t.Fatalf("Step1_Parse failed: %v", err)
	}

	g, err := Step2_Normalize(ir)
	if err != nil {
		t.Fatalf("Step2_Normalize failed: %v", err)
	}

	if len(g.Nodes) == 0 {
		t.Error("Expected at least one node in graph")
	}
	t.Logf("Graph: %d nodes, %d edges", len(g.Nodes), len(g.Edges))
}

func TestStep3_ToResources(t *testing.T) {
	data := loadSampleIR(t)

	ir, _ := Step1_Parse(data)
	g, _ := Step2_Normalize(ir)

	resources := Step3_ToResources(g)

	if len(resources) == 0 {
		t.Error("Expected at least one mapped resource")
	}

	for _, res := range resources {
		t.Logf("  resource id=%s type=%s name=%q region=%q parentId=%v",
			res.ID, res.Type.Name, res.Name, res.Region, res.ParentID)

		if res.Type.Name == "" {
			t.Errorf("Resource %s has empty Type.Name", res.ID)
		}
	}
}

func TestStep4_Validate(t *testing.T) {
	data := loadSampleIR(t)

	ir, _ := Step1_Parse(data)
	g, _ := Step2_Normalize(ir)
	resources := Step3_ToResources(g)
	results := Step4_Validate(resources)

	for _, res := range results {
		status := "✓ valid"
		if res.ValidationError != nil {
			status = "✗ " + res.ValidationError.Error()
		}
		t.Logf("  [%s] %s → %s", res.Resource.Type.Name, res.Resource.Name, status)
	}
}

// ────────────────────────────────────────────────────────────
// Full pipeline test
// ────────────────────────────────────────────────────────────

func TestOrchestrate_SampleDiagram(t *testing.T) {
	data := loadSampleIR(t)

	result, err := Orchestrate(data)
	if err != nil {
		t.Fatalf("Orchestrate returned a hard error: %v", err)
	}

	if result.Graph == nil {
		t.Fatal("Expected non-nil graph in result")
	}

	if len(result.Resources) == 0 {
		t.Fatal("Expected at least one resource in result")
	}

	t.Logf("Pipeline completed: %d resources, %d validation errors",
		len(result.Resources), len(result.Errors))

	for _, e := range result.Errors {
		t.Logf("  validation error: %s", e)
	}
}

// TestOrchestrate_InvalidPayload verifies that a completely broken JSON
// results in a hard parse error rather than a silent empty result.
func TestOrchestrate_InvalidPayload(t *testing.T) {
	_, err := Orchestrate([]byte(`{broken json`))
	if err == nil {
		t.Error("Expected parse error for invalid JSON input")
	}
}

// TestOrchestrate_EmptyDiagram verifies graceful handling of a diagram with no nodes.
func TestOrchestrate_EmptyDiagram(t *testing.T) {
	emptyDiagram := []byte(`{"nodes":[],"edges":[],"variables":[],"outputs":[],"timestamp":0}`)

	result, err := Orchestrate(emptyDiagram)
	if err != nil {
		t.Fatalf("Orchestrate returned unexpected hard error: %v", err)
	}

	if len(result.Resources) != 0 {
		t.Errorf("Expected 0 resources for empty diagram, got %d", len(result.Resources))
	}

	if !result.IsValid() {
		t.Error("Empty diagram should be considered valid")
	}
}

// TestNormalizeResourceType verifies the type normalizer covers common aliases.
func TestNormalizeResourceType(t *testing.T) {
	cases := map[string]string{
		"ec2":            "ec2_instance",
		"EC2":            "ec2_instance",
		"security-group": "security_group",
		"vpc":            "vpc",
		"route-table":    "route_table",
		"nat-gateway":    "nat_gateway",
	}

	for input, expected := range cases {
		got := normalizeResourceType(input)
		if got != expected {
			t.Errorf("normalizeResourceType(%q) = %q, want %q", input, got, expected)
		}
	}
}
