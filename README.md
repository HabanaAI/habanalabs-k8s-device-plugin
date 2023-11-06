# HABANA device plugin for Kubernetes

## Table of Contents

- [HABANA device plugin for Kubernetes](#habana-device-plugin-for-kubernetes)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Prerequisites](#prerequisites)
    - [The below sections detail existing plugins](#the-below-sections-detail-existing-plugins)
      - [Goya device plugin](#goya-device-plugin)
        - [Running Jobs](#running-jobs)
      - [Gaudi device plugin](#gaudi-device-plugin)
    - [With Docker](#with-docker)
      - [Build](#build)
    - [Build in CD](#build-in-cd)
      - [Deploy as Daemon Set:](#deploy-as-daemon-set)
  - [Changelog](#changelog)
    - [Version 0.9.1](#version-091)
    - [Version 0.8.1-beta1](#version-081-beta1)
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

### Build in CD
Requirements:
- go-hlml repo must be first downloaded from gerrit into the habanalabs-device-plugin repo.
  It is copied by the Dockerfile into the image during the build process.
  _(this is due a lack of go modules support in Gerrit v2)_


To build the image in the CD process use the `make build` build.
It accepts the following parameters:
- `base_image` -Image to use as the builder for the application
- `image` - Final full image name to deploy
- `version` - Image's tag

Full example showing usage of current jenkins variables(or parameters):
```
make build base_image=$baseDockerImage image=$pluginDockerImage version=$"{release_version}-${release_build_id}"
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


