import ujson
import os
import flickrapi
import random
import requests
import fdk

from cloudevents.sdk import converters
from cloudevents.sdk import marshaller
from cloudevents.sdk.converters import structured
from cloudevents.sdk.event import v02

#ssl._create_default_https_context = ssl._create_unverified_context

flickr = flickrapi.FlickrAPI(
    os.environ.get("FLICKR_API_KEY"),
    os.environ.get("FLICKR_API_SECRET"),
    token_cache_location='/tmp',
    format='parsed-json'
)

def handler(ctx, data=None, loop=None):
    
    # Scrape Flickr
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

        # For each photo
        for p in photos.get('photos', {'photo': {}}).get('photo', []):
            photo_url_tpl = 'https://farm{0}.staticflickr.com/{1}/{2}_{3}{4}.{5}'
            photo_url = photo_url_tpl.format(p['farm'], p['server'],p['id'], p['secret'], '_c', 'jpg')

            data = {"photo_url": photo_url}

            event = (
                v02.Event().
                SetContentType("application/json").
                SetData(data).
                SetEventID("my-id").
                SetSource("scraper").
                SetEventType("cloudevent.flickr.image")
            )
            event_json = ujson.dumps(event.Properties())
            print(event_json)

            process_url = "http://docker.for.mac.localhost:8080/t/cloudevents/image-processor"
            r = requests.post(process_url, data=event_json)


if __name__ == "__main__":
    fdk.handle(handler)
