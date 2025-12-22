FROM golang:latest AS builder

WORKDIR /app

COPY . .

RUN go mod download
# Adicionei -ldflags="-s -w" para diminuir o tamanho do binário (opcional)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

FROM alpine:latest

WORKDIR /app

# 1. Instala certificados CA (caso seu app faça requisições HTTPS para fora)
# e git/bash se precisar debuggar lá dentro, mas opcional.
RUN apk add --no-cache ca-certificates

# 2. CRUCIAL: Cria a pasta .ssh onde a chave será salva
RUN mkdir -p .ssh

COPY --from=builder /app/main .

# 3. Documenta a porta (apenas documentação, mas boa prática)
# O padrão do Wish é 23234, se você mudou no código, ajuste aqui.
EXPOSE 23234

CMD ["./main"]