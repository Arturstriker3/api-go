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
