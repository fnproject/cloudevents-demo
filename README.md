# CloudEvents demo

Demo rules: https://github.com/ac360/cloudevents-demo

## Configuration

All you need is to put Twitter account credentials to [app file](app.yaml).
Follow the instructions at https://developer.twitter.com/en/docs/basics/getting-started#get-started-app

## How to run

```bash
fn -v deploy --all --registry `whoaim`
```

## How to test

Using the following command you can test the application:
```bash
curl -v -X POST ${FN_API_URL}/r/cloudevents/cloudevent -d @tweet-entry/payloads/aws.payload.json
```

## Notes

A function that [does image processing](image-processor) has a config var: [`DETECT_SENSITIVITY`](image-processor/func.yaml), 
it has default value - `"0.3"`, it’s enough to detect most of the objects.
However, as lower you go then more objects detected but probability of the correct detection may vary from 0 to 100. 
So if you want to get more objects detected, just set that to `"0.1"`

## Acceptable payloads

[CloudEvent function](tweet-entry) accepts the following payloads:

 * [AWS CloudEvent](tweet-entry/payloads/aws.payload.json)
 * [Azure CloudEvent](tweet-entry/payloads/azure.payload.json)

Image processing function accepts the following payload:

 * [Batch image payload](image-processor/payload.sample.json)

## Setting up functions using Fn CLI (alternative to deploy)

```bash
# creates an application
fn apps create cloudevents

# sets application config
fn apps config set cloudevents TWITTER_CONSUMER_KEY ...
fn apps config set cloudevents TWITTER_CONSUMER_SECRET ...
fn apps config set cloudevents TWITTER_ACCESS_TOKEN_KEY ...
fn apps config set cloudevents TWITTER_ACCESS_TOKEN_SECRET ...

fn routes create cloudevents /cloudevent --image denismakogon/tweet-entry:0.0.1 --type async --format json --timeout 60 --idle-timeout 30
fn routes create cloudevents /image-processor --image denismakogon/image-processor:0.0.1 --type async --format json --timeout 3600 --idle-timeout 120 --memory 1024
fn routes config set cloudevents /image-processor DETECT_SENSITIVITY "0.3"
```
