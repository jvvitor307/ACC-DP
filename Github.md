# ACC Data Platform - Arquitetura do Projeto

## 1) Objetivo

Construir uma plataforma de dados baseada em telemetria do Assetto Corsa Competizione (ACC) para estudo e pratica de Engenharia de Dados, cobrindo ingestao em tempo real, streaming, armazenamento em data lake (bronze) como Parquet particionado, governanca de schemas e uma API backend de autenticacao/sessao para controle de usuario ativo por maquina.

## 2) Fontes de dados (Shared Memory ACC)

Leitura local via shared memory:

- `Local\\acpmf_physics` (size: 800, freq: ~333 Hz, intervalo 3ms)
- `Local\\acpmf_graphics` (size: 1588, freq: ~333 Hz, intervalo 3ms)
- `Local\\acpmf_static` (size: 784, freq: 0.2 Hz, intervalo 5s)

## 3) Arquitetura macro

Fluxo principal de dados:

1. App CLI interativo (Go) oferece menu com gestao de usuarios e acao de iniciar o producer.
2. Producer em Go le shared memory, normaliza e serializa em Avro (3 contratos: physics, graphics, static).
3. Producer publica em batch de 1 segundo no Redpanda.
4. Schema Registry do Redpanda armazena e valida compatibilidade dos schemas Avro.
5. Redpanda Connect le os topicos, desserializa Avro via Schema Registry, aplica transformations em Bloblang (flatten envelope + payload, derivar colunas de particao) e agrega em janelas de 5 minutos.
6. Redpanda Connect grava arquivos Parquet no MinIO (camada bronze) com particionamento por `usuario_id`, `session_id` e tempo.
7. Se houver falha/transiente no producer, eventos ficam em buffer local com BadgerDB para reprocessamento.

Fluxo de autenticacao (backend):

1. Aplicacao cliente (na maquina do usuario) gera e persiste `machine_id` local (via biblioteca `machineid`).
2. Cliente chama `POST /auth/login` enviando credenciais e `machine_id`.
3. Backend registra/atualiza a maquina no Postgres (upsert por `machine_uid`).
4. Backend revoga sessao ativa anterior do usuario e tambem qualquer sessao ativa da maquina.
5. Backend cria nova sessao e emite `access_token` (JWT) + `refresh_token` (rotativo).
6. Em `POST /auth/refresh`, o refresh token antigo e invalidado e um novo par de tokens e emitido.
7. Em `POST /auth/logout`, sessao e refresh tokens associados sao revogados.

## 4) Componentes

### 4.1 App CLI (Go) — Entry point principal

Responsabilidades:

- Fornecer menu interativo via terminal com 6 opcoes:
  1. Registrar usuario
  2. Login
  3. Trocar usuario
  4. Listar usuarios
  5. Quem sou eu
  6. Iniciar producer
- Delegar operacoes de usuario para o modulo `user/`.
- Iniciar o producer chamando `pipeline.Run(ctx)` do modulo `producer/`.

Submodulos:

- `app/main.go` (menu interativo, loop de entrada)
- `app/user.go` (register, login, switch, list, whoami usando `acc-dp/user`)
- `app/producer.go` (chamada ao pipeline do producer)

Dependencias principais:

- `acc-dp/user` (modulo workspace para gestao de usuarios)
- `acc-dp/producer/pipeline` (modulo workspace para captura e publicacao)
- `github.com/denisbrodbeck/machineid` (geracao de `machine_id` unico)
- `golang.org/x/term` (leitura de senha sem eco no terminal)

### 4.2 Modulo User (Go) — Standalone

Responsabilidades:

- Cliente HTTP para `POST /auth/register` e `POST /auth/login` no backend.
- Cache local em `users.json` com escrita atomica (`*.tmp` + `os.Rename`).
- Suporte a `UpsertUser`, `SetActive`, `Active`, `List`, `Reload`.
- Watcher que detecta mudancas no usuario ativo em runtime (polling a cada 5s).

Submodulos:

- `user/model.go` — structs `User` e `Cache` para persistencia JSON.
- `user/client.go` — cliente HTTP para backend auth.
- `user/store.go` — cache local em `users.json`.
- `user/watcher.go` — poller de arquivo para deteccao de mudanca do usuario ativo.
- `user/store_test.go` — testes cobrindo CRUD, persistencia e escrita atomica.

### 4.3 Producer (Go)

Responsabilidades:

