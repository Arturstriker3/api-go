# 🔄 GoMailer Certificate Auto-Renewal System

## 📧 **Entrega Automática de Certificados por Email**

Quando executado no Docker com SMTP configurado, o GoMailer envia automaticamente o certificado CA para seu email:

### **Quando os Certificados são Enviados:**

- ✅ **Geração Inicial**: Quando certificados são criados pela primeira vez
- ✅ **Renovação Automática**: Quando certificados são renovados automaticamente (30 dias antes do vencimento)
- ✅ **Renovação Manual**: Quando você executa manualmente o script de renovação

### **Requisitos para Email:**

O certificado será enviado automaticamente se TODAS as condições forem atendidas:

1. 🐳 **Executando no Docker** (detecta arquivo `/.dockerenv`)
2. 📧 **SMTP configurado** com estas variáveis de ambiente:
   - `SMTP_HOST` - Seu servidor SMTP
   - `SMTP_USER` - Seu endereço de email (destinatário)
   - `SMTP_PASSWORD` - Sua senha de email/senha de app

### **Conteúdo do Email:**

O email contém:

- 📜 **Conteúdo completo do certificado CA** pronto para salvar como `ca-cert.pem`
- 📋 **Instruções de uso** para aplicações cliente
- 🔍 **Detalhes do certificado** (data de expiração, organização, etc.)
- 💡 **Exemplos de integração** NestJS/Node.js

### **Exemplo de Configuração de Email:**

```env
# No seu arquivo .env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=seu-email@gmail.com
SMTP_PASSWORD=sua-senha-de-app
SMTP_FROM=seu-email@gmail.com
```

### **Benefícios:**

- 🚀 **Sem extração manual** de certificados dos containers Docker
- 📨 **Entrega instantânea** quando certificados são gerados/renovados
- 🔄 **Atualizações automáticas** - receba certificados renovados por email
- 💾 **Backup fácil** - certificados ficam salvos no seu email

---

## 🔄 **Renovação Automática de Certificados**

Sistema de renovação automática de certificados auto-assinados para produção **sem downtime**.

## 📋 **Scripts Disponíveis**

### **1. Geração Inicial**

```bash
# Gerar certificados auto-assinados para desenvolvimento
go run -tags generate_certs scripts/generate-self-signed-certs.go
```

### **2. Renovação Automática**

```bash
# Verificar e renovar certificados (se necessário)
go run -tags renew_certs scripts/auto-renew-certs.go
```

### **3. Automação via Cron (Linux/Mac)**

```bash
# Tornar executável
chmod +x scripts/cert-renewal-cron.sh

# Adicionar ao crontab (executa diariamente às 2h)
crontab -e
# Adicionar linha:
0 2 * * * /path/to/gomailer/scripts/cert-renewal-cron.sh
```

### **4. Automação via Task Scheduler (Windows)**

```powershell
# Executar manualmente
PowerShell -ExecutionPolicy Bypass -File "scripts\cert-renewal-task.ps1"

# Ou configurar no Task Scheduler:
# - Trigger: Daily at 2:00 AM
# - Action: PowerShell -ExecutionPolicy Bypass -File "C:\path\to\gomailer\scripts\cert-renewal-task.ps1"
```

## ⚙️ **Como Funciona**

### **🔍 Verificação Automática**

- ✅ Verifica se certificados existem
- ✅ Calcula dias até expiração
- ✅ Renova automaticamente se < 30 dias
- ✅ Mantém certificados válidos se > 30 dias

### **🔄 Renovação Sem Downtime**

1. **Geração Segura**: Novos certificados em diretório temporário
2. **Backup Automático**: Certificados antigos salvos em `certs/backup_TIMESTAMP/`
3. **Substituição Atômica**: Troca instantânea dos arquivos
4. **Zero Downtime**: API continua funcionando durante o processo

### **📁 Estrutura de Arquivos**

```
certs/
├── server.crt          # Certificado atual
├── server.key          # Chave privada atual
├── ca-cert.pem         # Certificado CA
├── backup_1234567890/  # Backup automático
│   ├── server.crt
│   ├── server.key
│   └── ca-cert.pem
└── temp/               # Diretório temporário (removido após uso)
```

## 🚀 **Configuração para Produção**

### **Opção 1: Cron Job (Linux)**

```bash
# Editar crontab
crontab -e

# Adicionar (executa diariamente às 2h da manhã)
0 2 * * * /path/to/gomailer/scripts/cert-renewal-cron.sh

# Verificar logs
tail -f /path/to/gomailer/logs/cert-renewal.log
```

### **Opção 2: Windows Task Scheduler**

