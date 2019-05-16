FROM golang:1.12.5-alpine

RUN apk add git

RUN mkdir /app

WORKDIR /app

COPY . .

CMD ["go", "run", ".", "agent"]