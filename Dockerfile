FROM golang:1.12

COPY . /potato

WORKDIR /potato

RUN go build -mod vendor -o /bin/potato .
RUN mkdir /storage

ENTRYPOINT [ "potato" ]