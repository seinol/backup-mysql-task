package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/robfig/cron"
)

func mysqlbackup() {
	cfenv, _ := cfenv.Current()
	services, _ := cfenv.Services.WithTag("mysql")
	for _, s := range services {
		dump := fmt.Sprintf("/home/vcap/app/.apt/usr/bin/mysqldump -u %s --password=%s -h %s --databases %s > /tmp/%s.sql", s.Credentials["username"], s.Credentials["password"], s.Credentials["hostname"], s.Credentials["name"], s.Name)
		exec.Command("sh", "-c", dump).Output()
		fmt.Println("Dumped on /tmp/" + s.Name + ".sql")
	}
	exec.Command("sh", "-c", "/home/vcap/app/.apt/usr/bin/s3cmd -c /home/vcap/app/.s3cfg put /tmp/*.sql s3://"+os.Getenv("BUCKET_NAME")).Output()
}

func main() {
	c := cron.New()
	c.AddFunc("@daily", mysqlbackup)
	c.Start()

	var port string
	if port = os.Getenv("PORT"); len(port) == 0 {
		port = "9000"
	}

	fmt.Println("Start " + port)
	http.ListenAndServe(":"+port, nil)
}
