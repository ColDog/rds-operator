FROM alpine:3.6

RUN adduser -D rds-operator
USER rds-operator

ADD tmp/_output/bin/rds-operator /usr/local/bin/rds-operator
