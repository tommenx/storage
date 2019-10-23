package server

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/api"
	"github.com/tommenx/storage/pkg/api/types"
	"github.com/tommenx/storage/pkg/store"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sync"
)

var (
	errFailToRead  = restful.NewError(http.StatusBadRequest, "unable to read request body")
	errFailToWrite = restful.NewError(http.StatusInternalServerError, "unable to write response")
)

type server struct {
	exec api.Executor
	lock sync.Mutex
}

// StartServer starts a kubernetes scheduler extender http apiserver
func StartServer(kubeCli kubernetes.Interface, informerFactory informers.SharedInformerFactory, port int, db store.EtcdInterface) {
	e := api.NewExecutor(kubeCli, informerFactory, db)
	svr := &server{exec: e}

	stopCh := make(chan struct{})
	go informerFactory.Start(stopCh)
	e.Run(stopCh)
	ws := new(restful.WebService)
	ws.
		Path("/scheduler").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/setonepod").To(svr.setOnePod).
		Doc("set pod").
		Operation("SetPod").
		Writes(types.SetPodResult{}))

	ws.Route(ws.POST("/setbatchpod").To(svr.setBatchPod).
		Doc("set batch pod").
		Operation("SetBatchPod").
		Writes(types.SetPodResult{}))

	ws.Route(ws.GET("/hello").To(svr.hello).
		Doc("hello").
		Operation("Hello").
		Writes(types.HelloResult{}))
	ws.Route(ws.GET("/util").To(svr.getUtil).
		Doc("util").
		Operation("Util").
		Writes(types.GetInstanceResult{}))

	restful.Add(ws)
	glog.Infof("start scheduler extender server, listening on 0.0.0.0:%d", port)
	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func (svr *server) setBatchPod(req *restful.Request, resp *restful.Response) {
	svr.lock.Lock()
	defer svr.lock.Unlock()
	args := &types.SetBatchPodArgs{}
	if err := req.ReadEntity(args); err != nil {
		errorResponse(resp, errFailToRead)
		return
	}
	setPodResult, err := svr.exec.SetBatchPod(args)
	if err != nil {
		errorResponse(resp, restful.NewError(http.StatusInternalServerError,
			fmt.Sprintf("unable to filter nodes: %v", err)))
		return
	}

	if err := resp.WriteEntity(setPodResult); err != nil {
		errorResponse(resp, errFailToWrite)
	}
}
func (svr *server) getUtil(req *restful.Request, resp *restful.Response) {
	svr.lock.Lock()
	defer svr.lock.Unlock()
	getInstanceResult, err := svr.exec.GetInstanceUtil(&types.GetInstanceArgs{})
	if err != nil {
		errorResponse(resp, restful.NewError(http.StatusInternalServerError,
			fmt.Sprintf("unable to get instace storage util : %v", err)))
		return
	}
	if err := resp.WriteEntity(getInstanceResult); err != nil {
		errorResponse(resp, errFailToWrite)
	}
}

func (svr *server) setOnePod(req *restful.Request, resp *restful.Response) {
	svr.lock.Lock()
	defer svr.lock.Unlock()
	args := &types.SetOnePodArgs{}
	if err := req.ReadEntity(args); err != nil {
		errorResponse(resp, errFailToRead)
		return
	}
	setPodResult, err := svr.exec.SetOnePod(args)
	if err != nil {
		errorResponse(resp, restful.NewError(http.StatusInternalServerError,
			fmt.Sprintf("unable to filter nodes: %v", err)))
		return
	}

	if err := resp.WriteEntity(setPodResult); err != nil {
		errorResponse(resp, errFailToWrite)
	}
}

func (svr *server) hello(req *restful.Request, resp *restful.Response) {
	svr.lock.Lock()
	defer svr.lock.Unlock()
	helloResult := &types.HelloResult{
		Hello: "hello",
	}
	if err := resp.WriteEntity(helloResult); err != nil {
		errorResponse(resp, errFailToWrite)
	}
}

func errorResponse(resp *restful.Response, svcErr restful.ServiceError) {
	glog.Error(svcErr.Message)
	if writeErr := resp.WriteServiceError(svcErr.Code, svcErr); writeErr != nil {
		glog.Errorf("unable to write error: %v", writeErr)
	}
}
