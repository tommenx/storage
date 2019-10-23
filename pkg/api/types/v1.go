package types

type Request struct {
	Pod   string `json:"pod"`
	Read  string `json:"read"`
	Write string `json:"write"`
}

type SetOnePodArgs struct {
	Namespace string    `json:"namespace"`
	Requests  []Request `json:"requests"`
}

type SetBatchPodArgs struct {
	Tag   string `json:"tag"`
	Val   string `json:"val"`
	Read  string `json:"read"`
	Write string `json:"write"`
}

type SetPodResult struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type HelloResult struct {
	Hello string `json:"hello"`
}

type GetInstanceArgs struct {
}

type Instance struct {
	Name  string `json:"name"`
	Read  string `json:"read"`
	Write string `json:"write"`
}

type GetInstanceResult struct {
	Instances []Instance `json:"instances"`
	Code      int32      `json:"code"`
	Message   string     `json:"message"`
}
