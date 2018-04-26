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

os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'
FN_PREFIX = "/function"
FAST_TF_GRAPH = FN_PREFIX + "/frozen_inference_graph.pb"
LABEL_MAP = FN_PREFIX + "/coco_label_map.json"
FN_API_URL = os.environ.get("FN_API_URL")
FN_APP_NAME = os.environ.get("FN_APP_NAME")
SENSITIVITY = float(os.environ.get("DETECT_SENSITIVITY", "0.3"))


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
    root.setLevel(logging.DEBUG)
    ch = logging.StreamHandler(sys.stderr)
    ch.setLevel(logging.INFO)
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


def with_graph(label_map):

    sess = tf.Session()
    sess.graph.as_default()

    def fn(ctx, data=None, loop=None):
        log = get_logger(ctx)
        log.info("tf graph imported")
        # Read and pre-process an image.
        data = ujson.loads(data)
        media_url = data.get("media_url")

        log.info("request data unmarshaled, media_url: %s" % media_url)
        resp = requests.get(media_url)
        resp.raise_for_status()

        img = cv.imdecode(
            np.array(bytearray(resp.content), dtype=np.uint8),
            cv.COLOR_GRAY2BGR
        )
        log.info("image loaded")

        rows = img.shape[0]
        cols = img.shape[1]
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
        num_detections = int(out[0][0])
        log.info("detection completed, objects found: %s" % num_detections)
        for i in range(num_detections):
            class_id = int(out[3][0][i])
            label = get_label_by_id(class_id, label_map)
            score = float(out[1][0][i])
            bbox = [float(v) for v in out[2][0][i]]
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
                cv.rectangle(
                    img,
                    (int(x), int(y)),
                    (int(right), int(bottom)),
                    (125, 255, 51), thickness=2
                )
                cv.putText(
                    img,
                    label.get("display_name"),
                    (int(x)+100, int(y)+40),
                    cv.FONT_HERSHEY_SIMPLEX,
                    2, (255, 255, 255),
                    2, cv.LINE_AA
                )

        log.info("image was processed and updated")
        h = hashlib.md5()
        h.update(media_url.encode("utf-8"))
        filename = 'poster_%s.jpeg' % h.hexdigest()
        cv.imwrite(filename, img)
        log.info("image was written to a file: {0}".format(filename))
        api = setup_twitter()

        with open(filename, "rb") as photo:
            event_id = data.get("event_id")
            event_type = data.get("event_type")
            status = "Event ID: {0}\nEvent type: {1}".format(event_id, event_type)
            resp = api.upload_media(media=photo)
            log.info("image posted as tweet")
            api.update_status(status=status, media_ids=[resp["media_id"], ])
            log.info("image tweet updated with status: {0}".format(status))

    return fn


if __name__ == "__main__":
    load_tf_graph()
    fdk.handle(with_graph(load_label_map()))
