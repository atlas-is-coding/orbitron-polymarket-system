package tui

import "github.com/charmbracelet/lipgloss"

var (
	// ── Deep Violet & Cyber Palette ──────────────────────────────────────────
	ColorBg        = lipgloss.Color("#05020A") // very dark violet
	ColorSurface   = lipgloss.Color("#0F081C") // panel bg
	ColorSurface2  = lipgloss.Color("#180D2E") // active rows, cards
	ColorPrimary   = lipgloss.Color("#8A2BE2") // violet — active tab bg
	ColorPrimary2  = lipgloss.Color("#411A7A") // darker violet — header bg
	ColorAccent    = lipgloss.Color("#00FFF9") // cyan — section titles, accents
	ColorGlow      = lipgloss.Color("#D480FF") // light purple — logo, special highlights
	ColorSuccess   = lipgloss.Color("#00FF9D") // neon green
	ColorWarning   = lipgloss.Color("#FFD700") // amber
	ColorError     = lipgloss.Color("#FF2A55") // neon red
	ColorMuted     = lipgloss.Color("#7E6A9A") // gray/purple — inactive text
	ColorBorder    = lipgloss.Color("#3A1F71") // border inactive
	ColorFg        = lipgloss.Color("#FFFFFF") // pure white
	ColorFgDim     = lipgloss.Color("#C2ADD9") // light violet-tinted dim
	ColorSelected  = lipgloss.Color("#330A66") // tab active bg
	ColorRowActive = lipgloss.Color("#29154D") // focused settings row bg

	// ── Custom Borders ───────────────────────────────────────────────────────
	BorderNormal = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}

	BorderRounded = lipgloss.RoundedBorder()

	// ── Header ───────────────────────────────────────────────────────────────
	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg).
			Background(ColorPrimary2).
			Padding(0, 1)

	StyleHeaderDot = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Background(ColorPrimary2).
			Bold(true)

	StyleHeaderGlow = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Background(ColorPrimary2).
			Bold(true)

	StyleHeaderMuted = lipgloss.NewStyle().
			Foreground(ColorFgDim).
			Background(ColorPrimary2)

	// ── Tab bar ───────────────────────────────────────────────────────────────
	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorAccent).
			Padding(0, 2)

	StyleTabInactive = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(ColorBg).
			Padding(0, 2)

	StyleTabSep = lipgloss.NewStyle().
			Foreground(ColorBorder).
			Background(ColorBg).
			SetString("│")

	StyleTabBar = lipgloss.NewStyle().
			Background(ColorBg).
			Padding(0, 0)

	StyleTabBarLine = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// ── Sub-tab (Trading) ─────────────────────────────────────────────────────
	StyleSubTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorGlow).
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
			Border(BorderRounded).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	StyleBorderActive = lipgloss.NewStyle().
			Border(BorderRounded).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	StyleBorderGlow = lipgloss.NewStyle().
			Border(BorderRounded).
			BorderForeground(ColorAccent).
			Padding(0, 1)

	StyleTooltipBox = lipgloss.NewStyle().
			Border(BorderRounded).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	// ── Toast ─────────────────────────────────────────────────────────────────
	StyleToastInfo = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(BorderRounded).
			BorderForeground(ColorAccent).
			Padding(0, 1).
			Bold(true)

	StyleToastSuccess = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(BorderRounded).
			BorderForeground(ColorSuccess).
			Padding(0, 1).
			Bold(true)

	StyleToastError = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(BorderRounded).
			BorderForeground(ColorError).
			Padding(0, 1).
			Bold(true)

	StyleToastWarning = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(BorderRounded).
			BorderForeground(ColorWarning).
			Padding(0, 1).
			Bold(true)

	// ── Text ─────────────────────────────────────────────────────────────────
	StyleTooltip = lipgloss.NewStyle().Foreground(ColorFgDim)
	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleError   = lipgloss.NewStyle().Foreground(ColorError)
	StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleBold    = lipgloss.NewStyle().Bold(true).Foreground(ColorFg)
	StylePrimary = lipgloss.NewStyle().Foreground(ColorPrimary)
	StyleAccent  = lipgloss.NewStyle().Foreground(ColorAccent)
	StyleGlow    = lipgloss.NewStyle().Foreground(ColorGlow).Bold(true)
	StyleFgDim   = lipgloss.NewStyle().Foreground(ColorFgDim)

	// ── Settings field widgets ────────────────────────────────────────────────
	StyleToggleOn = lipgloss.NewStyle().
			Foreground(ColorBg).
			Background(ColorSuccess).
			Bold(true).Padding(0, 1)

	StyleToggleOff = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(ColorSurface2).
			Padding(0, 1)

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
			Foreground(ColorGlow).
			PaddingBottom(1)

	// ── Empty state ───────────────────────────────────────────────────────────
	StyleEmptyState = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Align(lipgloss.Center)

	// ── Splash ────────────────────────────────────────────────────────────────
	StyleSplashLogo = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleSplashSubtitle = lipgloss.NewStyle().
			Foreground(ColorGlow)

	StyleSplashBox = lipgloss.NewStyle().
			Border(BorderNormal).
			BorderForeground(ColorPrimary).
			Background(ColorSurface).
			Padding(1, 4)

	// ── Sidebar ───────────────────────────────────────────────────────────────
	StyleSidebar = lipgloss.NewStyle().
			Background(ColorSurface).
			Foreground(ColorFgDim)

	StyleSidebarActive = lipgloss.NewStyle().
			Background(ColorAccent).
			Foreground(ColorBg).
			Bold(true)

	StyleSidebarInactive = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSidebarLogo = lipgloss.NewStyle().
			Foreground(ColorGlow).
			Bold(true)

	StyleSidebarSubtitle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSidebarSep = lipgloss.NewStyle().
			Foreground(ColorBorder)

	StyleSidebarLabel = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Bold(false)

	// ── Status bar ────────────────────────────────────────────────────────────
	StyleStatusBar = lipgloss.NewStyle().
			Background(ColorSurface).
			Foreground(ColorMuted).
			Padding(0, 1)

	// ── Panel chrome ─────────────────────────────────────────────────────────
	StylePanelTitle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StylePanelDivider = lipgloss.NewStyle().
			Foreground(ColorBorder)

	StylePanelActive = lipgloss.NewStyle().
			Border(BorderNormal).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	StylePanelInactive = lipgloss.NewStyle().
			Border(BorderNormal).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	StylePanelHelp = lipgloss.NewStyle().
			Border(BorderNormal).
			BorderForeground(ColorBorder).
			Foreground(ColorMuted).
			Padding(0, 1)
)
