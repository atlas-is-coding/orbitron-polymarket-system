package tui

import "github.com/charmbracelet/lipgloss"

var (
	// ── Deep Violet Palette (spec §2) ───────────────────────────────────────
	ColorBg        = lipgloss.Color("#0e0b1a") // was #0B0B0E
	ColorBgMid     = lipgloss.Color("#13102a") // new
	ColorBgLight   = lipgloss.Color("#1a1535") // new (replaces ColorSurface2/ColorRowActive)
	ColorSurface   = lipgloss.Color("#13102a") // compat alias → same as ColorBgMid
	ColorSurface2  = lipgloss.Color("#1a1535") // compat alias → same as ColorBgLight
	ColorRowActive = lipgloss.Color("#1a1535") // compat alias → same as ColorBgLight
	ColorSelected  = lipgloss.Color("#1a1535") // compat alias

	ColorPrimary    = lipgloss.Color("#7c3aed") // was #8A2BE2
	ColorPrimary2   = lipgloss.Color("#4A1580") // unchanged
	ColorPrimaryDim = lipgloss.Color("#555555") // new

	ColorBright = lipgloss.Color("#a78bfa") // new — prefer this in new code
	ColorAccent = lipgloss.Color("#a78bfa") // compat alias → prefer ColorBright in new code
	ColorGlow   = lipgloss.Color("#a78bfa") // compat alias → prefer ColorBright in new code

	ColorSuccess = lipgloss.Color("#34d399") // was #00FF9D
	ColorWarning = lipgloss.Color("#fbbf24") // was #FFD700
	ColorDanger  = lipgloss.Color("#f87171") // new
	ColorError   = lipgloss.Color("#f87171") // compat alias → same as ColorDanger

	ColorText   = lipgloss.Color("#e0e0e0") // new
	ColorFg     = lipgloss.Color("#e0e0e0") // compat alias → same as ColorText
	ColorFgDim  = lipgloss.Color("#888888") // was #9595A8
	ColorMuted  = lipgloss.Color("#888888") // was #4A4A60 — now visibly readable (used in help bars, empty states, metadata)
	ColorBorder = lipgloss.Color("#555555") // intentionally same as ColorPrimaryDim for this palette; separate if palette diverges

	// ── Borders ─────────────────────────────────────────────────────────────
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

	BorderThick = lipgloss.Border{
		Top:         "━",
		Bottom:      "━",
		Left:        "┃",
		Right:       "┃",
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
	}

	BorderRounded = lipgloss.RoundedBorder()

	// BorderSpec uses double vertical bars (║) with thin horizontal (─) per spec §6.
	BorderSpec = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "║",
		Right:       "║",
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
	}

	// ── Text Styles ─────────────────────────────────────────────────────────
	StyleTooltip = lipgloss.NewStyle().Foreground(ColorFgDim)
	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleError   = lipgloss.NewStyle().Foreground(ColorError)
	StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleBold    = lipgloss.NewStyle().Bold(true).Foreground(ColorFg)
	StylePrimary = lipgloss.NewStyle().Foreground(ColorPrimary)
	StyleAccent  = lipgloss.NewStyle().Foreground(ColorAccent)
	StyleGlow    = lipgloss.NewStyle().Foreground(ColorGlow).Bold(true)
	StyleFgDim      = lipgloss.NewStyle().Foreground(ColorFgDim)
	StyleFgDimBold  = lipgloss.NewStyle().Foreground(ColorFgDim).Bold(true)

	// ── Typography (spec §3) ─────────────────────────────────────────────────
	StylePageTitle   = lipgloss.NewStyle().Bold(true).Foreground(ColorBright)
	StyleSectionHead = lipgloss.NewStyle().Bold(true).Foreground(ColorText)
	StyleBody        = lipgloss.NewStyle().Foreground(ColorText)
	StyleValue       = lipgloss.NewStyle().Foreground(ColorBright)
	StylePositive    = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleNegative    = lipgloss.NewStyle().Foreground(ColorDanger)

	// ── Header & Tabs ───────────────────────────────────────────────────────
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
			Background(ColorBg)

	StyleTabBarLine = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleSubTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorPrimary).
			Padding(0, 2)

	StyleSubTabInactive = lipgloss.NewStyle().
			Foreground(ColorFgDim).
			Background(ColorSurface2).
			Padding(0, 2)

	// ── UI Components ───────────────────────────────────────────────────────
	StyleHelpBar = lipgloss.NewStyle().
			Background(ColorBgLight).
			BorderTop(true).
			BorderStyle(lipgloss.Border{Top: "─"}).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	StyleHelpKey = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Background(ColorSurface2).
			Padding(0, 1).
			Bold(true)

	StyleHelpDesc = lipgloss.NewStyle().
			Foreground(ColorFgDim).
			Padding(0, 1)

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

	// ── Toast Notifications ─────────────────────────────────────────────────
	StyleToastInfo = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(BorderThick).
			BorderForeground(ColorAccent).
			Padding(0, 2).
			Bold(true)

	StyleToastSuccess = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(BorderThick).
			BorderForeground(ColorSuccess).
			Padding(0, 2).
			Bold(true)

	StyleToastError = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(BorderThick).
			BorderForeground(ColorError).
			Padding(0, 2).
			Bold(true)

	StyleToastWarning = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorSurface2).
			Border(BorderThick).
			BorderForeground(ColorWarning).
			Padding(0, 2).
			Bold(true)

	// ── Settings Field Widgets ──────────────────────────────────────────────
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
			Foreground(ColorAccent).
			PaddingBottom(1)

	// ── Empty State ─────────────────────────────────────────────────────────
	StyleEmptyState = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Align(lipgloss.Center)

	// ── Splash Screen ───────────────────────────────────────────────────────
	StyleSplashLogo = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleSplashSubtitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleSplashBox = lipgloss.NewStyle().
			Border(BorderThick).
			BorderForeground(ColorPrimary).
			Background(ColorSurface).
			Padding(2, 4)

	// ── Sidebar ─────────────────────────────────────────────────────────────
	StyleSidebar = lipgloss.NewStyle().
			Background(ColorSurface).
			Foreground(ColorFgDim)

	StyleSidebarActive = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorBg).
			Bold(true)

	StyleSidebarInactive = lipgloss.NewStyle().
			Foreground(ColorFgDim)

	StyleSidebarLogo = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleSidebarSubtitle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSidebarSep = lipgloss.NewStyle().
			Foreground(ColorBorder)

	StyleSidebarLabel = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// ── Status Bar ──────────────────────────────────────────────────────────
	StyleStatusBar = lipgloss.NewStyle().
			Background(ColorSurface2).
			Foreground(ColorFgDim)

	StyleStatusMode = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorBg).
			Bold(true).
			Padding(0, 2)

	StyleStatusLive = lipgloss.NewStyle().
			Background(ColorSuccess).
			Foreground(ColorBg).
			Bold(true).
			Padding(0, 2)

	StyleStatusTime = lipgloss.NewStyle().
			Background(ColorSurface2).
			Foreground(ColorFg).
			Padding(0, 2)

	// ── Panel Chrome ────────────────────────────────────────────────────────
	StylePanelTitle = lipgloss.NewStyle().
			Foreground(ColorFgDim).
			Bold(true).
			Padding(0, 1)

	StylePanelTitleActive = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Padding(0, 1)

	StylePanelDivider = lipgloss.NewStyle().
			Foreground(ColorBorder)

	StylePanelActive = lipgloss.NewStyle().
			Border(BorderSpec).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	StylePanelInactive = lipgloss.NewStyle().
			Border(BorderSpec).
			BorderForeground(ColorPrimaryDim).
			Padding(0, 1)

	StylePanelHelp = lipgloss.NewStyle().
			Border(BorderRounded).
			BorderForeground(ColorBorder).
			Foreground(ColorMuted).
			Padding(0, 1)
)
