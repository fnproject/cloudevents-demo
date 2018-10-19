# All Rights Reserved.
#
#    Licensed under the Apache License, Version 2.0 (the "License"); you may
#    not use this file except in compliance with the License. You may obtain
#    a copy of the License at
#
#         http://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
#    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
#    License for the specific language governing permissions and limitations
#    under the License.

import time
import fdk
import numpy as np
import tensorflow as tf
import cv2 as cv
import requests
import ujson
import logging
import sys
import os
import twython
import hashlib
import slackclient
from PIL import Image


os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'
FN_PREFIX = "/function/tf_models"
FAST_TF_GRAPH = FN_PREFIX + "/frozen_inference_graph.pb"
LABEL_MAP = FN_PREFIX + "/coco_label_map.json"
FN_API_URL = os.environ.get("FN_API_URL")
FN_APP_NAME = os.environ.get("FN_APP_NAME")
SENSITIVITY = float(os.environ.get("DETECT_SENSITIVITY", "0.3"))
HaarPath = FN_PREFIX + "/haarcascade_frontalface_default.xml"
FaceCascade = cv.CascadeClassifier(HaarPath)
ThugMaskPath = FN_PREFIX + "/thuglife_mask.png"
FnLogo = FN_PREFIX + "/fn-logo.png"
SLACK_TOKEN = os.environ.get("SLACK_API_TOKEN")
SLACK_CHANNEL = os.environ.get("SLACK_CHANNEL")


def setup_twitter():
    consumer_key = os.environ.get("TWITTER_CONSUMER_KEY")
    consumer_secret = os.environ.get("TWITTER_CONSUMER_SECRET")
    access_token = os.environ.get("TWITTER_ACCESS_TOKEN_KEY")
    access_token_secret = os.environ.get("TWITTER_ACCESS_TOKEN_SECRET")

    return twython.Twython(
        consumer_key, consumer_secret,
        access_token, access_token_secret
    )


def get_logger(ctx):
    root = logging.getLogger()
    root.setLevel(logging.INFO)
    ch = logging.StreamHandler(sys.stderr)
    call_id = ctx.CallID()
    formatter = logging.Formatter(
        '[call: {0}] - '.format(call_id) +
        '%(asctime)s - '
        '%(name)s - '
        '%(levelname)s - '
        '%(message)s'
    )
    ch.setFormatter(formatter)
    root.addHandler(ch)
    return root


def load_label_map():
    with open(LABEL_MAP, "r") as f:
        label_map = ujson.load(f)
        return label_map


def get_label_by_id(label_id, label_map):
    for m in label_map:
        _id = m.get("id")
        if _id is None:
            return {"display_name": "unknown"}
        if int(_id) == label_id:
            return m


def load_tf_graph():
    with tf.gfile.FastGFile(FAST_TF_GRAPH, 'rb') as f:
        graph_def = tf.GraphDef()
        graph_def.ParseFromString(f.read())
        tf.import_graph_def(graph_def, name='')
    return graph_def


def process_media(sess, media_url, log):
    log.info("request data unmarshaled, media_url: %s" % media_url)
    resp = requests.get(media_url)
    resp.raise_for_status()

    img = cv.imdecode(
        np.array(bytearray(resp.content), dtype=np.uint8),
        cv.COLOR_GRAY2BGR
    )
    log.info("image loaded")

    inp = cv.resize(img, (300, 300))
    log.info("image resized")
    inp = inp[:, :, [2, 1, 0]]  # BGR2RGB
    log.info("image formatted as an input")
    # Run the model
    out = sess.run(
        [sess.graph.get_tensor_by_name('num_detections:0'),
         sess.graph.get_tensor_by_name('detection_scores:0'),
         sess.graph.get_tensor_by_name('detection_boxes:0'),
         sess.graph.get_tensor_by_name('detection_classes:0')],
        feed_dict={'image_tensor:0':
                   inp.reshape(1, inp.shape[0], inp.shape[1], 3)})
    return img, out


