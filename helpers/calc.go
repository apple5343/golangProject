package helpers

import (
	"fmt"
	"server/config"
	"server/server/db"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
)

type Calculator struct {
	toProcess chan *Subtask
	db        db.SqlDB
	Worker    Worker
}

type Worker struct {
	mu        sync.Mutex
	list      []map[string]interface{}
	Delays    map[string]int
	toProcess chan *Subtask
}

func NewCalculator(cfg *config.Config, db db.SqlDB) (*Calculator, error) {
	toProcess := make(chan *Subtask)
	list := []map[string]interface{}{}
	calculator := &Calculator{toProcess: toProcess, db: db}
	for i := 0; i < cfg.Workers; i++ {
		killCh := make(chan int)
		list = append(list, map[string]interface{}{"id": i + 1, "kill": killCh, "status": "in waiting", "expression": "", "expressionId": 0})
		go calculator.Worker.worker(toProcess, i+1, killCh)
	}
	delays, err := db.GetDelays()
	if err != nil {
		return calculator, err
	}
	calculator.Worker.Delays = delays
	calculator.Worker.list = list
	calculator.Worker.toProcess = toProcess
	return calculator, nil
}

func (w *Worker) ChangeStatus(id int, newStatus string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, k := range w.list {
		if k["id"] == id {
			k["status"] = newStatus
			return
		}
	}
}

func (w *Worker) ChangeExpression(id int, expression string, expressionId int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, k := range w.list {
		if k["id"] == id {
			k["expression"] = expression
			k["expressionId"] = expressionId
			return
		}
	}
}

func (w *Worker) GetWorkersInfo() []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, v := range w.list {
		result = append(result, map[string]interface{}{"id": v["id"], "status": v["status"], "expression": v["expression"], "expressionId": v["expressionId"]})
	}
	return result
}

func (w *Worker) RemoveWorker(id int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	workers := []map[string]interface{}{}
	for _, v := range w.list {
		if v["id"] != id {
			workers = append(workers, v)
		} else {
			v["kill"].(chan int) <- 0
		}
	}
	w.list = workers
	if len(w.list)-1 == 1 {
		i := w.list[len(w.list)-1][`id`].(int) + 1
		killCh := make(chan int)
		w.list = append(w.list, map[string]interface{}{"id": i, "kill": killCh, "status": "in waiting", "expression": "", "expressionId": 0})
		go w.worker(w.toProcess, i, killCh)
	}
}

type Task struct {
	db         db.SqlDB
	subtask    *Spliter
	toProcess  chan *Subtask
	Id         int
	Status     string
	Expression string
	Result     string
	Created    time.Time
	LastPing   time.Time
}

type Subtask struct {
	substack *Symbol
	pingCh   chan int
	wg       *sync.WaitGroup
	taskId   int
}

func (w *Worker) worker(in chan *Subtask, id int, killCh <-chan int) {
	for {
		select {
		case v := <-in:
			w.ChangeStatus(id, "at work")
			expression, _ := govaluate.NewEvaluableExpression(v.substack.value)
			w.ChangeExpression(id, v.substack.value, v.taskId)
			select {
			case <-time.After(time.Duration(w.Delays[v.substack.op]) * time.Second):
				result, _ := expression.Eval(nil)
				v.substack.result = strconv.FormatFloat(result.(float64), 'f', -1, 64)
				v.pingCh <- v.substack.id
				v.wg.Done()
				w.ChangeStatus(id, "in waiting")
				w.ChangeExpression(id, "", 0)
			case <-killCh:
				in <- v
				return
			}
		case <-killCh:
			return
		}
	}
}

func (c *Calculator) NewTask(expression string) (*Task, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	if !IsValidExpression(expression) {
		return nil, fmt.Errorf("выражение недопустимо")
	}
	spliter := NewSpliter(expression)
	task := &Task{subtask: spliter, toProcess: c.toProcess, Expression: expression, db: c.db, Created: time.Now()}
	return task, nil
}

func (t *Task) Start() {
	wg := &sync.WaitGroup{}
	for !t.subtask.done {
		t.subtask.Split()
		resultsCh := make(chan int)
		for _, v := range t.subtask.Symbols {
			if v.expressionType == "calculation" {
				go func(s *Symbol) {
					t.toProcess <- &Subtask{substack: s, pingCh: resultsCh, wg: wg, taskId: t.Id}
				}(v)
				wg.Add(1)
			}
		}
		go func() {
			wg.Wait()
			close(resultsCh)
		}()
		processed := []int{}
		for i := range resultsCh {
			processed = append(processed, i)
			updated := ""
			old := ""
			lastStep := ""
			for _, v := range t.subtask.Symbols {
				if v.id == i {
					i = 99999999999
					old += `<span class=old>` + v.value + `</span>`
					updated += `<span class=new>` + v.result + `</span>`
					lastStep += v.result
					continue
				}
				if v.result != "" && slices.IndexFunc(processed, func(i int) bool { return i == v.id }) >= 0 {
					updated += v.result
					old += v.result
					lastStep += v.result
				} else {
					lastStep += v.value
					updated += v.value
					old += v.value
				}
			}
			err := t.db.UpdatePing(t.Id, time.Now())
			if err != nil {
				fmt.Println(err)
			}
			err = t.db.AddSubtask(old, time.Now(), t.Id, updated)
			if err != nil {
				fmt.Println(err)
			}
			err = t.db.UpdateLastStep(t.Id, lastStep)
			if err != nil {
				fmt.Println(err)
			}
		}
		newExpression := ""
		for _, v := range t.subtask.Symbols {
			if v.result != "" {
				newExpression += v.result
			} else {
				newExpression += v.value
			}
		}
		t.subtask.Update(newExpression)
	}
	t.Result = t.subtask.result
	t.db.SetResult(t.Id, t.Result)
	t.db.SetStatus(t.Id, "completed")
}

func (c *Calculator) ContinueCalculations() {
	tasks, err := c.db.GetInterruptedTasks()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range tasks {
		lastStep, _ := c.db.GetTaskById(int(v["id"].(int64)))
		spliter := NewSpliter(lastStep["lastStep"].(string))
		t, err := time.Parse("2006-01-02 15:04:05", v["created"].(string))
		if err != nil {
			fmt.Println(err)
		}
		task := &Task{subtask: spliter, toProcess: c.toProcess, Expression: lastStep["lastStep"].(string), db: c.db, Created: t, Id: int(v["id"].(int64))}
		go task.Start()
	}
}
