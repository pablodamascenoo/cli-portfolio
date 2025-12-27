package content

import (
	"embed"

	"github.com/charmbracelet/glamour"
)

//go:embed *
var Files embed.FS

func GetContent(filename string) (string, error) {
	data, err := Files.ReadFile(filename)

	if err != nil {
		return "", err
	}

	formatedString, err := glamour.Render(string(data), "dark")

	if err == nil {
		return string(data), nil
	}

	return formatedString, nil
}
