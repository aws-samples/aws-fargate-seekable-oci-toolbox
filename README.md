# Seekable OCI on AWS Fargate Toolbox

A repository of tools to help you get started with SOCI and the [SOCI
Snapshotter](https://github.com/awslabs/soci-snapshotter).

For more information on SOCI and AWS Fargate see the:

- [Launch Blog](https://aws.amazon.com/blogs/aws/aws-fargate-enables-faster-container-startup-using-seekable-oci/)
- [Under the hood blog](https://aws.amazon.com/blogs/containers/under-the-hood-lazy-loading-container-images-with-seekable-oci-and-aws-fargate)

## Table of Contents

- [Generate a SOCI Index in a Container](./containerized-index-builder/) -
  useful when developing locally.
- [Generate a SOCI Index in a CI/CD Pipeline](./soci-codepipeline/)
- [Capture ECS Events to monitor Task startup times on AWS
  Fargate](./ecs-task-events/)
- Identify if the workload is running on the [SOCI Snapshotter on AWS
  Fargate](./am-i-lazy)

------

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

------

## License

This library is licensed under the MIT-0 License. See the LICENSE file.

------