# Triggo - Delayed timer trigger for IFTTT

Triggo allows you to setup delayed actions in IFTTT through Webhook actions.

Currently, Triggo is primarily used to delay actions incoming from a Google Assistant applet in IFTTT.

My main goal was to have the ability to execute commands such as:

* Hey Google, turn off the TV in 30 minutes
* Google, turn on the kitchen light in 5 minutes
* etc..

## Usage

Triggo has two modes: server mode and worker mode.

Both modes require a valid running Redis instance to enqueue and dequeue background jobs. To connect Triggo to the Redis instance:

* provide a URL through the `-redis-url` flag 
* set a `REDIS_URL` environment variable

## Server Mode

The web server mode is used to accept POST requests incoming from IFTTT webhooks.
When a valid request has been received, Triggo will queue a background job to execute after the delay specified in the request body.

#### Routes:
```
POST /create

# Creates and schedules an event ({trigger_type}_{device}) to be executed after the given delay

Request Body:
{
  "device": "Bedroom TV",
  "delay_mins: "30",
  "trigger_type": "turn_off"
  "secret_key": "yoursecretkey"
}
```

## Worker Mode

The background worker mode is in charge of firing off a webhook event back to IFTTT after a request's delay has expired.

To start Triggo in background mode, set the `-worker=true` flag. 

You will also need to provide a valid IFTTT API URL and IFTTT API Key.   

You have two ways of doing this:
* Using the `-ifttt-api-key` and `-ifttt-api-url` flags
* Setting the `IFTTT_API_KEY` and `IFTTT_API_URL` variables in your environment.

#### POSTing to IFTTT

Triggo will POST event payloads to IFTTT using the API URL and API Key provided.  

For example, given an initial IFTTT request was received (by Triggo in server mode) with the following data:

```
{
  "device": "Bedroom TV",
  "delay_mins: "30",
  "trigger_type": "turn_off"
  "secret_key": "yoursecretkey"
}
```

Triggo (worker mode) will send the following request to IFTTT after the delay (30 mins):

```
POST https://{IFTTT_API_URL}/trigger/{trigger_type}_{device}/with/key/{IFTTT_API_KEY}

Response (from IFTTT):

Congratulations! You've fired the {trigger_type}_{device} event
```

#### IFTTT Event Name / Key
From the previous example, we are sending IFTTT an event name in the following format:

```
{trigger_type}_{device} # This data was received from the initial IFTTT webhook
```

In a perfect world, this would be sufficient to be able to execute X command in Y device. Unfortunately, the Google Assistant will sometimes pick up extra words as part of applet command. 

E.g When saying "turn off the kitchen lights in 30 minutes", the following payload is generated:
```
{
  "device": "the kitchen lights",
  "delay_mins: "30",
  "trigger_type": "turn_off"
  "secret_key": "yoursecretkey"
}
```

As you can see, an extra 'the' was appended to the device name. Once our trigger was scheduled and ready to go back to IFTTT, the event name would be `turn_off_the_kitchen_lights`. 

In order to keep things normalized, Triggo does the following preprocessing of the device name before scheduling a trigger:

1. remove the following definite articles from the device name: a, an, and, the
2. remap using mappings.yaml e.g. the lights => bedroom_lamp
3. singularize device name e.g. bedroom lights => bedroom light, lamps => lamp, etc...
4. convert to underscore_case e.g. Bedroom TV => bedroom_tv

Once the device name has been normalized, the trigger will be scheduled and sent to IFTTT once the delay has passed.

Going back to the previous example, given the following incoming request:

```
{
  "device": "the kitchen lights",
  "delay_mins: "30",
  "trigger_type": "turn_off"
  "secret_key": "yoursecretkey"
}
```

Triggo will preprocess `the kitchen lights` into 'kitchen_light' and schedule the event `turn_off_kitchen_light` to be sent to IFTTT in 30 minutes.

#### Supported Device Mappings

Triggo supports mapping humanized device names to a standardized normalized device name to ensure the correct event type gets sent to IFTTT.

Given the following mappings.yaml:

```
tv: bedroom_tv
light: bedroom_lamp
lamp: bedroom_lamp
my_light: bedroom_lamp
kitchen_bulb: kitchen_light
```

Triggo will be able to convert the following requests:

```
1. Trigger Request:
{
  "device": "tv",
  "delay_mins: "30",
  "trigger_type": "turn_off"
  "secret_key": "yoursecretkey"
}
Maps To Event => "turn_off_bedroom_tv"


2. Trigger Request:
{
  "device": "light",
  "delay_mins: "30",
  "trigger_type": "turn_off"
  "secret_key": "yoursecretkey"
}
Maps To Event => "turn_off_bedroom_lamp"


3. Trigger Request:
{
  "device": "my_light",
  "delay_mins: "30",
  "trigger_type": "turn_off"
  "secret_key": "yoursecretkey"
}
Maps To Event => "turn_off_bedroom_lamp"


4. Trigger Request:
{
  "device": "kitchen_bulb",
  "delay_mins: "30",
  "trigger_type": "turn_on"
  "secret_key": "yoursecretkey"
}
Maps To Event => "turn_on_kitchen_light"
```


## Configuration

Triggo supports the following flag / configuration values:

```
▶ ./triggo -h
Usage of ./triggo:

-ifttt-api-key string
      API Key for IFTTT [required for worker]
-ifttt-api-url string
      API URL for IFTTT [required for worker]
-namespace string
      namespace used for redis work queues. Defaults to program name (arg[0]). (default "triggo")
-port string
      port used to listen for incoming http requests [required for server]
-redis-url string
      url of Redis instance [required]
-secret-key string
      key used to authenticate requests
-worker
      run as background worker node

``` 

## Deployment

You can deploy Triggo to any server that has the golang runtime >= 1.12. 

To get Triggo running, please make sure you have a Triggo instance running in server mode (to accept incoming IFTTT requests) and another in worker mode (to send requests to IFTTT)

I personally used [Dokku](http://dokku.viewdocs.io/dokku/) to deploy my personal Triggo instance. I supplied a very basic Procfile that can be used for dokku / herokuish / heroku. 

If deploying to a heroku-like PaaS, make sure you have at least one worker process scaled in your server / dyno:
```
(heroku|dokku) ps:scale worker=1
```

## IFTTT Setup

In my IFTTT account, I setup the following applet to send requests to Triggo:

```
APPLET NAME: 

If You say "Turn off the $ in # minutes", then Make a web request


IF THIS: 

[Google Assistant] Say a phrase with both a number and a text ingredient 

* What do you want to say?: Turn off the $ in # minutes 
* What's another way to say it?: Stop the $ in # minutes
* And another way?: Turn off $ in # minutes

THEN THAT: 

[Webhooks] Make a web request

* URL: {{TRIGGO URL}}
* Method: POST
* Content Type: application/json
* Body: { 
  "device": "{{TextField}}", 
  "delay_mins": "{{NumberField}}", 
  "created_time_str": "{{CreatedAt}}", 
  "trigger_type": "turn_off", 
  "secret_key": "yoursecretkey" 
}
```

To receive Triggo events, I also have several IFTTT applets of the following format:

```
# Example. YMMV depending on which IoT devices you have in your house

APPLET NAME:

If Maker Event "turn_off_bedroom_lamp", then turn off Bedroom Lamp

IF THIS:

[Webhooks] Receive a web request

* Event Name: 'turn_off_bedroom_lamp'


THEN THAT:

[Wemo] Turn off 

* Which switch?: Bedroom Lamp




