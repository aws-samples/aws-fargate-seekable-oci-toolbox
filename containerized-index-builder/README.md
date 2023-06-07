# Containerized SOCI Index Builder

Before a container image can be lazily loaded by the SOCI Snapshotter it needs
to be indexed. To do so, you can use the `soci`
[cli](https://github.com/awslabs/soci-snapshotter/tree/main/cmd/soci), and run
the `soci create` command against a local container image.

However, the local container image needs to exist in a containerd image store.
If you are running the Docker Engine (including Docker Desktop) the container
images downloaded with `docker pull` are not being stored in the containerd
image store, they are being stored in the Docker Engine image store, so they can
not be accessed by the `soci` cli.

This sample tool shows how you can still use the `soci` cli with the Docker
Engine by containerizing all of the prerequisites. This container runs
containerd within a container, a bit like Docker in Docker, but it is containerd
in Docker. The entrypoint of this container is a simple bash script that can be
customized. The current version does the following:

1. Start the containerd daemon
2. Download the container image defined in the `<command>` part of the `docker
   run` from a remote container registry.
3. Generate a SOCI Index
4. Push the SOCI Index back up to the remote container registry.

## Using this container image

First build the container image

```bash
docker build \
    --tag sociindexbuilder:latest \
    .
```

And then run the container image, passing in the workload image as the
command (hello-world in this example).

> The example below is fetching and generating an index for a container image
> stored in Amazon ECS (Hence the AWS_REGION environment variable and the
> `--volume` mount of the local AWS credentials).

```bash
docker run \
	--rm \
	--privileged \
	--env AWS_REGION=us-east-1 \
	--mount type=tmpfs,destination=/var/lib/containerd \
	--mount type=tmpfs,destination=/var/lib/soci-snapshotter-grpc \
	--volume ${HOME}/.aws:/root/.aws \
	sociindexbuilder:latest \
	111222333444.dkr.ecr.us-east-1.amazonaws.com/myworkload:latest
```

### Minimum Layer Size

There is an initial overhead when lazy loading a container image layer as the
SOCI artifacts need to be downloaded and the FUSE file system needs to be
configured. Therefore for small container image layer it may actually be quicker
to just do a pull the layer whole, rather then lazily loading it.

By default when running `soci create` if the container image layer is less then
10MB we will not create an index for it, therefore it will not be lazy loaded.
This value is adjustable, so you can pass in `--env MIN_LAYER_SIZE` to the
container image and set a new MB limit for the `soci create` command. So to
index all container image layers 5 MB or pass in `--env MIN_LAYER_SIZE=5`.

### Architecture

This index builder tool is not multi architecture aware. It will only create a
SOCI Index for a single architecture of a container image at a time. By default
the index builder tool expects to pull and create an Index for an x86 container
image. That value can be overridden with the `--env IMAGE_ARCH` variable. For
example to index an arm64 container image pass in `--env
IMAGE_ARCH=linux/arm64`.