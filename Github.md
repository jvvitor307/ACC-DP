# ACC Data Platform - Arquitetura do Projeto

## 1) Objetivo

Construir uma plataforma de dados baseada em telemetria do Assetto Corsa Competizione (ACC) para estudo e pratica de Engenharia de Dados, cobrindo ingestao em tempo real, streaming, armazenamento em data lake (bronze), governanca de schemas e uma API backend de autenticacao/sessao para controle de usuario ativo por maquina.

## 2) Fontes de dados (Shared Memory ACC)

Leitura local via shared memory:

- `Local\\acpmf_physics` (size: 800, freq: 333 Hz)
- `Local\\acpmf_graphics` (size: 1588, freq: 60 Hz)
- `Local\\acpmf_static` (size: 784, dados estaticos)

## 3) Arquitetura macro

Fluxo principal de dados:

1. Producer em Go le shared memory.
2. Producer normaliza e serializa em Avro (3 contratos: physics, graphics, static).
3. Producer publica em batch de 1 segundo no Redpanda.
4. Schema Registry do Redpanda armazena e valida compatibilidade dos schemas Avro.
5. Consumer em Go le os topicos, agrega janelas de 5 minutos.
6. Consumer grava arquivos Avro no MinIO (camada bronze) com particionamento por usuario e tempo.
7. Se houver falha/transiente no producer, eventos ficam em buffer local com BadgerDB para reprocessamento.

Fluxo de autenticacao (backend):

1. Aplicacao cliente (na maquina do usuario) gera e persiste `machine_id` local.
2. Cliente chama `POST /auth/login` enviando credenciais e `machine_id`.
3. Backend registra/atualiza a maquina no Postgres (upsert por `machine_uid`).
4. Backend revoga sessao ativa anterior do usuario e tambem qualquer sessao ativa da maquina.
5. Backend cria nova sessao e emite `access_token` (JWT) + `refresh_token` (rotativo).
6. Em `POST /auth/refresh`, o refresh token antigo e invalidado e um novo par de tokens e emitido.
7. Em `POST /auth/logout`, sessao e refresh tokens associados sao revogados.

## 4) Componentes

### 4.1 Producer (Go)

Responsabilidades:

- Ler `physics`, `graphics` e `static` da shared memory.
- Enriquecer metadados de usuario selecionado (`usuario_id`, `username`, timestamp de captura, fonte).
- Serializar mensagens em Avro (subject separado por tipo no Schema Registry).
- Publicar no Redpanda em lote a cada 1 segundo.
- Persistir no BadgerDB quando nao conseguir enviar (retry assíncrono).

Submodulos sugeridos:

- `internal/source/acc_shm` (leitura shared memory)
- `internal/avro` (serializacao + registro/lookup de schema)
- `internal/broker/redpanda` (producer kafka API compatível)
- `internal/buffer/badger` (durable local buffer)
- `internal/user` (cadastro e selecao de usuario)
- `internal/batch` (flush de 1s)

### 4.2 Redpanda

Topicos (1 por tipo):

- `acc.physics.v1`
- `acc.graphics.v1`
- `acc.static.v1`

Schema Registry:

- Subject strategy por topico (ex.: `acc.physics.v1-value`).
- Politica de compatibilidade recomendada: `BACKWARD`.
- Evolucao de schema com versionamento controlado (adicoes opcionais, sem quebra).

### 4.3 Consumer (Go)

Responsabilidades:

- Consumir os 3 topicos.
- Desserializar Avro com schema versionado.
- Agrupar por chave logica: `usuario_id + tipo + janela_5min`.
- Gerar arquivos Avro de saida com janela fixa de 5 minutos.
- Gravar no MinIO com path padronizado da bronze.

Submodulos sugeridos:

- `internal/consumer` (poll/commit/offset)
- `internal/window` (agregacao 5 minutos)
- `internal/sink/minio` (upload + idempotencia)
- `internal/filewriter/avro` (append/flush de arquivo)

### 4.4 MinIO (Data Lake Bronze)

Bucket sugerido: `datalake-bronze`

Paths:

