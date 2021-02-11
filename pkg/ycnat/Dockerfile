FROM alpine:3.12 as build

WORKDIR /src
RUN mkdir -p /out/
RUN apk add upx go
ADD main.go .
ADD go.mod .
ADD go.sum .
ADD docker-entrypoint.sh /out/usr/bin/

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/ycnat
RUN upx /out/ycnat

FROM scratch
COPY --from=build /out/ycnat /bin/ycnat
ENTRYPOINT ["/bin/ycnat"]
CMD [ "-rt", "/run/config/rt", "-ip", "/run/config/ip" ]
