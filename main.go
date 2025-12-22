package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "strings"
	"embed"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {

	m := initialModel()

	user := s.User()
	welcomeMsg := fmt.Sprintf("Olá, %s! Bem-vindo ao meu portfólio interativo via SSH!\n\nUse as setas para navegar e TAB para alternar o foco.\nPressione 'q' para sair.\n\n", user)

	m.greetings = welcomeMsg

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

// --- Estilos ---
var (
	docStyle     = lipgloss.NewStyle().Margin(1, 2)
	focusedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))  // Roxo
	blurredStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")) // Cinza

	progressStyle = lipgloss.NewStyle().Align(lipgloss.Center).AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center)

	// Estilo para o item selecionado na lista
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

//go:embed content
var contents embed.FS

func renderProjectCard(typeOfApp string, title string, desc string) string {

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

func loadContent(filename string) string {

	data, err := contents.ReadFile(filename)

	if err != nil {
		return "Falha ao carregar arquivo " + err.Error()
	}

	rendered, err := glamour.Render(string(data), "dark")

	if err != nil {
		return string(data)
	}

	return rendered
}

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

func initialModel() model {
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
		progressStyle = progressStyle.Width(m.width).Height(m.height)

		// Altura (deixando espaço para margens verticais)
		contentHeight := m.height - 4

		m.list.SetSize(menuWidth, contentHeight)

		if !m.loaded {
			m.viewport = viewport.New(viewportWidth, contentHeight)
			// Carrega conteúdo inicial
			content := loadContent("content/" + m.list.SelectedItem().(item).content)
			m.viewport.SetContent(content)
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
			content := loadContent("content/" + selected.content)
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

				return m, tea.Batch(cmds...)

				var projectCards []string
				for _, proj := range m.projects {
					card := renderProjectCard(proj.desc, proj.title, proj.content)
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
		return progressStyle.Render(m.greetings + "\n\n" + m.progress.View() + "\n\nCarregando portfólio...")
	}

	// Estiliza as bordas baseadas no foco
	listView := blurredStyle.Render(m.list.View())
	viewportView := blurredStyle.Render(m.viewport.View())

	if m.focused == 0 {
		listView = focusedStyle.Render(m.list.View())
	} else {
		viewportView = focusedStyle.Render(m.viewport.View())
	}

	// Junta horizontalmente: [ MENU ] + [ VIEWPORT ]
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, viewportView)
}

func main() {
	// Configuração da porta (23234 é padrão para apps Wish)
	port := "23234"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	// 2. Configura o Servidor SSH
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf(":%s", port)),

		// IMPORTANTE: Isso cria uma chave de host persistente.
		// Sem isso, seus usuários verão aquele aviso de "WARNING: REMOTE HOST ID CHANGED"
		// toda vez que você reiniciar o servidor.
		wish.WithHostKeyPath(".ssh/id_ed25519"),

		wish.WithMiddleware(
			// O "Glue" (Cola) que liga o SSH ao Bubble Tea
			bm.Middleware(teaHandler),

			// Middleware de Log (opcional, mas bom para debug)
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Error("não foi possível criar servidor", "err", err)
	}

	// 3. Inicia o Servidor
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Iniciando servidor SSH", "host", "0.0.0.0", "port", port)

	go func() {
		if err = s.ListenAndServe(); err != nil && err != ssh.ErrServerClosed {
			log.Error("não foi possível iniciar", "err", err)
			done <- nil
		}
	}()

	<-done // Espera o sinal de Ctrl+C para parar
	log.Info("Parando servidor SSH...")

	// Shutdown gracioso (dá 30s para conexões ativas fecharem)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Error("erro no shutdown", "err", err)
	}
}
