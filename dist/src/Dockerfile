FROM golang:1.12.0-alpine

RUN mkdir /casino
COPY ./www /casino
WORKDIR /casino

EXPOSE 80
EXPOSE 5001
EXPOSE 9000

RUN apk update && apk add gcc && apk add libc-dev && apk add apache2
RUN apk add git && apk add openrc
COPY ./frontend /var/www/localhost/htdocs

RUN export GOPROXY=https://goproxy.io && export GO111MODULE=on && go build -o main
RUN chmod +x main
RUN chmod +x waf

ENTRYPOINT ["sh", "/casino/run.sh"]
