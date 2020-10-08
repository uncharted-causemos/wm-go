#
# Use docker multistage builds by having intermediate builder image to keep final image size small
# Check https://docs.docker.com/develop/develop-images/multistage-build/ for more details
#
############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

ARG GITLAB_LOGIN
ARG GITLAB_TOKEN

RUN apk update && apk add --no-cache git && apk add --no-cach make && apk --no-cache add ca-certificates
# Gitlab reads following login information from ~/.netrc file
RUN echo "machine gitlab.uncharted.software login ${GITLAB_LOGIN} password ${GITLAB_TOKEN}" > ~/.netrc
RUN cat ~/.netrc

WORKDIR /go/src/wm-go

COPY . .

RUN make install && make build

############################
# STEP 2 build an image
############################
FROM scratch

# Copy certificate from the builder image. It is required to make https requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/wm-go/bin /

ENTRYPOINT ["/wm"]

EXPOSE 4200
