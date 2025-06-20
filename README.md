# GoMailer

<div align="center">

[🇧🇷 Português](#português) | [🇺🇸 English](README.en.md)

</div>

# Português

Um microsserviço para manipulação de envio de emails através de uma fila RabbitMQ, construído com Go.

## Funcionalidades

- API REST para enfileiramento de emails
- Integração com RabbitMQ para enfileiramento confiável de mensagens
- Envio de emails via SMTP com suporte a HTML
- Configuração baseada em variáveis de ambiente
- Suporte a Docker para RabbitMQ
- Conexão TCP segura para integração com outros serviços

## Pré-requisitos

- Go 1.24 ou superior
- Docker e Docker Compose
- Credenciais de servidor SMTP (ex: Gmail SMTP)

## Configuração

1. Clone o repositório:

```bash
git clone <repository-url>
cd gomailer
```

2. Instale as dependências:

```bash
go mod download
```

3. Configure as variáveis de ambiente:

   - Copie o arquivo de exemplo de ambiente:

   ```bash
   cp env.example .env
   ```

   - Edite o arquivo `.env` com suas configurações

4. Inicie o RabbitMQ usando Docker Compose:

```bash
docker-compose up -d
```

5. Execute a aplicação:

```bash
go run cmd/main.go
```

O serviço iniciará nas portas configuradas (padrão: HTTP 8080, TCP 9000).

## Variáveis de Ambiente

### Variáveis Obrigatórias

- `SMTP_USER`: Usuário do servidor SMTP (obrigatório)
- `SMTP_PASSWORD`: Senha do servidor SMTP (obrigatório)
- `SMTP_FROM`: Endereço de email de envio (obrigatório)
- `TCP_AUTH_SECRET`: Chave secreta para autenticação TCP (obrigatório)

### Variáveis Opcionais com Valores Padrão

- `SMTP_HOST`: Host do servidor SMTP (padrão: "smtp.gmail.com")
- `SMTP_PORT`: Porta do servidor SMTP (padrão: 587)
- `RABBITMQ_HOST`: Host do RabbitMQ (padrão: "localhost")
- `RABBITMQ_PORT`: Porta do RabbitMQ (padrão: "5672")
- `RABBITMQ_USER`: Usuário do RabbitMQ (padrão: "admin")
- `RABBITMQ_PASSWORD`: Senha do RabbitMQ (padrão: "admin")
- `API_PORT`: Porta do servidor API (padrão: "8080")
- `TCP_PORT`: Porta do servidor TCP (padrão: "9000")

## Uso da API

### Enfileirar um Email via HTTP

```http
POST /email
Content-Type: application/json

{
  "to": ["destinatario@exemplo.com"],
  "subject": "Olá",
  "body": "<h1>Olá Mundo</h1><p>Este é um email de teste.</p>"
}
```

Resposta:

```json
{
  "message": "Email enfileirado com sucesso"
}
```

### Integração via TCP

Para integrar outros serviços com o GoMailer, você pode usar o cliente TCP fornecido:

```go
package main

import (
    "log"
    "gomailer/pkg/client"
)

func main() {
    // Criar cliente de email
    emailClient := client.NewEmailClient(
        "localhost",           // Host do serviço
        "9000",               // Porta TCP
        "seu-segredo-aqui",   // Chave de autenticação
    )

    // Preparar requisição de email
    request := &client.EmailRequest{
        To:      []string{"destinatario@exemplo.com"},
        Subject: "Teste via TCP",
        Body:    "<h1>Olá</h1><p>Este é um teste via TCP</p>",
    }

    // Enviar email
    if err := emailClient.SendEmail(request); err != nil {
        log.Fatalf("Erro ao enviar email: %v", err)
    }
}
```

Para usar o cliente em outro projeto:

1. Adicione o GoMailer como dependência:

```bash
go get github.com/seu-usuario/gomailer
```

2. Configure as variáveis de ambiente no seu serviço:

```env
GOMAILER_HOST=localhost
GOMAILER_PORT=9000
GOMAILER_AUTH_SECRET=seu-segredo-aqui
```

## Arquitetura

O serviço segue um padrão de arquitetura limpa com os seguintes componentes:

- `cmd/main.go`: Ponto de entrada da aplicação
- `config/`: Estruturas de configuração e manipulação de ambiente
- `internal/api/`: Manipuladores da API HTTP
- `internal/email/`: Serviço de envio de email
- `internal/queue/`: Implementação do consumidor RabbitMQ
- `internal/tcp/`: Servidor TCP para integração com outros serviços
- `pkg/client/`: Cliente TCP para integração externa

## Tratamento de Erros

O serviço implementa um tratamento robusto de erros:

- Validação de variáveis de ambiente
- Validação de entrada para requisições de email
- Tratamento de erros de conexão com a fila
- Tratamento de erros de envio SMTP com reenvio para a fila
- Autenticação e validação de conexões TCP
- Desligamento gracioso em sinais do sistema

## Desenvolvimento

Para executar o serviço em modo de desenvolvimento:

1. Copie e configure as variáveis de ambiente:

```bash
cp env.example .env
# Edite .env com suas configurações
```

2. Inicie o RabbitMQ:

```bash
docker-compose up -d
```

3. Execute o serviço:

```bash
go run cmd/main.go
```

## Implantação em Produção

Para implantação em produção:

1. Compile o binário:

```bash
go build -o gomailer cmd/main.go
```

2. Configure as variáveis de ambiente no seu ambiente de produção
3. Configure um gerenciador de processos (ex: systemd)
4. Configure monitoramento e logging adequados
5. Use um serviço SMTP de nível de produção
6. Configure firewalls para permitir apenas conexões TCP confiáveis

## Licença

MIT License
