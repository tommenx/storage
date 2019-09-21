package types

type SetOnePodArgs struct {
	Namespace    string `json:"namespace"`
	Pod          string `json:"pod"`
	StorageLabel string `json:"storage_label"`
}

type SetBatchPodArgs struct {
	Tag          string `json:"tag"`
	Val          string `json:"val"`
	StorageLabel string `json:"storage_label"`
}

type SetPodResult struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type HelloResult struct {
	Hello string `json:"hello"`
}
