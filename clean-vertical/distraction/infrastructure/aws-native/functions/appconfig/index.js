const AWSXRay = require('aws-xray-sdk-core')
const AWS = AWSXRay.captureAWS(require('aws-sdk'))

const appconfig = new AWS.AppConfig({
    apiVersion: "2019-10-09"
})

async function handler() {

    const segment = AWSXRay.getSegment()

    try {
        segment.addAnnotation("ghkey", "some value")
        segment.addMetadata("ghMETAkey", "some META value")

        const result = await appconfig.getConfiguration({
            Application: "test-app",
            Configuration: "test-clean-vertical-conf-profile",
            Environment: "dev-env",
            // ClientConfigurationVersion: "1",
            ClientId: "cv-appconfig-lambda"
        }).promise()

        const buffer = Buffer.from(result.Content, 'base64').toString('utf8')
        const config = JSON.parse(buffer)

        return {
            statusCode: 200,
            body: JSON.stringify({result, buffer, config}),
        }
    } catch (e) {
        segment.addError(e)
    } finally {
        segment.close()
    }
}

exports.handler = handler