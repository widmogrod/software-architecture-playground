import urllib.request
import json

def handler(event, context):
    url = f'http://localhost:2772/applications/test-app/environments/dev-env/configurations/test-clean-vertical-conf-profile'
    try:
        with urllib.request.urlopen(url) as response:
            config = response.read()
            return {
                'statusCode': 200,
                'body': config,
            }
    except Exception as e:
        return {
            'statusCode': 400,
            'body': e.reason,
        }