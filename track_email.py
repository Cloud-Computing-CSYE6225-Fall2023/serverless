import os
import boto3


def track_emails_to_db(data):
    try:
        region = os.environ["region"]
        table_name = os.environ["dynamodb_table_name"]

        # Set up DynamoDB client
        dynamodb = boto3.resource('dynamodb', region_name=region)
        table = dynamodb.Table(table_name)
        # Put the item into the DynamoDB table
        response = table.put_item(Item=data)

        # Check the response
        if response['ResponseMetadata']['HTTPStatusCode'] == 200:
            return {
                "statusCode": 200,
                "msg": "Successfully uploaded to dynanodb"
            }

    except Exception as err:
        print("dynamo_err: ", str(err))
        return {
            "statusCode": 500,
            "msg": "Error creating item in dynanodb: " + str(err)
        }
