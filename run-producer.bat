@echo off
setlocal EnableExtensions EnableDelayedExpansion

rem Executa sempre a partir da pasta onde o .bat esta
cd /d "%~dp0"

if not exist ".env" (
  echo ERRO: arquivo .env nao encontrado na raiz do projeto.
  exit /b 1
)

rem Carrega variaveis do .env (ignora linhas em branco e comentarios)
for /f "usebackq tokens=1,* delims==" %%A in (".env") do (
  set "k=%%A"
  if not "!k!"=="" if not "!k:~0,1!"=="#" set "%%A=%%B"
)

echo Subindo docker compose...
docker compose --env-file .env -f infra/docker-compose.yml up -d
if errorlevel 1 (
  echo ERRO ao subir container.
  exit /b 1
)

echo Iniciando producer...
go run ./producer/cmd/producer

endlocal
