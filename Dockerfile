ARG OPENCV_VERSION=v4.5.1

FROM golang:alpine as gobuilder


FROM cyrilix/opencv-buildstage:${OPENCV_VERSION} as builder

LABEL maintainer="Cyrille Nofficial"

COPY --from=gobuilder /usr/local/go /usr/local/go
ENV GOPATH /go
ENV PATH /usr/local/go/bin:$GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "/src $GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

ENV PKG_CONFIG_PATH /usr/local/lib/pkgconfig:/usr/local/lib64/pkgconfig
ENV CGO_CPPFLAGS -I/usr/local/include
ENV CGO_CXXFLAGS "--std=c++1z"

WORKDIR /src
ADD . .

RUN CGO_LDFLAGS="$(pkg-config --libs opencv4)" \
    CGO_ENABLED=1 CGO_CPPFLAGS=${CGO_CPPFLAGS} CGO_CXXFLAGS=${CGO_CXXFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build -a ./cmd/rc-camera/




FROM cyrilix/opencv-runtime:${OPENCV_VERSION}

ENV LD_LIBRARY_PATH /usr/local/lib:/usr/local/lib64

USER 1234
COPY --from=builder /src/rc-camera /go/bin/rc-camera
ENTRYPOINT ["/go/bin/rc-camera"]