- Ler `physics`, `graphics` e `static` da shared memory com intervalos configuraveis (3ms, 3ms, 5s).
- Enriquecer metadados de usuario selecionado (`usuario_id`, `username`, `session_id`, timestamp de captura, fonte).
- Serializar mensagens em Avro (subject separado por tipo no Schema Registry).
- Publicar no Redpanda em lote a cada 1 segundo.
- Persistir no BadgerDB quando nao conseguir enviar (retry assincrono).

Submodulos implementados:

- `cmd/producer/main.go` (entry point — chama `pipeline.Run`)
- `pipeline/pipeline.go` (orquestracao central, 427 linhas)
- `internal/source/acc_shm` (leitura shared memory)
- `internal/service/avro` (serializacao + registro/lookup de schema)
- `internal/service/normalizer` (normalizacao de eventos + session tracking)
- `internal/broker/redpanda` (producer kafka API compativel)
- `internal/buffer/badger` (durable local buffer)
- `internal/batch` (flush de 1s)
- `internal/worker` (retry assincrono de eventos em buffer)
- `internal/config` (carregamento de env vars)
- `internal/domain/event` (modelos de dominio dos eventos)
- `internal/metrics` (coleta e report de metricas)

### 4.4 Redpanda

Topicos (1 por tipo):

- `acc.physics.v1`
- `acc.graphics.v1`
- `acc.static.v1`

Schema Registry:

- Subject strategy por topico (ex.: `acc.physics.v1-value`).
- Politica de compatibilidade recomendada: `BACKWARD`.
- Evolucao de schema com versionamento controlado (adicoes opcionais, sem quebra).

### 4.5 Redpanda Connect (Consumer)

Responsabilidades:

- Consumir os 3 topicos com input `kafka` nativo do Redpanda Connect.
- Desserializar Avro via Schema Registry integrado (`schema_registry_enabled: true`).
- Aplicar processors Bloblang para:
  - Flattening do envelope + payload em um unico registro plano.
  - Derivar colunas de particao: `year`, `month`, `day`, `hour` a partir de `event_time`.
- Agregar em janelas tumbling de 5 minutos com processor `window` (por composite key `usuario_id + session_id`).
- Gravar arquivos Parquet no MinIO via output `s3` com particionamento por `usuario_id`, `session_id`, `year`, `month`, `day`, `hour`.
- Gerenciamento de offsets e consumer groups nativo do Redpanda Connect.
- Retry automatico e resiliencia configurados nativamente pelo framework.

Configuracao (YAML):

- `consumer/configs/acc_physics.yaml` — Pipeline para topico `acc.physics.v1`
- `consumer/configs/acc_graphics.yaml` — Pipeline para topico `acc.graphics.v1`
- `consumer/configs/acc_static.yaml` — Pipeline para topico `acc.static.v1`

Cada pipeline segue o padrao:

```yaml
input:
  kafka:
    addresses: ["${REDPANDA_BROKERS}"]
    topics: ["${TOPIC}"]
    consumer_group: "acc-dp-connect"
    start_from_oldest: true
    schema_registry_enabled: true
    schema_registry_url: "${SCHEMA_REGISTRY_URL}"

processors:
  - mapping: |
      root = this
      root = root.merge(this.payload)
      root.payload = deleted()
      root.year = (this.event_time / 1000).number().floor()
      # ... derivar month, day, hour do event_time
  
  - window:
      timestamp_path: meta.kafka_timestamp
      size: 5m

output:
  s3:
    endpoint: "${MINIO_ENDPOINT}"
    bucket: "${MINIO_BUCKET}"
    path: "bronze/${TABLE}/${! json(\"usuario_id\") }/${! json(\"session_id\") }/year=${! json(\"year\") }/month=${! json(\"month\") }/day=${! json(\"day\") }/hour=${! json(\"hour\") }/"
    encoding: parquet
    force_path_style: true
```

Vantagens sobre o consumer Python anterior:

- Sem codigo customizado — configuracao declarativa em YAML.
- Resiliencia nativa (retry, backoff, checkpoint de offsets).
- Sem necessidade de WAL — o proprio Kafka garante at-least-once via consumer groups.
- Binario unico e leve — sem dependencias Python, PyArrow, etc.
- Facil de escalar horizontalmente (basta aumentar `consumer_group` instances).

### 4.6 MinIO (Data Lake Bronze)

Bucket: `datalake-bronze`

Tabelas Parquet:

