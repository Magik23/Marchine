package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/magik23/marchine/internal/config"
	"github.com/magik23/marchine/internal/library"
)

const (
	// Marchine keeps the full two-column launcher intact down to this size.
	// Below it we hide the application instead of progressively collapsing it.
	minTerminalWidth  = 124
	minTerminalHeight = 48
)

func (m Model) View() tea.View {
	w := m.width
	if w <= 0 {
		w = minTerminalWidth
	}
	h := m.height
	if h <= 0 {
		h = minTerminalHeight
	}

	if terminalTooSmall(w, h) {
		v := tea.NewView(m.renderTerminalTooSmall(w, h))
		v.AltScreen = true
		v.WindowTitle = "Marchine — Game Machinery."
		return v
	}

	// One explicit terminal-cell gutter on every edge. Marchine renders into
	// the exact inner rectangle; frameViewport then pads/crops that rectangle
	// so left/right and top/bottom can never drift apart as layouts change.
	contentWidth := w - 2
	contentHeight := h - 2
	header := m.renderHeader(contentWidth, contentHeight)
	tabs := m.renderSourceTabs(contentWidth)

	var parts []string
	switch m.screen {
	case screenSetup:
		footer := m.renderSetupFooter(contentWidth)
		bodyH := max(12, contentHeight-lipgloss.Height(header)-lipgloss.Height(tabs)-lipgloss.Height(footer)-3)
		parts = []string{header, tabs, m.renderSetup(contentWidth, bodyH), footer}
	case screenInfo:
		footer := m.renderInfoFooter(contentWidth)
		bodyH := max(12, contentHeight-lipgloss.Height(header)-lipgloss.Height(tabs)-lipgloss.Height(footer)-3)
		parts = []string{header, tabs, m.renderInfo(contentWidth, bodyH), footer}
	case screenCredits:
		footer := m.renderCreditsFooter(contentWidth)
		bodyH := max(12, contentHeight-lipgloss.Height(header)-lipgloss.Height(tabs)-lipgloss.Height(footer)-3)
		parts = []string{header, tabs, m.renderCredits(contentWidth, bodyH), footer}
	case screenCategories:
		footer := m.renderCategoriesFooter(contentWidth)
		bodyH := max(12, contentHeight-lipgloss.Height(header)-lipgloss.Height(tabs)-lipgloss.Height(footer)-3)
		parts = []string{header, tabs, m.renderCategories(contentWidth, bodyH), footer}
	default:
		search := m.renderSearch(contentWidth)
		footer := m.renderFooter(contentWidth)

		// Five vertical blocks produce four separator rows. Count those rows
		// explicitly; older builds accidentally let the main area run long.
		mainH := max(
			9,
			contentHeight-
				lipgloss.Height(header)-
				lipgloss.Height(tabs)-
				lipgloss.Height(search)-
				lipgloss.Height(footer)-
				4,
		)

		parts = []string{
			header,
			tabs,
			search,
			m.renderMain(contentWidth, mainH),
			footer,
		}
	}

	// Every screen owns the same fixed-height inner viewport. The final footer
	// is anchored to the last row of that viewport so moving between Library,
	// Info, Credits, Categories and Setup can never make the bottom menu jump.
	content := composePage(parts, contentHeight)
	content = frameViewport(content, w, h)

	v := tea.NewView(content)
	v.AltScreen = true
	v.WindowTitle = "Marchine — Game Machinery."
	return v
}

func composePage(parts []string, targetHeight int) string {
	if len(parts) == 0 || targetHeight <= 0 {
		return ""
	}

	if len(parts) == 1 {
		return lipgloss.NewStyle().
			Height(targetHeight).
			Render(parts[0])
	}

	upper := strings.Join(parts[:len(parts)-1], "\n")
	footer := parts[len(parts)-1]
	used := lipgloss.Height(upper) + 1 + lipgloss.Height(footer)

	if used < targetHeight {
		// Put any renderer rounding/slack immediately before the footer rather
		// than below it. The visible page baseline therefore stays invariant.
		upper += strings.Repeat("\n", targetHeight-used)
	}

	return upper + "\n" + footer
}

func frameViewport(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	if width < 3 || height < 3 {
		return lipgloss.Place(
			width,
			height,
			lipgloss.Left,
			lipgloss.Top,
			content,
		)
	}

	innerW := width - 2
	innerH := height - 2

	lines := strings.Split(content, "\n")

	if len(lines) > innerH {
		lines = lines[:innerH]
	}

	for len(lines) < innerH {
		lines = append(lines, "")
	}

	blank := strings.Repeat(" ", width)
	out := make([]string, 0, height)
	out = append(out, blank)

	for _, line := range lines {
		line = ansi.Truncate(line, innerW, "")
		padW := max(0, innerW-lipgloss.Width(line))

		out = append(
			out,
			" "+line+strings.Repeat(" ", padW)+" ",
		)
	}

	out = append(out, blank)

	return strings.Join(out, "\n")
}

