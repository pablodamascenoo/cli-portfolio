FROM golang:latest AS builder

WORKDIR /app

COPY . .

RUN go mod download
# Adicionei -ldflags="-s -w" para diminuir o tamanho do binário (opcional)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

FROM alpine:latest

WORKDIR /app

# --- ADICIONE ISTO ---
# Diz ao Lip Gloss que o terminal suporta 256 cores
ENV TERM=xterm-256color
# Opcional: Para cores reais (TrueColor/RGB), se seu terminal suportar
ENV COLORTERM=truecolor
# ---------------------
# 1. Instala certificados CA (caso seu app faça requisições HTTPS para fora)
# e git/bash se precisar debuggar lá dentro, mas opcional.
# Instala ncurses-terminfo-base para suporte completo a cores
RUN apk add --no-cache ca-certificates ncurses-terminfo-base

# 2. CRUCIAL: Cria a pasta .ssh onde a chave será salva
RUN mkdir -p .ssh

COPY --from=builder /app/main .

# 3. Documenta a porta (apenas documentação, mas boa prática)
# O padrão do Wish é 23234, se você mudou no código, ajuste aqui.
EXPOSE 23234

CMD ["./main"]
