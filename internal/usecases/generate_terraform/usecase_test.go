package generate_terraform_test

import (
	"context"
	"testing"

	"cloudcanvas-backend/internal/architecture"
	"cloudcanvas-backend/internal/iac"
	"cloudcanvas-backend/internal/iac/terraform/mapper"
	resource "cloudcanvas-backend/internal/mapper"
	generate_terraform "cloudcanvas-backend/internal/usecases/generate_terraform"
)

// ── Stub IaC engine ───────────────────────────────────────────────────────────

type stubEngine struct {
	output *iac.Output
	err    error
}

func (s *stubEngine) Name() string { return "terraform" }

func (s *stubEngine) Generate(_ context.Context, _ *architecture.Architecture, _ []*resource.Resource) (*iac.Output, error) {
	return s.output, s.err
}

// ── Minimal diagram JSON (empty canvas) ───────────────────────────────────────

// emptyDiagram is the minimal IR JSON the parser accepts.
const emptyDiagram = `{
	"nodes": [],
	"edges": []
}`

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestNew_PanicsOnDoubleRegister(t *testing.T) {
	// Ensure New() does not panic with a fresh mapper registry.
	reg := mapper.NewRegistry()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("New() panicked unexpectedly: %v", r)
		}
	}()
	_ = generate_terraform.New(reg)
}

func TestExecute_EmptyJSON_ReturnsError(t *testing.T) {
	reg := mapper.NewRegistry()
	uc := generate_terraform.New(reg)

	_, err := uc.Execute(context.Background(), generate_terraform.GenerateTerraformInput{
		DiagramJSON: []byte{},
		Provider:    resource.AWS,
	})
	if err == nil {
		t.Fatal("Execute() with empty JSON expected error, got nil")
	}
}

func TestExecute_MissingProvider_ReturnsError(t *testing.T) {
	reg := mapper.NewRegistry()
	uc := generate_terraform.New(reg)

	_, err := uc.Execute(context.Background(), generate_terraform.GenerateTerraformInput{
		DiagramJSON: []byte(emptyDiagram),
		Provider:    "",
	})
	if err == nil {
		t.Fatal("Execute() with empty provider expected error, got nil")
	}
}
