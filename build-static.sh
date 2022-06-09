#! /bin/bash

IMAGE_NAME=robocar-camera
BINARY_NAME=rc-camera

TAG=$(git describe)
FULL_IMAGE_NAME=docker.io/cyrilix/${IMAGE_NAME}:${TAG}
OPENCV_VERSION=4.5.5
SRC_CMD=./cmd/$BINARY_NAME
GOLANG_VERSION=1.18

image_build(){
  #local platform=$1
  local containerName=builder

  GOPATH=/go

  buildah from --name ${containerName} docker.io/cyrilix/opencv-buildstage-static:${OPENCV_VERSION}
  buildah config --label maintainer="Cyrille Nofficial" "${containerName}"

  buildah copy --from=docker.io/library/golang:${GOLANG_VERSION} "${containerName}" /usr/local/go /usr/local/go
  buildah config --env GOPATH=/go \
                 --env PATH=/usr/local/go/bin:$GOPATH/bin:/usr/local/go/bin:/usr/bin:/bin \
                 "${containerName}"

  buildah run \
    --env GOPATH=${GOPATH} \
    "${containerName}" \
    mkdir -p /src "$GOPATH/src" "$GOPATH/bin"

  buildah run \
    --env GOPATH=${GOPATH} \
    "${containerName}" \
    chmod -R 777 "$GOPATH"


  #buildah config --env PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/local/lib64/pkgconfig "${containerName}"
  buildah config --workingdir /src/ "${containerName}"

  buildah add "${containerName}" . .

  #for platform in "linux/amd64" "linux/arm64" "linux/arm/v7"
  for platform in "linux/arm64" "linux/arm/v7"
  do

    GOOS=$(echo "$platform" | cut -f1 -d/) && \
    GOARCH=$(echo "$platform" | cut -f2 -d/) && \
    GOARM=$(echo "$platform" | cut -f3 -d/ | sed "s/v//" )

    case $GOARCH in
      "amd64")
        ARCH=amd64
        ARCH_LIB_DIR=/usr/lib/x86_64-linux-gnu
        EXTRA_LIBS="-lopencv_alphamat"
        CC=gcc
        CXX=g++
      ;;
      "arm64")
        ARCH=arm64
        ARCH_LIB_DIR=/usr/lib/aarch64-linux-gnu
        EXTRA_LIBS="-ltbb -ltegra_hal -lavcodec -lavformat -lavutil -lswscale"
        CC=aarch64-linux-gnu-gcc
        CXX=aarch64-linux-gnu-g++
      ;;
      "arm")
        ARCH=armhf
        ARCH_LIB_DIR=/usr/lib/arm-linux-gnueabihf
        EXTRA_LIBS="-ltbb -ltegra_hal -lavcodec -lavformat -lavutil -lswscale"
        CC=arm-linux-gnueabihf-gcc
        CXX=arm-linux-gnueabihf-g++
      ;;
    esac


    ### TO remove
      buildah run "$containerName" dpkg --add-architecture ${ARCH}
  buildah run "$containerName" apt-get install -y \
      libavcodec-dev:${ARCH} \
      libavdevice-dev:${ARCH}

