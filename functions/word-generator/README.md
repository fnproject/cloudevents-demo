Word generator demo
-------------------

Idea
====

Generate missing words (part of speech) based on CloudEvent event type.

Formats
=======

This function works in the following CloudEvent formats:
 
 - structured (content-type: `application/cloudevents+json`)
 - binary (parse CloudEvent from HTTP headers)

Workflow
========

CloudEvent emitter sends the CloudEvent in one of the formats, a function responds with the result in the same format as an inbound CloudEvent.
The result of the execution is a CloudEvent, more information you may find [here](https://docs.google.com/document/d/1Vkrmz0vLyiJnUmHUeJfmFbBldDyD-DOFcBNOU-eEKeg/edit#).

How to call a function with binary CloudEvent
=============================================
```bash
curl -v -X POST -H 'ce-specversion: 0.1' \
    -H 'ce-time: 2018-10-23T12:28:22.4579346Z' \
    -H 'ce-id: 96fb5f0b-001e-0108-6dfe-da6e2806f124' \
    -H 'ce-source: http://srcdog.com/cedemo' \
    -H 'ce-type: word.found.name' \
    -H 'content-type: application/json' `fn inspect context | grep api | awk '{print $2}'`/t/cncf/word-generator-trigger
```

See [binary helper](call_binary.sh).

How to call a function with structured CloudEvent
=================================================

```bash
curl -v -X POST -H 'content-type: application/cloudevents+json' `fn inspect context | grep api | awk '{print $2}'`/t/cncf/word-generator-trigger -d '{"specversion": "0.1", "type": "word.found.exclamation", "id": "16fb5f0b-211e-1102-3dfe-ea6e2806f124", "time":"2018-10-23T12:28:23.3464579Z", "contenttype": "application/json"}'
```

See [structured helper](call_structured.sh).

Configuration
=============

By default a function will attempt to post the result to a callback URL supplied as `X-Callback-Url` request header.
In order to make function return the result to a caller a function needs additional configuration:

```bash
fn config fn cncf word-generator SYNC_MODE 1
```

How to deploy
=============

```bash
fn use context <your-context>
fn --verbose deploy --app cncf
```

To make a function return the result:
```bash
fn config fn cncf word-generator SYNC_MODE 1
```

Caution
=======

If you have a plan to use `fn invoke` do not forget to specify the content type:

`--content-type application/cloudevents+json`

With `fn invoke` you'd call a function in structured format, to call it in binary mode use cURL.
