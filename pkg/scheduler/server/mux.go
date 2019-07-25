package server

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/scheduler"
	schedulerapiv1 "k8s.io/kubernetes/pkg/scheduler/api/v1"
	"net/http"
	"sync"
)

var (
	errFailToRead  = restful.NewError(http.StatusBadRequest, "unable to read request body")
	errFailToWrite = restful.NewError(http.StatusInternalServerError, "unable to write response")
)

type server struct {
	scheduler scheduler.Scheduler
	lock      sync.Mutex
}

// StartServer starts a kubernetes scheduler extender http apiserver
func StartServer(port int) {
	s := scheduler.NewScheduler()
	svr := &server{scheduler: s}

	ws := new(restful.WebService)
	ws.
		Path("/scheduler").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/filter").To(svr.filterNode).
		Doc("filter nodes").
		Operation("filterNodes").
		Writes(schedulerapiv1.ExtenderFilterResult{}))

	ws.Route(ws.POST("/prioritize").To(svr.prioritizeNode).
		Doc("prioritize nodes").
		Operation("prioritizeNodes").
		Writes(schedulerapiv1.HostPriorityList{}))

	ws.Route(ws.POST("/bind").To(svr.bindNode).
		Doc("bind node").
		Operation("bind Node").
		Writes(schedulerapiv1.ExtenderBindingResult{}))

	restful.Add(ws)

	glog.Infof("start scheduler extender server, listening on 0.0.0.0:%d", port)
	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func (svr *server) filterNode(req *restful.Request, resp *restful.Response) {
	svr.lock.Lock()
	defer svr.lock.Unlock()

	args := &schedulerapiv1.ExtenderArgs{}
	if err := req.ReadEntity(args); err != nil {
		errorResponse(resp, errFailToRead)
		return
	}

	filterResult, err := svr.scheduler.Filter(args)
	if err != nil {
		errorResponse(resp, restful.NewError(http.StatusInternalServerError,
			fmt.Sprintf("unable to filter nodes: %v", err)))
		return
	}

	if err := resp.WriteEntity(filterResult); err != nil {
		errorResponse(resp, errFailToWrite)
	}
}

func (svr *server) prioritizeNode(req *restful.Request, resp *restful.Response) {
	args := &schedulerapiv1.ExtenderArgs{}
	if err := req.ReadEntity(args); err != nil {
		errorResponse(resp, errFailToRead)
		return
	}

	priorityResult, err := svr.scheduler.Priority(args)
	if err != nil {
		errorResponse(resp, restful.NewError(http.StatusInternalServerError,
			fmt.Sprintf("unable to priority nodes: %v", err)))
		return
	}

	if err := resp.WriteEntity(priorityResult); err != nil {
		errorResponse(resp, errFailToWrite)
	}
}

func (svr *server) bindNode(req *restful.Request, resp *restful.Response) {
	args := schedulerapiv1.ExtenderBindingArgs{}
	if err := req.ReadEntity(args); err != nil {
		errorResponse(resp, errFailToRead)
		return
	}
	bindResult, err := svr.scheduler.Bind(args)
	if err != nil {
		errorResponse(resp, restful.NewError(http.StatusInternalServerError,
			fmt.Sprintf("unable to bind nodes: %v", err)))
		return
	}
	if err := resp.WriteEntity(bindResult); err != nil {
		errorResponse(resp, errFailToWrite)
	}
}

func errorResponse(resp *restful.Response, svcErr restful.ServiceError) {
	glog.Error(svcErr.Message)
	if writeErr := resp.WriteServiceError(svcErr.Code, svcErr); writeErr != nil {
		glog.Errorf("unable to write error: %v", writeErr)
	}
}
