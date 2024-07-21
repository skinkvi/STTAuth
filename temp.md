version: '3'

services:
  app:
    image: golang:latest
    volumes:
      - ./:/app
    working_dir: /app/cmd/sso
    command: bash -c "./docker-entrypoint.sh go run main.go --config=/app/config/local.yaml"
    ports:
      - "11011:11011"
    environment:
      - POSTGRES_HOST=db
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=1234
      - POSTGRES_DB=STTDB
    depends_on:
      - db
  db:
    image: postgres:13.3
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=1234
      - POSTGRES_DB=STTDB
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./db/migration:/docker-entrypoint-initdb.d

  nginx:
    image: nginx:latest
    container_name: STTNginx
    restart: always
    ports:
      - "80:80"
    volumes:
      - ../../nginx:/etc/nginx/conf.d

volumes:
  postgres_data:
