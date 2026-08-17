FROM golang:1.26.6-alpine

COPY ./src /usr/app

COPY ./entrypoint.sh /usr/app/entrypoint.sh
WORKDIR /usr/app

RUN chmod +x entrypoint.sh

EXPOSE 8080

ENTRYPOINT [ "./entrypoint.sh" ]


