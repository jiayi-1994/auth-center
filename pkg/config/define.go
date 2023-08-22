package config

const (
	ContainerReadOnly     = "auth:readonly"
	ContainerReadWrite    = "auth:readwrite"
	Uid                   = "uid"
	RbName                = "rb-auth-%s"
	FinalizerAuthCenter   = "authCenter.finalizers"
	FinalizerHarbor       = "harbor.finalizers"
	FinalizerNsAuthCenter = "jiayi.com/authcenter"

	HarborEmailFormat = "%s@authCenter.test.com"
)
