package server

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/api"
	"github.com/tommenx/storage/pkg/api/types"
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
	exec api.Exec
	lock sync.Mutex
}

// StartServer starts a kubernetes scheduler extender http apiserver
func StartServer(kubeCli kubernetes.Interface, informerFactory informers.SharedInformerFactory, port int) {
	e := api.NewExec(kubeCli, informerFactory)
	svr := &server{exec: e}

	stopCh := make(chan struct{})
	go informerFactory.Start(stopCh)
	e.Run(stopCh)
	ws := new(restful.WebService)
	ws.
		Path("/scheduler").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/setpod").To(svr.setPod).
		Doc("set pod").
		Operation("SetPod").
		Writes(types.SetPodResult{}))

	ws.Route(ws.GET("/hello").To(svr.hello).
		Doc("hello").
		Operation("Hello").
		Writes(types.HelloResult{}))

	restful.Add(ws)
	fmt.Println("aaaaaa")

	glog.Infof("start scheduler extender server, listening on 0.0.0.0:%d", port)
	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func (svr *server) setPod(req *restful.Request, resp *restful.Response) {

	args := &types.SetPodArgs{}
	if err := req.ReadEntity(args); err != nil {
		errorResponse(resp, errFailToRead)
		return
	}
	setPodResult, err := svr.exec.SetPod(args)
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
