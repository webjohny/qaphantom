package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type ConfigExtra struct {
	DeepPaa bool `json:"deep_paa"`
	RedirectMethod bool `json:"redirect_method"`
	FastParsing bool `json:"fast_parsing"`
	CountStreams int `json:"count_streams"`
	LimitStreams int `json:"limit_streams"`
	CmdStreams string `json:"cmd_streams"`
}

func (t *MysqlFreeTask) MergeTask(task MysqlTask) {
	t.Id = int(task.Id.Int64)
	t.Keyword = task.Keyword.String
	t.CatId = int(task.CatId.Int64)
	t.Cat = task.Cat.String
	t.TryCount = int(task.TryCount.Int32)
}

func (t *MysqlFreeTask) MergeSite(site MysqlSite){
	t.SiteId = int(site.Id.Int64)
	t.Language = site.Language.String
	t.Theme = site.Theme.String
	t.Domain = site.Domain.String
	t.Login = site.Login.String
	t.Password = site.Password.String
	t.From = int(site.From.Int64)
	t.To = int(site.To.Int64)
	t.QstsLimit = int(site.QstsLimit.Int64)
	t.Linking = int(site.Linking.Int64)
	t.Header = int(site.Header.Int64)
	t.SubHeaders = int(site.SubHeaders.Int64)
	t.ParseDates = int(site.ParseDates.Int64)
	t.ParseDoubles = int(site.ParseDoubles.Int64)
	t.PubImage = int(site.PubImage.Int64)
	t.VideoStep = int(site.VideoStep.Int64)
	t.QaCountFrom = int(site.QaCountFrom.Int32)
	t.QaCountTo = int(site.QaCountTo.Int32)
	t.ParseFast = int(site.ParseFast.Int32)
	t.ParseSearch4 = int(site.ParseSearch4.Int32)
	t.ImageKey = int(site.ImageKey.Int64)
	t.H1 = int(site.H1.Int32)
	t.ShOrder = int(site.ShOrder.Int32)
	t.ShFormat = int(site.ShFormat.Int32)
	t.ImageSource = int(site.ImageSource.Int64)
	t.CountRows = int(site.CountRows.Int64)
	t.MoreTags = site.MoreTags.String
	t.SymbMicroMarking = site.SymbMicroMarking.String
	t.Extra = ConfigExtra{}

	var extra map[string]interface{}
	_ = json.Unmarshal([]byte(site.Extra.String), &extra)
	if v, ok := extra["deep_paa"] ; ok {
		t.Extra.DeepPaa = v.(bool)
	}
	if v, ok := extra["redirect_method"] ; ok {
		t.Extra.RedirectMethod = v.(bool)
	}
	if v, ok := extra["count_streams"] ; ok {
		t.Extra.CountStreams = v.(int)
	}
	if v, ok := extra["limit_streams"] ; ok {
		t.Extra.LimitStreams = v.(int)
	}
	if v, ok := extra["cmd_streams"] ; ok {
		t.Extra.CmdStreams = v.(string)
	}
}

func (t *MysqlFreeTask) SetFinished(status int, errorMsg string) {
	now := time.Now()
	formattedDate := now.Format("2006-01-02 15:04:05")

	lastLog := ""
	if len(t.Log) > 0 {
		lastLog = t.Log[len(t.Log)-1]
	}

	data := map[string]interface{}{}
	data["status"] = strconv.Itoa(status)
	data["log"] = strings.Join(t.Log, "\n")
	data["log_last"] = lastLog
	data["error"] = errorMsg
	data["parser"] = "NULL"
	data["timeout"] = "NULL"
	data["parse_date"] = formattedDate

	_, err := MYSQL.UpdateTask(data, t.Id)
	if err != nil {
		log.Println("MysqlFreeTask.SetFinished.HasError", err)
	}
}

