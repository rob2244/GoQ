FROM golang:1.14-alpine3.12 AS Builder

WORKDIR /build
COPY . .

RUN mkdir dist; \
    cd ./cmd/queuemanager; \
    go build; \
    cp queuemanager /build/dist


FROM alpine:3.12

WORKDIR /opt/GoQ/
COPY --from=Builder /build/dist/queuemanager .
RUN ls
ENTRYPOINT [ "/opt/GoQ/queuemanager",  "-tls=false" ]
