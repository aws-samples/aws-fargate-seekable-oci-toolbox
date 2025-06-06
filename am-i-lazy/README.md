# Am I Lazy?

This directory contains an Init Container that runs on AWS Fargate which tells
you if a container has been lazy loaded by reading the ECS Task Metadata
endpoint.

The quick sample application will provide a log output look:

```json
{
    "Cluster": "arn:aws:ecs:us-east-1:111222333444:cluster/default",
    "TaskARN": "arn:aws:ecs:us-east-1:111222333444:task/default/cdf1415e0a764dfa9f4f3270f413ea58",
    "Family": "nginxdemo",
    "Revision": "1",
    "TaskCpu": 512,
    "TaskMemory": 1024,
    "ImagePullTime": 2.380300311,
    "Containers": [
        {
            "Name": "amilazy",
            "Image": "111222333444.dkr.ecr.us-east-1.amazonaws.com/amilazy:v0.2",
            "ImageId": "sha256:e8377cb653de43ab399046b7e9be38890c288269b450992bb52deefc327d6a16",
            "Snapshotter": "overlayfs"
        },
        {
            "Name": "nginx",
            "Image": "111222333444.dkr.ecr.us-east-1.amazonaws.com/nginx-demo:latest",
            "ImageId": "sha256:e8377cb653de43ab399046b7e9be38890c288269b450992bb52deefc327d6a16",
            "Snapshotter": "soci"
        }
    ]
}
```

The assumption is that this `amilazy` init container is added to your Tasks as a
non essential container. It will run, query the Task Metadata Endpoint, and by
emitting a entry to stout, send the data to Cloudwatch Logs.

## Building the Am I Lazy Container

Before we can run this container within our Tasks, we need to build the
container and push it to ECR.

```bash
AWS_ACCOUNT_ID=111222333444
AWS_REGION=us-east-1

finch build \
    --file=Dockerfile \
    --platform=linux/arm64,linux/amd64 \
    --output=type=image,oci-mediatypes=true,compression=zstd,force-compression=true,name=${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/amilazy:v0.3,push=true \
    .
```

## Running the Am I Lazy Container

After the Image has been built, we then need to add the amilazy container into
the existing Task definitions. The json snippet for a non essential container
would look like:

```json
{
    "family": "nginxdemo",
    "executionRoleArn": "arn:aws:iam::111222333444:role/ecsTaskExecutionRole",
    "networkMode": "awsvpc",
    "containerDefinitions": [
        {
            "name": "amilazy",
            "image": "111222333444.dkr.ecr.us-east-1.amazonaws.com/amilazy:v0.2",
            "essential": false,
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "/aws/ecs/amilazy",
                    "awslogs-region": "us-east-1",
                    "awslogs-stream-prefix": "amilazy"
                }
            }
        },
        {
            "name": "nginx",
            "image": "111222333444.dkr.ecr.us-east-1.amazonaws.com/nginx-demo:latest",
            "essential": true,
            "linuxParameters": {
                "initProcessEnabled": true
            },
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "/aws/ecs/service/nginx-demo",
                    "awslogs-region": "us-east-1",
                    "awslogs-stream-prefix": "nginx"
                }
            }
        }
    ],
    "cpu": "512",
    "memory": "1024"
}
```

## Log Insights Query

All of your amilazy containers should log to a centralized Log Group, once that
is done you can now query all the metrics to get a high level view of what
snapshotters are being used within the environment with [CloudWatch Log
Insights](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/AnalyzingLogData.html).

A sample query would be:

```
filter @logStream like /^amilazy/
| fields
    Family,
    Revision,
    Containers.0.Name as ContainerOneName,
    Containers.0.Snapshotter as ContainerOneSnapshotter,
    Containers.1.Name as ContainerTwoName,
    Containers.1.Snapshotter as ContainerTwoSnapshotter
| stats count(*) as Count by Family, Revision, ContainerOneName, ContainerOneSnapshotter, ContainerTwoName, ContainerTwoSnapshotter
| sort Count desc
| limit 20
```

This should produce a table like:

| Family | Revision | ContainerOneName | ContainerOneSnapshotter | ContainerTwoName | ContainerTwoSnapshotter | Count |
| --- | --- | --- | --- | --- | --- | --- |
| nginx | 2 | nginxdemo | soci | amilazy | overlayfs | 8 |
| nginx | 1 | nginxdemo | soci | amilazy | overlayfs | 8 |
| nginx | 3 | nginxdemo | soci | amilazy | overlayfs | 8 |
| nginx | 4 | nginxdemo | soci | amilazy | overlayfs | 5 |
