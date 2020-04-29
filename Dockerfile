############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git && apk add --no-cach make

WORKDIR /go/src/wm-go

COPY . .

RUN make install && make build

############################
# STEP 2 build an image
############################
FROM scratch

COPY --from=builder /go/src/wm-go/bin /

ENTRYPOINT ["/wm"]

EXPOSE 4200
