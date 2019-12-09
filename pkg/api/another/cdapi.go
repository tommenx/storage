package another

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/api/types"
	"github.com/tommenx/storage/pkg/httputil"
	"net/http"
	"time"
)

var (
	qpsPrefix             = "qps"
	allocationPrefix      = "allocation"
	timeCompletionPrefix  = "timecompletion"
	settingPrefix         = "setting"
	instanceUseFreePrefix = "utilfree"
)

type cdClient struct {
	url        string
	httpClient *http.Client
}
type CDClient interface {
	GetQPS(which string) (*types.QueryResult, error)
	GetAllocation(which string) (*types.QueryResult, error)
	GetTimeCompletion(which string) (*types.ResourceCompletionResult, error)
	PutSetting(which, key, val string) (*types.SetPodResult, error)
	GetInstanceUseFree(which string) (*types.GetInstanceUseFreeResult, error)
}

func NewCDClient(url string, timeout time.Duration) CDClient {
	httpClient := &http.Client{Timeout: timeout}
	return &cdClient{
		url:        url,
		httpClient: httpClient,
	}
}

func (cd *cdClient) GetQPS(which string) (*types.QueryResult, error) {
	apiURL := fmt.Sprintf("%s/%s/%s", cd.url, qpsPrefix, which)
	body, err := httputil.GetBodyOK(cd.httpClient, apiURL)
	if err != nil {
		return nil, err
	}
	res := &types.QueryResult{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (cd *cdClient) GetAllocation(which string) (*types.QueryResult, error) {
	apiURL := fmt.Sprintf("%s/%s/%s", cd.url, allocationPrefix, which)
	body, err := httputil.GetBodyOK(cd.httpClient, apiURL)
	if err != nil {
		return nil, err
	}
	res := &types.QueryResult{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (cd *cdClient) GetTimeCompletion(which string) (*types.ResourceCompletionResult, error) {
	apiURL := fmt.Sprintf("%s/%s/%s", cd.url, timeCompletionPrefix, which)
	body, err := httputil.GetBodyOK(cd.httpClient, apiURL)
	if err != nil {
		return nil, err
	}
	res := &types.ResourceCompletionResult{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (cd *cdClient) GetInstanceUseFree(which string) (*types.GetInstanceUseFreeResult, error) {
	apiURL := fmt.Sprintf("%s/%s/%s", cd.url, instanceUseFreePrefix, which)
	body, err := httputil.GetBodyOK(cd.httpClient, apiURL)
	if err != nil {
		return nil, err
	}
	res := &types.GetInstanceUseFreeResult{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	return res, err
}

// http://10.77.110.131:8888/setting/old?key=qps&val=900
func (cd *cdClient) PutSetting(which, key, val string) (*types.SetPodResult, error) {
	apiURL := fmt.Sprintf("%s/%s/%s?key=%s&val=%s", cd.url, settingPrefix, which, key, val)
	glog.Infof("url is %s", apiURL)
	resp := &types.SetPodResult{}
	body, err := httputil.GetBodyOK(cd.httpClient, apiURL)
	if err != nil {
		resp.Code = 1
		glog.Errorf("error is %+v", err)
		return resp, err
	}

	err = json.Unmarshal(body, resp)
	if err != nil {
		resp.Code = 1
		glog.Errorf("error is %+v", err)
		return resp, err
	}
	return resp, nil
}
