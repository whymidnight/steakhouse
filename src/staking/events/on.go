package events

import (
	"encoding/json"
	"io/ioutil"
)

func (sub *Subscription) Commit() {
	fileBytes, err := json.Marshal(sub)
	if err != nil {
		panic(nil)
	}
	ioutil.WriteFile(sub.FilePath, fileBytes, 0755)
}

func (sub *Subscription) SetScheduled(sched bool) {
	sub.IsScheduled = sched
	sub.Commit()
}
func (sub *Subscription) SetProcessed(proc bool) {
	sub.IsProcessed = proc
	sub.Commit()
}
