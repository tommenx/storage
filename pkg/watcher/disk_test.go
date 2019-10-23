package watcher

import (
	"github.com/tommenx/storage/pkg/config"
	"github.com/tommenx/storage/pkg/rpc"
	"testing"
)

func TestGetRemainingResource(t *testing.T) {
	_, err := GetRemainingResource("sda")
	if err != nil {
		t.Errorf("error is %+v", err)
	}
}

func TestReportRemainingResource(t *testing.T) {
	rpc.Init(":50051")
	config.Init("../../config.toml")
	err := ReportRemainingResource()
	if err != nil {
		t.Errorf("remaining resource error, err=%+v", err)
		return
	}
	t.Logf("success")
}

//func TestFormatIostatResult(t *testing.T) {
//	str := `Linux 3.10.0-957.5.1.el7.x86_64 (localhost.localdomain) 	10/18/2019 	_x86_64_	(4 CPU)
//
//avg-cpu:  %user   %nice %system %iowait  %steal   %idle
//           8.97    0.07    7.98    0.24    0.00   82.73
//
//Device:         rrqm/s   wrqm/s     r/s     w/s    rMB/s    wMB/s avgrq-sz avgqu-sz   await r_await w_await  svctm  %util
//sda               0.00     0.32    0.15   17.45     0.01     0.10    12.68     0.02    1.36    0.99    1.37   0.91   1.61
//sdb               0.00     0.33    0.00    0.01     0.00     0.00   693.89     0.00   93.23    2.66  127.49   2.82   0.00
//centos-root       0.00     0.00    0.15   17.19     0.01     0.10    12.85     0.03    1.47    1.02    1.48   0.93   1.61
//centos-swap       0.00     0.00    0.00    0.00     0.00     0.00    73.63     0.00    0.86    0.86    0.00   0.56   0.00
//centos-home       0.00     0.00    0.00    0.00     0.00     0.00    50.45     0.00    1.24    0.79    1.59   0.78   0.00
//dock-lvol0        0.00     0.00    0.00    0.00     0.00     0.00    48.19     0.00    1.52    1.52    0.00   1.46   0.00
//vgdata-pvc--5e183f9e--ecc6--11e9--9231--309c23e8d374     0.00     0.00    0.00    0.06     0.00     0.00    46.00     0.01  250.56    4.48  251.65   0.14   0.00
//vgdata-pvc--693bb6be--ecc6--11e9--9231--309c23e8d374     0.00     0.00    0.00    0.06     0.00     0.00    17.62     0.01  256.60    2.76  257.71   0.07   0.00
//vgdata-pvc--35b0db63--ef41--11e9--9231--309c23e8d374     0.00     0.00    0.00    0.06     0.00     0.00    17.62     0.03  534.67    1.84  536.98   0.10   0.00
//vgdata-pvc--414cb127--ef41--11e9--9231--309c23e8d374     0.00     0.00    0.00    0.06     0.00     0.00    17.63     0.02  275.19    2.00  276.29   0.05   0.00`
//	data := formatIostatResult(str)
//	for key, usage := range data {
//		t.Logf("%s : read %s write %s \n", key, usage[0], usage[1])
//	}
//}