- `bronze/acc_physics/usuario_id=123/year=2026/month=03/day=22/hour=17/0500_0959_8f2a1b9d.avro`
- `bronze/acc_graphics/usuario_id=123/year=2026/month=03/day=22/hour=17/0500_0959_8f2a1b9d.avro`
- `bronze/acc_statics/usuario_id=123/year=2026/month=03/day=22/hour=17/0500_0959_8f2a1b9d.avro`

Padrao de nome de arquivo (gold standard):

- `<minuto_segundo_inicio>_<minuto_segundo_fim>_<uuid_ou_worker_id>.avro`

Exemplo:

- `0500_0959_8f2a1b9d.avro`

### 4.5 Backend Auth (Go + Gin + Postgres)

Responsabilidades:

- Expor endpoints `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout` e `GET /healthz`.
- Registrar/atualizar maquina automaticamente no login a partir do `machine_id` enviado pelo cliente.
- Garantir regra de negocio: um usuario ativo em um unico computador por vez.
- Garantir regra complementar: uma maquina com um unico usuario ativo por vez.
- Emitir JWT de acesso curto e refresh token rotativo (one-time).
- Persistir estado de autenticacao nas tabelas `users`, `machines`, `sessions` e `refresh_tokens`.
- Publicar documentacao Swagger em `GET /swagger` e especificacao OpenAPI em `GET /swagger/openapi.yaml`.

Submodulos implementados:

- `backend/cmd/server`
- `backend/internal/adapter/http`
- `backend/internal/service/auth`
- `backend/internal/repository`
- `backend/internal/database/migrations`

## 5) Contratos de dados (Avro)

Arquivos de schema (um por tipo):

- `schemas/acc_physics.avsc`
- `schemas/acc_graphics.avsc`
- `schemas/acc_static.avsc`

Campos comuns recomendados em todos os schemas:

- `event_id` (string, UUID)
- `event_time` (long, timestamp millis)
- `ingestion_time` (long, timestamp millis)
- `usuario_id` (string)
- `username` (string)
- `source` (string, ex.: `acpmf_physics`)
- `schema_version` (int)
- `payload` (record especifico por tipo)

Observacao:

- Campos do `payload` devem refletir 1:1 os dados do shared memory para manter fidelidade da bronze.

## 6) Estrategia de buffering e confiabilidade

Producer (1 segundo):

- Coleta continua dos sinais.
- Acumula eventos por tipo durante 1s.
- Publica lote no topico correspondente.

Fallback BadgerDB:

- Se publish falhar: grava evento/lote localmente.
- Worker de retry tenta reenviar com backoff.
- Somente remove do Badger apos confirmacao de envio.

Consumer (5 minutos):

- Acumula por janela fixa de 5 minutos.
- No fechamento da janela, grava arquivo Avro no MinIO.
- Commit de offset somente apos upload concluido com sucesso.

## 7) Sistema de usuario no Producer

Escopo simples (local):

- Cadastro com `nome` e `username`.
- Seletor de usuario ativo antes de iniciar captura.
- Usuario ativo entra como metadado em todos os eventos.

Persistencia sugerida:

- Arquivo local simples (`users.json`) ou BadgerDB em namespace separado.

## 8) Particionamento e chaveamento

Recomendacoes:

- Chave da mensagem no broker: `usuario_id`.
- Beneficio: ordenacao por usuario e melhor localidade de consumo.
- Particoes por topico: iniciar com 3 a 6 e ajustar por throughput.

## 9) Qualidade e observabilidade

Metricas minimas:

- Eventos lidos por tipo (physics/graphics/static)
- Latencia de ingestao e de flush (1s e 5min)
- Taxa de erro de serializacao/publicacao
- Tamanho do backlog no BadgerDB
- Tempo de upload no MinIO

Logs estruturados:

- `event_id`, `usuario_id`, `topic`, `partition`, `offset`, `window_start`, `window_end`, `object_key`

## 10) Estrutura de repositorio atual

```text
acc-data-platform/
	backend/
		cmd/server/main.go
		internal/
			adapter/http/
			service/auth/
			repository/
			database/migrations/
		Dockerfile
	producer/
		cmd/producer/main.go
		internal/
			source/acc_shm/
			avro/
			broker/redpanda/
			batch/
			buffer/badger/
			user/
		schemas/
			acc_physics.avsc
			acc_graphics.avsc
			acc_static.avsc
	consumer/
		cmd/consumer/main.go
		internal/
			consumer/
			avro/
			window/
			filewriter/avro/
			sink/minio/
	infra/
		docker-compose.yml
	go.work
	.env.example
```

