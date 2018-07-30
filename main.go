package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
	_ "net/http/pprof"
	"os/exec"
	"fmt"
	"time"
)
const (
	SUCCESS            int = 0
	ILLEGAL_DATAFORMAT int = 41003

	INVALID_METHOD int = 42001
	INVALID_PARAMS int = 42002

	INTERNAL_ERROR  int = 45002
	TOO_EARLY  int = 45003
)

var applyhistory = make(map[string]time.Time)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/echo", Echo),
		rest.Get("/tokenapply/:address", tokenapply),
	)
	if err != nil {
		fmt.Println("StartRestfulServer err=", err)
	}
	api.SetApp(router)
	err = http.ListenAndServe(":8080", api.MakeHandler())
	if err != nil {
		fmt.Println("ListenAndServe: ", err.Error())
	}
	fmt.Println("StartRestfulServer completed.")
}

func ResponsePack(errCode int) map[string]interface{} {
	resp := map[string]interface{}{
		"Action":  "",
		"Result":  "",
		"Error":   errCode,
		"Desc":    "",
		"Version": "1.0.0",
	}
	return resp
}

func Echo(w rest.ResponseWriter, r *rest.Request) {
	rsp := ResponsePack(SUCCESS)
	rsp["Action"] = "Echo"
	rsp["Result"] = "Echo"
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteJson(rsp)
	return
}

func tokenapply(w rest.ResponseWriter, r *rest.Request) {
	address := r.PathParam("address")
	if lastime, ok := applyhistory[address]; ok {
		if getHourDiffer(lastime,time.Now()) <24{
			rsp := ResponsePack(TOO_EARLY)
			rsp["Action"] = "tokenapply"
			rsp["Result"] = fmt.Sprint("you have applied today, can you try in next 24 hour, thanks")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteJson(rsp)
		}
		return
	}

	if address == "" {
		rest.Error(w, "address required.", http.StatusInternalServerError)
		return
	}

	c := fmt.Sprintf(`echo 1|./ontology asset transfer --asset=ont --to=%s  --from=1 --amount=1000`,address)
	cmd := exec.Command("sh", "-c", c)
	out, err := cmd.Output()
	if err!=nil{
		rsp := ResponsePack(INTERNAL_ERROR)
		rsp["Action"] = "tokenapply"
		rsp["Result"] = fmt.Sprintf("failed.out=%s,error=%s",out,err)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteJson(rsp)
		return
	}
	c2 := fmt.Sprintf(`echo 1|./ontology asset transfer --asset=ong --to=%s  --from=1 --amount=1000`,address)
	cmd2 := exec.Command("sh", "-c", c2)
	out2, err := cmd2.Output()
	if err!=nil{
		rsp := ResponsePack(INTERNAL_ERROR)
		rsp["Action"] = "tokenapply"
		rsp["Result"] = fmt.Sprintf("failed.out=%s,error=%s.",out2,err)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteJson(rsp)
		return
	}
	applyhistory[address]=time.Now()
	rsp := ResponsePack(SUCCESS)
	rsp["Action"] = "tokenapply"
	rsp["Result"] = fmt.Sprintf("Success.\n ONT transfer hash=%s\n,ONG transfer hash=%s\n",out,out2)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteJson(rsp)
}


//获取相差时间
func getHourDiffer(t1, t2 time.Time) int64 {
	var hour int64
	var err error
	//t1, err := time.ParseInLocation("2006-01-02 15:04:05", start_time, time.Local)
	//t2, err := time.ParseInLocation("2006-01-02 15:04:05", end_time, time.Local)
	if err == nil && t1.Before(t2) {
		diff := t2.Unix() - t1.Unix() //
		hour = diff / 3600
		return hour
	} else {
		return hour
	}
}