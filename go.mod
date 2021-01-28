module github.com/HabanaAI/habanalabs-k8s-device-plugin

go 1.15

replace github.com/HabanaAI/gohlml => ./vendor/github.com/HabanaAI/gohlml

require (
	github.com/fsnotify/fsnotify v1.4.9
	google.golang.org/grpc v1.35.0
	k8s.io/kubelet v0.20.2
)