- `bronze/acc_physics/` (arquivos Parquet particionados por `usuario_id`, `session_id`, `year`, `month`, `day`, `hour`)
- `bronze/acc_graphics/` (arquivos Parquet particionados por `usuario_id`, `session_id`, `year`, `month`, `day`, `hour`)
- `bronze/acc_statics/` (arquivos Parquet particionados por `usuario_id`, `session_id`, `year`, `month`, `day`, `hour`)

### 4.7 Backend Auth (Go + Gin + Postgres)

Responsabilidades:

- Expor endpoints `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout` e `GET /healthz`.
- Registrar/atualizar maquina automaticamente no login a partir do `machine_id` enviado pelo cliente.
- Garantir regra de negocio: um usuario ativo em um unico computador por vez.
- Garantir regra complementar: uma maquina com um unico usuario ativo por vez.
- Emitir JWT de acesso curto e refresh token rotativo (one-time).
- Persistir estado de autenticacao nas tabelas `users`, `machines`, `sessions` e `refresh_tokens`.
- Publicar documentacao Swagger em `GET /swagger` e especificacao OpenAPI em `GET /swagger/openapi.yaml`.

Submodulos implementados:

- `cmd/server/main.go`
- `internal/adapter/http` (router Gin, handlers, swagger)
- `internal/service/auth` (logica de negocio, JWT, sessoes)
- `internal/repository` (acesso a dados Postgres)
- `internal/database/migrations` (migracoes SQL)
- `internal/database/postgres.go` (conexao)
- `internal/database/migrate.go` (execucao de migracoes)
- `internal/config` (carregamento de configuracao)
- `internal/domain` (modelos + erros de dominio)

## 5) Contratos de dados (Avro)

Schemas sao registrados programaticamente no Schema Registry pelo producer (nao existem arquivos `.avsc` no repositorio).

Campos comuns em todos os envelopes:

- `event_id` (string, UUID)
- `event_time` (long, timestamp millis)
- `ingestion_time` (long, timestamp millis)
- `session_id` (string, UUID da sessao ACC)
- `usuario_id` (string)
- `username` (string)
- `source` (string, ex.: `acpmf_physics`, `acpmf_graphics`, `acpmf_static`)
- `payload` (record especifico por tipo — campos refletem 1:1 os dados do shared memory)

Observacoes:

- A versao do schema e resolvida dinamicamente pelo Schema Registry no Redpanda Connect, nao faz parte do envelope.
- Cada topico tem seu proprio subject no Schema Registry (ex.: `acc.physics.v1-value`).

## 6) Estrategia de buffering e confiabilidade

### Producer (1 segundo)

- Coleta continua dos sinais com intervalos configuraveis (`PRODUCER_PHYSICS_INTERVAL`, `PRODUCER_GRAPHICS_INTERVAL`, `PRODUCER_STATIC_INTERVAL`).
- Acumula eventos por tipo durante 1s.
- Publica lote no topico correspondente.

Fallback BadgerDB:

- Se publish falhar: grava evento/lote localmente.
- Worker de retry tenta reenviar com backoff.
- Somente remove do Badger apos confirmacao de envio.

### Consumer — Redpanda Connect (5 minutos)

- Redpanda Connect consome os topicos com consumer group nativo.
- Janela tumbling de 5 minutos configurada via processor `window`.
- Processors Bloblang fazem flattening do envelope + payload e derivam colunas de particao.
- Output S3 escreve arquivos Parquet particionados no MinIO.
- Resiliencia nativa: retry, backoff e checkpoint de offsets gerenciados pelo framework.
- Sem necessidade de WAL — o Kafka garante at-least-once via consumer groups e offset commits automaticos.

## 7) Sistema de usuario

Integracao com backend auth + cache local via modulo standalone `user/`:

- Registro: opcao "1" no menu interativo chama `POST /auth/register` no backend e persiste usuario no cache local (`data/users.json`).
- Login: opcao "2" chama `POST /auth/login` no backend, armazena identidade no cache local e define como ativo. Envia `machine_id` gerado automaticamente.
- Troca de usuario: opcao "3" (alias para login). Se o producer estiver rodando, o watcher detecta a mudanca no `users.json` via polling a cada 5s e atualiza o usuario ativo sem reiniciar.
- Consulta: opcao "4" lista usuarios em cache com marcador de ativo; opcao "5" mostra o usuario ativo atual.

Fluxo de startup:

1. App exibe menu interativo.
2. Usuario gerencia contas (opcoes 1-5) ou inicia o producer (opcao 6).
3. Ao iniciar producer: carrega cache local (`users.json`), busca usuario ativo, inicia watcher de polling.
4. Constroi `normalizer.Identity` a partir do usuario ativo.
5. Captura dados com a identidade do usuario ativo, trocando automaticamente se o watcher detectar mudanca.

