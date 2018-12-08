import fdk
import ujson
import os
import random
import flickrapi
import ssl
import sys
import requests
from fdk import fixtures

ssl._create_default_https_context = ssl._create_unverified_context


flickr = flickrapi.FlickrAPI(
    os.environ.get("FLICKR_API_KEY"),
    os.environ.get("FLICKR_API_SECRET"),
    token_cache_location='/tmp',
    format='parsed-json'
)

PHOTO_SOURCE_URL = 'https://farm{0}.staticflickr.com/{1}/{2}_{3}{4}.{5}'

def get_image_url(photo_dict):
    return PHOTO_SOURCE_URL.format(
        photo_dict['farm'], photo_dict['server'],
        photo_dict['id'], photo_dict['secret'],
        '_c', 'jpg'
    )


def photo_to_payload(body, photo_dict):
    return {
        "id": photo_dict.get('id'),
        "image_url": get_image_url(photo_dict),
        "countrycode": body.get("countrycode"),
        "bucket": body.get("bucket", "")
    }


def handler(ctx, data=None, loop=None):
    payloads = []
    if data and len(data) > 0:
        body = ujson.loads(data)
        photos = flickr.photos.search(
            text=body.get("query", "baby smile"),
            per_page=int(body.get("num", "5")),
            page=int(body.get("page", int(random.uniform(1, 50)))),
            extras="original_format",
            safe_search="1",
            content_type="1",
        )

        for p in photos.get('photos', {'photo': {}}).get('photo', []):
            payloads.append(photo_to_payload(body, p))

    for p in payloads:
        this_payload = {
            "media":[p.get("image_url")], 
            "event_id": "test_id",
            "event_type": "test_type",
            "ran_on": "<fn-hostname-or-dns>"
        }
        process_url = "http://docker.for.mac.localhost:8080/t/cloudevents/image-processor"
        r = requests.post(process_url, data=ujson.dumps(this_payload))
    
    return {"result": payloads}


if __name__ == "__main__":
    fdk.handle(handler)
