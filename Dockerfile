FROM alpine:3.3
MAINTAINER Alehano <halyapin@gmail.com>
RUN apk --update add ca-certificates
RUN mkdir /bot
RUN mkdir /bot/data
COPY ./radiotbot /bot/
WORKDIR /bot
CMD /bot/radiotbot