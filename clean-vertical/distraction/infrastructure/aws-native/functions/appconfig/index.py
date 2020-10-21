import urllib.request


def handler(event, context):
    url = f'http://localhost:2772/applications/test-app/environments/dev-env/configurations/test-clean-vertical-conf-profile'
    with urllib.request.urlopen(url) as response:
        config = response.read()
        return {
            StatusCode: 200,
            Body: config
        }