## 8) Particionamento e chaveamento

Recomendacoes:

- Chave da mensagem no broker: `usuario_id`.
- Beneficio: ordenacao por usuario e melhor localidade de consumo.
- Particoes por topico: iniciar com 3 a 6 e ajustar por throughput.

Particionamento no data lake (Parquet):

- Colunas de particao: `usuario_id`, `session_id`, `year`, `month`, `day`, `hour`.
- `year`, `month`, `day`, `hour` sao derivados do `event_time` via Bloblang no Redpanda Connect.
- Estrutura de path: `bronze/<tabela>/usuario_id=<id>/session_id=<id>/year=<y>/month=<m>/day=<d>/hour=<h>/<arquivo>.parquet`

## 9) Qualidade e observabilidade

Metricas minimas:

- Eventos lidos por tipo (physics/graphics/static)
- Latencia de ingestao e de flush (1s e 5min)
- Taxa de erro de serializacao/publicacao
- Tamanho do backlog no BadgerDB
- Tempo de upload no MinIO
- Records consumidos, erros e offsets por topico (Redpanda Connect metrics)

Logs estruturados:

- `event_id`, `usuario_id`, `session_id`, `topic`, `partition`, `offset`, `window_start`, `window_end`, `object_key`

## 10) Estrutura de repositorio atual

```text
acc-data-platform/
    app/                         # CLI interativo (entry point principal)
        main.go                  # Menu com 6 opcoes
        user.go                  # Operacoes de usuario via modulo user/
        producer.go              # Inicializacao do pipeline
        go.mod                   # Modulo acc-dp/app
    backend/                     # API de autenticacao (Go + Gin)
        cmd/server/main.go
        internal/
            adapter/http/        # Router, handlers, swagger
            service/auth/        # Logica de negocio
            repository/          # Acesso a dados Postgres
            database/            # Conexao, migracoes
                migrations/
            config/              # Configuracao
            domain/              # Modelos, erros
        Dockerfile
    producer/                    # Captura + publicacao (Go)
        cmd/producer/main.go     # Chama pipeline.Run()
        pipeline/
            pipeline.go          # Orquestracao central
        internal/
            source/acc_shm/      # Leitura shared memory
            service/avro/        # Serializacao Avro + Schema Registry
            service/normalizer/  # Normalizacao + session tracking
            broker/redpanda/     # Producer Kafka
            batch/               # Flush de 1s
            buffer/badger/       # Buffer local resiliente
            worker/              # Retry assincrono
            config/              # Env vars
            domain/event/        # Modelos de dominio
            metrics/             # Metricas
    consumer/                    # Redpanda Connect (configuracao YAML)
        configs/
            acc_physics.yaml      # Pipeline: physics -> flatten -> window -> Parquet/S3
            acc_graphics.yaml     # Pipeline: graphics -> flatten -> window -> Parquet/S3
            acc_static.yaml       # Pipeline: static -> flatten -> window -> Parquet/S3
        Dockerfile               # Imagem Redpanda Connect + configs
        go.mod                   # (legado — consumer agora e Redpanda Connect)
    user/                        # Modulo standalone de gestao de usuarios (Go)
        client.go                # Cliente HTTP para backend
        model.go                 # Structs User e Cache
        store.go                 # Cache local em users.json
        store_test.go            # Testes
        watcher.go               # Poller para mudanca de usuario ativo
        go.mod                   # Modulo acc-dp/user
    infra/
        docker-compose.yml       # Orquestracao completa
    data/                        # Dados runtime (gitignore)
        users.json               # Cache local de usuarios
        badger/                  # Buffer BadgerDB do producer
    go.work                      # Workspace Go (5 modulos)
    .env.example                 # Template de variaveis de ambiente
    Makefile                     # Helpers Docker
    ACC_MAP_RAW.md               # Documentacao shared memory ACC
```

## 11) MVP (entrega atual)

1. Ler shared memory dos 3 blocos no producer.
2. Publicar 3 topicos Avro no Redpanda com batch de 1s.
3. Registrar schemas no Schema Registry com compatibilidade backward.
4. Redpanda Connect consome topicos, faz flatten, windowing de 5min e grava Parquet particionado no MinIO.
5. Implementar sistema de usuario integrado com backend auth (menu interativo no app/ + cache local em `users.json` + watcher para troca em runtime).
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

