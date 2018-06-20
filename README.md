DocBase Go Hook
=================

DocBase and Slack Comment Hook Tool

* [Requirement](#requirement)
* [Preparation](#preparation)

Requirement
-------------

- DocBase Account
- GCP Account


Preparation
-------------

## Mapping DocbaseAccount and SlackAccount

`$ cp config.json.sample config.json`

and set like this

```
{
  "accounts": {
    "DocBaseAccoount": "SlackUserID",
  }
}
```

## Setting secret.yaml

`$ cp secret.yaml.sample secret.yaml`

please set your incoming webhook url

```
env_variables:
  SLACK_INCOMING_WEBHOOK: "/TXXXXXXXX/XXXXXXXXX/XXXXXXXXXXXXXXXXXXXXXXXX"
  ```

## Login GCP account

```
$ gcloud config set account $GACCOUNT
$ gcloud config set project $GOOGLE_PROJECT_ID
```

## GAE deploy

```
$ gcloud app deploy
```

## Setting DocBase

After deploy, you can get GAE Endpoint like `https://xxx.appspot.com`,
You need to set DocBase Webhook URL , https://xxx.appspot.com/docbase/events
