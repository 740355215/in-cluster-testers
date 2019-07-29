package controller

import (
	"cloud.tencent.com/tke/in-cluster-tester/testcase-operator/pkg/controller/testcase"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, testcase.Add)
}
