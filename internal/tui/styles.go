package tui

import "github.com/charmbracelet/lipgloss"

var (
	// ── Deep Violet Palette ───────────────────────────────────────────────────
	ColorBg        = lipgloss.Color("#0D0A14") // near black-purple bg
	ColorSurface   = lipgloss.Color("#1A1429") // panel bg
	ColorSurface2  = lipgloss.Color("#241D38") // active rows, cards
	ColorPrimary   = lipgloss.Color("#7C3AED") // violet — active tab bg
	ColorPrimary2  = lipgloss.Color("#4C1D95") // darker violet — header bg
	ColorAccent    = lipgloss.Color("#A855F7") // purple — section titles, accents
	ColorGlow      = lipgloss.Color("#D946EF") // magenta — logo, special highlights
	ColorSuccess   = lipgloss.Color("#34D399") // green
	ColorWarning   = lipgloss.Color("#FBBF24") // amber
	ColorError     = lipgloss.Color("#F87171") // red
	ColorMuted     = lipgloss.Color("#6B7280") // gray — inactive text
	ColorBorder    = lipgloss.Color("#2D2249") // border inactive
	ColorFg        = lipgloss.Color("#F9FAFB") // near white
	ColorFgDim     = lipgloss.Color("#C4B5FD") // light violet-tinted dim
	ColorSelected  = lipgloss.Color("#4C1D95") // tab active bg
	ColorRowActive = lipgloss.Color("#241D38") // focused settings row bg

	// ── Header ───────────────────────────────────────────────────────────────
	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg).
			Background(ColorPrimary2).
			Padding(0, 2)

	StyleHeaderDot = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Background(ColorPrimary2)

	StyleHeaderGlow = lipgloss.NewStyle().
			Foreground(ColorGlow).
			Background(ColorPrimary2).
			Bold(true)

	StyleHeaderMuted = lipgloss.NewStyle().
				Foreground(ColorFgDim).
				Background(ColorPrimary2)

	// ── Tab bar ───────────────────────────────────────────────────────────────
	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg).
			Background(ColorSelected).
			Padding(0, 2)

	StyleTabInactive = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Background(ColorBg).
				Padding(0, 2)

	StyleTabSep = lipgloss.NewStyle().
			Foreground(ColorBorder).
			Background(ColorBg).
			SetString(" ")

	StyleTabBar = lipgloss.NewStyle().
			Background(ColorBg).
			Padding(0, 0)

	StyleTabBarLine = lipgloss.NewStyle().
				Foreground(ColorBorder)

	// ── Sub-tab (Trading) ─────────────────────────────────────────────────────
	StyleSubTabActive = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent).
				Underline(true).
				Padding(0, 1)

	StyleSubTabInactive = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 1)

	// ── Misc ─────────────────────────────────────────────────────────────────
	StyleHelpBar = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(ColorBg).
			Padding(0, 1)

	StyleHelpKey = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	StyleBorderActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary)

	StyleBorderGlow = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGlow)

	StyleTooltipBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	// ── Toast ─────────────────────────────────────────────────────────────────
	StyleToastInfo = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Padding(0, 2).
			Bold(true)

	StyleToastSuccess = lipgloss.NewStyle().
				Foreground(ColorFg).
				Background(ColorSurface2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess).
				Padding(0, 2).
				Bold(true)

	StyleToastError = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(0, 2).
			Bold(true)

	StyleToastWarning = lipgloss.NewStyle().
				Foreground(ColorFg).
				Background(ColorSurface2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorWarning).
				Padding(0, 2).
				Bold(true)

	// ── Text ─────────────────────────────────────────────────────────────────
	StyleTooltip = lipgloss.NewStyle().Foreground(ColorFgDim)
	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleError   = lipgloss.NewStyle().Foreground(ColorError)
	StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleBold    = lipgloss.NewStyle().Bold(true)
	StylePrimary = lipgloss.NewStyle().Foreground(ColorPrimary)
	StyleAccent  = lipgloss.NewStyle().Foreground(ColorAccent)
	StyleGlow    = lipgloss.NewStyle().Foreground(ColorGlow).Bold(true)
	StyleFgDim   = lipgloss.NewStyle().Foreground(ColorFgDim)

	// ── Settings field widgets ────────────────────────────────────────────────
	StyleToggleOn = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	StyleToggleOff = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleEnumArrow = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleEnumValue = lipgloss.NewStyle().
			Foreground(ColorFg).
			Bold(true)

	StyleFieldActive = lipgloss.NewStyle().
				Background(ColorRowActive).
				Foreground(ColorFg)

	StyleSectionTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent).
				PaddingBottom(1)

	// ── Empty state ───────────────────────────────────────────────────────────
	StyleEmptyState = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Align(lipgloss.Center)

	// ── Splash ────────────────────────────────────────────────────────────────
	StyleSplashLogo = lipgloss.NewStyle().
			Foreground(ColorGlow).
			Bold(true)

	StyleSplashSubtitle = lipgloss.NewStyle().
				Foreground(ColorFgDim)

	StyleSplashBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Background(ColorSurface).
			Padding(1, 4)
)
