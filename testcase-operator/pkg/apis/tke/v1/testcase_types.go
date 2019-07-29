package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TestCaseSpec defines the desired state of TestCase
type TestCaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Image    string   `json:"image"`
	Commands []string `json:"commands"`
}

// TestCaseStatus defines the observed state of TestCase
type TestCaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Result       string `json:"result"`
	Message      string `json:"message,omitempty"`
	CompleteTime string `json:"complete_time"`
	// ${NAMESPACE}/${NAME}
	PodName string `json:"pod_name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TestCase is the Schema for the testcases API
// +k8s:openapi-gen=true
type TestCase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TestCaseSpec   `json:"spec,omitempty"`
	Status TestCaseStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TestCaseList contains a list of TestCase
type TestCaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TestCase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TestCase{}, &TestCaseList{})
}
