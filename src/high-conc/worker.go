package high_conc

type Job interface {
	Do()
}

var JobQueue chan Job

func init() {
	JobQueue = make(chan Job, 40)
	dispatcher := NewDispatcher(20)
	dispatcher.Run()
}

type Woker struct {
	WokerPool  chan chan Job
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job) Woker {
	return Woker{
		WokerPool:  workerPool,
		JobChannel: make(chan Job, 1),
		quit:       make(chan bool, 1),
	}
}

func (w Woker) Start() {
	go func() {
		for {
			w.WokerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				job.Do()
			case <-w.quit:
				return
			}
		}
	}()
}

func (w Woker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type Dispatcher struct {
	WorkerPool chan chan Job
	maxWorkers int
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{
		WorkerPool: pool,
		maxWorkers: maxWorkers,
	}
}

func (d *Dispatcher) Dispatch() {
	for {
		select {
		case job := <-JobQueue:
			go func(job Job) {
				jobChannel := <-d.WorkerPool
				jobChannel <- job
			}(job)
		}
	}
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}
	go d.Dispatch()
}

