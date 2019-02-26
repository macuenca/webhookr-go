FROM alpine:3.9
RUN apk update && apk add ca-certificates
COPY home.html /home.html
COPY alert.ogg /alert.ogg
ADD webhookr /bin
CMD /bin/webhookr
