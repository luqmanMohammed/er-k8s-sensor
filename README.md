# EventsRunner-K8s-Sensor <!-- omit in toc -->
[![Go Report Card](https://goreportcard.com/badge/github.com/luqmanMohammed/eventsrunner-k8s-sensor)](https://goreportcard.com/report/github.com/luqmanMohammed/eventsrunner-k8s-sensor)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/luqmanMohammed/eventsrunner-k8s-sensor/CodeQL?label=CodeQL&logo=Github)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/luqmanMohammed/eventsrunner-k8s-sensor/Build%20and%20Test?logo=Github)

Config driven sensor for Kubernetes events.

---
## Table of Contents <!-- omit in toc -->

- [1. What is it?](#1-what-is-it)
- [2. Usage](#2-usage)
- [3. Configuration](#3-configuration)
  - [3.1. Sensor configuration](#31-sensor-configuration)
  - [3.2. Rules](#32-rules)
    - [3.2.1. Rule collection from ConfigMaps](#321-rule-collection-from-configmaps)
- [4. Executor (Processor/Action)](#4-executor-processoraction)
  - [4.1. log executor](#41-log-executor)
  - [4.2. script executor](#42-script-executor)
  - [4.3. eventsrunner executor](#43-eventsrunner-executor)
    - [4.3.1. mTLS authentication](#431-mtls-authentication)
    - [4.3.2. JWT authentication](#432-jwt-authentication)
- [5. License](#5-license)

---
## 1. What is it?
[Events Runner](https://github.com/luqmanMohammed/eventsrunner) is a config-driven automation server that can be used to do event-driven automation where it follows the sensor-runner architecture. Sensors are intended to send events from several sources to the server utilizing its Rest API. Events Runner will process the received events and trigger actions based on configs.

This project implements a config-driven sensor to listen to Kubernetes events and trigger actions. The sensor is implemented in Go and uses the [k8s.io/client-go](https://pkg.go.dev/k8s.io/client-go) library to interface with Kubernetes. The sensor can be configured using the Rules construct to define what k8s events to listen to. The actions taken by the sensor can be either forwarding the event to Events Runner or running a script.

---

## 2. Usage
## 3. Configuration

### 3.1. Sensor configuration
>[Viper](https://github.com/spf13/viper) is used to manage sensor configuration.

Sensor configuration will be collected from the following > sources in the following order:

1. Defaults
2. Environment variables
    - Environment variables must have the prefix `ER_K8S_SENSOR_`
3. Configuration file
    - Configuration file must be named `config.yaml`
    - Config file will be loaded from the following locations in the following order.
        1. /etc/er-k8s-sensor/config.yaml
        2. $HOME/.er-k8s-sensor/config.yaml
    - Config file location can be explicitly specified using the `--config` flag.
    - Configuration file must be in YAML format

Configs used by the sensor are:

|         **Key**          | **Type** |                                                                **Description**                                                                 |       **Default**        |
| :----------------------: | :------: | :--------------------------------------------------------------------------------------------------------------------------------------------: | :----------------------: |
|       LogVerbosity       |   int    |                                          Log verbosity to be used. Can be set using the `-v` flag too                                          |           None           |
|        SensorName        |  string  |              Name of the sensor, used to ignore specific objects from being processed by labeling them with `<SensorName>=ignore`              |      er-k8s-sensor       |
|      KubeConfigPath      |  string  |                                                       Absolute path to Kubernetes config                                                       |       $HOME/.kube        |
|     SensorNamespace      |  string  |                                                  Rules will be collected from this namespace                                                   |       eventsrunner       |
| SensorRuleConfigMapLabel |  string  |                                           Label to identify Rules configmaps in the SensorNamespace                                            | er-k8s-sensor-rules=true |
|       WorkerCount        |   int    |                                         Number of workers to be started to process items in the queue                                          |            10            |
|       MaxTryCount        |   int    |                                           Number of times failed events executions should be retried                                           |            5             |
|       RequeueDelay       |   time   |                                                  Time to wait before requeuing a failed event                                                  |           30s            |
|       ExecutorType       |  string  |                Type of executor to be used to process events in the queue. Valid options are `log`, `script` or `eventsrunner`                 |           log            |
|        ScriptDir         |  string  |                                    If script executor is used, Absolute path to the location of the scripts                                    |           None           |
|       ScriptPrefix       |  string  |                                        If script executor is used, Prefix that all scripts should have                                         |           None           |
|         AuthType         |  string  | If eventsrunner executor is used, Type of authentication to use when connecting to the Events Runner server. Valid options are `jwt` or `mTLS` |           None           |
|   EventsRunnerBaseURL    |  string  |                                       If eventsrunner executor is used, URL of the Events Runner server                                        |           None           |
|      RequestTimeout      |   time   |                      If eventsrunner executor is used, Request timeout to be configured when calling Events Runner server                      |           None           |
|        CaCertPath        |  string  |                If eventsrunner executor is used and auth type mTLS and/or HTTPS endpoint is used, Absolute path to the CA cert                 |           None           |
|         JWTToken         |  string  |                 If eventsrunner executor is used and auth type JWT is used, JWT token to be added in the authorization header                  |           None           |
|      ClientCertPath      |  string  |                         If eventsrunner executor is used and auth type mTLS  is used, Absolute path to the client cert                         |           None           |
|      ClientKeyPath       |  string  |                          If eventsrunner executor is used and auth type mTLS is used, Absolute path to the client key                          |           None           |

Example config file to be provided to the sensor when using the script executor (Refer above table to see all supported configs):
```yaml
---
sensorNamespace: "eventsrunner"
executorType: "script"
scriptDir:    "/tmp/test-scripts"
scriptPrefix: "script"
```

### 3.2. Rules
Rules are defined in the JSON format.

```json
{
    "id": "cm-rule-1",
    "group": "",
    "version": "v1",
    "resource": "configmaps",
    "namespaces": ["default"],
    "eventTypes": ["ADDED", "MODIFIED"],
    "updatesOn": ["data"],
    "fieldFilter": "metadata.name=test",
    "labelFilter": "testKey=testValue",
}
```
- `id` (Required): Identifier of the rule. **Should be unique across the entire sensor**. If multiple have the same id, the last one will be used. If an id is not provided, the rule will be ignored. The above example rule has the id `cm-rule-1`.
- `group` (Required): Group of the resource to be watched. The above example rule has the group `` since ConfigMap is part of the core API group.
- `version` (Required): API Version of the resource to be watched. Example rule has the version `v1` since ConfigMaps are part of APIVersion v1.
- `resource` (Required): Resource to be watched. The above example rule is set to watch the resource `configmaps`.
- `namespace` (Optional): List of namespaces where the resource should be watched. Remove namespaces for cluster-wide resources. If empty for namespace-bound resources, the sensor will watch resources in all namespaces. The above example rule has the namespaces `default`.
- `eventTypes` (Optional): List of event types to be considered. If empty, all event types will be considered. Valid options are `ADDED`, `MODIFIED`, `DELETED`.  The above example rule has the event types `ADDED` and `MODIFIED` configured.
- `updatesOn` (Optional): List of fields to be considered for updates. If empty, all fields will be considered. The above example shows that the sensor will process the event only if the `data` field is updated.
- `fieldFilter` (Optional): Filter to be applied on the field. If empty, no filter will be applied. All valid Kubernetes field selectors are supported. The above example shows that the sensor will consider the event if the `metadata.name` field is `test`
- `labelFilter` (Optional): Filter to be applied on the label. If empty, no filter will be applied. All valid Kubernetes label selectors are supported. The above example shows that the sensor will consider the event if the `testKey` label is present and the value is `testValue`

#### 3.2.1. Rule collection from ConfigMaps
Kubernetes Sensor supports loading rules from ConfigMaps. ConfigMaps in the `SensorNamespace` ([configurable](#31-sensor-configuration)) with the label `SensorRuleConfigMapLabel` ([configurable](#31-sensor-configuration)) will be considered as rules ConfigMaps. Rule ConfigMap's data field should have the key `rules` and the value should be a JSON array of rules JSON objects as depicted above. The sensor will automatically reload the affected rules if any of the rule ConfigMaps are updated.

---

## 4. Executor (Processor/Action)
Executor is the component that will be used to process the events. Event structure as follows:
```json
{
    "eventType": "added",
    "ruleID": "cm-rule-1",
    "objects": [
        {
            "kind": "ConfigMap",
        }
    ]
}
```
- eventType: Type of the event
- ruleID: Identifier of the rule that was triggered
- objects: List of objects that triggered the event
    - ADDED: The object that was added
    - MODIFIED: Pre-update object will be stored at index 0 and post-update object at index 1
    - DELETED: The object that was deleted

Specific executor can be configured using the `ExecutorType` ([configurable](#31-sensor-configuration)) config.

### 4.1. log executor
Log executor/action is the simplest executor which would log the rule's id. Intended to be used for debugging/testing/POC purposes.

### 4.2. script executor
> :warning: Scripts executed by the sensor should be vetted before allowing the sensor to run it.

> Required Configs: `ScriptDir` and `ScriptPrefix`


Script executor can be used to execute a script on the event inside the sensor environment. Scripts should be located in the `ScriptDir` ([configurable](#31-sensor-configuration)) directory. Scripts should have the following naming convention: `<ScriptPrefix>-<Rule.ID>.sh`. If the script is not valid nor executable, the execution will return an error. The relevant event will be passed to the script as an environment variable named `EVENT` which would be a base64 encoded JSON object. The executor would consider the execution a success; if the exit code is 0.

### 4.3. eventsrunner executor
> NOTE: [Events Runner](https://github.com/luqmanMohammed/eventsrunner) is not ready for any use yet. Watch for updates.

> Required Configs: `AuthType` and `EventsRunnerBaseURL`


> Optional Configs: `RequestTimeout`


Events Runner executor can be used to forward the events to the Events Runner server to be processed. Executor supports both mTLS and JWT authentication methodologies to authenticate with the Events Runner server. Events are forwarded to the following endpoint of the Events Runner server: `<EventsRunnerBaseURL>/api/v1/events`.

Authentication methodology can be configured using the `AuthType` ([configurable](#31-sensor-configuration)) config.

#### 4.3.1. mTLS authentication
> Required Configs: `CaCertPath`, `ClientCertPath` and `ClientKeyPath`


mTLS authentication is used to authenticate with the Events Runner server. The sensor will use the provided client cert and key to authenticate with the Events Runner server. The sensor will use the provided CA cert to validate the server's certificate.

#### 4.3.2. JWT authentication
> Required Configs: `JWTToken`


JWT authentication is used to authenticate with the Events Runner server. The sensor will use the provided JWT token to authenticate with the Events Runner server. Token will be added as a Bearer token in the request Authorization header. The sensor will utilize the CA cert provided via `CACertPath` to validate the server's certificate if it's exposing a TLS enabled endpoint.

---

## 5. License
This software is licensed under the MIT license.