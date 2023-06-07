import json
import logging
import os
import sys
from datetime import datetime

import boto3
import pandas as pd
from botocore.exceptions import ClientError

root = logging.getLogger()
if root.handlers:
    for handler in root.handlers:
        root.removeHandler(handler)
logging.basicConfig(format='%(asctime)s %(message)s', level=logging.INFO)


def main():

    # Ensure the User has set the LOG GROUP NAME for the container.
    if "LOG_GROUP_NAME" not in os.environ:
        logging.info(
            "LOG_GROUP_NAME has not been set in environment variables")
        sys.exit(1)

    log_group_name = os.environ.get("LOG_GROUP_NAME")
    logging.info('Using Log Group %s', log_group_name)

    # Ensure the User has set the AWS REGION for the container.
    if "AWS_REGION" not in os.environ:
        logging.info(
            "AWS_REGION has not been set in environment variables")
        sys.exit(1)

    aws_region = os.environ.get("AWS_REGION")
    logging.info('Using AWS Region %s', aws_region)

    # Get all of the Log Streams from the Cloudwatch logs API
    client = boto3.client("logs", region_name=aws_region)
    try:
        logging.info('Getting Log Streams for Log Group %s', log_group_name)
        response = client.describe_log_streams(
            logGroupName=log_group_name
        )
    except ClientError as error:
        logging.error(error)
        sys.exit(1)

    logstreams = response['logStreams']

    # For Each Log Group Stream find all of the log events. For Each event
    # format the timestamps, retrieve the Task Id and the the Family and add
    # them to an array of all of the Tasks.
    all_tasks = []
    for logstream in logstreams:
        logging.info('Getting Log Events for Log Stream %s',
                     logstream['logStreamName'])
        try:
            response = client.get_log_events(
                logGroupName=log_group_name,
                logStreamName=logstream['logStreamName']
            )
        except ClientError as error:
            logging.error(error)
            sys.exit(1)

        events = response['events']

        for event in events:
            event_raw = json.loads(event['message'])

            format = "%Y-%m-%dT%H:%M:%S.%fZ"
            created_at = datetime.strptime(
                event_raw['detail']['pullStoppedAt'], format)
            started_at = datetime.strptime(
                event_raw['detail']['startedAt'], format)

            delta = started_at - created_at

            taskfamily_raw = event_raw['detail']['taskDefinitionArn'].split(
                "/")[1]
            taskfamily = taskfamily_raw.split(":")[0]
            taskfamilyrev = taskfamily_raw.split(":")[1]

            task = {
                "taskId": event_raw['detail']['taskArn'].split("/")[2],
                "taskFamily": taskfamily,
                "taskFamilyRev": taskfamilyrev,
                "createdAt": created_at.strftime("%H:%M:%S:%f"),
                "startedAt": started_at.strftime("%H:%M:%S:%f"),
                "startTimedeltaSeconds": delta.total_seconds()
            }

            all_tasks.append(task)

    # To make the data easier to visualize, I am going to convert the array into
    # a DataFrame.
    df = pd.DataFrame(all_tasks)

    # This shows all of the raw data in my table.
    print("Printing Raw Table")
    print(df.to_markdown())

    # If I wanted to look at the average pull time, I can group the DataFrame by
    # task family.
    print("Printing Average Pull Time Grouped By Task Family")
    df2 = df[['taskFamily', 'taskFamilyRev', 'startTimedeltaSeconds']
             ].groupby(['taskFamily', 'taskFamilyRev']).mean()
    print(df2.to_markdown())


if __name__ == "__main__":
    main()
