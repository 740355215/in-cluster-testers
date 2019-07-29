package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SummarySpec defines the desired state of Summary
type SummarySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// SummaryStatus defines the observed state of Summary
type SummaryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	TotalNumber  int `json:"total_number"`
	PassedNumber int `json:"passed_number"`
	FailedNumber int `json:"failed_number"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Summary is the Schema for the summaries API
// +k8s:openapi-gen=true
type Summary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SummarySpec   `json:"spec,omitempty"`
	Status SummaryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SummaryList contains a list of Summary
type SummaryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Summary `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Summary{}, &SummaryList{})
}
