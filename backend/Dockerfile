FROM golang:1.24-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod ./
COPY . ./

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/backend ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/backend /usr/local/bin/backend

EXPOSE 8088

ENTRYPOINT ["backend"]