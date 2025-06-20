# GoMailer

<div align="center">

[🇧🇷 Português](#português) | [🇺🇸 English](README.en.md)

</div>

# Português

Um microsserviço para manipulação de envio de emails através de uma fila RabbitMQ, construído com Go. Fornece uma interface TCP segura para integração com outros serviços.

## Funcionalidades

- Servidor TCP para integração com serviços
- Integração com RabbitMQ para enfileiramento confiável de mensagens
- Envio de emails via SMTP com suporte a HTML
- Configuração baseada em variáveis de ambiente
- Suporte a Docker para RabbitMQ
- Métricas Prometheus e dashboards Grafana

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

4. Inicie a infraestrutura usando Docker Compose:

```bash
docker-compose up -d
```

5. Execute a aplicação:

```bash
go run cmd/main.go
```

O serviço iniciará o servidor TCP na porta 9000 (padrão) e métricas na porta 9091.

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
- `TCP_PORT`: Porta do servidor TCP (padrão: "9000")
- `METRICS_PORT`: Porta das métricas Prometheus (padrão: "9091")

## Integração via TCP

Para integrar outros serviços com o GoMailer, você pode usar o cliente TCP fornecido:

```go
package main

import (
    "log"
    "os"
    "gomailer/pkg/client"
)

func main() {
    // Criar cliente de email
    emailClient := client.NewEmailClient(
        os.Getenv("GOMAILER_HOST"),     // Host do serviço
        os.Getenv("GOMAILER_PORT"),     // Porta TCP
        os.Getenv("GOMAILER_AUTH_SECRET"), // Chave de autenticação
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
go get github.com/Arturstriker3/gomailer
```

2. Configure as variáveis de ambiente no seu serviço:

```env
GOMAILER_HOST=localhost
GOMAILER_PORT=9000
GOMAILER_AUTH_SECRET=seu-segredo-aqui
```

### Exemplo de Integração com NestJS

Aqui está como integrar o GoMailer em uma aplicação NestJS:

1. Crie um serviço de cliente TCP:

```typescript
// src/services/gomailer.service.ts
import { Injectable, OnModuleInit } from "@nestjs/common";
import { Socket } from "net";

interface EmailRequest {
  to: string[];
  subject: string;
  body: string;
}

@Injectable()
export class GomailerService implements OnModuleInit {
  private client: Socket;
  private connected: boolean = false;

  constructor() {
    this.client = new Socket();
  }

  async onModuleInit() {
    await this.connect();
  }

  private connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.client.connect(
        {
          host: process.env.GOMAILER_HOST || "localhost",
          port: parseInt(process.env.GOMAILER_PORT || "9000"),
        },
        () => {
          this.connected = true;
          // Enviar autenticação
          const auth = { secret: process.env.GOMAILER_AUTH_SECRET };
          this.client.write(JSON.stringify(auth));
          resolve();
        }
      );

      this.client.on("error", (error) => {
        this.connected = false;
        reject(error);
      });

      this.client.on("close", () => {
        this.connected = false;
      });
    });
  }

  async sendEmail(request: EmailRequest): Promise<void> {
    if (!this.connected) {
      await this.connect();
    }

    return new Promise((resolve, reject) => {
      this.client.write(JSON.stringify(request));

      this.client.once("data", (data) => {
        const response = JSON.parse(data.toString());
        if (response.error) {
          reject(new Error(response.error));
        } else {
          resolve();
        }
      });
    });
  }

  onModuleDestroy() {
    if (this.client) {
      this.client.destroy();
    }
  }
}
```

2. Registre o serviço no seu módulo:

```typescript
// src/app.module.ts
import { Module } from "@nestjs/common";
import { ConfigModule } from "@nestjs/config";
import { GomailerService } from "./services/gomailer.service";

@Module({
  imports: [
    ConfigModule.forRoot(), // Para variáveis de ambiente
  ],
  providers: [GomailerService],
  exports: [GomailerService],
})
export class AppModule {}
```

3. Use o serviço nos seus controllers/services:

```typescript
// src/controllers/email.controller.ts
import { Controller, Post, Body } from "@nestjs/common";
import { GomailerService } from "../services/gomailer.service";

@Controller("email")
export class EmailController {
  constructor(private readonly gomailerService: GomailerService) {}

  @Post()
  async sendEmail(
    @Body() emailData: { to: string[]; subject: string; body: string }
  ) {
    try {
      await this.gomailerService.sendEmail(emailData);
      return { message: "Email enfileirado com sucesso" };
    } catch (error) {
      throw new Error(`Erro ao enviar email: ${error.message}`);
    }
  }
}
```

4. Configure as variáveis de ambiente no seu `.env`:

```env
GOMAILER_HOST=localhost
GOMAILER_PORT=9000
GOMAILER_AUTH_SECRET=seu-segredo-aqui
```

O serviço NestJS gerencia:

- Conexão automática
- Autenticação com GoMailer
- Reconexão em falhas
- Desligamento limpo
- Segurança de tipos com TypeScript

## Monitoramento

O serviço expõe métricas Prometheus e inclui um dashboard Grafana pré-configurado:

- Métricas Prometheus: http://localhost:9091/metrics
- Dashboard Grafana: http://localhost:3000 (credenciais padrão: admin/admin)

O dashboard inclui:

- Taxa de enfileiramento e envio de emails
- Tamanho da fila e latência de processamento
- Métricas de conexões TCP
- Taxa de erros

## Arquitetura

O serviço segue um padrão de arquitetura limpa com os seguintes componentes:

- `cmd/main.go`: Ponto de entrada da aplicação
- `config/`: Estruturas de configuração e manipulação de ambiente
- `internal/email/`: Serviço de envio de email
- `internal/queue/`: Implementação do consumidor RabbitMQ
- `internal/tcp/`: Servidor TCP para integração com outros serviços
- `pkg/client/`: Cliente TCP para integração externa

## Tratamento de Erros

O serviço implementa um tratamento robusto de erros:

- Validação de variáveis de ambiente
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

2. Inicie a infraestrutura:

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
