
# Kapacitor Alerts API

## Contents

* [Description](#description)

* [Installation and Usage](#installation-and-usage)

* [Variables](#variables)

* [5xx](#5xx)
  * [Get All Tasks](#1-get-all-tasks)
  * [Get Task](#2-get-task)
  * [Get Task State](#3-get-task-state)
  * [Create Task](#4-create-task)
  * [Update Task](#5-update-task)
  * [Delete Task](#6-delete-task)

* [Crashed](#crashed)
  * [Get All Tasks](#1-get-all-tasks-1)
  * [Get Task](#2-get-task-1)
  * [Create Task](#3-create-task)
  * [Update Task](#4-update-task)
  * [Delete Task](#5-delete-task)

* [Memory](#memory)
  * [Get All Tasks](#1-get-all-tasks-2)
  * [Get All Tasks for App](#2-get-all-tasks-for-app)
  * [Get Task](#3-get-task)
  * [Create Task](#4-create-task-1)
  * [Update Task](#5-update-task-1)
  * [Delete Task](#6-delete-task-1)

* [Release](#release)
  * [Get All Tasks](#1-get-all-tasks-3)
  * [Get Task](#2-get-task-2)
  * [Create Task](#3-create-task-1)
  * [Update Task](#4-update-task-1)
  * [Delete Task](#5-delete-task-1)

--------

## Description

This API communicates with Influx's Kapacitor alerts/monitoring API to enable monitoring and alerting on Akkeris apps based on certain criteria/events. Alerts can be configured for HTTP 5xx status codes, Memory usage, Akkeris Releases, and when an app crashes.

## API Variables

These are the variables used in this documentation:

- *KAPACITOR_ALERTS_API* : URI of the running instance
- *APP_NAME*: App to act on
- *SLACK_CHANNEL*: Slack channel to notify
- *EMAIL*: Email address to notify
- *POST*: URL to notify via webhook

## 5xx

### 1. Get All Tasks

Get a list of the configuration of the 5xx event monitoring on all apps

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/tasks/5xx
```


### 2. Get Task

Get the configuration of the 5xx event monitoring on an app

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/task/5xx/{{APP_NAME}}
```



### 3. Get Task State

Get the current state of the 5xx event monitoring on an app

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/task/5xx/{{APP_NAME}}/state
```



### 4. Create Task

Begin monitoring an app for 5xx events

***Endpoint:***

```bash
Method: POST
URL: {{KAPACITOR_ALERTS_API}}/task/5xx
```


***Headers:***

| Key | Value | Description |
| --- | ------|-------------|
| Content-Type | application/json |  |


***Body:***

```js        
{
	"app": "{{APP_NAME}}",		// App to monitor
	"tolerance": "low",		// Tolerance (low | medium | high)
	"slack": "{{SLACK_CHANNEL}}",	// *Optional* slack channel to notify
	"email": "{{EMAIL}}",		// *Optional* email address to notify
	"post": "{{POST}}"		// *Optional* URL to post a webhook to
}
```



### 5. Update Task

Update the configuration for 5xx monitoring on an app

***Endpoint:***

```bash
Method: PATCH
URL: {{KAPACITOR_ALERTS_API}}/task/5xx
```


***Headers:***

| Key | Value | Description |
| --- | ------|-------------|
| Content-Type | application/json |  |


***Body:***

```js        
{
	"app": "{{APP_NAME}}",		// App to monitor
	"tolerance": "medium",		// Tolerance (low | medium | high)
	"slack": "{{SLACK_CHANNEL}}",	// *Optional* slack channel to notify
	"email": "{{EMAIL}}",		// *Optional* email address to notify
	"post": "{{POST}}"		// *Optional* URL to post a webhook to
}
```


### 6. Delete Task

Stop monitoring an app for 5xx events

***Endpoint:***

```bash
Method: DELETE
URL: {{KAPACITOR_ALERTS_API}}/task/5xx/{{APP_NAME}}
```



## Crashed

Send an alert to a Slack channel, an email address, or as a webhook when an app crashes.

### 1. Get All Tasks

Get a list of the configuration of the crash event monitoring on all apps

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/tasks/crashed
```


### 2. Get Task

Get the configuration of the crash event monitoring on an app

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/task/crashed/{{APP_NAME}}
```



### 3. Create Task

Begin monitoring an app for crash events

***Endpoint:***

```bash
Method: POST
URL: {{KAPACITOR_ALERTS_API}}/task/crashed
```


***Headers:***

| Key | Value | Description |
| --- | ------|-------------|
| Content-Type | application/json |  |

***Body:***

```js        
{
	"app": "{{APP_NAME}}",		// App to monitor
	"slack": "{{SLACK_CHANNEL}}",	// *Optional* slack channel to notify
	"email": "{{EMAIL}}",		// *Optional* email address to notify
	"post": "{{POST}}"		// *Optional* URL to post a webhook to
}
```


### 4. Update Task

Update the configuration for crash monitoring on an app

***Endpoint:***

```bash
Method: PATCH
URL: {{KAPACITOR_ALERTS_API}}/task/crashed
```


***Headers:***

| Key | Value | Description |
| --- | ------|-------------|
| Content-Type | application/json |  |


***Body:***

```js        
{
	"app": "{{APP_NAME}}",		// App to monitor
	"slack": "{{SLACK_CHANNEL}}",	// *Optional* slack channel to notify
	"email": "{{EMAIL}}",		// *Optional* email address to notify
	"post": "{{POST}}"		// *Optional* URL to post a webhook to
}
```


### 5. Delete Task

Stop monitoring an app for crash events

***Endpoint:***

```bash
Method: DELETE
URL: {{KAPACITOR_ALERTS_API}}/task/crashed/{{APP_NAME}}
```



## Memory

Send an alert to a Slack channel, email address, or as a webhook when an app uses more than the specified amount of memory.

### 1. Get All Tasks

Get the configuration of the release monitoring for all dynos on all apps

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/tasks/memory
```

### 2. Get All Tasks for App

Get the configuration of the release monitoring for all dynos on an app

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/tasks/memory/{{APP_NAME}}
```

### 3. Get Task

Get the configuration of the memory usage monitoring on an app and dyno

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/tasks/memory/{{APP_NAME}}/{{TASK_ID}}
```

### 4. Create Task

Begin monitoring an app for memory usage

***Endpoint:***

```bash
Method: POST
URL: {{KAPACITOR_ALERTS_API}}/task/memory
```


***Headers:***

| Key | Value | Description |
| --- | ------|-------------|
| Content-Type | application/json |  |


***Body:***

```js        
{
	"app": "{{APP_NAME}}",		// App to monitor
	"dynotype": "web",		// Dyno to monitor
	"warn": "200",			// Warning threshold (in MB)
	"crit": "500",			// Critical threshold (in MB)
	"window": "12h",		// Window to use for results
	"every": "1m",			// How often to check
	"slack": "{{SLACK_CHANNEL}}",	// *Optional* slack channel to notify
	"email": "{{EMAIL}}",		// *Optional* email address to notify
	"post": "{{POST}}"		// *Optional* URL to post a webhook to
}
```

### 5. Update Task

Update the configuration for memory usage monitoring on an app

***Endpoint:***

```bash
Method: PATCH
URL: {{KAPACITOR_ALERTS_API}}/task/memory
```

***Headers:***

| Key | Value | Description |
| --- | ------|-------------|
| Content-Type | application/json |  |


***Body:***

```js        
{
	"app": "{{APP_NAME}}",		// App to monitor
	"dynotype": "web",		// Dyno to monitor
	"warn": "500",			// Warning threshold (in MB)
	"crit": "700",			// Critical threshold (in MB)
	"window": "12h",		// Window to use for results
	"every": "1m",			// How often to check
	"slack": "{{SLACK_CHANNEL}}",	// *Optional* slack channel to notify
	"email": "{{EMAIL}}",		// *Optional* email address to notify
	"post": "{{POST}}"		// *Optional* URL to post a webhook to
}
```

### 6. Delete Task

Stop monitoring an app for memory usage

***Endpoint:***

```bash
Method: DELETE
URL: {{KAPACITOR_ALERTS_API}}/task/memory/{{APP_NAME}}/{{TASK_ID}}
```

***Headers:***

| Key | Value | Description |
| --- | ------|-------------|
| Content-Type | application/json |  |


## Release

Send an alert to a Slack channel, email address, or as a webhook when a new version of an app is released.

### 1. Get All Tasks

Get a list of the configuration of all release tasks

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/tasks/release
```

### 2. Get Task

Get the configuration of the release monitoring on an app

***Endpoint:***

```bash
Method: GET
URL: {{KAPACITOR_ALERTS_API}}/task/release/{{APP_NAME}}
```

### 3. Create Task

Begin monitoring an app for releases

***Endpoint:***

```bash
Method: POST
URL: {{KAPACITOR_ALERTS_API}}/task/release
```

***Headers:***

| Key | Value | Description |
| --- | ------|-------------|
| Content-Type | application/json |  |


***Body:***

```js        
{
	"app": "{{APP_NAME}}",		// App to monitor
	"slack": "{{SLACK_CHANNEL}}",	// *Optional* slack channel to notify
	"email": "{{EMAIL}}",		// *Optional* email address to notify
	"post": "{{POST}}"		// *Optional* URL to post a webhook to
}
```

### 4. Update Task

Update the configuration for new release monitoring on an app

***Endpoint:***

```bash
Method: PATCH
URL: {{KAPACITOR_ALERTS_API}}/task/release
```

***Headers:***

| Key | Value | Description |
| --- | ------|-------------|
| Content-Type | application/json |  |


***Body:***

```js        
{
	"app": "{{APP_NAME}}",		// App to monitor
	"slack": "{{SLACK_CHANNEL}}",	// *Optional* slack channel to notify
	"email": "{{EMAIL}}",		// *Optional* email address to notify
	"post": "{{POST}}"		// *Optional* URL to post a webhook to
}
```

### 5. Delete Task

Stop monitoring an app for releases

***Endpoint:***

```bash
Method: DELETE
URL: {{KAPACITOR_ALERTS_API}}/task/release/{{APP_NAME}}
```


---
[Back to top](#kapacitor-alerts-api)