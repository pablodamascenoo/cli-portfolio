package tui

import (
	"pablodamascenoo/form-bubble/content"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Dados ---
type item struct {
	title, desc string
	content     string // O conteúdo detalhado que aparecerá na direita
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// --- Modelo Principal ---
type model struct {
	greetings string
	progress  progress.Model
	list      list.Model
	projects  []item
	viewport  viewport.Model
	focused   int // 0: List, 1: Viewport
	loaded    bool
	width     int
	height    int
}

func (m model) SetGreetings(greet string) model {
	m.greetings = greet
	return m
}

func InitialModel() model {
	// Populando o menu
	items := []list.Item{
		item{title: "Sobre Mim", desc: "Quem sou eu?", content: "about.md"},
		item{title: "Skills", desc: "Tech Stack", content: "skills.md"},
		item{title: "Projetos", desc: "Projetos criados", content: "projects.md"},
	}

	projects := []item{
		{title: "Projeto 1", desc: "web", content: "Projeto web 1"},
		{title: "Projeto 2", desc: "api", content: "Projeto API 2"},
		{title: "Projeto 3", desc: "desktop", content: "Projeto desktop 3"},
	}

	m := model{
		list:     list.New(items, list.NewDefaultDelegate(), 0, 0),
		progress: progress.New(progress.WithDefaultGradient()),
		projects: projects,
	}
	m.list.Title = "Meu Portfólio"
	return m
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*350, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			// Alterna o foco entre Menu e Conteúdo
			m.focused = (m.focused + 1) % 2
			return m, nil

			// case "up", "k", "down", "j":
			// 	selected := m.list.SelectedItem().(item)
			// 	m.viewport.SetContent(selected.content)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// --- Responsividade Mágica ---
		// Calcula tamanhos: Menu (30%), Viewport (restante)
		menuWidth := int(float64(m.width) * 0.3)
		viewportWidth := m.width - menuWidth - 6 // -6 para bordas e margens
		ProgressStyle = ProgressStyle.Width(m.width).Height(m.height)

		// Altura (deixando espaço para margens verticais)
		contentHeight := m.height - 4

		m.list.SetSize(menuWidth, contentHeight)

		if !m.loaded {
			m.viewport = viewport.New(viewportWidth, contentHeight)
			// Carrega conteúdo inicial
			content, err := content.GetContent(m.list.SelectedItem().(item).content)
			if err == nil {
				m.viewport.SetContent(content)
			}
			m.loaded = true
		} else {
			m.viewport.Width = viewportWidth
			m.viewport.Height = contentHeight
		}

	case tickMsg:
		if m.progress.Percent() == 1.0 {
			return m, nil
		}

		// Note that you can also use progress.Model.SetPercent to set the
		// percentage value explicitly, too.
		cmd := m.progress.IncrPercent(0.10)
		return m, tea.Batch(tickCmd(), cmd)

	case progress.FrameMsg:

		progressUpdate, cmd := m.progress.Update(msg)
		m.progress = progressUpdate.(progress.Model)
		return m, cmd
	}

	// Roteamento de eventos (Update apenas do componente focado ou ambos)
	if m.focused == 0 {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		if selected, ok := m.list.SelectedItem().(item); ok {
			content, err := content.GetContent("content/" + selected.content)
			if err != nil {
				content = "Erro ao carregar conteúdo."
			}
			m.viewport.SetContent(content)

			if selected.title == "Projetos" {
				// Renderiza os cartões de projetos

				onGoingContentStyle := lipgloss.NewStyle().
					Italic(true).
					Align(lipgloss.Center).
					AlignVertical(lipgloss.Center).
					Width(m.viewport.Width).
					Height(m.viewport.Height)

				m.viewport.SetContent(onGoingContentStyle.Render("Em construção...")) // Limpa antes

				// return m, tea.Batch(cmds...)

				var projectCards []string
				for _, proj := range m.projects {
					card := RenderProjectCard(proj.desc, proj.title, proj.content)
					projectCards = append(projectCards, card)
				}
				combined := lipgloss.JoinVertical(lipgloss.Top, append([]string{content}, projectCards...)...)
				m.viewport.SetContent(combined)
			}
		}
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.loaded {
		return "Carregando..."
	}

	if m.progress.Percent() < 1.0 {
		return ProgressStyle.Render(m.greetings + "\n\n" + m.progress.View() + "\n\nCarregando portfólio...")
	}

	// Estiliza as bordas baseadas no foco
	listView := BlurredStyle.Render(m.list.View())
	viewportView := BlurredStyle.Render(m.viewport.View())

	if m.focused == 0 {
		listView = FocusedStyle.Render(m.list.View())
	} else {
		viewportView = FocusedStyle.Render(m.viewport.View())
	}

	// Junta horizontalmente: [ MENU ] + [ VIEWPORT ]
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, viewportView)
}
