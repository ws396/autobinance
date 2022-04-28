FROM golang:1.17.5 as base

FROM base as dev

ENV GO111MODULE=on

RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

WORKDIR /opt/app/api

RUN go mod init github.com/ws396/autobinance
RUN air init

CMD ["air"]