// Package generate_terraform provides the use case for converting a canvas
// diagram JSON payload into one or more Terraform source files.
//
// # Layer contract
//
//	JSON bytes
//	  → Step1: parser.ParseIRDiagram   (diagram/parser)
//	  → Step2: parser.NormalizeToGraph (diagram/parser)
//	  → Step3: architecture.MapDiagramToArchitecture (architecture aggregate)
//	  → Step4: codegen/terraform.Compiler.Compile  (codegen layer)
//	  → iac.Output  (iac layer)
//
// The use case is provider-aware: callers supply the target CloudProvider so
// that the correct architecture generator and Terraform mapper are selected.
package generate_terraform

import (
	"context"
	"fmt"

	"cloudcanvas-backend/internal/architecture"
	codegenTF "cloudcanvas-backend/internal/codegen/terraform"
	"cloudcanvas-backend/internal/diagram/parser"
	"cloudcanvas-backend/internal/iac"
	"cloudcanvas-backend/internal/iac/registry"
	"cloudcanvas-backend/internal/iac/terraform/generator"
	"cloudcanvas-backend/internal/iac/terraform/mapper"
	resource "cloudcanvas-backend/internal/mapper"
)

// GenerateTerraformInput holds everything the use case needs from the caller.
type GenerateTerraformInput struct {
	// DiagramJSON is the raw canvas JSON produced by the frontend (IR format).
	DiagramJSON []byte

	// Provider identifies the cloud provider to generate code for.
	// Currently only resource.AWS ("aws") is fully supported.
	Provider resource.CloudProvider
}

// GenerateTerraformOutput wraps the generated IaC files returned by the engine.
type GenerateTerraformOutput struct {
	// Files contains the rendered Terraform source files (main.tf, variables.tf, …).
	// Each entry includes the relative path and file content.
	Files []iac.GeneratedFile

	// Architecture is the intermediate Architecture aggregate built from the diagram.
	// Callers may inspect it for debugging or further processing.
	Architecture *architecture.Architecture
}

// UseCase is the primary entry point for the Terraform generation pipeline.
//
// Construct with New(), which wires the Terraform engine registry and the
// provider-specific mapper registry.
type UseCase struct {
	engineRegistry *registry.EngineRegistry
}

// New creates a UseCase with a pre-configured Terraform engine backed by the
// supplied mapper registry.
//
// Usage:
//
//	mapperReg := mapper.NewRegistry()
//	mapperReg.Register(awsmapper.New()) // register AWS Terraform mapper
//	uc := generate_terraform.New(mapperReg)
//	out, err := uc.Execute(ctx, generate_terraform.GenerateTerraformInput{...})
func New(terraformMappers *mapper.MapperRegistry) *UseCase {
	engineReg := registry.New()

	// Build and register the Terraform engine with the supplied mapper registry.
	tfEngine := generator.NewEngine(terraformMappers)
	if err := engineReg.Register(tfEngine); err != nil {
		// This only fails if the engine was already registered, which cannot
		// happen here since we just created the registry. Panic to surface bugs early.
		panic(fmt.Sprintf("generate_terraform: failed to register terraform engine: %v", err))
	}

	return &UseCase{engineRegistry: engineReg}
}

// NewWithEngine creates a UseCase with a custom pre-built IaC engine.
// Use this variant in tests or when you need to swap the engine implementation.
func NewWithEngine(engine iac.Engine) *UseCase {
	engineReg := registry.New()
	if err := engineReg.Register(engine); err != nil {
		panic(fmt.Sprintf("generate_terraform: failed to register engine: %v", err))
	}
	return &UseCase{engineRegistry: engineReg}
}

// Execute runs the full JSON → Terraform pipeline and returns the generated files.
//
// The pipeline:
//  1. Parse the raw diagram JSON into an IR representation.
//  2. Normalize the IR into a DiagramGraph (resolves parent-child, edges, etc.).
//  3. Convert the graph into an Architecture aggregate using the provider-specific
//     generator (if registered) or the default mapper.
//  4. Compile the Architecture into Terraform HCL via the codegen layer, which
//     performs topological sort then calls the IaC engine.
//
// Returns GenerateTerraformOutput on success, or a descriptive error on failure.
func (uc *UseCase) Execute(ctx context.Context, input GenerateTerraformInput) (*GenerateTerraformOutput, error) {
	if len(input.DiagramJSON) == 0 {
		return nil, fmt.Errorf("generate_terraform: diagram JSON is empty")
	}
	if input.Provider == "" {
		return nil, fmt.Errorf("generate_terraform: provider must be specified")
	}

	// ── Step 1: Parse IR JSON ─────────────────────────────────────────────────
	ir, err := parser.ParseIRDiagram(input.DiagramJSON)
	if err != nil {
		return nil, fmt.Errorf("generate_terraform: parse diagram JSON: %w", err)
	}

	// ── Step 2: Normalize to DiagramGraph ────────────────────────────────────
	diagramGraph, err := parser.NormalizeToGraph(ir)
	if err != nil {
		return nil, fmt.Errorf("generate_terraform: normalize diagram: %w", err)
	}

	// ── Step 3: Build Architecture aggregate ─────────────────────────────────
	arch, err := architecture.MapDiagramToArchitecture(diagramGraph, input.Provider)
	if err != nil {
		return nil, fmt.Errorf("generate_terraform: map diagram to architecture: %w", err)
	}

	// ── Step 4: Compile to Terraform via the codegen layer ───────────────────
	tfEngine, ok := uc.engineRegistry.Get("terraform")
	if !ok {
		return nil, fmt.Errorf("generate_terraform: terraform engine not registered")
	}

	compiler := codegenTF.NewCompiler(tfEngine)
	iacOutput, err := compiler.Compile(ctx, arch)
	if err != nil {
		return nil, fmt.Errorf("generate_terraform: compile to terraform: %w", err)
	}

	return &GenerateTerraformOutput{
		Files:        iacOutput.Files,
		Architecture: arch,
	}, nil
}
