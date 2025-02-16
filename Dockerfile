FROM golang:1.23.6 AS build
COPY . ./
RUN go install github.com/a-h/templ/cmd/templ@v0.3.819
RUN templ generate
RUN CGO_ENABLED=1 go build -o "/bin/openuem-console" .

FROM debian:latest
COPY --from=build /bin/openuem-console /bin/openuem-console
COPY ./assets /bin/assets
RUN apt-get update
RUN apt install -y ca-certificates
EXPOSE 1323
EXPOSE 1324
WORKDIR /bin
ENTRYPOINT ["/bin/openuem-console"]