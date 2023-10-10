FROM alpine:latest AS build-stage

RUN apk add ffmpeg

RUN ffmpeg -f lavfi -i testsrc=d=600  -an -c:v libx264 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 /output.h264
RUN ffmpeg -f lavfi -i sine=d=600  -vn -c:a libopus -page_duration 20000 /output.ogg

FROM scratch AS export-stage
COPY --from=build-stage /output.h264 /output.h264
COPY --from=build-stage /output.ogg /output.ogg