FROM ubuntu:20.04

COPY ./bin/gouploadserver-linux-amd64 /app/gouploadserver

EXPOSE 8080

VOLUME [ "/app" ]

ENTRYPOINT [ "/app/gouploadserver" ]

WORKDIR /app

CMD [ "-p", "8080", "." ]