package widget

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *AttestationChip satisfies component.Component.
var _ component.Component = (*AttestationChip)(nil)

// AttestationState represents the attestation/signing state of a plugin,
// commit, or session artifact, per 002-provider-base-contract.md and
// 014-tui-operator-surfaces.md.
type AttestationState string

const (
	// AttestationSigned indicates the artifact carries a valid cryptographic
	// signature but the key has not been independently verified.
	AttestationSigned AttestationState = "signed"

	// AttestationUnsigned indicates the artifact has no signature.
	AttestationUnsigned AttestationState = "unsigned"

	// AttestationVerified indicates the signature has been verified against a
	// trusted key store (e.g. sigstore-equivalent verification).
	AttestationVerified AttestationState = "verified"
)

// AttestationChip renders the signing / verification state of an artifact
// together with an optional truncated key fingerprint.
//
// Example renderings:
//
//	✓ verified  ed25519:abc1234…d4f2
//	~ signed
//	✗ unsigned
//
// In no-color mode the chip renders with plain ASCII and no color escapes.
type AttestationChip struct {
	state       AttestationState
	fingerprint string // raw fingerprint; will be truncated for display
	accessLabel string
	t           theme.Theme
	noColor     bool
}

// AttestationChipOption configures an AttestationChip during construction.
type AttestationChipOption func(*AttestationChip)

// WithAttestationState sets the attestation state.
func WithAttestationState(s AttestationState) AttestationChipOption {
	return func(c *AttestationChip) { c.state = s }
}

// WithAttestationFingerprint sets the raw key fingerprint displayed in
// truncated form alongside the state.  The format is expected to be
// "algo:hexbytes" (e.g. "ed25519:abcd1234...").  Truncation is applied
// by the widget; pass the full fingerprint.
func WithAttestationFingerprint(fp string) AttestationChipOption {
	return func(c *AttestationChip) { c.fingerprint = fp }
}

// WithAttestationTheme sets the Theme.
func WithAttestationTheme(t theme.Theme) AttestationChipOption {
	return func(c *AttestationChip) { c.t = t }
}

// WithAttestationNoColor forces symbol-first, no-ANSI rendering.
func WithAttestationNoColor(nc bool) AttestationChipOption {
	return func(c *AttestationChip) { c.noColor = nc }
}

// WithAttestationAccessibleLabel sets the accessible label.
func WithAttestationAccessibleLabel(label string) AttestationChipOption {
	return func(c *AttestationChip) { c.accessLabel = label }
}

// NewAttestationChip constructs an AttestationChip with default theme.
func NewAttestationChip(opts ...AttestationChipOption) *AttestationChip {
	c := &AttestationChip{
		state: AttestationUnsigned,
		t:     theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetTheme updates the theme.
func (c *AttestationChip) SetTheme(t theme.Theme) { c.t = t }

// AccessibleLabel returns the accessible label.
func (c *AttestationChip) AccessibleLabel() string {
	if c.accessLabel != "" {
		return c.accessLabel
	}
	label := "attestation: " + string(c.state)
	if c.fingerprint != "" {
		label += " (" + c.fingerprint + ")"
	}
	return label
}

// truncateFingerprint renders the fingerprint in "algo:prefix…suffix" form.
// If the fingerprint is short enough it is returned as-is.
func truncateFingerprint(fp string) string {
	const maxLen = 20
	if len(fp) <= maxLen {
		return fp
	}
	// Find the colon separating algo prefix from hex body
	for i, ch := range fp {
		if ch == ':' && i < len(fp)-1 {
			algo := fp[:i+1]
			body := fp[i+1:]
			if len(body) > 8 {
				return algo + body[:4] + "…" + body[len(body)-4:]
			}
		}
	}
	return fp[:maxLen-1] + "…"
}

type attestationVisual struct {
	symbol string
	color  color.Color
	label  string
}

func (c *AttestationChip) visual() attestationVisual {
	switch c.state {
	case AttestationVerified:
		return attestationVisual{"✓", c.t.StatusSuccess, "verified"}
	case AttestationSigned:
		return attestationVisual{"~", c.t.StatusWarning, "signed"}
	default:
		return attestationVisual{"✗", c.t.StatusError, "unsigned"}
	}
}

// ViewString renders the chip as a plain string.
func (c *AttestationChip) ViewString() string {
	v := c.visual()
	fp := ""
	if c.fingerprint != "" {
		fp = "  " + truncateFingerprint(c.fingerprint)
	}
	if c.noColor {
		return v.symbol + " " + v.label + fp
	}
	style := lipgloss.NewStyle().Foreground(v.color)
	rendered := style.Render(v.symbol+" "+v.label) + lipgloss.NewStyle().Foreground(c.t.TextTertiary).Render(fp)
	return rendered
}

// Init satisfies tea.Model.
func (c *AttestationChip) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (c *AttestationChip) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return c, nil }

// View renders the chip as a tea.View.
func (c *AttestationChip) View() tea.View { return tea.NewView(c.ViewString()) }

// SetSize is a no-op.
func (c *AttestationChip) SetSize(width, height int) {}

// Focus is a no-op.
func (c *AttestationChip) Focus() {}

// Blur is a no-op.
func (c *AttestationChip) Blur() {}
