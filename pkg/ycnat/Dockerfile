FROM alpine:3.12 as build

WORKDIR /src
RUN mkdir -p /out/usr/bin
RUN apk add upx go
ADD main.go .
ADD go.mod .
ADD go.sum .
ADD docker-entrypoint.sh /out/usr/bin/

RUN mkdir -p /out/etc/apk && cp -r /etc/apk/* /out/etc/apk/
RUN apk add --no-cache --initdb -p /out \
    alpine-baselayout
RUN rm -rf /out/etc/apk /out/lib/apk /out/var/cache

RUN go build -ldflags "-s -w" -o /out/usr/bin/усnat
RUN upx /out/usr/bin/усnat

FROM scratch
COPY --from=build /out /
ENV CFG=/run/config/rtid
ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]