func (m Model) renderHeader(width, height int) string {
	t := m.theme

	border := lipgloss.NewStyle().
		Foreground(t.Border)

	accent := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	brandTitleStyle := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	if height >= 30 {
		logoRaw := wordmark()

		// A little more visual mass without turning the secondary identity into
		// another block-art logo: letter spacing makes ordinary terminal text
		// read larger while keeping the header to the same compact height.
		brandTitleRaw := "G  A  M  E     M  A  C  H  I  N  E  R  Y  ."
		brandSubRaw := "BROWSE     /     SELECT     /     LAUNCH"

		brandW := max(
			lipgloss.Width(brandTitleRaw),
			lipgloss.Width(brandSubRaw),
		)

		wideNeeded := lipgloss.Width(logoRaw) + brandW + 8

		if width >= wideNeeded {
			logo := accent.Render(logoRaw)
			brandTitle := brandTitleStyle.Render(brandTitleRaw)

			// Terminals cannot use a different point size for one text run, so
			// the operating line is made slightly stronger with bold weight and
			// roomier spacing while remaining ordinary terminal text.
			brandSubStyle := lipgloss.NewStyle().
				Foreground(t.Text).
				Bold(true)

			brandSub := brandSubStyle.Render(brandSubRaw)

			brandBlockW := max(
				lipgloss.Width(brandTitle),
				lipgloss.Width(brandSub),
			)

			brandBlock := lipgloss.NewStyle().
				Width(brandBlockW).
				Align(lipgloss.Center).
				Render(
					"\n" +
						brandTitle +
						"\n" +
						brandSub +
						"\n",
				)

			// The identity block is centered in the space that remains AFTER the
			// MARCHINE wordmark. This is the visual center the header actually has.
			contentW := max(1, width-4)
			logoW := lipgloss.Width(logoRaw)
			remainingW := max(
				brandBlockW,
				contentW-logoW,
			)

			brandArea := lipgloss.NewStyle().
				Width(remainingW).
				Align(lipgloss.Center).
				Render(brandBlock)

			row := lipgloss.JoinHorizontal(
				lipgloss.Top,
				logo,
				brandArea,
			)

			return border.Render(
				"┌"+strings.Repeat("─", max(0, width-2))+"┐",
			) + "\n" +
				lipgloss.NewStyle().
					Padding(0, 1).
					Width(width-2).
					Render(row) +
				"\n" +
				border.Render(
					"└"+strings.Repeat("─", max(0, width-2))+"┘",
				)
		}
	}

	left := accent.Render("MARCHINE") +
		muted.Render("  //  GAME MACHINERY.")

	right := text.
		Bold(true).
		Render("BROWSE / SELECT / LAUNCH")

	gap := max(
		1,
		width-
			lipgloss.Width(left)-
			lipgloss.Width(right)-
			2,
	)

	return border.Render("─") +
		" " +
		left +
		strings.Repeat(" ", gap) +
		right
}

func (m Model) renderSourceTabs(width int) string {
	t := m.theme

	border := lipgloss.NewStyle().
		Foreground(t.Border)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	active := lipgloss.NewStyle().
		Foreground(t.SelectText).
		Background(t.Selection).
		Bold(true).
		Padding(0, 1)

	inactive := lipgloss.NewStyle().
		Foreground(t.Text).
		Padding(0, 1)

	// Each label always remains attached to its real source.
	gameTab := inactive.Render("GAMES")
	arcadeTab := inactive.Render("ARCADE")

	if m.screen == screenLibrary {
		if m.source == library.SourceArcade {
			arcadeTab = active.Render("ARCADE")
		} else {
			gameTab = active.Render("GAMES")
		}
	}

	// Only the VISUAL ORDER is changed:
	// Arcade first, Games second.
	line := border.Render("[ ") +
		arcadeTab +
		border.Render(" ]   [ ") +
		gameTab +
		border.Render(" ]")

	if m.loading {
		loadingLabel := "INDEXING ARCADE…"

		if strings.Contains(m.status, "REFRESH") {
			loadingLabel = "REFRESHING ARCADE…"
		}

		line += muted.Render(
			"     " + loadingLabel,
		)
	}

	return line
}

