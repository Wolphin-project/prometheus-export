FROM golang:1.9.1-alpine as buildbase

RUN apk add --no-cache git gcc

ARG SRCPATH=/go/src/git.rnd.alterway.fr/Wolphin-project/partners/prometheus-export

WORKDIR ${SRCPATH}
COPY . ${SRCPATH}
RUN CGO_ENABLED=0 GOOS=linux  go build -a -installsuffix cgo -o /prometheus-exporter .

# --------------------

FROM alpine:latest as runbase

RUN apk --no-cache add ca-certificates
COPY --from=buildbase ${SRCPATH}/prometheus-exporter /usr/bin/
ENTRYPOINT ["/usr/bin/prometheus-exporter"]
