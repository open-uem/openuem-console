FROM golang:latest AS build
COPY . ./
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate
RUN go build -o "/bin/openuem-console" .

FROM debian:latest
COPY --from=build /bin/openuem-console /bin/openuem-console
COPY ./assets /bin/assets
RUN apt-get update
RUN apt install -y ca-certificates
EXPOSE 1323
EXPOSE 1324
WORKDIR /bin
ENTRYPOINT ["/bin/openuem-console"]