func (t *MysqlFreeTask) FreeTask() {
	lastLog := ""
	if len(t.Log) > 0 {
		lastLog = t.Log[len(t.Log)-1]
	}

	if t.TryCount > 0 {
		t.TryCount -= 1
	}

	data := map[string]interface{}{}
	data["log"] = strings.Join(t.Log, "\n")
	data["log_last"] = lastLog
	data["parser"] = "NULL"
	data["timeout"] = "NULL"
	data["try_count"] = t.TryCount

	_, err := MYSQL.UpdateTask(data, t.Id)
	if err != nil {
		log.Println("MysqlFreeTask.FreeTask.HasError", err)
	}
}

func (t *MysqlFreeTask) SetTimeout(parser int) {
	now := time.Now().Local().Add(time.Minute * time.Duration(5))
	formattedDate := now.Format("2006-01-02 15:04:05")

	lastLog := ""
	if len(t.Log) > 0 {
		lastLog = t.Log[len(t.Log)-1]
	}

	data := map[string]interface{}{}
	data["log"] = strings.Join(t.Log, "\n")
	data["log_last"] = lastLog
	data["parser"] = strconv.Itoa(parser)
	data["timeout"] = formattedDate

	_, err := MYSQL.UpdateTask(data, t.Id)
	if err != nil {
		log.Println("MysqlFreeTask.SetTimeout.HasError", err)
	}
}

func (t *MysqlFreeTask) SetError(error string) {
	if error == "" {
		return
	}
	now := time.Now().Local().Add(time.Minute * time.Duration(5))
	formattedDate := now.Format("2006-01-02 15:04:05")
	t.SetLog(error)

	data := map[string]interface{}{}
	data["log"] = strings.Join(t.Log, "\n")
	data["log_last"] = error
	data["error"] = error
	data["status"] = 2
	data["parser"] = ""
	data["timeout"] = "NULL"
	data["parse_date"] = formattedDate

	_, err := MYSQL.UpdateTask(data, t.Id)
	if err != nil {
		log.Println("MysqlFreeTask.SetError.HasError", err)
	}
}

func (t *MysqlFreeTask) SetLog(text string) {
	if text == "" {
		return
	}

	timePoint := time.Now()
	text = timePoint.Format("2006-01-02 15:04:05") + " #" + strconv.Itoa(t.Id) + ": " + text
	fmt.Println(text)
	t.Log = append(t.Log, text)
	t.SaveLog()
}

func (t *MysqlFreeTask) SaveLog() {
	data := map[string]interface{}{}
	data["log"] = strings.Join(t.Log, "\n")
	data["log_last"] = t.Log[len(t.Log) - 1]

	_, err := MYSQL.UpdateTask(data, t.Id)
	if err != nil {
		log.Println("MysqlFreeTask.SaveLog.HasError", err)
	}
}

func (t *MysqlFreeTask) GetRandDomain() string {
	domains := t.Domain
	if domains != "" && domains != "[]" {
		var arr []string
		err := json.Unmarshal([]byte(domains), &arr)
		if err != nil {
			log.Println("MysqlFreeTask.GetRandDomain.HasError", err)
		}else {
			return UTILS.ArrayRand(arr)
		}
	}
	return ""
}

func (t *MysqlFreeTask) GetRandSymb() string {
	symbs := t.SymbMicroMarking
	if symbs != "" && symbs != "[]" {
		var arr []string
		err := json.Unmarshal([]byte(symbs), &arr)
		if err != nil {
			log.Println("MysqlFreeTask.GetRandSymb.HasError", err)
		}else {
			return UTILS.ArrayRand(arr)
		}
	}
	return ""
}

func (t *MysqlFreeTask) GetRandTag() string {
	moreTags := t.MoreTags
	if moreTags != "" && moreTags != "[]" {
		var arr []string
		err := json.Unmarshal([]byte(moreTags), &arr)
		if err != nil {
			log.Println("MysqlFreeTask.GetRandTag.HasError", err)
		}else {
			return UTILS.ArrayRand(arr)
		}
	}
	return ""
}
