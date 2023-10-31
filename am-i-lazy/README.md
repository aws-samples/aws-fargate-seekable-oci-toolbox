# Am I Lazy?

This directory contains an Init Container that runs on AWS Fargate which tells
you if the container has been lazy loaded by reading the ECS Task Metadata
endpoint.

The quick sample application will provide a log output look:

```json
{
    "Cluster": "arn:aws:ecs:us-east-1:11112222333444:cluster/default",
    "TaskARN": "arn:aws:ecs:us-east-1:11112222333444:task/default/59cc60c6e63f487fae626ccb83501f2b",
    "Family": "amilazy",
    "Revision": "1",
    "TaskCpu": 256,
    "TaskMemory": 512,
    "ImagePullTime": 1.546018865,
    "Snapshotter": "soci"
}
```

The assumption is that this `amilazy` init container is added to your Tasks as a
non essential container. It will run, query the Task Metadata Endpoint, and by
emitting a entry in the stout log, send the data to Cloudwatch Logs.

## Building the Am I Lazy Container

Before we can run this container within our Tasks, we need to build the
container and push it to ECR.

```bash
AWS_ACCOUNT_ID=11112222333444
AWS_REGION=us-east-1

docker buildx \
    build \
    --platform linux/arm64,linux/amd64 \
    --push \
    --tag $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/amilazy:v0.1 \
    .
```

We also need to generate a SOCI Index for this container. When using SOCI on AWS
Fargate, all containers in the Task need to be indexed. For this step we are
going to use the [containerized-index-builder](../containerized-index-builder/)
tool also container in this repo.

```bash
# Arm64 Image
docker run \
	--rm \
	--privileged \
	--env AWS_REGION=us-east-1 \
	--mount type=tmpfs,destination=/var/lib/containerd \
	--mount type=tmpfs,destination=/var/lib/soci-snapshotter-grpc \
    --env MIN_LAYER_SIZE=5 \
    --env IMAGE_ARCH=linux/arm64 \
	--volume ${HOME}/.aws:/root/.aws \
	sociindexbuilder:latest \
	$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/amilazy:v0.1

# Amd64 Image
docker run \
	--rm \
	--privileged \
	--env AWS_REGION=us-east-1 \
	--mount type=tmpfs,destination=/var/lib/containerd \
	--mount type=tmpfs,destination=/var/lib/soci-snapshotter-grpc \
    --env MIN_LAYER_SIZE=5 \
    --env IMAGE_ARCH=linux/amd64 \
	--volume ${HOME}/.aws:/root/.aws \
	sociindexbuilder:latest \
	$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/amilazy:v0.1
```

## Running the Am I Lazy Container

After the Image has been built, we then need to add the amilazy container into
the existing Task definitions. The json snippet for a non essential container
would look like:

```bash
{
    "name": "amilazy",
    "image": "11112222333444.dkr.ecr.us-east-1.amazonaws.com/amilazy:v0.1",
    "essential": false,
    "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
            "awslogs-group": "/aws/ecs/service/amilazy",
            "awslogs-region": "us-east-1",
            "awslogs-stream-prefix": "amilazy",
            "awslogs-create-group": "true"
        }
    }
}
```

## Log Insights Query

All of your amilazy containers should log to a centralized Log Group, once that
is done you can now query all the metrics to get a high level view of what
snapshotters are being used within the environment with Log Insights Query.

```
fields @timestamp, @message
| sort @timestamp desc
| stats count(*) by Cluster,Family,Revision,Snapshotter
| limit 20
```

This query should produce a table like:

| Cluster                                              | Family  | Revision | Snapshotter | count(*) |
| ---------------------------------------------------- | ------- | -------- | ----------- | -------- |
| arn:aws:ecs:us-east-1:11112222333444:cluster/default | amilazy | 1        | soci        | 10       |
