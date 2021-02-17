module github.com/HabanaAI/habanalabs-k8s-device-plugin

go 1.15

require (
	github.com/HabanaAI/gohlml v1.3.0
	github.com/fsnotify/fsnotify v1.4.9
	google.golang.org/grpc v1.35.0
	k8s.io/kubelet v0.19.7
)

// uncomment below if developing with a local copy of gohlml
//replace github.com/HabanaAI/gohlml v1.3.0 => ./pkg/gohlml
