# Monitoring Amazon ECS Task Events

When Lazy Loading Container Images on AWS Fargate with the SOCI Snapshotter,
there are multiple ways you can monitor the performance improvement, examples include:

1. Comparing the `startTime` of the container images within an ECS Task.
2. If the Application is a long running Application, for example a web service,
   comparing the time taken to respond to the first request.
3. If the Application is a short running, for example a batch processing
  workload, comparing the total time taken to complete the work.

In this directory, there is a sample solution for number 1. We will deploy a
solution that collects the [Amazon ECS
Events](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_cwe_events.html),
and stores them in Amazon CloudWatch Logs. Each time an Amazon ECS Task moves
through the [lifecycle
states](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-lifecycle.html),
events are emitted to Amazon EventBridge. By capturing, filtering, and
forwarding these events on to Amazon CloudWatch Logs, we have a long term
archive of the events allowing us to compare the `startTime` across multiple tasks.

## Comparing the Pull Time of Amazon ECS Tasks

Within an Amazon ECS [task state change
event](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_cwe_events.html#ecs_task_events),
there are metrics for `createdAt` and `startedAt`. If we take the `createdAt`
value away from the `startedAt` value, we will be able to see a combined time in
seconds for all of the container images to be pulled and the ECS Task to be
ready to start the application. Giving us the data for the first performance
metric mentioned above, the `startTime` of a Task.

1. We will deploy a CloudFormation template to capture ECS Task Events in
   EventBridge and forward them on to CloudWatch logs.

   If you are running short running workloads and want to capture the total run
   time of the Task, update line 28 in the template, from `RUNNING` events to
   `STOPPED` events, this will allow the metric `stoppedAt` to be collected and
   stored in Cloudwatch for additional queries.

> By default this is capturing events for the "default" ECS Cluster. If you would
> like to get events from a different ECS Cluster, then ensure to update or
> override the parameter.

```
aws cloudformation update-stack \
  --stack-name ecsevents \
  --template-body file://captureecsevents.yaml \
  --parameters ParameterKey=ECSClusterName,ParameterValue=default
```

2. Start running some Tasks! If you are trying to monitor the effectiveness of
   SOCI, run some Tasks with SOCI Indexes and some without, allowing you to get
   the before and after `startTime` for the workload.

3. Once the Events are being forwarded to CloudWatch Logs we want to analyze the
   data. Unfortunately [CloudWatch Log
   Insights](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/AnalyzingLogData.html)
   is unable to parse the timestamps in a way for us to work with, so instead we
   are using python and [pandas](https://pandas.pydata.org/) to parse the data
   locally.

   The script to query CloudWatch Logs and parse the data is available in the
   [query](/query) sub directory and can be ran locally with `python query/app.py`,
   however for convenience it can also be packaged into a container image along
   with the panda dependencies.

    ```bash
    # Build the Log query container image.
    docker build \
        --file Dockerfile \
        --tag logquery:latest \
        .

    # Note that we mount in the local AWS credentials file into the container.
    docker run \
      --rm \
      --env AWS_REGION=us-east-1 \
      --env LOG_GROUP_NAME=/ecs/cluster/events/default \
      --volume ${HOME}/.aws:/root/.aws \
      logquery:latest
    ```

   The output of the script is 2 tables, one showing the raw data of each Task,
   and one summarizing the data by Task Definition Family and Task Definition
   Family Revision.

   1. The raw data of each ECS Task:

    |    | taskId                           | taskFamily                     |   taskFamilyRev | createdAt       | startedAt       |   startTimedeltaSeconds |
    |---:|:---------------------------------|:-------------------------------|----------------:|:----------------|:----------------|------------------------:|
    |  1 | d31c70e879444e16a0f54db2f7e1ebd7 | socidemopytorchtraining        |               3 | 11:24:56:866000 | 11:24:57:536000 |                   0.67  |
    |  2 | 40059ec1332e4df783105e2471007432 | socidemopytorchtrainingnoindex |               2 | 11:26:33:359000 | 11:26:38:206000 |                   4.847 |
    |  5 | 40059ec1332e4df783105e2471007432 | socidemopytorchtrainingnoindex |               2 | 11:26:33:359000 | 11:26:38:206000 |                   4.847 |
    |  6 | b24c1acff4574c3ebec84cce91273485 | socidemopytorchtrainingnoindex |               2 | 11:26:22:952000 | 11:26:27:185000 |                   4.233 |
    |  8 | 92499e7b4c0644eda740709baacec6a5 | socidemopytorchtraining        |               3 | 11:24:57:491000 | 11:24:58:377000 |                   0.886 |
    |  9 | a4cc719d20d541f3a42cd4a2260ad04a | socidemopytorchtraining        |               3 | 11:24:57:081000 | 11:24:58:197000 |                   1.116 |
    | 11 | 92499e7b4c0644eda740709baacec6a5 | socidemopytorchtraining        |               3 | 11:24:57:491000 | 11:24:58:377000 |                   0.886 |
    | 12 | 796b9aa1e0d04fca8438a4da25659d32 | socidemopytorchtraining        |               3 | 11:24:56:230000 | 11:24:57:311000 |                   1.081 |
    | 13 | b44ffbc752da479ba875ff7b583c8519 | socidemopytorchtrainingnoindex |               2 | 11:26:33:348000 | 11:26:38:191000 |                   4.843 |
    | 15 | b27e649c93c34deab17ddb6ec0141fb0 | socidemopytorchtraining        |               3 | 11:24:57:824000 | 11:24:59:140000 |                   1.316 |
    | 16 | eb9b959db48e413d908cfafe6e210505 | socidemopytorchtrainingnoindex |               2 | 11:26:39:209000 | 11:26:42:001000 |                   2.792 |
    | 17 | 2a74e58766a44f3b887a8059340a0c2b | socidemopytorchtrainingnoindex |               2 | 11:26:37:380000 | 11:26:40:516000 |                   3.136 |
    | 18 | 796b9aa1e0d04fca8438a4da25659d32 | socidemopytorchtraining        |               3 | 11:24:56:230000 | 11:24:57:311000 |                   1.081 |
    | 19 | 7ac323e104804abca66d9ffd65eaa384 | socidemopytorchtrainingnoindex |               2 | 11:26:34:934000 | 11:26:37:862000 |                   2.928 |
    | 20 | eb9b959db48e413d908cfafe6e210505 | socidemopytorchtrainingnoindex |               2 | 11:26:39:209000 | 11:26:42:001000 |                   2.792 |
    | 21 | 64717ab435374c21a2c3b2853f68795d | socidemopytorchtrainingnoindex |               2 | 11:26:38:115000 | 11:26:41:464000 |                   3.349 |
    | 22 | 3eab0a09012844b997fab53c9c556677 | socidemopytorchtrainingnoindex |               2 | 11:26:19:768000 | 11:26:23:585000 |                   3.817 |
    | 24 | 64717ab435374c21a2c3b2853f68795d | socidemopytorchtrainingnoindex |               2 | 11:26:38:115000 | 11:26:41:464000 |                   3.349 |
    | 26 | 98d9ffc31582455eb0c865f70ae5e8cd | socidemopytorchtrainingnoindex |               2 | 11:26:37:032000 | 11:26:42:036000 |                   5.004 |
    | 27 | d31c70e879444e16a0f54db2f7e1ebd7 | socidemopytorchtraining        |               3 | 11:24:56:866000 | 11:24:57:536000 |                   0.67  |
    | 28 | b27e649c93c34deab17ddb6ec0141fb0 | socidemopytorchtraining        |               3 | 11:24:57:824000 | 11:24:59:140000 |                   1.316 |
    | 29 | 521ebe92e8114439a6c6368ff36ff05f | socidemopytorchtrainingnoindex |               2 | 11:26:34:538000 | 11:26:39:265000 |                   4.727 |
    | 30 | a4cc719d20d541f3a42cd4a2260ad04a | socidemopytorchtraining        |               3 | 11:24:57:081000 | 11:24:58:197000 |                   1.116 |

    2. A Summary grouping by Task Definition Family and Task Family Definition Revision

    |                                         |   startTimedeltaSeconds |
    |:----------------------------------------|------------------------:|
    | ('socidemopytorchtraining', '3')        |                1.0138   |
    | ('socidemopytorchtrainingnoindex', '2') |                3.89723  |
