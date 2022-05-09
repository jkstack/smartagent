package exec

import (
	"encoding/json"

	"github.com/jkstack/anet"
)

func (ex *Executor) sendError(taskID, errmsg string) {
	var msg anet.Msg
	msg.Type = anet.TypeError
	msg.Important = true
	msg.TaskID = taskID
	msg.ErrorMsg = errmsg
	data, _ := json.Marshal(msg)
	ex.chWrite <- data
}
