package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "strings"

	"pablodamascenoo/form-bubble/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {

	m := tui.InitialModel()

	user := s.User()
	welcomeMsg := fmt.Sprintf("Olá, %s! Bem-vindo ao meu portfólio interativo via SSH!\n\nUse as setas para navegar e TAB para alternar o foco.\nPressione 'q' para sair.\n\n", user)

	m = m.SetGreetings(welcomeMsg)

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func main() {
	// Configuração da porta (23234 é padrão para apps Wish)
	lipgloss.SetColorProfile(termenv.TrueColor)

	print("Inicializando o servidor SSH do portfólio...\n")

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
