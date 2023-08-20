package config

const (
	ContainerReadOnly     = "auth:readonly"
	ContainerReadWrite    = "auth:readwrite"
	Uid                   = "uid"
	RbName                = "rb-auth-%s"
	FinalizerContainer    = "container.finalizers"
	FinalizerHarbor       = "harbor.finalizers"
	FinalizerNsAuthCenter = "jiayi.com/authcenter"
)
