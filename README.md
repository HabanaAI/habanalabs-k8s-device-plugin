# Habana Device Plugin for Kubernetes

The Habana device plugin for Kubernetes, operating as a DaemonSet, enables the automatic registration of
Habana devices within your Kubernetes cluster, while also monitoring the health status of these devices.
This integration ensures seamless management and monitoring of Habana devices within the Kubernetes ecosystem,
enhancing operational efficiency and reliability.

This repository contains Habana official implementation of the [Kubernetes device plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/).

## Table of Contents
- [Habana Device Plugin for Kubernetes](#habana-device-plugin-for-kubernetes)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
  - [Gaudi Device Registration](#gaudi-device-registration)
  - [Building and Running Locally Using Docker](#building-and-running-locally-using-docker)


## Prerequisites

The below lists the prerequisites needed for running Habana device plugin:
- Habana Drivers
- Kubernetes version >= 1.19
- [Habana-container-runtime](https://github.com/HabanaAI/habana-container-runtime)


## Gaudi Device Registration

Once the prerequisites mentioned earlier have been established in the nodes,
you can then activate support in your cluster by deploying the Daemonset:

```shell
$ kubectl create -f habanalabs-device-plugin-gaudi.yaml
```


## Building and Running Locally Using Docker

To build and run using a docker, employ the following options according to your specific scenario: 

- To pull the prebuilt image, run:
```shell
$ docker pull vault.habana.ai/docker-k8s-device-plugin/docker-k8s-device-plugin:1.14.0
```

- To build without cloning the repository, run:
```shell
$ docker build -t vault.habana.ai/docker-k8s-device-plugin:devel -f Dockerfile https://github.com/HabanaAI/habanalabs-k8s-device-plugin.git#1.14.0
```

- To modify the code, run: 
```shell
$ git clone https://github.com/HabanaAI/habanalabs-k8s-device-plugin.git && cd habanalabs-k8s-device-plugin
$ docker build -t vault.habana.ai/docker-k8s-device-plugin:devel -f Dockerfile .
```
