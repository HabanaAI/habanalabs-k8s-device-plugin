# HABANA device plugin for Kubernetes

## Table of Contents

- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Existing Plugins](#existing-plugins)
  - [Goya device plugin](#Goya-device-plugin)
  - [Gaudi device plugin](#Gaudi-device-plugin)
- [Changelog](#changelog)
- [Issues](#issues)


## Introduction

The HABANA device plugin for Kubernetes is a Daemonset that allows you to automatically:
- Enables the registration of HABANA devices in your Kubernetes cluster.
- Keep track of the health of your Device

## Prerequisites
The list of prerequisites for running the HABANA device plugin is described below:
- HABANA drivers
- Kubernetes version >= 1.10

### The below sections detail existing plugins

#### Goya device plugin

Once you have enabled this option on *all* the nodes you wish to use,
you can then enable support in your cluster by deploying the following Daemonset:

```shell
$ kubectl create -f habanalabs-device-plugin.yaml
```

##### Running Jobs

Can now be consumed via container level resource requirements using the resource name habana.com/goya:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: habanalabs-goya-demo0
spec:
  nodeSelector:
    accelerator: habanalabs
  containers:
    - name: habana-ai-base-container
      image: habanai/goya-demo:0.9.1-43-debian9.8
      workingDir: /home/user1
      securityContext:
        capabilities:
          add: ["SYS_RAWIO"]
      command: ["sleep"]
      args: ["10000"]
      resources:
        limits:
          habana.ai/goya: 1
  imagePullSecrets:
    - name: regcred
```

#### Gaudi device plugin

Once you have enabled this option on *all* the nodes you wish to use,
you can then enable support in your cluster by deploying the following Daemonset:

```shell
$ kubectl create -f habanalabs-device-plugin-gaudi.yaml
```

### With Docker

#### Build
Option 1, pull the prebuilt image from [Docker Hub](https://hub.docker.com/r/habanai/k8s-device-plugin):
```shell
$ docker pull habanai/k8s-device-plugin:0.9.1
```

Option 2, build without cloning the repository:
```shell
$ docker build --network=host --no-cache -t habanai/k8s-device-plugin:0.9.1  habanalabs-k8s-device-plugin
```

Option 3, if you want to modify the code:
```shell
https://github.com/HabDevops/habanalabs-k8s-device-plugin
$ git clone https://github.com/HabDevops/habanalabs-k8s-device-plugin.git && cd habanalabs-k8s-device-plugin
$ git checkout v0.9.1
$ docker build -t habanai/k8s-device-plugin:0.9.1 .
```

#### Deploy as Daemon Set:
```shell
$ kubectl create -f habanalabs-device-plugin.yaml
```

## Changelog

### Version 0.9.1
- New HLML SW 0.9.1-43 debian9.8

### Version 0.8.1-beta1
- Support k8s plugin for Gaudi
- New HLML SW 0.8.1-55 debian9.8
- Add new resource namespace e.g: goya/gaudi 
- Refactor device plugin to eventually handle multiple resource types
- Move plugin error retry to event loop so we can exit with a signal

# Issues
* You can report a bug by [filing a new issue](https://github.com/HabDevops/habanalabs-k8s-device-plugin/issues/new)
* You can contribute by opening a [pull request](https://help.github.com/articles/using-pull-requests/)