def apply_mask(img, x1, y1, x2, y2):
    img_pil = Image.fromarray(cv.cvtColor(img, cv.COLOR_BGR2RGB)).convert('RGB')

    mask = Image.open(ThugMaskPath)

    person = img[y1:y2, x1:x2]
    person_pil = Image.fromarray(cv.cvtColor(person, cv.COLOR_BGR2RGB))

    gray = cv.cvtColor(person, cv.COLOR_BGR2GRAY)
    faces = FaceCascade.detectMultiScale(gray, 1.15)
    for (x, y, w, h) in faces:
        mask = mask.resize((w, h), Image.ANTIALIAS)
        offset = (x, y)
        person_pil.paste(mask, offset, mask=mask)

    img_pil.paste(person_pil, (x1, y1))

    return cv.cvtColor(np.array(img_pil), cv.COLOR_RGB2BGR)


def process_detection(out, img, label_map, detection_index, log):
    rows = img.shape[0]
    cols = img.shape[1]
    class_id = int(out[3][0][detection_index])
    label = get_label_by_id(class_id, label_map)
    score = float(out[1][0][detection_index])
    bbox = [float(v) for v in out[2][0][detection_index]]
    if score > SENSITIVITY:
        log.info("\nobject class id: {0}"
                 "\nobject display name: {1}"
                 "\nscore: {2}\n"
                 .format(class_id,
                         label.get("display_name"),
                         score))
        x = bbox[1] * cols
        y = bbox[0] * rows
        right = bbox[3] * cols
        bottom = bbox[2] * rows
        if label.get("display_name") == "person":
            img = apply_mask(img, int(x), int(y), int(right), int(bottom))
        
        cv.rectangle(
            img,
            (int(x), int(y)),
            (int(right), int(bottom)),
            (125, 255, 51), thickness=2
        )
        cv.putText(
            img,
            label.get("display_name"),
            (int(x), int(y)+20),
            cv.FONT_HERSHEY_SIMPLEX,
            0.5, (255, 255, 255),
            2, cv.LINE_AA
        )

    return class_id, score, img


def setup_img_path(media_url):
    h = hashlib.md5()
    h.update(media_url.encode("utf-8"))
    return '/tmp/poster_%s.jpeg' % h.hexdigest()


def post_media(slack_client, slack_channel,
               filename, entities, status, log):
    medias = entities.get("media", [])
    for media in medias:
        if "media_url_https" in media:
            url = media.get("media_url_https")
            if url is not None:
                response = post_image_to_slack(
                    slack_client, slack_channel, filename, url, status
                )
                if "ok" in response and response["ok"]:
                    log.info("message posted to Slack successfully "
                             "from image: {0}".format(url))
                else:
                    if "headers" in response:
                        hs = response["headers"]
                        if "Retry-After" in hs:
                            delay = int(response["headers"]["Retry-After"])
                            time.sleep(delay)
                            post_image_to_slack(
                                slack_client, slack_channel,
                                filename, url, status
                            )
                        else:
                            raise Exception(ujson.dumps(response))


def post_image_to_slack(slack_client, slack_channel,
                        filename, img_url, status):
    return slack_client.api_call(
        "chat.postMessage",
        channel=slack_channel,
        text=status,
        attachments=ujson.dumps([{
            "title": filename,
            "image_url": img_url
        }])
    )


def post_image(ctx, status, media_url, img):
    log = get_logger(ctx)
    slack_client = None
    if SLACK_TOKEN is not None or SLACK_CHANNEL is not None:
        slack_client = slackclient.SlackClient(SLACK_TOKEN)
    else:
        log.warning("missing slack token or channel is missing, skipping...")
    twitter_api = setup_twitter()
    log.info("image was processed and updated")
    filename = setup_img_path(media_url)
    cv.imwrite(filename, img)
    log.info("image was written to a file: {0}".format(filename))
    with open(filename, "rb") as photo:
        resp = twitter_api.upload_media(media=photo)
        log.info("response content type: {0}".format(type(resp)))
        log.info("after twitter image upload:\n\n\n")
        log.info(ujson.dumps(resp))
        log.info("\n\n\nimage posted as tweet")
        tweet = twitter_api.update_status(status=status, media_ids=[resp["media_id"], ])
        log.info("response content type: {0}".format(type(tweet)))
        log.info("\n\n\nafter twitter status updated\n\n\n")
        log.info(tweet)
        log.info("\n\n\nimage tweet updated with status: {0}\n\n\n".format(status))
        same_tweet = twitter_api.show_status(id=tweet["id"])
        log.info("response content type: {0}".format(type(same_tweet)))
        log.info("same tweet later:\n\n\n")
        log.info(same_tweet)
        log.info("\n\n\n")

        if slack_client is not None and SLACK_CHANNEL is not None:
            if "entities" in tweet:
                entities = tweet["entities"]
                if "media" in entities:
                    post_media(
                        slack_client, SLACK_CHANNEL,
                        filename, entities, status, log
                    )


