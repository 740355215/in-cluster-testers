package apis

import (
	"cloud.tencent.com/tke/in-cluster-tester/testcase-operator/pkg/apis/tke/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1.SchemeBuilder.AddToScheme)
}
