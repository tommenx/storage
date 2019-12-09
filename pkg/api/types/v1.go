package types

type SetOnePodArgs struct {
	Namespace string     `json:"namespace"`
	Requests  []Instance `json:"requests"`
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

type PutSettingArgs struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type QueryResult struct {
	Val     string `json:"val"`
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type ResourceCompletionResult struct {
	ResourceTime int    `json:"resource_time"`
	Completion   int    `json:"completion"`
	Code         int32  `json:"code"`
	Message      string `json:"message"`
}

type InstanceUseFree struct {
	Name string `json:"name"`
	Use  int    `json:"use"`
	Free int    `json:"free"`
}

type GetInstanceUseFreeResult struct {
	Instances []*InstanceUseFree `json:"instances"`
	Code      int32              `json:"code"`
	Message   string             `json:"message"`
}
