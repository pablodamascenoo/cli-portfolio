package tui

import "github.com/charmbracelet/lipgloss"

func RenderProjectCard(typeOfApp string, title string, desc string) string {

	// 1. Defina o Ícone baseado no tipo
	var iconStr string
	var borderColor lipgloss.Color

	switch typeOfApp {
	case "api":
		iconStr = "  ●  ●  ●  \n ┌───────┐ \n │|||||||│ \n └───────┘ \n  [SERVER] "
		borderColor = lipgloss.Color("#FF5F87") // Rosa/Vermelho
	case "web":
		iconStr = "  ●  ○  ○  \n ┌───────┐ \n │ www.  │ \n └───────┘ \n           "
		borderColor = lipgloss.Color("#00D787") // Verde
	case "desktop":
		iconStr = " >_ bin    \n           \n $ run...  \n ░░░░░99%  \n           "
		borderColor = lipgloss.Color("#5F87FF") // Azul
	}

	// 2. Estilo da CAIXA DO ÍCONE (Esquerda)
	iconStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()). // Borda arredondada fica lindo
		BorderForeground(borderColor).
		MarginLeft(6).
		Padding(0, 1). // Espaço interno
		Width(12).     // Largura fixa para alinhar todos
		Align(lipgloss.Center)

	// 3. Estilo do CONTEÚDO (Direita)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(borderColor)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#767676")) // Cinza escuro

	// Renderiza o bloco de texto da direita
	textBlock := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		descStyle.Render(desc),
	)

	// Adiciona uma margem à esquerda do texto para separar do ícone
	textStyle := lipgloss.NewStyle().PaddingLeft(2).Width(30)

	// 4. MÁGICA FINAL: Junta o Ícone e o Texto lado a lado
	return lipgloss.JoinHorizontal(lipgloss.Center,
		iconStyle.Render(iconStr),
		textStyle.Render(textBlock),
	)
}
