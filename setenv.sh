#!/usr/bin/env bash

set -xe

fn update fn $1 image-processor --config TWITTER_CONSUMER_KEY=$TWITTER_CONF_KEY
fn update fn $1 image-processor --config TWITTER_CONSUMER_SECRET=$TWITTER_CONF_SECRET
fn update fn $1 image-processor --config TWITTER_ACCESS_TOKEN_KEY=$TWITTER_TOKEN_KEY
fn update fn $1 image-processor --config TWITTER_ACCESS_TOKEN_SECRET=$TWITTER_TOKEN_SECRET
fn update fn $1 image-processor --config SLACK_API_TOKEN=$SLACK_API_TOKEN
fn update fn $1 image-processor --config SLACK_CHANNEL=$SLACK_CHANNEL
fn update fn $1 scraper --config FLICKR_API_KEY=$FLICKR_API_KEY
fn update fn $1 scraper --config FLICKR_API_SECRET=$FLICKR_API_SECRET
fn update app $1 --syslog-url $SYSLOG_URL
