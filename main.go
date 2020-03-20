package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/webjohny/qaphantom/config"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

func cmdTimer(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "%s %s %s \n", r.Method, r.URL, r.Proto)
	//Iterate over all header fields
	for k, v := range r.Header {
		fmt.Fprintf(w, "Header field %q, Value %q\n", k, v)
	}

	fmt.Fprintf(w, "Host = %q\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr= %q\n", r.RemoteAddr)
	//Get value for a specified token
	fmt.Fprintf(w, "\n\nFinding value of \"Accept\" %q", r.Header["Accept"])

	w.Header().Set("Content-Type", "application/json")

	commandExec := r.FormValue("cmd")

	result := map[string]interface{}{
		"status": true,
	}

	var limit int64 = 1000

	if len(r.FormValue("limit")) > 0 {
		limit, _ = strconv.ParseInt(r.FormValue("limit"), 0, 64)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1000 * limit) * time.Millisecond)
	defer cancel()
	//php -f /var/www/html/cron.php parser cron sleeping 5
	out, err := exec.CommandContext(ctx, "bash", "-c", commandExec).Output()

	if err != nil {
		// This will fail after 100 milliseconds. The 5 second sleep
		// will be interrupted.
		result["status"] = false
		fmt.Println(err)
	}
	fmt.Println(string(out))

	json.NewEncoder(w).Encode(result)
}

func main() {
	conf := config.Create()

	r := mux.NewRouter()
	r.HandleFunc("/cmd-timer", cmdTimer).Methods("POST")
	log.Fatal(http.ListenAndServe(":" + string(conf.Port), r))
}