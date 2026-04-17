package widget

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// styledRun is a contiguous text slice paired with the SGR-derived lipgloss
// style that should render it. Produced by sgrParser.
type styledRun struct {
	Text  string
	Style lipgloss.Style
}

// sgrState captures the active SGR attributes between chunks so that
// sequences split across Parse calls resume cleanly. Nil fg/bg means the
// terminal default (no explicit colour applied to the lipgloss style).
type sgrState struct {
	fg            color.Color
	bg            color.Color
	hasFG         bool
	hasBG         bool
	bold          bool
	dim           bool
	italic        bool
	underline     bool
	reverse       bool
	strikethrough bool
}

// reset clears all SGR attributes to their defaults (no colour, no attrs).
func (s *sgrState) reset() {
	*s = sgrState{}
}

// style materialises the current SGR state as a lipgloss.Style marked
// inline (so it neither pads nor introduces block behaviour).
func (s sgrState) style() lipgloss.Style {
	st := lipgloss.NewStyle().Inline(true)
	if s.hasFG {
		st = st.Foreground(s.fg)
	}
	if s.hasBG {
		st = st.Background(s.bg)
	}
	if s.bold {
		st = st.Bold(true)
	}
	if s.dim {
		st = st.Faint(true)
	}
	if s.italic {
		st = st.Italic(true)
	}
	if s.underline {
		st = st.Underline(true)
	}
	if s.reverse {
		st = st.Reverse(true)
	}
	if s.strikethrough {
		st = st.Strikethrough(true)
	}
	return st
}

// sgrParser tokenises byte streams containing ANSI SGR escape sequences
// into styled runs. It retains decoder state across Parse calls so that
// sequences split between chunks (e.g. "\x1b[31" then "mhello") are
// applied to the correct trailing text rather than being dropped.
//
// The parser is not safe for concurrent use.
type sgrParser struct {
	decoder    *ansi.Parser
	state      byte // ansi decode-state carried between Parse calls
	sgr        sgrState
	carryState byte // same as state; kept as separate field for clarity
}

// newSGRParser returns a parser ready to consume log output. The returned
// parser retains SGR colour/attr state until Reset is called.
func newSGRParser() *sgrParser {
	p := &sgrParser{
		decoder: ansi.NewParser(),
	}
	p.decoder.SetParamsSize(32)
	p.decoder.SetDataSize(1024)
	return p
}

// Reset clears any carried SGR state and decoder mid-sequence state. Use
// this when starting a new, unrelated stream.
func (p *sgrParser) Reset() {
	p.sgr.reset()
	p.state = 0
	p.carryState = 0
	p.decoder.Reset()
}

// Parse tokenises s applying the parser's carried SGR state. It returns
// one styledRun per contiguous text slice; each run's style reflects the
// SGR attributes active at the moment that text was emitted. Non-SGR CSI
// sequences (cursor movement, erase, private modes) are silently dropped.
// Any partial escape sequence at the tail of s is held in internal state
// so a subsequent Parse call can complete it.
func (p *sgrParser) Parse(s string) []styledRun {
	if s == "" {
		return nil
	}

	var runs []styledRun
	var buf strings.Builder
	currentStyle := p.sgr.style()

	flush := func() {
		if buf.Len() == 0 {
			return
		}
		runs = append(runs, styledRun{Text: buf.String(), Style: currentStyle})
		buf.Reset()
	}

	input := s
	for len(input) > 0 {
		seq, _, n, newState := ansi.DecodeSequence(input, p.state, p.decoder)

		if n == 0 {
			// Defensive: decoder made no progress. Drop one byte to avoid
			// infinite loops on malformed input.
			input = input[1:]
			continue
		}

		switch {
		case isCsiTerminator(newState, p.state, seq):
			// Completed a CSI sequence. If it's SGR ('m'), update state;
			// otherwise drop it. We only inspect the decoder command when
			// a sequence just finished (i.e. we returned to NormalState).
			cmd := ansi.Cmd(p.decoder.Command())
			if cmd.Final() == 'm' && cmd.Prefix() == 0 {
				flush()
				applySGR(&p.sgr, p.decoder.Params())
				currentStyle = p.sgr.style()
			}
			// Non-SGR CSI sequences (including private modes with prefix)
			// are silently stripped.
		case isTextSequence(seq):
			buf.WriteString(seq)
		default:
			// Lone ESC, completed non-CSI escape (e.g. "\x1bM"),
			// completed OSC/DCS/APC string, or continuation of an
			// in-progress sequence — drop from output.
		}

		p.state = newState
		input = input[n:]
	}

	flush()
	p.carryState = p.state
	return runs
}

// isTextSequence reports whether seq is renderable text (printable ASCII,
// a multi-byte grapheme, or a bare control char like newline/tab that we
// want to preserve verbatim in log output).
func isTextSequence(seq string) bool {
	if len(seq) == 0 {
		return false
	}
	b := seq[0]
	// ESC / CSI / DCS / OSC / APC / SOS / PM prefixes indicate an escape
	// sequence — not text.
	switch b {
	case ansi.ESC, ansi.CSI, ansi.DCS, ansi.OSC, ansi.APC, ansi.SOS, ansi.PM:
		return false
	}
	return true
}

