# GoMailer

<div align="center">
  <a href="#" onclick="switchLanguage('en')" id="en-button">
    <img src="https://img.shields.io/badge/lang-en-red.svg" alt="English">
  </a>
  <a href="#" onclick="switchLanguage('pt-br')" id="pt-br-button">
    <img src="https://img.shields.io/badge/lang-pt--br-green.svg" alt="Português">
  </a>
</div>

<div id="readme-en">

A microservice for handling email sending through a RabbitMQ queue, built with Go.

## Features

- REST API for queuing emails
- RabbitMQ integration for reliable message queuing
- SMTP email sending with HTML support
- Environment-based configuration
- Docker support for RabbitMQ

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose
- SMTP server credentials (e.g., Gmail SMTP)

## Setup

1. Clone the repository:

```bash
git clone <repository-url>
cd gomailer
```

2. Install dependencies:

```bash
go mod download
```

3. Set up environment variables:

   - Copy the example environment file:

   ```bash
   cp env.example .env
   ```

   - Edit the `.env` file with your configuration

4. Start RabbitMQ using Docker Compose:

```bash
docker-compose up -d
```

5. Run the application:

```bash
go run cmd/main.go
```

The service will start on the configured port (default: 8080).

## Environment Variables

### Required Variables

- `SMTP_USER`: SMTP server username (required)
- `SMTP_PASSWORD`: SMTP server password (required)
- `SMTP_FROM`: Email address to send from (required)

### Optional Variables with Defaults

- `SMTP_HOST`: SMTP server host (default: "smtp.gmail.com")
- `SMTP_PORT`: SMTP server port (default: 587)
- `RABBITMQ_HOST`: RabbitMQ host (default: "localhost")
- `RABBITMQ_PORT`: RabbitMQ port (default: "5672")
- `RABBITMQ_USER`: RabbitMQ username (default: "admin")
- `RABBITMQ_PASSWORD`: RabbitMQ password (default: "admin")
- `API_PORT`: API server port (default: "8080")

## API Usage

### Queue an Email

```http
POST /email
Content-Type: application/json

{
  "to": ["recipient@example.com"],
  "subject": "Hello",
  "body": "<h1>Hello World</h1><p>This is a test email.</p>"
}
```

Response:

```json
{
  "message": "Email queued successfully"
}
```

## Architecture

The service follows a clean architecture pattern with the following components:

- `cmd/main.go`: Application entry point
- `config/`: Configuration structures and environment handling
- `internal/api/`: HTTP API handlers
- `internal/email/`: Email sending service
- `internal/queue/`: RabbitMQ consumer implementation

## Error Handling

The service implements robust error handling:

- Environment variable validation
- Input validation for email requests
- Queue connection error handling
- SMTP sending error handling with message requeuing
- Graceful shutdown on system signals

## Development

To run the service in development mode:

1. Copy and configure environment variables:

```bash
cp env.example .env
# Edit .env with your settings
```

2. Start RabbitMQ:

```bash
docker-compose up -d
```

3. Run the service:

```bash
go run cmd/main.go
```

## Production Deployment

For production deployment:

1. Build the binary:

```bash
go build -o gomailer cmd/main.go
```

2. Set up environment variables in your production environment
3. Configure a process manager (e.g., systemd)
4. Set up proper monitoring and logging
5. Use a production-grade SMTP service

## License

MIT License

</div>

<div id="readme-pt-br" style="display: none;">

Um microsserviço para manipulação de envio de emails através de uma fila RabbitMQ, construído com Go.

## Funcionalidades

- API REST para enfileiramento de emails
- Integração com RabbitMQ para enfileiramento confiável de mensagens
- Envio de emails via SMTP com suporte a HTML
- Configuração baseada em variáveis de ambiente
- Suporte a Docker para RabbitMQ

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

O serviço iniciará na porta configurada (padrão: 8080).

## Variáveis de Ambiente

### Variáveis Obrigatórias

- `SMTP_USER`: Usuário do servidor SMTP (obrigatório)
- `SMTP_PASSWORD`: Senha do servidor SMTP (obrigatório)
- `SMTP_FROM`: Endereço de email de envio (obrigatório)

### Variáveis Opcionais com Valores Padrão

- `SMTP_HOST`: Host do servidor SMTP (padrão: "smtp.gmail.com")
- `SMTP_PORT`: Porta do servidor SMTP (padrão: 587)
- `RABBITMQ_HOST`: Host do RabbitMQ (padrão: "localhost")
- `RABBITMQ_PORT`: Porta do RabbitMQ (padrão: "5672")
- `RABBITMQ_USER`: Usuário do RabbitMQ (padrão: "admin")
- `RABBITMQ_PASSWORD`: Senha do RabbitMQ (padrão: "admin")
- `API_PORT`: Porta do servidor API (padrão: "8080")

## Uso da API

### Enfileirar um Email

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

## Arquitetura

O serviço segue um padrão de arquitetura limpa com os seguintes componentes:

- `cmd/main.go`: Ponto de entrada da aplicação
- `config/`: Estruturas de configuração e manipulação de ambiente
- `internal/api/`: Manipuladores da API HTTP
- `internal/email/`: Serviço de envio de email
- `internal/queue/`: Implementação do consumidor RabbitMQ

## Tratamento de Erros

O serviço implementa um tratamento robusto de erros:

- Validação de variáveis de ambiente
- Validação de entrada para requisições de email
- Tratamento de erros de conexão com a fila
- Tratamento de erros de envio SMTP com reenvio para a fila
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

## Licença

MIT License

</div>

<script>
function switchLanguage(lang) {
    document.getElementById('readme-en').style.display = lang === 'en' ? 'block' : 'none';
    document.getElementById('readme-pt-br').style.display = lang === 'pt-br' ? 'block' : 'none';
    
    document.getElementById('en-button').style.opacity = lang === 'en' ? '1' : '0.5';
    document.getElementById('pt-br-button').style.opacity = lang === 'pt-br' ? '1' : '0.5';
}

// Initialize with English
switchLanguage('en');
</script>
