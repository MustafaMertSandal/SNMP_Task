# ---- build stage ----
FROM golang:1.25-alpine AS build
WORKDIR /src

# git gerekebilir (go mod bazı durumlarda kullanır)
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# main package: ./main
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/snmp-task ./main

# ---- runtime stage ----
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/snmp-task /app/snmp-task

# Config'i compose ile mount edeceğiz; burada sadece default path ayarlıyoruz
EXPOSE 8080
ENTRYPOINT ["/app/snmp-task"]
CMD ["-config", "/app/config.yaml"]
