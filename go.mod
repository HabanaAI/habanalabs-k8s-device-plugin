module github.com/HabanaAI/habanalabs-k8s-device-plugin

go 1.21

toolchain go1.21.5

require (
	github.com/HabanaAI/gohlml v1.3.0
	github.com/fsnotify/fsnotify v1.4.9
	google.golang.org/grpc v1.35.0
	k8s.io/kubelet v0.19.7
)

require (
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/sys v0.0.0-20201112073958-5cba982894dd // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

// uncomment below if developing with a local copy of gohlml
replace github.com/HabanaAI/gohlml v1.3.0 => ../go-hlml
