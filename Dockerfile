FROM alpine:edge

RUN apk add tzdata

WORKDIR /app

ADD auditstream /app/


CMD [ "./auditstream" ]
