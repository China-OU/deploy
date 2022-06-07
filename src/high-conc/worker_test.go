package high_conc
// 运行方式，cd 到当前目录，go test -v worker_test.go worker.go

import (
	"testing"
	"time"
	"fmt"
	"strconv"
)

func TestWoker(t *testing.T){
	for i := 0; i < 100; i++ {
		//job := SampleJob{
		//	ST: i,
		//}
		job := SampleJobWithTimeout{
			ST: i,
		}
		JobQueue <- &job
	}
	time.Sleep(110*time.Second)
	fmt.Println("结束")
}

type SampleJob struct {
	ST int
}

func (self *SampleJob) Do() {
	time.Sleep(time.Second * 10)
	now := time.Now().Format("15:04:05")
	fmt.Println(strconv.Itoa(self.ST) + " " + now)
}

type SampleJobWithTimeout struct {
	ST int
}

func (self *SampleJobWithTimeout) Do() {
	timeout := time.After(5 * time.Second)
	exec_result_ch := make(chan bool, 1)
	go func() {
		exec_result := DealFunc(self.ST)
		exec_result_ch <- exec_result
	}()
	select {
	case <-exec_result_ch:
		fmt.Println("执成成功")
	case <-timeout:
		fmt.Println("请求超时")
	}
}

func DealFunc(st int) bool {
	time.Sleep(time.Second * 10)
	//now := time.Now().Format("15:04:05")
	//fmt.Println(strconv.Itoa(st) + " " + now)
	return true
}