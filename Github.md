# ACC Data Platform - Arquitetura do Projeto

## 1) Objetivo

Construir uma plataforma de dados baseada em telemetria do Assetto Corsa Competizione (ACC) para estudo e prática de Engenharia de Dados, cobrindo ingestao em tempo real, streaming, armazenamento em data lake (bronze) e governanca de schemas.

## 2) Fontes de dados (Shared Memory ACC)

Leitura local via shared memory:

- `Local\\acpmf_physics` (size: 800, freq: 333 Hz)
- `Local\\acpmf_graphics` (size: 1588, freq: 60 Hz)
- `Local\\acpmf_static` (size: 784, dados estaticos)

## 3) Arquitetura macro

Fluxo principal:

1. Producer em Go le shared memory.
2. Producer normaliza e serializa em Avro (3 contratos: physics, graphics, static).
3. Producer publica em batch de 1 segundo no Redpanda.
4. Schema Registry do Redpanda armazena e valida compatibilidade dos schemas Avro.
5. Consumer em Go le os topicos, agrega janelas de 5 minutos.
6. Consumer grava arquivos Avro no MinIO (camada bronze) com particionamento por usuario e tempo.
7. Se houver falha/transiente no producer, eventos ficam em buffer local com BadgerDB para reprocessamento.

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

## 7) Sistema de login e identidade por maquina no Producer

Objetivo:

- Permitir autenticacao local/remota do operador antes da captura de telemetria.
- Suportar multiplos usuarios ativos ao mesmo tempo em diferentes maquinas.
- Enriquecer eventos Avro com o usuario correto por instancia de reader.
- Manter trilha minima de seguranca (senha hasheada, sessao com expiracao, auditoria basica).

### 7.1 Componentes

- `internal/auth` (regras de login, hash de senha, controle de sessao)
- `internal/user` (cadastro e gestao de usuario ativo por maquina)
- `internal/repository/postgres` (persistencia de usuarios, sessoes e ativos por maquina)
- `internal/handler/api` (handlers e middleware de autenticacao)
- `cmd/reader` (resolucao de identidade e cache local de `machine_id`)
- `web/templates` (frontend server-side para login e gestao de usuario)

### 7.2 Fluxo de autenticacao e ativacao

1. Operador acessa `GET /login`.
2. Frontend envia `POST /auth/login` com `username` e `password`.
3. Backend valida credenciais no Postgres (senha hasheada).
4. Backend cria sessao e devolve cookie HttpOnly + Secure (quando HTTPS ativo).
5. Reader identifica sua instancia via `machine_id` persistido localmente.
6. Operador escolhe usuario ativo para uma maquina em `POST /users/active` (`user_id` + `machine_id`).
7. Reader consulta o usuario ativo da propria maquina e usa esse contexto para preencher eventos.

### 7.3 Modelo de dados (PostgreSQL)

Tabela `users`:

- `id` (uuid, pk)
- `username` (text, unique, not null)
- `name` (text, not null)
- `password_hash` (text, not null)
- `role` (text, default `operator`)
- `created_at`, `updated_at` (timestamptz)

Tabela `user_sessions`:

- `id` (uuid, pk)
- `user_id` (uuid, fk -> users.id)
- `token_hash` (text, not null)
- `expires_at` (timestamptz, not null)
- `created_at` (timestamptz)

Tabela `active_users`:

- `machine_id` (text, pk)
- `user_id` (uuid, fk -> users.id)
- `updated_at` (timestamptz, not null)


### 7.4 Regras de seguranca

- Nunca persistir senha em texto puro.
- Usar `bcrypt` (custo configuravel) ou `argon2id` para hash de senha.
- Guardar apenas hash do token de sessao (nao token puro).
- Expirar sessao por tempo e invalidar no logout.
- Aplicar limite de tentativa de login por IP/usuario.

### 7.5 Integracao com pipeline

- O reader resolve identidade por prioridade: flags manuais -> ativo por `machine_id` -> fallback.
- `machine_id` e persistido localmente para estabilidade entre reinicios.
- Cada maquina pode operar com usuario diferente sem sobrescrever estado global.
- Metadados no envelope Avro continuam padronizados (`usuario_id`, `username`, `event_time`, `ingestion_time`).

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

## 10) Estrutura de repositorio sugerida

```text
acc-data-platform/
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
		redpanda/
		minio/
	docs/
		architecture.md
```

## 11) MVP (primeira entrega)

1. Ler shared memory dos 3 blocos no producer.
2. Publicar 3 topicos Avro no Redpanda com batch de 1s.
3. Registrar schemas no Schema Registry com compatibilidade backward.
4. Consumir e gravar Avro no MinIO em janela de 5min, com path particionado.
5. Implementar cadastro/selecao simples de usuario no producer.
6. Ativar fallback de envio com BadgerDB no producer.

## 12) Evolucao depois do MVP

1. Camada silver (normalizacao e deduplicacao).
2. Camada gold (KPIs de corrida: ritmo, consistencia, degradacao de pneu).
3. Catalogacao de dados (metastore + data contracts mais rigidos).
4. Validacao automatica de schema em CI.
5. Dashboards operacionais e de dominio.

## 13) Dependencias configuradas

### 13.1 Go workspace e modulos

- Workspace Go criado em `go.work` com dois modulos:
	- `producer/go.mod`
	- `consumer/go.mod`

Bibliotecas principais do producer:

- `github.com/twmb/franz-go` (cliente Kafka para Redpanda)
- `github.com/hamba/avro/v2` (serializacao Avro)
- `github.com/dgraph-io/badger/v4` (buffer local resiliente)
- `github.com/jackc/pgx/v5` e `github.com/jmoiron/sqlx` (acesso PostgreSQL)
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
- PostgreSQL (persistencia de login, usuarios e ativo por maquina)

### 13.3 Variaveis de ambiente

Arquivo criado:

- `.env.example`

Inclui:

- Endpoints Redpanda e Schema Registry
- Nomes dos topicos
- Credenciais e endpoint do MinIO
- Configuracao do PostgreSQL e `DATABASE_URL`
- Variaveis para identidade de maquina no reader (ex.: `ACCDP_MACHINE_ID`, `ACCDP_MACHINE_ID_PATH`)
- Janela do producer (1s) e consumer (5m)
- Caminhos locais de BadgerDB

### 13.4 Comandos de bootstrap

Subir infraestrutura:

```powershell
cd infra
docker compose --env-file ../.env up -d
```

Baixar dependencias Go:

```powershell
cd ..\producer
go mod tidy
cd ..\consumer
go mod tidy
```

Verificar modulos no workspace:

```powershell
cd ..
go work sync
```

