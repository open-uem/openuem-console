FROM golang:latest AS build
COPY . ./
RUN go build -o "/bin/openuem-console" .

FROM debian:latest
COPY --from=build /bin/openuem-console /bin/openuem-console
COPY ./assets /tmp/assets
EXPOSE 1323
EXPOSE 1324
WORKDIR /tmp
ENTRYPOINT ["/bin/openuem-console"]