package tui

import "github.com/charmbracelet/lipgloss"

var (
	// ── Modern Sleek Palette ────────────────────────────────────────────────
	ColorBg        = lipgloss.Color("#0B0B0E") // Deep dark gray/black
	ColorSurface   = lipgloss.Color("#13131A") // Slightly lighter for panels
	ColorSurface2  = lipgloss.Color("#1E1E28") // Active rows
	ColorPrimary   = lipgloss.Color("#8A2BE2") // Main purple accent
	ColorPrimary2  = lipgloss.Color("#4A1580") // Darker purple
	ColorAccent    = lipgloss.Color("#00E6F0") // Cyan glow
	ColorGlow      = lipgloss.Color("#D480FF") // Soft purple glow
	ColorSuccess   = lipgloss.Color("#00FF9D") // Bright neon green
	ColorWarning   = lipgloss.Color("#FFD700") // Amber
	ColorError     = lipgloss.Color("#FF3366") // Bright neon red
	ColorMuted     = lipgloss.Color("#4A4A60") // Inactive gray
	ColorBorder    = lipgloss.Color("#2A2A37") // Subtle border
	ColorFg        = lipgloss.Color("#EBEBEB") // Off-white
	ColorFgDim     = lipgloss.Color("#9595A8") // Dimmed text
	ColorSelected  = lipgloss.Color("#2A2A40") // Tab active bg
	ColorRowActive = lipgloss.Color("#252538") // Focused row

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
	StyleFgDim   = lipgloss.NewStyle().Foreground(ColorFgDim)

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
			Foreground(ColorMuted).
			Background(ColorSurface).
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
			Background(ColorSurface2).
			Bold(true).
			Padding(0, 1)

	StylePanelDivider = lipgloss.NewStyle().
			Foreground(ColorBorder)

	StylePanelActive = lipgloss.NewStyle().
			Border(BorderThick).
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

