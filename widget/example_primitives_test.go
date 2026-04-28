package widget

import (
	"fmt"
	"time"

	"github.com/RenseiAI/tui-components/theme"
)

// ExampleCapabilityChip_noColor demonstrates a CapabilityChip in no-color mode.
func ExampleCapabilityChip_noColor() {
	chip := NewCapabilityChip(
		WithCapabilityValue("active-cpu"),
		WithCapabilityHumanLabel("Billed for active CPU only"),
		WithCapabilityNoColor(true),
	)
	fmt.Println(chip.ViewString())
	// Output:
	// ◆ active-cpu  Billed for active CPU only
}

// ExampleCapabilityChip_themed demonstrates a CapabilityChip with the default theme.
func ExampleCapabilityChip_themed() {
	chip := NewCapabilityChip(
		WithCapabilityValue("dial-in"),
		WithCapabilityHumanLabel("Dial-in transport"),
		WithCapabilityTheme(theme.DefaultTheme()),
	)
	// Lipgloss rendering contains ANSI sequences; omit Output:.
	_ = chip.ViewString()
}

// ExampleScopePill_noColor demonstrates ScopePill in no-color mode for all scopes.
func ExampleScopePill_noColor() {
	for _, scope := range []Scope{ScopeProject, ScopeOrg, ScopeTenant, ScopeGlobal} {
		p := NewScopePill(WithScopeValue(scope), WithScopeNoColor(true))
		fmt.Println(p.ViewString())
	}
	// Output:
	// [project]
	// [org]
	// [tenant]
	// [global]
}

// ExampleAttestationChip_noColor demonstrates AttestationChip rendering.
func ExampleAttestationChip_noColor() {
	states := []AttestationState{AttestationVerified, AttestationSigned, AttestationUnsigned}
	for _, s := range states {
		chip := NewAttestationChip(
			WithAttestationState(s),
			WithAttestationNoColor(true),
		)
		fmt.Println(chip.ViewString())
	}
	// Output:
	// ✓ verified
	// ~ signed
	// ✗ unsigned
}

// ExampleProviderHealthDot_noColor demonstrates ProviderHealthDot rendering.
func ExampleProviderHealthDot_noColor() {
	for _, h := range []ProviderHealth{ProviderHealthReady, ProviderHealthDegraded, ProviderHealthUnhealthy} {
		d := NewProviderHealthDot(
			WithProviderHealth(h),
			WithProviderHealthShowLabel(true),
			WithProviderHealthNoColor(true),
		)
		fmt.Println(d.ViewString())
	}
	// Output:
	// ● ready
	// ◐ degraded
	// ✗ unhealthy
}

// ExampleToolchainChip_noColor demonstrates ToolchainChip for multi-toolchain demands.
func ExampleToolchainChip_noColor() {
	chip := NewToolchainChip(
		WithToolchainSpecs(
			ToolchainSpec{"java", "17"},
			ToolchainSpec{"node", "20.x"},
		),
		WithToolchainNoColor(true),
	)
	fmt.Println(chip.ViewString())
	// Output:
	// ⚙ java=17, node=20.x
}

// ExampleSandboxCapacityGauge_unlimited demonstrates the unlimited capacity gauge.
func ExampleSandboxCapacityGauge_unlimited() {
	g := NewSandboxCapacityGauge(
		WithGaugeCurrent(3),
		WithGaugeMax(0), // unlimited
		WithGaugeNoColor(true),
	)
	fmt.Println(g.ViewString())
	// Output:
	// 3 / ∞
}

// ExamplePolicyDecisionBanner_noColor demonstrates policy decision banners.
func ExamplePolicyDecisionBanner_noColor() {
	b := NewPolicyDecisionBanner(
		WithPolicyDecision(PolicyBlocked),
		WithPolicyDescription("agent cannot write to /etc/"),
		WithPolicyReason("path-allowlist"),
		WithPolicyBannerNoColor(true),
	)
	fmt.Println(b.ViewString())
	// Output:
	// ✗ BLOCKED    agent cannot write to /etc/  [path-allowlist]
}

// ExampleAuditChain_noColor demonstrates the AuditChain widget.
func ExampleAuditChain_noColor() {
	ts := time.Date(2026, 4, 28, 14, 23, 1, 0, time.UTC)
	e := NewAuditEntry(
		WithAuditEventKind("session.start"),
		WithAuditActor("worker-01"),
		WithAuditTimestamp(ts),
		WithAuditAttestation(AttestationVerified),
		WithAuditNoColor(true),
	)
	chain := NewAuditChain(
		WithChainEntries(e),
		WithChainIntegrity(ChainIntegrityOK),
		WithAuditChainNoColor(true),
	)
	// Chain integrity header + entry; omit Output: since timestamp formatting
	// and column alignment depends on entry count.
	_ = chain.ViewString()
}

// ExampleFleetGrid_noColor demonstrates FleetGrid rendering.
func ExampleFleetGrid_noColor() {
	g := NewFleetGrid(
		WithFleetWorkers(
			FleetWorker{
				ID:           "w-1",
				MachineGroup: "mac-01",
				Status:       WorkerStatusBusy,
				Region:       "iad1",
				LoadFraction: 0.75,
				BillingModel: "active-cpu",
			},
		),
		WithFleetNoColor(true),
	)
	// Output contains ANSI-free text; omit Output: due to bar character width.
	_ = g.ViewString()
}

// ExampleKitDetectResult_noColor demonstrates kit detection output.
func ExampleKitDetectResult_noColor() {
	r := NewKitDetectResult(
		WithKitMatches(
			KitMatch{Name: "spring-java", Version: "v1.2.0", Order: 1},
		),
		WithKitDetectNoColor(true),
	)
	// Omit Output: due to column padding variability.
	_ = r.ViewString()
}
