import os
import boto3
import requests
from requests_aws4auth import AWS4Auth

region = os.environ['AWS_REGION']
service = 'es'
credentials = boto3.Session().get_credentials()
awsauth = AWS4Auth(credentials.access_key, credentials.secret_key, region, service, session_token=credentials.token)

host = os.getenv('OPENSEARCH_HOST')
index = 'lambda-index'
datatype = '_doc'
url = host + '/' + index + '/' + datatype + '/'

headers = { "Content-Type": "application/json" }

def handler(event, context):
    count = 0
    for record in event['Records']:
        # Get the primary key for use as the OpenSearch ID
        id = record['dynamodb']['Keys']['ID']['S'] + ":" + record['dynamodb']['Keys']['Type']['S']

        if record['eventName'] == 'REMOVE':
            r = requests.delete(url + id, auth=awsauth)
            print("remove: ", r.text)
        else:
            document = record['dynamodb']['NewImage']
            r = requests.put(url + id, auth=awsauth, json=document, headers=headers)
            # this is needed to make the document immediately available for search
            # but because open search is build from dynamo db stream, it is not needed
            # since, data will always eventually be available
            # synchronous indexing it's a different story
            # r = requests.put(url + id + "?refresh=true", auth=awsauth, json=document, headers=headers)
            print("new: ", r.text)
        count += 1
    return str(count) + ' records processed.'