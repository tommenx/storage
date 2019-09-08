package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TidbCluster is the control script's spec
type StorageLabel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	// Spec defines the behavior of a tidb cluster
	Spec StorageLabelSpec `json:"spec"`
}

type StorageLabelSpec struct {
	ReadBpsDevice   int64 `json:"read_bps_device,omitempty"`
	ReadIopsDevice  int64 `json:"read_iops_device,omitempty"`
	WriteBpsDevice  int64 `json:"write_bps_device,omitempty"`
	WriteIopsDevice int64 `json:"write_iops_device,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StorageLabelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StorageLabel `json:"items"`
}