func (m Model) renderCategories(width, height int) string {
	t := m.theme

	border := lipgloss.NewStyle().
		Foreground(t.Border)

	accent := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	selected := lipgloss.NewStyle().
		Foreground(t.SelectText).
		Background(t.Selection).
		Bold(true)

	filters := m.filters()

	if len(filters) == 0 {
		return lipgloss.Place(
			width,
			height,
			lipgloss.Center,
			lipgloss.Center,
			muted.Render("NO CATEGORIES"),
		)
	}

	counts := map[string]int{}

	for _, filter := range filters {
		for _, g := range m.arcade {
			if matchesArcadeFilter(
				g,
				filter,
				m.cfg.Arcade.HideJunk,
			) {
				counts[filter]++
			}
		}
	}

	const innerW = 40

	// The complete CatVer-backed category set can be longer than one terminal.
	// Keep a stable scrolling viewport around the cursor instead of clipping.
	visibleRows := max(5, height-8)
	visibleRows = min(visibleRows, len(filters))

	start := 0

	if m.categoryCursor >= visibleRows {
		start = m.categoryCursor - visibleRows + 1
	}

	if start+visibleRows > len(filters) {
		start = max(
			0,
			len(filters)-visibleRows,
		)
	}

	end := min(
		len(filters),
		start+visibleRows,
	)

	var b strings.Builder

	b.WriteString(
		accent.Render("ARCADE CATEGORIES"),
	)

	b.WriteByte('\n')

	categoryMode := "COMPLETE LIBRARY"

	if m.cfg.Arcade.HideJunk {
		categoryMode = "NON-ARCADE HIDDEN"
	}

	b.WriteString(
		muted.Render(
			fmt.Sprintf(
				"%d indexed games · CatVer/local metadata · %s",
				len(m.arcade),
				categoryMode,
			),
		),
	)

	b.WriteByte('\n')

	b.WriteString(
		border.Render(
			strings.Repeat("─", innerW),
		),
	)

	b.WriteByte('\n')

	for i := start; i < end; i++ {
		category := filters[i]
		marker := "  "

		if i == m.categoryCursor {
			marker = "> "
		}

		count := fmt.Sprintf(
			"%d",
			counts[category],
		)

		nameW := innerW - len(count) - 3

		row := marker +
			pad(
				trim(category, nameW-2),
				nameW-2,
			) +
			" " +
			count

		row = pad(row, innerW)

		if i == m.categoryCursor {
			b.WriteString(
				selected.Render(row),
			)
		} else if category == m.activeFilter() {
			b.WriteString(
				accent.Render(row),
			)
		} else {
			b.WriteString(
				text.Render(row),
			)
		}

		if i < end-1 {
			b.WriteByte('\n')
		}
	}

	if start > 0 || end < len(filters) {
		b.WriteByte('\n')

		b.WriteString(
			muted.Render(
				fmt.Sprintf(
					"%d–%d / %d",
					start+1,
					end,
					len(filters),
				),
			),
		)
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		Padding(1, 2).
		Width(innerW + 6).
		Render(b.String())

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

func (m Model) renderCategoriesFooter(width int) string {
	t := m.theme

	key := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	item := func(k, label string) string {
		return key.Render(k) +
			text.Render(" "+label)
	}

	line := strings.Join(
		[]string{
			item("↑↓", "SELECT"),
			item("ENTER", "APPLY"),
			item("ESC", "RETURN"),
		},
		muted.Render("  ·  "),
	)

	if lipgloss.Width(line) <= width {
		return line
	}

	return item("ENTER", "APPLY") +
		muted.Render("  ·  ") +
		item("ESC", "RETURN")
}

func (m Model) renderSearch(width int) string {
	t := m.theme

	border := lipgloss.NewStyle().
		Foreground(t.Border)

	promptStyle := lipgloss.NewStyle().
		Foreground(t.Accent2)

	textStyle := lipgloss.NewStyle().
		Foreground(t.Text)

	cursorStyle := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	q := m.search

	if q == "" {
		q = "search…"

		textStyle = lipgloss.NewStyle().
			Foreground(t.Muted)
	}

	cursor := ""

	if m.searchMode {
		cursor = cursorStyle.Render("▌")
	}

	inside := promptStyle.Render("/ ") +
		textStyle.Render(q) +
		cursor

	insideW := max(8, width-2)

	spaces := max(
		0,
		insideW-1-lipgloss.Width(inside),
	)

	return border.Render(
		"┌"+strings.Repeat("─", insideW)+"┐",
	) + "\n" +
		border.Render("│") +
		" " +
		inside +
		strings.Repeat(" ", spaces) +
		border.Render("│") +
		"\n" +
		border.Render(
			"└"+strings.Repeat("─", insideW)+"┘",
		)
}

func (m Model) renderMain(width, height int) string {
	// The detail column follows the actual embedded ANSI width. Replacing
	// fightstick.ans with a smaller sculpture therefore makes the container
	// smaller too; Marchine never stretches the panel around an old hard-coded
	// 88-column assumption. `height` is the exact viewport height reserved by
	// View(), not the terminal height.
	rightW := sculpturePanelWidth()

	const minListW = 44

	if width < minListW+rightW+2 {
		// This path is only a defensive fallback. Normal operation reaches the
		// hard minimum-size screen before the two-column launcher can collapse.
		list := m.renderList(
			width,
			max(9, height/2),
		)

		detailW := min(width, rightW)

		detail := m.renderDetail(
			detailW,
			max(9, height-height/2),
		)

		if detailW < width {
			detail = lipgloss.NewStyle().
				Width(width).
				Align(lipgloss.Center).
				Render(detail)
		}

		return lipgloss.JoinVertical(
			lipgloss.Left,
			list,
			detail,
		)
	}

	leftW := width - rightW - 2

	left := m.renderList(
		leftW,
		max(9, height),
	)

	right := m.renderDetail(
		rightW,
		max(9, height),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		"  ",
		right,
	)
}

func (m Model) renderList(width, height int) string {
	t := m.theme

	borderStyle := lipgloss.NewStyle().
		Foreground(t.Border)

	textStyle := lipgloss.NewStyle().
		Foreground(t.Text)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	selected := lipgloss.NewStyle().
		Foreground(t.SelectText).
		Background(t.Selection).
		Bold(true)

	panel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		Width(max(20, width)).
		Height(max(7, height))

	items := m.filtered()
	rows := max(4, height-4)

	start := 0

	if m.cursor >= rows {
		start = m.cursor - rows + 1
	}

	end := min(
		len(items),
		start+rows,
	)

	innerW := max(18, width-2)

	var b strings.Builder

	if m.source == library.SourceArcade {
		// Let wide windows benefit all three Arcade columns instead of giving
		// nearly every extra cell to NAME. YEAR stays comfortably fixed while
		// MANUFACTURER receives about a third of the available row width.
		yearW := 7
		makerW := max(16, innerW*34/100)
		nameW := innerW - yearW - makerW - 2

		// Defensive fallback for unusually narrow layouts. Marchine normally
		// reaches its minimum-terminal gate before this becomes necessary.
		if nameW < 20 {
			makerW = max(10, innerW-yearW-22)
			nameW = innerW - yearW - makerW - 2
		}

		b.WriteString(
			muted.Render(
				pad("NAME", nameW) +
					" " +
					pad("YEAR", yearW) +
					" " +
					pad("MANUFACTURER", makerW),
			),
		)

		b.WriteByte('\n')

		b.WriteString(
			borderStyle.Render(
				strings.Repeat("─", innerW),
			),
		)

		b.WriteByte('\n')

		for i := start; i < end; i++ {
			g := items[i]
			marker := "  "

			if i == m.cursor {
				marker = "> "
			}

			row := marker +
				pad(
					trim(g.Name, nameW-2),
					nameW-2,
				) +
				" " +
				pad(
					trim(g.Year, yearW),
					yearW,
				) +
				" " +
				trim(g.Manufacturer, makerW)

			row = pad(row, innerW)

			if i == m.cursor {
				b.WriteString(
					selected.Render(row),
				)
			} else {
				b.WriteString(
					textStyle.Render(row),
				)
			}

			if i < end-1 {
				b.WriteByte('\n')
			}
		}
	} else {
		b.WriteString(
			muted.Render(
				fmt.Sprintf(
					"GAMES  //  %d",
					len(items),
				),
			),
		)

		b.WriteByte('\n')

		b.WriteString(
			borderStyle.Render(
				strings.Repeat("─", innerW),
			),
		)

		b.WriteByte('\n')

		for i := start; i < end; i++ {
			g := items[i]
			marker := "  "

			if i == m.cursor {
				marker = "> "
			}

			row := pad(
				marker+
					trim(
						g.Name,
						innerW-2,
					),
				innerW,
			)

			if i == m.cursor {
				b.WriteString(
					selected.Render(row),
				)
			} else {
				b.WriteString(
					textStyle.Render(row),
				)
			}

			if i < end-1 {
				b.WriteByte('\n')
			}
		}
	}

	if len(items) == 0 {
		if m.loading &&
			m.source == library.SourceArcade {
			b.WriteString(
				"\n" +
					muted.Render(
						"INDEXING LOCAL MAME LIBRARY…",
					),
			)
		} else {
			b.WriteString(
				"\n" +
					muted.Render(
						"NO GAMES IN THIS VIEW",
					),
			)
		}
	}

	return panel.Render(b.String())
}

func (m Model) renderDetail(width, height int) string {
	t := m.theme

	panel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		Width(max(20, width)).
		Height(max(7, height))

	accent := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	warn := lipgloss.NewStyle().
		Foreground(t.Warning)

	innerW := max(18, width-2)

	_, artH := sculptureDimensions()

	// Give the ANSI only the rows it actually owns. Extra panel height belongs
	// to the game information below it, not to artificial black bands around
	// the sculpture. At unusually short heights the stable center crop remains.
	sculptH := min(
		artH,
		max(
			minSculptureVisibleH,
			height-8,
		),
	)

	var b strings.Builder

	if sculpture := FightstickSculpture(innerW, sculptH, m.theme); sculpture != "" {
		b.WriteString(sculpture)
		b.WriteByte('\n')

		b.WriteString(
			muted.Render(
				strings.Repeat("─", innerW),
			),
		)

		b.WriteByte('\n')
	}

	if g, ok := m.selected(); ok {
		b.WriteByte('\n')

		b.WriteString(
			accent.Render(
				trim(g.Name, innerW),
			),
		)

		b.WriteByte('\n')

		b.WriteString(
			muted.Render(
				strings.Repeat(
					"─",
					min(
						innerW,
						max(
							12,
							len([]rune(g.Name)),
						),
					),
				),
			),
		)

		b.WriteString("\n\n")

		if g.Source == library.SourceArcade {
			meta := strings.Trim(
				strings.Join(
					[]string{
						g.Year,
						g.Manufacturer,
					},
					"  ·  ",
				),
				" ·",
			)

			b.WriteString(
				text.Render(
					trim(meta, innerW),
				),
			)

			b.WriteByte('\n')
		}
	} else {
		b.WriteString(
			"\n" +
				muted.Render("NO SELECTION"),
		)
	}

	// Normal READY/search/theme/sculpture chatter stays out of this clean panel.
	// Only information that asks for attention is surfaced here.
	status := strings.TrimSpace(m.status)

	if status != "" &&
		status != "READY" &&
		status != "SEARCH" &&
		!strings.HasPrefix(status, "THEME ·") {

		if strings.Contains(status, "FAILED") ||
			strings.Contains(status, "ERROR") ||
			strings.Contains(status, "NOT FOUND") ||
			strings.Contains(status, "NO ROMS") {

			b.WriteByte('\n')

			b.WriteString(
				warn.Render(
					trim(status, innerW),
				),
			)
		} else if strings.Contains(status, "INDEX") ||
			strings.Contains(status, "REFRESH") ||
			strings.Contains(status, "LAUNCH") ||
			strings.Contains(status, "RETURNED") {

			b.WriteByte('\n')

			b.WriteString(
				muted.Render(
					trim(status, innerW),
				),
			)
		}
	}

	return panel.Render(b.String())
}

func (m Model) renderFooter(width int) string {
	t := m.theme

	key := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	item := func(k, label string) string {
		return key.Render(k) +
			text.Render(" "+label)
	}

	credits := item("O", "CREDITS")

	full := []string{
		item("ENTER", "PLAY"),
		item("/", "SEARCH"),
		item("TAB", "SOURCE"),
	}

	if m.source == library.SourceArcade {
		full = append(
			full,
			item("C", "CATEGORIES"),
		)
	}

	full = append(
		full,
		item("I", "INFO"),
		credits,
	)

	// Credits remains the final item. Refresh and Setup stay discoverable from
	// INFO rather than bloating the normal launcher footer.
	medium := []string{
		item("ENTER", "PLAY"),
		item("/", "SEARCH"),
		item("I", "INFO"),
		credits,
	}

	compact := []string{
		item("ENTER", "PLAY"),
		item("I", "INFO"),
		credits,
	}

	separator := muted.Render("  ·  ")

	for _, items := range [][]string{
		full,
		medium,
		compact,
	} {
		line := strings.Join(
			items,
			separator,
		)

		if lipgloss.Width(line) <= width {
			return line
		}
	}

	return item("I", "INFO") +
		separator +
		credits
}

func (m Model) renderSetup(width, height int) string {
	t := m.theme

	panel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border)

	accent := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	accent2 := lipgloss.NewStyle().
		Foreground(t.Accent2)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	warn := lipgloss.NewStyle().
		Foreground(t.Warning)

	selected := lipgloss.NewStyle().
		Foreground(t.SelectText).
		Background(t.Selection).
		Bold(true)

	innerW := max(46, width-8)

	var b strings.Builder

	b.WriteString(
		accent.Render("ARCADE SETUP"),
	)

	b.WriteString(
		muted.Render(
			"  // MAME DRIVER · LOCAL PATHS",
		),
	)

	b.WriteString("\n\n")

	b.WriteString(
		muted.Render(
			"EMULATOR EXECUTABLE",
		),
	)

	b.WriteByte('\n')

	b.WriteString(
		m.renderSetupField(
			setupFieldMAME,
			innerW,
			selected,
			text,
			muted,
		),
	)

	b.WriteString("\n\n")

	b.WriteString(
		muted.Render("PRIMARY ROM PATH"),
	)

	b.WriteByte('\n')

	b.WriteString(
		m.renderSetupField(
			setupFieldROM,
			innerW,
			selected,
			text,
			muted,
		),
	)

	b.WriteString("\n\n")

	emulatorState := "NOT FOUND"
	emulatorStyle := warn

	mameCandidate := config.ExpandPath(
		strings.TrimSpace(m.setupMAME),
	)

	if mameCandidate == "" {
		mameCandidate = "mame"
	}

	if p, err := exec.LookPath(mameCandidate); err == nil {
		emulatorState = "OK  " + p
		emulatorStyle = accent2
	}

	romState := "AUTO-DETECT FROM MAME"
	romStyle := muted

	if strings.TrimSpace(m.setupROM) != "" {
		romCandidate := config.ExpandPath(
			strings.TrimSpace(m.setupROM),
		)

		if st, err := os.Stat(romCandidate); err == nil &&
			st.IsDir() {
			romState = "OK  " + romCandidate
			romStyle = accent2
		} else {
			romState = "NOT FOUND  " + romCandidate
			romStyle = warn
		}
	}

	b.WriteString(
		muted.Render("EMU   "),
	)

	b.WriteString(
		emulatorStyle.Render(
			trim(
				emulatorState,
				innerW-6,
			),
		),
	)

	b.WriteByte('\n')

	b.WriteString(
		muted.Render("ROMS  "),
	)

	b.WriteString(
		romStyle.Render(
			trim(
				romState,
				innerW-6,
			),
		),
	)

	b.WriteByte('\n')

	if m.arcadeMeta.MAMEVersion != "" {
		b.WriteString(
			muted.Render("LIBRARY "),
		)

		b.WriteString(
			text.Render(
				fmt.Sprintf(
					"%d games  ·  %s",
					len(m.arcade),
					trim(
						m.arcadeMeta.MAMEVersion,
						36,
					),
				),
			),
		)

		b.WriteByte('\n')
	}

	b.WriteString(
		muted.Render("CONFIG "),
	)

	b.WriteString(
		text.Render(
			trim(
				m.configPath,
				innerW-7,
			),
		),
	)

	b.WriteString("\n\n")

	b.WriteString(
		muted.Render(
			"Enter edits the selected field. Empty primary path = MAME rompath auto-detect; extra config paths are preserved.",
		),
	)

	b.WriteByte('\n')

	b.WriteString(
		muted.Render(
			"Ctrl+S saves. R saves and immediately refreshes the Arcade library.",
		),
	)

	return panel.
		Width(max(50, width)).
		Height(max(16, height)).
		Padding(1, 2).
		Render(b.String())
}

func (m Model) renderSetupField(
	field,
	width int,
	selected,
	text,
	muted lipgloss.Style,
) string {
	value := m.setupMAME
	hint := "mame, /usr/bin/mame, or emulator path"

	if field == setupFieldROM {
		value = m.setupROM
		hint = "/path/to/MAME/ROMs"
	}

	isHint := value == ""

	if isHint {
		value = hint
	}

	if m.setupEditing &&
		m.setupField == field {
		value = m.setupBuffer + "▌"
		isHint = false
	}

	row := "> " +
		pad(
			trim(
				value,
				width-2,
			),
			width-2,
		)

	if m.setupField == field {
		return selected.Render(row)
	}

	if isHint {
		return muted.Render(
			"  " +
				pad(
					trim(
						value,
						width-2,
					),
					width-2,
				),
		)
	}

	return text.Render(
		"  " +
			pad(
				trim(
					value,
					width-2,
				),
				width-2,
			),
	)
}

func (m Model) renderSetupFooter(width int) string {
	t := m.theme

	key := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	item := func(k, label string) string {
		return key.Render(k) +
			text.Render(" "+label)
	}

	sep := muted.Render("  ·  ")

	if m.setupEditing {
		full := strings.Join(
			[]string{
				item("ENTER", "APPLY"),
				item("ESC", "CANCEL"),
				item("CTRL+U", "CLEAR"),
			},
			sep,
		)

		if lipgloss.Width(full) <= width {
			return full
		}

		return item("ENTER", "APPLY") +
			sep +
			item("ESC", "CANCEL")
	}

	full := strings.Join(
		[]string{
			item("↑↓/TAB", "FIELD"),
			item("ENTER", "EDIT"),
			item("CTRL+S", "SAVE"),
			item("R", "SAVE+REFRESH"),
			item("T", "THEME"),
			item("I", "INFO"),
			item("ESC", "RETURN"),
		},
		sep,
	)

	if lipgloss.Width(full) <= width {
		return full
	}

	medium := strings.Join(
		[]string{
			item("ENTER", "EDIT"),
			item("CTRL+S", "SAVE"),
			item("R", "REFRESH"),
			item("T", "THEME"),
			item("I", "INFO"),
			item("ESC", "RETURN"),
		},
		sep,
	)

	if lipgloss.Width(medium) <= width {
		return medium
	}

	return item("CTRL+S", "SAVE") +
		sep +
		item("I", "INFO") +
		sep +
		item("ESC", "RETURN")
}

func (m Model) renderInfo(width, height int) string {
	t := m.theme

	panel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		Width(max(50, width)).
		Height(max(18, height)).
		Padding(1, 2)

	accent := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	accent2 := lipgloss.NewStyle().
		Foreground(t.Accent2).
		Bold(true)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	section := func(title, body string) string {
		return accent.Render(title) +
			"\n\n" +
			text.Render(body)
	}

	how := "Marchine is a local terminal launcher. GAMES is a flat list of things you can execute. ARCADE indexes a ROM directory through the active emulator driver, then launches the selected game directly.\n\nEnter hands the terminal to the game. When the game exits, Marchine returns exactly where you left it.\n\nNo artwork scraping. No ratings. No account layer. No ROM-file modification."

	emu := "The setup field is deliberately called EMULATOR EXECUTABLE. The current automatic Arcade driver is MAME, but the executable + ROM-path model is intentionally reusable for future command-line emulator drivers.\n\nMarchine keeps MAME's local category metadata and filters the installed library without becoming an emulator encyclopedia."

	shortcuts := "Enter         Launch\n/             Search\nTab           Games / Arcade\nC             Categories (Arcade)\nR             Refresh Arcade Library\nS             Setup\nI             Info\nO             Credits\nEsc / I       Return from Info"

	toggleState := "OFF"

	if m.cfg.Arcade.HideJunk {
		toggleState = "ON"
	}

	filterLine := "H   HIDE NON-ARCADE   [ " +
		toggleState +
		" ]"

	filterHelp := "Casino, gambling, mahjong and other non-arcade entries are removed from the Arcade playlist and category menu when this is ON. The complete indexed library is preserved; turn it OFF to expose ALL and the non-arcade categories again."

	filterBlock := accent.Render(
		"ARCADE FILTER",
	) + "\n\n" +
		accent2.Render(filterLine) +
		"\n\n" +
		muted.Render(filterHelp)

	if width >= 96 {
		innerW := max(50, width-6)
		gapW := 6

		if width >= 128 {
			gapW = 10
		}

		leftW := max(
			40,
			(innerW-gapW)*47/100,
		)

		rightW := max(
			36,
			innerW-leftW-gapW,
		)

		left := lipgloss.NewStyle().
			Width(leftW).
			Render(
				section(
					"HOW IT WORKS",
					how,
				) +
					"\n\n" +
					section(
						"EMULATOR MODEL",
						emu,
					),
			)

		right := lipgloss.NewStyle().
			Width(rightW).
			Render(
				section(
					"SHORTCUTS",
					shortcuts,
				) +
					"\n\n" +
					filterBlock,
			)

		body := lipgloss.JoinHorizontal(
			lipgloss.Top,
			left,
			strings.Repeat(" ", gapW),
			right,
		)

		return panel.Render(body)
	}

	body := section(
		"HOW IT WORKS",
		how,
	) + "\n\n" +
		section(
			"EMULATOR MODEL",
			emu,
		) + "\n\n" +
		section(
			"SHORTCUTS",
			shortcuts,
		) + "\n\n" +
		filterBlock

	return panel.Render(body)
}

func (m Model) renderInfoFooter(width int) string {
	t := m.theme

	key := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	return key.Render("I / ESC") +
		text.Render(" RETURN") +
		muted.Render("  ·  ") +
		key.Render("H") +
		text.Render(" HIDE NON-ARCADE") +
		muted.Render("  ·  ") +
		key.Render("O") +
		text.Render(" CREDITS") +
		muted.Render("  ·  ") +
		key.Render("T") +
		text.Render(" THEME")
}

func (m Model) renderCredits(width, height int) string {
	t := m.theme

	panel := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		Width(max(50, width)).
		Height(max(18, height)).
		Padding(1, 2)

	accent := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	accent2 := lipgloss.NewStyle().
		Foreground(t.Accent2)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	creditsText := "MARCHINE — Game Machinery.\n\nA HyprStage project conceived, directed and designed by\nPierre Dionne / Albenoir Studio.\n\nBuilt in Go with Bubble Tea and Lip Gloss.\n\nMarchine is intentionally small and local-first: no artwork\nscraping, no account layer, no ratings wall, and no\nencyclopedia sprawl. It indexes what you own and launches it\ncleanly.\n\nReleased under the MIT License."

	artRaw := creditsBBSArt()
	artW := lipgloss.Width(artRaw)
	innerW := max(44, width-6)

	body := accent.Render("CREDITS") +
		"\n\n"

	// Credits is its own screen now. On wide terminals, copy and BBS artwork
	// share one row so the page stays compact and the art lives on the right.
	const minCopyW = 42
	const creditGap = 7

	if innerW >= minCopyW+creditGap+artW {
		copyW := innerW -
			creditGap -
			artW

		copyBlock := lipgloss.NewStyle().
			Width(copyW).
			Render(
				muted.Render(creditsText),
			)

		artBlock := lipgloss.NewStyle().
			Width(artW).
			Align(lipgloss.Left).
			Render(
				accent2.Render(artRaw),
			)

		body += lipgloss.JoinHorizontal(
			lipgloss.Top,
			copyBlock,
			strings.Repeat(
				" ",
				creditGap,
			),
			artBlock,
		)
	} else {
		body += muted.Render(
			creditsText,
		) + "\n\n"

		if artW <= innerW {
			body += lipgloss.NewStyle().
				Width(innerW).
				Align(lipgloss.Center).
				Render(
					accent2.Render(
						artRaw,
					),
				)
		} else {
			body += accent2.Render(
				artRaw,
			)
		}
	}

	return panel.Render(body)
}

func (m Model) renderCreditsFooter(width int) string {
	t := m.theme

	key := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	muted := lipgloss.NewStyle().
		Foreground(t.Muted)

	return key.Render("O / ESC") +
		text.Render(" RETURN") +
		muted.Render("  ·  ") +
		key.Render("I") +
		text.Render(" INFO") +
		muted.Render("  ·  ") +
		key.Render("T") +
		text.Render(" THEME")
}

func terminalTooSmall(width, height int) bool {
	return width < minTerminalWidth ||
		height < minTerminalHeight
}

func (m Model) renderTerminalTooSmall(
	width,
	height int,
) string {
	t := m.theme

	accent := lipgloss.NewStyle().
		Foreground(t.Accent).
		Bold(true)

	text := lipgloss.NewStyle().
		Foreground(t.Text)

	if width <= 0 || height <= 0 {
		return ""
	}

	// Keep even the fallback stable at absurdly tiny dimensions.
	if width < 28 || height < 7 {
		return lipgloss.Place(
			width,
			height,
			lipgloss.Center,
			lipgloss.Center,
			trim(
				"Terminal too small.",
				width,
			),
		)
	}

	message := accent.Render(
		"Terminal too small.",
	) + "\n\n" +
		text.Render(
			fmt.Sprintf(
				"Resize to at least %d × %d.",
				minTerminalWidth,
				minTerminalHeight,
			),
		)

	boxW := min(
		48,
		max(
			34,
			width-6,
		),
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		Width(boxW).
		Padding(1, 2).
		Align(lipgloss.Center).
		Render(message)

	// A single centered container replaces the launcher until there is enough
	// room for the normal two-column layout again.
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

func wordmark() string {
	letters := map[rune][]string{
		'M': {
			"█   █",
			"██ ██",
			"█ █ █",
			"█   █",
			"█   █",
		},
		'A': {
			" ███ ",
			"█   █",
			"█████",
			"█   █",
			"█   █",
		},
		'R': {
			"████ ",
			"█   █",
			"████ ",
			"█  █ ",
			"█   █",
		},
		'C': {
			" ████",
			"█    ",
			"█    ",
			"█    ",
			" ████",
		},
		'H': {
			"█   █",
			"█   █",
			"█████",
			"█   █",
			"█   █",
		},
		'I': {
			"█████",
			"  █  ",
			"  █  ",
			"  █  ",
			"█████",
		},
		'N': {
			"█   █",
			"██  █",
			"█ █ █",
			"█  ██",
			"█   █",
		},
		'E': {
			"█████",
			"█    ",
			"████ ",
			"█    ",
			"█████",
		},
	}

	word := "MARCHINE"
	rows := make([]string, 5)

	for _, ch := range word {
		for i := 0; i < 5; i++ {
			if rows[i] != "" {
				rows[i] += " "
			}

			rows[i] += letters[ch][i]
		}
	}

	return strings.Join(
		rows,
		"\n",
	)
}

func trim(s string, width int) string {
	if width <= 0 {
		return ""
	}

	if ansi.StringWidth(s) <= width {
		return s
	}

	return ansi.Truncate(s, width, "…")
}

func pad(s string, width int) string {
	if width <= 0 {
		return ""
	}

	n := ansi.StringWidth(s)
	if n >= width {
		return trim(s, width)
	}

	return s + strings.Repeat(" ", width-n)
}