// isCsiTerminator reports whether the most recent DecodeSequence call
// emitted a completed CSI sequence. That's signalled by returning to
// NormalState from a non-normal state while processing a CSI-prefixed
// slice.
func isCsiTerminator(newState, prevState byte, seq string) bool {
	if newState != 0 { // ansi.NormalState == 0
		return false
	}
	if prevState == 0 && !ansi.HasCsiPrefix(seq) {
		return false
	}
	// Sequence starts with CSI or ESC[ (if we entered from NormalState)
	// OR we were carrying a CSI mid-state across the call boundary.
	return ansi.HasCsiPrefix(seq) || isCsiCarryState(prevState)
}

// isCsiCarryState reports whether prevState represents a CSI parsing
// state (prefix / params / intermediate). These are the states for which
// the ansi decoder has already consumed the leading CSI bytes on a
// previous call.
func isCsiCarryState(prevState byte) bool {
	switch prevState {
	case ansi.PrefixState, ansi.ParamsState, ansi.IntermedState:
		return true
	}
	return false
}

// applySGR mutates state according to a completed SGR (CSI ... m)
// sequence. Unknown parameters are ignored; malformed 38/48 extended
// colour sub-sequences fall through without changing colour.
func applySGR(state *sgrState, params ansi.Params) {
	// An empty parameter list is equivalent to a single 0 (full reset).
	if len(params) == 0 {
		state.reset()
		return
	}

	for i := 0; i < len(params); i++ {
		n := params[i].Param(0)
		switch {
		case n == 0:
			state.reset()
		case n == 1:
			state.bold = true
		case n == 2:
			state.dim = true
		case n == 3:
			state.italic = true
		case n == 4:
			state.underline = true
		case n == 7:
			state.reverse = true
		case n == 9:
			state.strikethrough = true
		case n == 22:
			state.bold = false
			state.dim = false
		case n == 23:
			state.italic = false
		case n == 24:
			state.underline = false
		case n == 27:
			state.reverse = false
		case n == 29:
			state.strikethrough = false
		case n >= 30 && n <= 37:
			state.fg = lipgloss.ANSIColor(n - 30) //nolint:gosec
			state.hasFG = true
		case n == 38:
			consumed, c, ok := readExtendedColor(params[i+1:])
			i += consumed
			if ok {
				state.fg = c
				state.hasFG = true
			}
		case n == 39:
			state.fg = nil
			state.hasFG = false
		case n >= 40 && n <= 47:
			state.bg = lipgloss.ANSIColor(n - 40) //nolint:gosec
			state.hasBG = true
		case n == 48:
			consumed, c, ok := readExtendedColor(params[i+1:])
			i += consumed
			if ok {
				state.bg = c
				state.hasBG = true
			}
		case n == 49:
			state.bg = nil
			state.hasBG = false
		case n >= 90 && n <= 97:
			state.fg = lipgloss.ANSIColor(n - 90 + 8) //nolint:gosec
			state.hasFG = true
		case n >= 100 && n <= 107:
			state.bg = lipgloss.ANSIColor(n - 100 + 8) //nolint:gosec
			state.hasBG = true
		}
	}
}

// readExtendedColor parses the tail of a 38/48 SGR sub-sequence. It
// supports both semicolon and colon variants (both arrive as successive
// Params). Returns the number of parameters consumed beyond the 38/48
// marker plus the resolved colour, or ok=false for malformed input.
func readExtendedColor(rest ansi.Params) (consumed int, c color.Color, ok bool) {
	if len(rest) == 0 {
		return 0, nil, false
	}
	mode := rest[0].Param(0)
	switch mode {
	case 5: // 256-colour indexed
		if len(rest) < 2 {
			return len(rest), nil, false
		}
		idx := rest[1].Param(0)
		if idx < 0 || idx > 255 {
			return 2, nil, false
		}
		return 2, lipgloss.ANSIColor(idx), true //nolint:gosec
	case 2: // truecolor R;G;B
		if len(rest) < 4 {
			return len(rest), nil, false
		}
		r := clampByte(rest[1].Param(0))
		g := clampByte(rest[2].Param(0))
		b := clampByte(rest[3].Param(0))
		return 4, lipgloss.Color(rgbHex(r, g, b)), true
	}
	return 0, nil, false
}

// clampByte limits an int SGR parameter to the 0-255 byte range.
func clampByte(n int) uint8 {
	if n < 0 {
		return 0
	}
	if n > 255 {
		return 255
	}
	return uint8(n) //nolint:gosec
}

// rgbHex formats r/g/b as a #RRGGBB hex string consumable by
// lipgloss.Color.
func rgbHex(r, g, b uint8) string {
	const hexdigits = "0123456789abcdef"
	buf := [7]byte{'#'}
	buf[1] = hexdigits[r>>4]
	buf[2] = hexdigits[r&0x0f]
	buf[3] = hexdigits[g>>4]
	buf[4] = hexdigits[g&0x0f]
	buf[5] = hexdigits[b>>4]
	buf[6] = hexdigits[b&0x0f]
	return string(buf[:])
}
