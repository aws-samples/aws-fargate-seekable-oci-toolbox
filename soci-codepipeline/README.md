# Generating SOCI Indexes within a CI/CD Pipeline

Before a container image can be lazy loaded it needs to be indexed. To do so,
you can use the `soci`
[cli](https://github.com/awslabs/soci-snapshotter/tree/main/cmd/soci), and run
the `soci create` command against a local container image.

However, most container images are built as part of a CI/CD pipeline.
Therefore this repository attempts to provide a blueprint of how to build a SOCI
Index as part of a pipeline. The example in this repository is based on [AWS
CodePipeline](https://aws.amazon.com/codepipeline/) and [AWS
CodeBuild](https://aws.amazon.com/codebuild/), but the hope is that this
blueprint and the stages of the pipeline can be transferred to other CI/CD
Pipelines.

## Deploying the Sample Pipeline.

You can deploy the AWS CodePipeline in your AWS Account with the following steps:

1. Deploy the CloudFormation Template.

    ```bash
    aws cloudformation \
        create-stack \
        --stack-name soci-pipeline \
        --template-body file://cloudformation.yaml \
        --capabilities CAPABILITY_IAM
    ```

    This cloudformation template will create the following resources:

    * AWS CodeCommit repository to store the sample application.
    * Amazon ECR repository to store the Container Image and SOCI
      Index.
    * Amazon EventBridge rule to trigger the pipeline on a commit
      being pushed to AWS CodeCommit.
    * AWS CodePipeline with the AWS CodeBuild discussed above.

    Log into the AWS Console, to verify the cloudformation template has been
    deployed correctly, and the various resources in the AWS Code Suite has been
    deployed successfully.

2. Next we push the Sample Application into the CodeCommit Repository.

   If you have not configured you're local git client to authenticate and push to
   AWS CodeCommit, see the [AWS CodeCommit
   documentation](https://docs.aws.amazon.com/codecommit/latest/userguide/setting-up-ssh-unixes.html).

    ```bash
    git clone https://github.com/ollypom/mysfits.git
    git remote add codecommit ssh://git-codecommit.eu-west-1.amazonaws.com/v1/repos/socidemoapp
    git push codecommit master
    ```

    You can monitor the status of the CodePipeline in the AWS Console. Ensuring
    it has been triggered and the 2 CodeBuilds stages execute successfully.

3. In the AWS Console or using the AWS CLI, check to see that the container
   image along with a SOCI Index have been pushed to Amazon ECR.

   ```bash
   aws ecr list-images --repository socidemoapp
   {
       "imageIds": [
           {
               "imageDigest": "sha256:91bf19a70b1806ae97476fd19a17eea7977c068a0c1e361037cba457c2810d4a"
           },
           {
               "imageDigest": "sha256:a4239937f44e35d790dd8396fc3a215f59d758541d8370459eefc6a286650474",
               "imageTag": "1d2ca213-2121-447f-9a00-9633bb8941be"
           },
           {
               "imageDigest": "sha256:18d7b8d6b11eeae4e4265380fedf51720f3f47fb209cbeefdc395f6d18908693",
               "imageTag": "sha256-a4239937f44e35d790dd8396fc3a215f59d758541d8370459eefc6a286650474"
           }
       ]
   }
   ```

### Walkthrough of the buildspec file

The [buildspec file](https://docs.aws.amazon.com/codebuild/latest/userguide/build-spec-ref.html) file for CodeBuild can be be found in the cloudformation template,
however in this `README.md` we have added some commentary.

**Pre Build**

In the Pre Build stage we download the latest version of the `soci` cli and
ensure it is in the right place in the users path. We also log in to Amazon ECR
and export the token to a variable for future use, this is because containerd's
`ctr` does not respect a Docker credential file (`~/.docker/config.json`) and
instead needs credentials passed in with `--user` flag.

```bash
- echo Download the SOCI Binaries
- wget --quiet https://github.com/awslabs/soci-snapshotter/releases/download/v${SociVersion}/soci-snapshotter-${SociVersion}-linux-amd64.tar.gz
- tar xvzf soci-snapshotter-${SociVersion}-linux-amd64.tar.gz soci
- mv soci /usr/local/bin/soci
- echo Logging in to Amazon ECR...
- export PASSWORD=$(aws ecr get-login-password --region ${AWS::Region})
```

**Build**

To create a SOCI index for the container image, the container image needs to be
stored in containerd's image store, not the Docker Engine image store. Therefore
in our build stage we create the container image, export it as a tarball and then
load it into containerd.

Containerd is already running in our CodeBuild instance because it sits
underneath the Docker Engine included in the CodeBuild container image. To tell
`ctr` and `soci` to use the existing containerd, we have set an environment
variable for our instance `CONTAINERD_ADDRESS="/var/run/docker/containerd/containerd.sock"`

Steps defined in the buildspec file:

1. Leveraging Moby's [buildkit](https://github.com/moby/buildkit) (exposed by
   `docker buildx`) we build the container image. There are a number of
   [exporters](https://docs.docker.com/build/exporters/) for buildkit, but given
   that the end result we want is a tarball, we decided to use a
   `docker-container` builder because it can export directly to an `.tar`. I'm
   also leveraging a `type=oci` so that my container image conforms to the OCI
   v1 spec, but this is not mandatory, and `type=docker` (the Docker v2.2 spec)
   would also work for SOCI.

   Alternatively you could use the default `docker` builder provided out of the
   box by the Docker Engine, and then run a `docker save -o image.tar
   $IMAGE_URI:$IMAGE_TAG` to get to the image out of the Docker Engine image store.

    ```bash
    - echo Building the container image
    - docker buildx create --driver=docker-container --use
    - docker buildx build --quiet --tag $IMAGE_URI:$IMAGE_TAG --file Dockerfile.v2 --output type=oci,dest=./image.tar .
    ```

2. Import the container image into containerd's image store.

    ```bash
    - echo Import the container image to containerd
    - ctr image import ./image.tar
    ```

3. Index the container image using the `soci` CLI.

    ```bash
    - echo Generating SOCI index
    - soci create $IMAGE_URI:$IMAGE_TAG
    ```

4. Push both the container image and the SOCI index up to ECR using the user
   credentials we retrieved in the pre build.

    ```bash
    - echo Pushing the container image
    - ctr image push --user AWS:$PASSWORD $IMAGE_URI:$IMAGE_TAG
    - echo Push the SOCI index to ECR
    - soci push --user AWS:$PASSWORD $IMAGE_URI:$IMAGE_TAG
    ```

## Building a Multi Architecture Pipeline

When building container images it is common to build for both x86 and arm64
architectures at the same time. While you could build arm64 container images on
x86 hosts through emulation, it is a lot more performant to natively build arm64
images on arm64 hosts. Therefore we have included an example CodePipeline that
will concurrently build two container images, one x86 image, one arm64 image in
separate CodeBuild jobs on the native architectures.

After both container images have been built, we then need to create a [Docker
manifest list](https://distribution.github.io/distribution/#manifest-list) (or
an [OCI image
index](https://github.com/opencontainers/image-spec/blob/main/image-index.md)).
This additional metadata file is used as a signpost by the container runtime
when deciding which container image to use. I.e. a container runtime on a x86
host, will read the manifest list to discover the image digest for the x86
image. Once its retrieved the x86 container image digest, it will go and
download that image from the repository.

You can create a Docker manifest list in a CodeBuild container based builder
with the `docker manifest`
[docs](https://docs.docker.com/engine/reference/commandline/manifest/) or
`docker buildx imagetools`
([docs](https://docs.docker.com/engine/reference/commandline/buildx_imagetools/)).
There is a great example
[here](https://github.com/aws-samples/aws-multiarch-container-build-pipeline/blob/b1060d397751b1c9113a2c1c86c2d5565faa5f85/lib/build-manifest.ts#L70)
using `docker manifest`.

However in our example we have leveraged the new [CodeBuild Lambda based builder](https://aws.amazon.com/about-aws/whats-new/2023/11/aws-codebuild-lambda-compute/),
to build the manifest list using the `manifest-tool` cli
([source](https://github.com/estesp/manifest-tool)).

To deploy the multi architecture pipeline:

```bash
aws cloudformation \
    create-stack \
    --stack-name soci-multi-arch-pipeline \
    --template-body file://multiarch-cloudformation.yaml \
    --capabilities CAPABILITY_IAM
```

### Buildx work around

There is a difference in the [CodeBuild Amazon Linux
images](https://github.com/aws/aws-codebuild-docker-images) for x86 and arm64.
The arm64 image [does not include docker
buildx](https://github.com/aws/aws-codebuild-docker-images/issues/640),
therefore the [method](#walkthrough-of-the-buildspec-file) used in our x86
buildspec file can not be reused.

Instead we build the container image using `docker build`, and then export the
container image out of the Docker Engine image store with `docker save`, ready
to be imported into the containerd image store. Importing into the containerd
image store and generating the SOCI index is identical to the x86 buildspec
file.

Snippet from the arm64 buildspec file

```yaml
- echo Building the container image
- docker build --quiet --tag $IMAGE_URI:$IMAGE_TAG --file Dockerfile.v2 .
- echo Export the container image
- docker save --output ./image.tar $IMAGE_URI:$IMAGE_TAG
```