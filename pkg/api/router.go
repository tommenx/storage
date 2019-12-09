package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/api/another"
	"github.com/tommenx/storage/pkg/api/types"
	"github.com/tommenx/storage/pkg/store"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sync"
	"time"
)

type server struct {
	exec     Executor
	lock     sync.Mutex
	router   *gin.Engine
	port     int
	me       string
	cdClient another.CDClient
}

func NewServer(kubeCli kubernetes.Interface, informerFactory informers.SharedInformerFactory,
	port int, db store.EtcdInterface, me string, anotherURL string) *server {
	router := gin.Default()
	e := NewExecutor(kubeCli, informerFactory, db)
	cdClient := another.NewCDClient(fmt.Sprintf("http://%s", anotherURL), time.Second*5)
	svr := &server{
		exec:     e,
		router:   router,
		port:     port,
		cdClient: cdClient,
		me:       me,
	}
	stopCh := make(chan struct{})
	go informerFactory.Start(stopCh)
	e.Run(stopCh)
	router.LoadHTMLGlob("./templates/*")
	router.POST("/setonepod", svr.SetOnePod)
	router.POST("/setbatchpod", svr.SetBatchPod)
	router.GET("/allocation/:which", svr.GetResourceAllocation)
	router.GET("/time/:which", svr.GetResourceTime)
	router.GET("/util", svr.GetInstanceUtil)
	router.GET("/utilfree/:which", svr.GetInstanceUseFree)
	router.GET("/qps/:which", svr.GetQPS)
	router.GET("/requestqps/:which", svr.GetRequestQPS)
	router.GET("/timecompletion/:which", svr.GetResourceCompletion)
	router.GET("/setting/:which", svr.PutSetting)
	router.GET("/", svr.Index)
	router.POST("/test", svr.Test)
	return svr
}

func (s *server) SetOnePod(c *gin.Context) {
	args := &types.SetOnePodArgs{}
	c.MustBindWith(args, binding.JSON)
	res, _ := s.exec.SetOnePod(args)
	c.JSON(http.StatusOK, res)
}
func (s *server) Test(c *gin.Context) {
	args := &types.SetOnePodArgs{}
	c.MustBindWith(args, binding.JSON)
	glog.Infof("%+v", args)
	res := &types.SetPodResult{
		Code:    0,
		Message: "success",
	}
	c.JSON(http.StatusOK, res)
}

func (s *server) SetBatchPod(c *gin.Context) {
	args := &types.SetBatchPodArgs{}
	c.MustBindWith(args, binding.JSON)
	res, _ := s.exec.SetBatchPod(args)
	c.JSON(http.StatusOK, res)
}

func (s *server) GetInstanceUtil(c *gin.Context) {
	res, _ := s.exec.GetInstanceUtil()
	c.JSON(http.StatusOK, res)
}

func (s *server) GetResourceAllocation(c *gin.Context) {
	which := c.Param("which")
	var res *types.QueryResult
	if which == s.me {
		res, _ = s.exec.Query("resource_allocation", which)
	} else {
		res, _ = s.cdClient.GetAllocation(which)
	}
	c.JSON(http.StatusOK, res)
}

func (s *server) GetResourceTime(c *gin.Context) {
	which := c.Param("which")
	var res *types.QueryResult
	res, _ = s.exec.Query("resource_time", which)
	c.JSON(http.StatusOK, res)
}

func (s *server) GetQPS(c *gin.Context) {
	which := c.Param("which")
	var res *types.QueryResult
	if which == s.me {
		res, _ = s.exec.Query("qps", which)
	} else {
		res, _ = s.cdClient.GetQPS(which)
	}
	c.JSON(http.StatusOK, res)
}

func (s *server) GetRequestQPS(c *gin.Context) {
	which := c.Param("which")
	res, _ := s.exec.Query("request_qps", which)
	c.JSON(http.StatusOK, res)
}

func (s *server) GetResourceCompletion(c *gin.Context) {
	which := c.Param("which")
	var res *types.ResourceCompletionResult
	if which == s.me {
		res, _ = s.exec.GetResourceTimeCompletion(which)
	} else {
		res, _ = s.cdClient.GetTimeCompletion(which)
	}
	c.JSON(http.StatusOK, res)
}
func (s *server) PutSetting(c *gin.Context) {
	which := c.Param("which")
	var resp *types.SetPodResult
	key := c.Query("key")
	val := c.Query("val")
	if which == s.me {
		resp, _ = s.exec.PutSetting(key, val)
	} else {
		resp, _ = s.cdClient.PutSetting(which, key, val)
	}
	c.JSON(http.StatusOK, resp)

}

func (s *server) GetInstanceUseFree(c *gin.Context) {
	which := c.Param("which")
	var resp = &types.GetInstanceUseFreeResult{}
	if which == s.me {
		resp, _ = s.exec.GetInstanceUseFree()
	} else {
		resp, _ = s.cdClient.GetInstanceUseFree(which)
	}
	c.JSON(http.StatusOK, resp)
}

func (s *server) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", nil)
}

func (s *server) Run() {
	s.router.Run(fmt.Sprintf(":%d", s.port))
}
