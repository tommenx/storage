## Pym
## 目录结构
```bash
├── Gopkg.lock
├── Gopkg.toml
├── README.md
├── bin
├── build
├── cmd        实际程序入口
├── config.toml
├── deploy
├── hack
├── pkg        调用的入口
├── templates
└── vendor     第三方的包 
```
## 准备环境
1. kubernetes的版本在1.14及以上
2. 安装lvm插件
3. 创建名字为vgdata的卷组（VG）
## 编译指南
可以运行 `./hack/test.sh`脚本编译所有插件
## 部署指南
### CSI 辅助容器及创建storageclass
1. 先部署 `deploy/csi-plugin/`下的yaml文件创建CSI的许可以及自定义资源
2. 再部署`deploy/csi-plugin/lvm`下的yaml文件，创建辅助pod
### 定义新的CRD
1. 创建`deploy/crd.yaml`创建storagelabel自定义资源（历史遗留问题）
### 部署Pym资源管理系统
#### 部署ETCD
有两种方式部署ETCD
1. 在kubernetes中部署，以NodePort方式暴露服务，使用[ETCD-Operator](https://github.com/coreos/etcd-operator)
2. 在本机上部署
#### 在master上部署coordinator
```bash
# 把etcd后换成你的etcd服务地址
./coordinator --etcd 127.0.0.1:30085 --port 8888
```
#### 在每一台机器上部署csi-plugin
```bash
# coordinator后换成coordinator的ip
./csi-plugin --endpoint unix://var/lib/kubelet/plugins/lvmplugin.csi.alibabacloud.com/csi.sock --nodeid `hostname` --coordinator master:50051 &
```
#### 在worker节点（或每一台机器）部署catch-up
```bash
# coordinator后换成coordinator的ip
 ./catch-up --coordinator master:50051 --node `hostname` &
# 需要一个config文件在同级目录，参考 config.toml
# 根据hostname修改里面[node]的内容，其他暂时不用修改
```


