# build stage
FROM harbor.home.starkenberg.net/hub/library/golang:alpine AS build-env

WORKDIR /app

COPY . .

RUN apk update && \
    apk add git && \
    export GIT_COMMIT=$(git rev-list -1 HEAD) && \
    export BUILD_TIME=$(date --utc +%FT%TZ) && \
    GOOS=linux CGO_ENABLED=0 go build -mod vendor -a -installsuffix cgo -ldflags="-X main.GitHash=$GIT_COMMIT -X main.GoVer=$GOLANG_VERSION -X main.BuildTime=$BUILD_TIME -w -s" -o app

#Non Root User Configuration
RUN addgroup -S -g 10001 appGrp \
    && adduser -S -D -u 10000 -s /sbin/nologin -h /app -G appGrp app \
    && chown -R 10000:10001 /app

# package stage
FROM scratch
# Import the user and group files from the builder.
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /etc/ssl/certs /etc/ssl/certs
COPY --from=build-env /app .

#Override as non-root user
USER 10000:10001

EXPOSE 8080

ENTRYPOINT ["/app"]