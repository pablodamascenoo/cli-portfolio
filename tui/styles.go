package tui

import "github.com/charmbracelet/lipgloss"

// --- Estilos ---
var (
	DocStyle     = lipgloss.NewStyle().Margin(1, 2)
	FocusedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))  // Roxo
	BlurredStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")) // Cinza

	ProgressStyle = lipgloss.NewStyle().Align(lipgloss.Center).AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center)

	// Estilo para o item selecionado na lista
	ItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)
