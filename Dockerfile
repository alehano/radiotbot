FROM golang:alpine

ADD . /go/src/github.com/alehano/radiotbot
RUN apk --update add ca-certificates
RUN \
 cd /go/src/github.com/alehano/radiotbot && \
 go build -o /srv/search-bot && \
 mkdir /srv/data && \
 rm -rf /go/src/*

EXPOSE 8080
WORKDIR /srv
CMD ["/srv/search-bot"]
