FROM public.ecr.aws/docker/library/alpine:latest as socicli

ARG SOCIVER=0.4.0
ARG BUILDARCH
RUN wget https://github.com/awslabs/soci-snapshotter/releases/download/v$SOCIVER/soci-snapshotter-v$SOCIVER-linux-$BUILDARCH-static.tar.gz && \
    tar -zxvf soci-snapshotter-v$SOCIVER-linux-$BUILDARCH-static.tar.gz soci

FROM public.ecr.aws/amazonlinux/amazonlinux:2

RUN yum install -y \
    unzip

# Replace with the relevant architecture
ENV ARCH=aarch64
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-$ARCH.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install

COPY --from=socicli /soci /usr/local/bin/soci

COPY --from=public.ecr.aws/docker/library/docker:latest /usr/local/bin/ctr /usr/local/bin/ctr
COPY --from=public.ecr.aws/docker/library/docker:latest /usr/local/bin/containerd /usr/local/bin/containerd
COPY --from=public.ecr.aws/docker/library/docker:latest /usr/local/bin/runc /usr/local/bin/runc

WORKDIR /
COPY script.sh /script.sh
ENTRYPOINT ["./script.sh"]