## 11) MVP (primeira entrega)

1. Ler shared memory dos 3 blocos no producer.
2. Publicar 3 topicos Avro no Redpanda com batch de 1s.
3. Registrar schemas no Schema Registry com compatibilidade backward.
4. Consumir e gravar Avro no MinIO em janela de 5min, com path particionado.
5. Implementar cadastro/selecao simples de usuario no producer.
6. Ativar fallback de envio com BadgerDB no producer.
7. Subir backend auth com Postgres e endpoints de autenticacao.
8. Publicar Swagger/OpenAPI do backend para teste e integracao.

## 12) Evolucao depois do MVP

1. Camada silver (normalizacao e deduplicacao).
2. Camada gold (KPIs de corrida: ritmo, consistencia, degradacao de pneu).
3. Catalogacao de dados (metastore + data contracts mais rigidos).
4. Validacao automatica de schema em CI.
5. Dashboards operacionais e de dominio.

## 13) Dependencias configuradas

### 13.1 Go workspace e modulos

- Workspace Go criado em `go.work` com tres modulos:
	- `backend/go.mod`
	- `producer/go.mod`
	- `consumer/go.mod`

Bibliotecas principais do backend:

- `github.com/gin-gonic/gin` (API HTTP)
- `github.com/golang-jwt/jwt/v5` (tokens JWT)
- `github.com/jackc/pgx/v5` (driver Postgres)
- `github.com/google/uuid` (identificadores)
- `golang.org/x/crypto` (bcrypt para senha)

Bibliotecas principais do producer:

- `github.com/twmb/franz-go` (cliente Kafka para Redpanda)
- `github.com/hamba/avro/v2` (serializacao Avro)
- `github.com/dgraph-io/badger/v4` (buffer local resiliente)
- `github.com/spf13/cobra` e `github.com/spf13/viper` (CLI e configuracao)
- `go.uber.org/zap` (logging)
- `github.com/google/uuid` (identificadores)
- `golang.org/x/sys` (suporte baixo nivel para Windows/shared memory)

Bibliotecas principais do consumer:

- `github.com/twmb/franz-go` (consumo Redpanda)
- `github.com/hamba/avro/v2` (desserializacao/serializacao)
- `github.com/minio/minio-go/v7` (envio para MinIO)
- `go.uber.org/zap` (logging)
- `github.com/google/uuid` (identificadores de arquivo/worker)

### 13.2 Infra local com containers

Arquivo criado:

- `infra/docker-compose.yml`

Servicos inclusos:

- Redpanda broker
- Redpanda Console
- MinIO
- MinIO client (`mc`) para criar bucket `datalake-bronze`
- Postgres
- Backend Auth

### 13.3 Variaveis de ambiente

Arquivo criado:

- `.env.example`

Inclui:

- Endpoints Redpanda e Schema Registry
- Nomes dos topicos
- Credenciais e endpoint do MinIO
- Variaveis de Postgres
- Variaveis de autenticacao do backend (JWT e TTLs)
- Porta externa do backend
- Janela do producer (1s) e consumer (5m)
- Caminhos locais de usuarios e BadgerDB

### 13.4 Comandos de bootstrap

Subir infraestrutura:

```powershell
cd C:\Programming\ACC-DP
docker compose --env-file .env -f infra/docker-compose.yml up -d
```

Subir somente backend + postgres:

```powershell
cd C:\Programming\ACC-DP
docker compose --env-file .env -f infra/docker-compose.yml up -d postgres backend
```

Baixar dependencias Go:

```powershell
cd C:\Programming\ACC-DP\backend
go mod tidy
cd ..\producer
go mod tidy
cd ..\consumer
go mod tidy
```

Verificar modulos no workspace:

```powershell
cd C:\Programming\ACC-DP
go work sync
```

Validar backend:

```powershell
curl.exe -sS http://localhost:18088/healthz
start http://localhost:18088/swagger
```

