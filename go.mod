module habanai/habanalabs-k8s-device-plugin

go 1.15

require (
	github.com/HabanaAI/gohlml v1.2.1
	github.com/fsnotify/fsnotify v1.4.9
	google.golang.org/grpc v1.34.0
	k8s.io/api v0.19.3 // indirect
	k8s.io/kubelet v0.19.3
)

replace (
	k8s.io/kubelet => k8s.io/kubelet v0.19.3
)
