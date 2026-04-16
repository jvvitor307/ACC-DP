# ACC-DP Backend Auth

Backend de autenticacao com Gin + Postgres.

## Recursos

- Cadastro de usuario
- Login com JWT de acesso e refresh token rotativo
- Logout
- Sessao ativa unica por usuario
- Maquina ativa unica por vez
- Suporte a machine_id vindo do cliente

## Estrutura

- cmd/server: bootstrap HTTP
- internal/adapter/http: handlers e rotas
- internal/service/auth: regras de negocio de autenticacao
- internal/repository: acesso ao Postgres
- internal/database/migrations: schema SQL

## Tabelas

- users
- machines
- sessions
- refresh_tokens

## Regra de sessao

- Um usuario so pode ter uma sessao ativa por vez.
- Uma maquina so pode ter um usuario ativo por vez.
- Novo login revoga sessao ativa anterior do usuario.

## Variaveis de ambiente

Use os valores adicionados em .env.example:

- BACKEND_HTTP_ADDR
- BACKEND_JWT_SECRET
- BACKEND_ACCESS_TOKEN_TTL
- BACKEND_REFRESH_TOKEN_TTL
- BACKEND_SESSION_TTL
- POSTGRES_HOST
- POSTGRES_PORT
- POSTGRES_USER
- POSTGRES_PASSWORD
- POSTGRES_DB
- POSTGRES_SSLMODE

## Subir Postgres

Use o compose existente do projeto:

- docker compose -f infra/docker-compose.yml up -d postgres

## Rodar backend

No modulo backend:

- go run ./cmd/server

Ou via Docker Compose (recomendado para rodar junto com Postgres):

- docker compose --env-file .env.example -f infra/docker-compose.yml up -d postgres backend

API disponivel em:

- http://localhost:18088/healthz

Swagger disponivel em:

- http://localhost:18088/swagger
- http://localhost:18088/swagger/openapi.yaml

## Machine ID

O `machine_id` deve ser gerado pela aplicacao que roda na maquina do usuario e enviado nas rotas de auth.
O backend apenas valida e vincula esse identificador a sessoes e refresh tokens.

## Endpoints

- POST /auth/register
- POST /auth/login
- POST /auth/refresh
- POST /auth/logout
- GET /healthz

### Payload register

{
  "email": "user@example.com",
  "display_name": "User Name",
  "password": "strong-pass"
}

### Payload login

{
  "email": "user@example.com",
  "password": "strong-pass",
  "machine_id": "machine-unique-id",
  "device_name": "my-pc"
}

### Payload refresh

{
  "refresh_token": "token",
  "machine_id": "machine-unique-id"
}

### Payload logout

{
  "refresh_token": "token",
  "machine_id": "machine-unique-id"
}