#            libavcodec-dev:${ARCH} libavformat-dev:${ARCH} libswscale-dev:${ARCH} libv4l-dev:${ARCH} \
#            libxvidcore-dev:${ARCH} libx264-dev:${ARCH} \
#            crossbuild-essential-${ARCH} \
#            libjpeg62-turbo:${ARCH} \
#            libpng16-16:${ARCH} \
#            libwebp6:${ARCH} \
#            libwebp-dev:${ARCH} \
#            libtiff5:${ARCH} \
#            libavc1394-0:${ARCH} \
#            libavc1394-dev:${ARCH} \
#            libopenblas0:${ARCH} \
#            libopenblas-dev:${ARCH} \
#            liblapack-dev:${ARCH} \
#            liblapack3:${ARCH} \
#            libatlas3-base:${ARCH} \
#            libatlas-base-dev:${ARCH} \
#            libgphoto2-6:${ARCH} \
#            libgphoto2-dev:${ARCH} \
#            libgstreamer1.0-0:${ARCH} \
#            libgstreamer1.0-dev:${ARCH} \
#            libopenjp2-7:${ARCH} \
#            libopenjp2-7-dev:${ARCH} \
#            opencl-dev:${ARCH} \
#            libglib2.0-0:${ARCH} \
#            libglib2.0-dev:${ARCH} \
#            libtiff-dev:${ARCH} zlib1g-dev:${ARCH} \
#            libjpeg-dev:${ARCH} libpng-dev:${ARCH} \
#            libavcodec-dev:${ARCH} libavformat-dev:${ARCH} libswscale-dev:${ARCH} libv4l-dev:${ARCH} \
#            libxvidcore-dev:${ARCH} libx264-dev:${ARCH} \
#
#    #### End



    # shellcheck disable=SC2027
    CGO_LDFLAGS="-static -L/opt/opencv/${ARCH}/lib/opencv4/3rdparty -L/opt/opencv/${ARCH}/lib -L${ARCH_LIB_DIR} ${EXTRA_LIBS} -lopencv_gapi -lopencv_stitching -lopencv_aruco -lopencv_barcode -lopencv_bgsegm -lopencv_bioinspired -lopencv_ccalib -lopencv_dnn_objdetect -lopencv_dnn_superres -lopencv_dpm -lopencv_face -lopencv_fuzzy -lopencv_hfs -lopencv_img_hash -lopencv_intensity_transform -lopencv_line_descriptor -lopencv_mcc -lopencv_quality -lopencv_rapid -lopencv_reg -lopencv_rgbd -lopencv_saliency -lopencv_stereo -lopencv_structured_light -lopencv_phase_unwrapping -lopencv_superres -lopencv_optflow -lopencv_surface_matching -lopencv_tracking -lopencv_highgui -lopencv_datasets -lopencv_text -lopencv_plot -lopencv_videostab -lopencv_videoio -lopencv_xfeatures2d -lopencv_shape -lopencv_ml -lopencv_ximgproc -lopencv_video -lopencv_dnn -lopencv_xobjdetect -lopencv_objdetect -lopencv_calib3d -lopencv_imgcodecs -lopencv_features2d -lopencv_flann -lopencv_xphoto -lopencv_photo -lopencv_imgproc -lopencv_core -lquirc -llibprotobuf -lade -lIlmImf -littnotify -llibjpeg-turbo -llibopenjp2 -llibpng -llibtiff -llibwebp -lquirc -lzlib -ldl -lm -lpthread -lrt"

    printf "\nBuild binary for %s\n" "${platform}"
    printf "\tos:%s arch:%s variant:%s cc:%s cxx:%s\n" "$GOOS" "$GOARCH" "$GOARM" $CC $CXX
    printf "\tLDFLAGS:%s\n" "$CGO_LDFLAGS"
    buildah run \
      --env CGO_ENABLED=1 \
      --env CC=${CC} \
      --env CXX=${CXX} \
      --env GOOS=${GOOS} \
      --env GOARCH=${GOARCH} \
      --env GOARM=${GOARM} \
      --env CGO_CPPFLAGS="-I/opt/opencv/${ARCH}/include/opencv4/ -I/usr/lib/aarch64-linux-gnu/include" \
      --env CGO_LDFLAGS="${CGO_LDFLAGS}" \
      --env CGO_CXXFLAGS="--std=c++1z" \
      "${containerName}" \
      go build  -tags netgo,customenv -a -o ${BINARY_NAME}.${ARCH} ${SRC_CMD}
      #-lade -littnotify -llibpeg-turbo -llibopenjp2 -llibpng -llibprotobuf -llibtiff -llibtiff -llibwebp -llibquirc -ltbb -ltegra_hal -lzlib" \
      #-L/usr/lib/x86_64-linux-gnu
      # -L/usr/lib/gcc/x86_64-linux-gnu/10/
  done
  buildah commit --rm ${containerName} ${IMAGE_NAME}-builder
}

image_final(){
  local containerName=runtime

  for platform in "linux/amd64" "linux/arm64" "linux/arm/v7"
  do

    GOOS=$(echo $platform | cut -f1 -d/) && \
    GOARCH=$(echo $platform | cut -f2 -d/) && \
    GOARM=$(echo $platform | cut -f3 -d/ | sed "s/v//" )
    VARIANT="--variant $(echo $platform | cut -f3 -d/  )"

    if [[ -z "$GOARM" ]] ;
    then
      VARIANT=""
    fi

    if [[ "${GOARCH}" == "arm" ]]
    then
      BINARY="${BINARY_NAME}.armhf"
    else
      BINARY="${BINARY_NAME}.${GOARCH}"
    fi

    buildah from --name "${containerName}" --os "${GOOS}" --arch "${GOARCH}" ${VARIANT} docker.io/library/debian:stable-slim
    buildah copy --from ${IMAGE_NAME}-builder  "$containerName" "/src/${BINARY}" /usr/local/bin/${BINARY_NAME}

    buildah config --label maintainer="Cyrille Nofficial" "${containerName}"
    buildah config --user 1234 "$containerName"
    buildah config --cmd '' "$containerName"
    buildah config --entrypoint '[ "/usr/local/bin/'${BINARY_NAME}'" ]' "$containerName"

    buildah commit --rm --manifest ${IMAGE_NAME} ${containerName}
  done
}
#buildah rmi localhost/$IMAGE_NAME
#buildah manifest rm localhost/${IMAGE_NAME}

#image_build linux/amd64
#image_build linux/arm64
#image_build linux/arm/v7

image_build

# TODO: enable neon

# push image
#printf "\n\nPush manifest to %s\n\n" ${FULL_IMAGE_NAME}
image_final
#buildah manifest push --rm -f v2s2 "localhost/$IMAGE_NAME" "docker://$FULL_IMAGE_NAME" --all