def add_fn_logo(img):
    height, width, _ = img.shape
    img_pil = Image.fromarray(cv.cvtColor(img, cv.COLOR_BGR2RGB)).convert('RGB')
    mask = Image.open(FnLogo)
    mask_width, mask_height = mask.size
    img_ratio = float(height/width)
    mask_ratio = float(mask_height/mask_width)
    mask_scale_ration = 1
    if img_ratio > mask_ratio:
        # if original image is big - we need to scale up logo
        mask_scale_ration = img_ratio / mask_ratio
    if mask_ratio > img_ratio:
        # if original image is small - we need to scale down logo
        mask_scale_ration = mask_ratio / img_ratio

    custom_ratio = 3

    if 4 * mask_height > height:
        custom_ratio *= 1.5
    if 3 * mask_width > width:
        custom_ratio *= 2

    if height > 4 * mask_height:
        custom_ratio /= 1.5

    if width > 3 * mask_width:
        custom_ratio /= 2

    mask = mask.resize(
        (int(mask_width * mask_scale_ration / custom_ratio),
         int(mask_height * mask_scale_ration / custom_ratio)), Image.ANTIALIAS)
    img_pil.paste(mask, (10, 10), mask=mask)

    return cv.cvtColor(np.array(img_pil), cv.COLOR_RGB2BGR)


def process_single_media_file(ctx, sess, media_url, label_map,
                              event_id, event_type, ran_on):
    log = get_logger(ctx)
    scores = []
    classes = []
    img, out = process_media(sess, media_url, log)
    num_detections = int(out[0][0])
    log.info("detection completed, objects found: %s" % num_detections)
    for i in range(num_detections):
        class_id, score, img = process_detection(
            out, img, label_map, i, log)
        scores.append(score)
        classes.append(class_id)

    max_score = max(scores)
    max_score_label = classes[scores.index(max_score)]
    status = (
        'Event ID: {0}\nSource: {1}\n'
        'Ran On: {2}\nClassifier: {4}\nScore: {3}\n'
        .format(event_id, event_type, ran_on,
                str(max_score)[:3],
                get_label_by_id(
                    max_score_label,
                    label_map).get("display_name").upper())[:140]
    )
    log.info("status: {0}".format(status))
    return img, status


def with_graph(label_map):

    sess = tf.Session()
    sess.graph.as_default()

    def fn(ctx, data=None, loop=None):
        log = get_logger(ctx)
        log.info("tf graph imported")
        if data is not None or len(data) !=0:
            data = ujson.loads(data)
            log.info("incoming data: {0}".format(ujson.dumps(data)))
            media = data.get("media", [])
            event_id = data.get("event_id")
            event_type = data.get("event_type", "")
            if event_type.startswith("Microsoft"):
                event_type = "Azure"
                event_id = event_id.replace("-", "")
            ran_on = data.get("ran_on", "Fn Project on Oracle Cloud")

            for media_url in media:
                img, status = process_single_media_file(
                    ctx, sess, media_url, label_map,
                    event_id, event_type, ran_on
                )

                post_image(ctx, status, media_url, add_fn_logo(img))
            else:
                log.info("missing data")

    return fn


if __name__ == "__main__":
    load_tf_graph()
    fdk.handle(with_graph(load_label_map()))