1. Abrir **Task Scheduler**
2. **Create Basic Task**
3. **Name**: GoMailer Certificate Renewal
4. **Trigger**: Daily at 2:00 AM
5. **Action**: Start a program
   - **Program**: `PowerShell`
   - **Arguments**: `-ExecutionPolicy Bypass -File "C:\path\to\gomailer\scripts\cert-renewal-task.ps1"`

### **Opção 3: Docker Cron**

```dockerfile
# Adicionar ao Dockerfile
RUN apt-get update && apt-get install -y cron
COPY scripts/cert-renewal-cron.sh /etc/cron.daily/gomailer-certs
RUN chmod +x /etc/cron.daily/gomailer-certs
```

### **Requisitos para Docker**

Antes de rodar com Docker, crie um arquivo `.env` baseado no `env.example`:

```bash
# Copiar o exemplo
cp env.example .env

# Editar com suas configurações
nano .env
```

**Variáveis obrigatórias para Docker:**

```env
# Configuração SMTP
SMTP_HOST=seu-smtp-host
SMTP_PORT=587
SMTP_USER=seu-email@dominio.com
SMTP_PASSWORD=sua-senha
SMTP_FROM=seu-email@dominio.com

# Configuração RabbitMQ
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=admin

# Configuração TCP/TLS
TCP_PORT=9000
TCP_AUTH_SECRET=sua-chave-secreta
TCP_ENABLED=false
TCP_TLS_ENABLED=true
TCP_TLS_CERT_PATH=certs/server.crt
TCP_TLS_KEY_PATH=certs/server.key
TCP_TLS_CA_PATH=certs/ca-cert.pem

# Métricas
METRICS_PORT=9091
```

### **Opção 4: Kubernetes CronJob**

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: gomailer-cert-renewal
spec:
  schedule: "0 2 * * *" # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: cert-renewal
              image: gomailer:latest
              command:
                [
                  "go",
                  "run",
                  "-tags",
                  "renew_certs",
                  "scripts/auto-renew-certs.go",
                ]
          restartPolicy: OnFailure
```

## 📊 **Monitoramento**

### **Métricas Prometheus**

- `gomailer_tls_certificate_expiry_days` - Dias até expiração

### **Logs**

```bash
# Ver logs de renovação
tail -f logs/cert-renewal.log

# Verificar última execução
ls -la certs/backup_*
```

### **Alertas Grafana**

Configure alertas quando:

- Certificado expira em < 7 dias
- Renovação falha
- Processo de renovação não executou

## 🔧 **Configurações Avançadas**

### **Alterar Threshold de Renovação**

```go
// Em scripts/auto-renew-certs.go, linha 45
renewThreshold := 30.0  // Alterar para dias desejados
```

### **Configurar Domínios Personalizados**

```go
// Em scripts/auto-renew-certs.go, linha 88
DNSNames: []string{"localhost", "meudominio.com"},
```

### **Hot Reload da API**

Para recarregar certificados sem reiniciar:

```go
// Implementar em internal/tcp/server.go
func (s *Server) ReloadCertificates() error {
    // Recarregar certificados TLS
    cert, err := tls.LoadX509KeyPair(s.config.TCP.TLS.CertPath, s.config.TCP.TLS.KeyPath)
    if err != nil {
        return err
    }

    // Atualizar configuração TLS
    s.tlsConfig.Certificates = []tls.Certificate{cert}
    log.Println("🔄 TLS certificates reloaded successfully")
    return nil
}
```

## ⚠️ **Considerações de Segurança**

### **✅ Vantagens dos Auto-Assinados**

- ✅ **Controle Total**: Você gerencia a renovação
- ✅ **Sem Dependências**: Não depende de serviços externos
- ✅ **Renovação Automática**: Sistema próprio de renovação
- ✅ **Zero Downtime**: Troca sem interrupção

### **⚠️ Limitações**

- ⚠️ **Browsers**: Mostram aviso de certificado não confiável
- ⚠️ **Clientes**: Precisam configurar para aceitar certificados auto-assinados
- ⚠️ **Produção Pública**: Não recomendado para APIs públicas

### **🔒 Para Produção Pública**

Se precisar de certificados confiáveis publicamente:

1. Use **Let's Encrypt** com Certbot
2. Use **AWS Certificate Manager**
3. Use **Cloudflare SSL**
4. Use certificados corporativos

## 🎯 **Resumo**

Este sistema permite usar **certificados auto-assinados em produção** com:

- ✅ **Renovação automática** (30 dias antes do vencimento)
- ✅ **Zero downtime** (substituição atômica)
- ✅ **Backup automático** (rollback se necessário)
- ✅ **Logs completos** (auditoria e debug)
- ✅ **Multiplataforma** (Linux, Windows, Docker, Kubernetes)

**Ideal para**: APIs internas, microserviços, ambientes corporativos onde o controle total é mais importante que a confiança pública dos certificados.
