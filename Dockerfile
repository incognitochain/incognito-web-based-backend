FROM golang:1.20-alpine3.16 AS build

RUN apk update && apk add gcc musl-dev gcompat libc-dev linux-headers
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -tags=jsoniter -o incognito-web-based-backend

FROM alpine:3.16
EXPOSE 8080

WORKDIR /app
COPY devenv/service/run.sh /app/run.sh
COPY --from=build /app/incognito-web-based-backend /app/webservice

RUN chmod +x /app/run.sh /app/webservice
ENTRYPOINT [ "/app/run.sh" ]
