docker compose -f deployments/compose/docker-compose.yml up --build -d

go build -o bin/api ./apps/api


$env:GOOS='linux'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'; go build -o bin/api ./apps/api


docker compose -f deployments/compose/docker-compose.yml up --build -d