Workspace Go em `go.work` com cinco modulos:

- `app/go.mod` — `acc-dp/app` (CLI interativo)
- `backend/go.mod` — `acc-dp/backend` (API auth)
- `producer/go.mod` — `acc-dp/producer` (captura e publicacao)
- `consumer/go.mod` — `acc-dp/consumer` (legado Go, consumer agora e Redpanda Connect)
- `user/go.mod` — `acc-dp/user` (gestao de usuarios, cache e watcher)

Bibliotecas principais do app:

- `github.com/denisbrodbeck/machineid` (identificador unico da maquina)
- `golang.org/x/term` (leitura de senha sem eco)

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
- `go.uber.org/zap` (logging)
- `github.com/google/uuid` (identificadores)
- `golang.org/x/sys` (suporte baixo nivel para Windows/shared memory)

Bibliotecas principais do consumer:

- Redpanda Connect — binario unico com suporte nativo a Kafka (Avro + Schema Registry), processors Bloblang, windowing e output S3/Parquet.
- Sem dependencias Python — todo o pipeline e declarativo em YAML.

Bibliotecas principais do user:

- (sem dependencias externas alem da stdlib)

### 13.2 Infra local com containers

Arquivo: `infra/docker-compose.yml`

Servicos inclusos (9):

- `redpanda` — Broker Kafka-compatible (v24.3.9) com Schema Registry e PandaProxy
- `redpanda-console` — UI web para administracao do Redpanda
- `minio` — Object storage S3-compatible
- `minio-mc` — MinIO client para criar bucket `datalake-bronze` automaticamente
- `postgres` — Banco de dados para o backend auth (Postgres 16-alpine)
- `backend` — API de autenticacao (Go, depende de postgres)
- `redpanda-init` — Criacao automatica dos topicos `acc.physics.v1`, `acc.graphics.v1`, `acc.static.v1`
- `consumer` — Redpanda Connect com pipelines YAML (depende de redpanda-init e minio-mc)

Volumes:

- `redpanda-data` — Dados do broker
- `minio-data` — Dados do MinIO
- `postgres-data` — Dados do Postgres

### 13.3 Variaveis de ambiente

Arquivo: `.env.example`

Inclui:

- Imagens Docker e nomes de containers
- Portas externas (Redpanda: 19092, Schema Registry: 18081, PandaProxy: 18082, Admin: 19644, Console: 18080, MinIO API: 19000, MinIO Console: 19001, Postgres: 15432, Backend: 18088)
- Nomes dos topicos (`acc.physics.v1`, `acc.graphics.v1`, `acc.static.v1`)
- Credenciais e endpoint do MinIO
- Variaveis de Postgres (host, porta, user, password, db, sslmode)
- Variaveis de autenticacao do backend (JWT secret, TTLs de access token, refresh token e sessao)
- Intervalos de captura do producer (`PRODUCER_PHYSICS_INTERVAL=3ms`, `PRODUCER_GRAPHICS_INTERVAL=3ms`, `PRODUCER_STATIC_INTERVAL=5s`)
- Janela do producer (`PRODUCER_FLUSH_INTERVAL=1s`) e window do Redpanda Connect (`CONNECT_WINDOW=5m`)
- Caminhos locais de usuarios (`USER_STORAGE_PATH`) e BadgerDB (`BADGER_PATH`)
- URL do backend para o app (`BACKEND_URL`)

### 13.4 Comandos de bootstrap

Subir infraestrutura completa:

```powershell
cd C:\Programming\ACC-DP
docker compose --env-file .env -f infra/docker-compose.yml up -d
```

Subir somente backend + postgres:

```powershell
cd C:\Programming\ACC-DP
docker compose --env-file .env -f infra/docker-compose.yml up -d postgres backend
```

Subir infra + Redpanda Connect:

```powershell
cd C:\Programming\ACC-DP
docker compose --env-file .env -f infra/docker-compose.yml up -d
```

Baixar dependencias Go:

```powershell
cd C:\Programming\ACC-DP\backend
go mod tidy
cd ..\producer
go mod tidy
cd ..\app
go mod tidy
cd ..\user
go mod tidy
```

Verificar modulos no workspace:

```powershell
cd C:\Programming\ACC-DP
go work sync
```

Executar o app (menu interativo):

```powershell
cd C:\Programming\ACC-DP
go run ./app
```

Validar backend:

```powershell
curl.exe -sS http://localhost:18088/healthz
start http://localhost:18088/swagger
```
