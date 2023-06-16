FROM golang:1.20-alpine as build
RUN apk add -u git
WORKDIR /app
COPY . .
RUN go build -o /crd-to-yaml

FROM alpine
RUN apk add -u ca-certificates
COPY --from=build /crd-to-yaml /app/

EXPOSE 9998

WORKDIR /app/
ENTRYPOINT [ "/app/crd-to-yaml" ]
