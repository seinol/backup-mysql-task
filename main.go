package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/robfig/cron"
)

func dbdumper() {
	fmt.Println("starting db dumper")

	env, _ := cfenv.Current()
	mariadbs, _ := env.Services.WithLabel("mariadbent")
	mongodbs, _ := env.Services.WithLabel("mongodbent")

	//mysqldump for all mariadbs
	for _, s := range mariadbs {
		_, err := exec.Command("sh", "-c", "mkdir /tmp/"+s.Name).Output()
		if err != nil {
			log.Fatal("creating tmp dir "+s.Name+" failed: ", err.Error())
			return
		}
		command := fmt.Sprintf("/home/vcap/deps/0/apt/usr/bin/mysqldump --single-transaction --routines --triggers --no-create-db --skip-add-locks -u %s --password=%s -h %s --databases %s > /tmp/%s/%s_%s.sql", s.Credentials["username"], s.Credentials["password"], s.Credentials["hostname"], s.Credentials["name"], s.Name, time.Now().Format("2006-01-02"), s.Name)
		_, err = exec.Command("sh", "-c", command).Output()
		if err != nil {
			fmt.Println("creating "+s.Name+" mysqldump failed: ", err.Error())
		} else {
			fmt.Println("dumped on /tmp/" + s.Name + "/" + time.Now().Format("2006-01-02") + "_" + s.Name + ".sql")
		}
	}

	//mongodump for all mongodbs
	for _, s := range mongodbs {
		_, err := exec.Command("sh", "-c", "mkdir -p /tmp/"+s.Name).Output()
		if err != nil {
			log.Fatal("creating tmp dir for "+s.Name+" failed: ", err.Error())
			return
		}

		hosts := strings.Split(s.Credentials["host"].(string), ",")
		command := fmt.Sprintf("/home/vcap/deps/0/apt/usr/bin/mongodump -u %s -p %s -h %s/%s:%s,%s:%s,%s:%s -d %s -o /tmp/%s", s.Credentials["username"], s.Credentials["password"], s.Credentials["replica_set"], hosts[0], s.Credentials["port"], hosts[1], s.Credentials["port"], hosts[2], s.Credentials["port"], s.Credentials["database"], s.Name)
		_, err = exec.Command("sh", "-c", command).Output()
		if err != nil {
			fmt.Println("creating "+s.Name+" mongodump failed: ", err.Error())
		} else {
			fmt.Println("dumped on /tmp/" + s.Name + "_" + time.Now().Format("2006-01-02"))
		}
	}

	//create bucket and put to bucket with s3cmd for mariadbs
	for _, s := range mariadbs {

		//creating buckets automatically by the app is actually not possible because of s3cmd errors: 403 (SignatureDoesNotMatch) - problem unresolved
		_, err := exec.Command("sh", "-c", "/home/vcap/deps/0/apt/usr/bin/s3cmd -c /home/vcap/app/.s3cfg mb s3://"+s.Name).Output()
		if err != nil {
			log.Fatal("creating bucket for "+s.Name+" mariadb failed: ", err.Error())
			return
		}

		_, err = exec.Command("sh", "-c", "/home/vcap/deps/0/apt/usr/bin/s3cmd -c /home/vcap/app/.s3cfg put /tmp/"+s.Name+"/*.sql s3://"+s.Name).Output()
		if err != nil {
			fmt.Println("putting "+s.Name+" mariadbdump to s3 bucket failed: ", err.Error())
		} else {
			fmt.Println("pushed " + s.Name + " mariadb sql dump to s3 dynamic storage")
		}
	}
	//create bucket and put to bucket with s3cmd for mongodbs
	for _, s := range mongodbs {

		//creating buckets automatically by the app is actually not possible because of s3cmd errors: 403 (SignatureDoesNotMatch) - problem unresolved
		_, err := exec.Command("sh", "-c", "/home/vcap/deps/0/apt/usr/bin/s3cmd -c /home/vcap/app/.s3cfg mb s3://"+s.Name).Output()
		if err != nil {
			log.Fatal("creating bucket for "+s.Name+" mongodb failed: ", err.Error())
			return
		}
		_, err = exec.Command("sh", "-c", "cd /tmp/"+s.Name+"/"+s.Credentials["database"].(string)+"/ && for f in *; do mv \"$f\" \""+time.Now().Format("2006-01-02")+"_$f\"; done && cd").Output()
		if err != nil {
			fmt.Println("renaming "+s.Name+" dump files failed: ", err.Error())
		}
		_, err = exec.Command("sh", "-c", "/home/vcap/deps/0/apt/usr/bin/s3cmd -c /home/vcap/app/.s3cfg put /tmp/"+s.Name+"/"+s.Credentials["database"].(string)+"/* s3://"+s.Name).Output()
		if err != nil {
			fmt.Println("putting "+s.Name+" mongodb dump files to s3 bucket failed: ", err.Error())
		} else {
			fmt.Println("pushed " + s.Name + " mongodb dump files to s3 dynamic storage")
		}
	}

	//delete all local mariadb dumps
	for _, s := range mariadbs {
		_, err := exec.Command("sh", "-c", "rm -rf /tmp/"+s.Name).Output()
		if err != nil {
			log.Fatal("removing tmp-dir for "+s.Name+" failed: ", err.Error())
			return
		}
		fmt.Println("deleted " + s.Name + " local mariadb sql dump in /tmp")
	}
	//delete all local mongodb dumps
	for _, s := range mongodbs {
		_, err := exec.Command("sh", "-c", "rm -rf /tmp/"+s.Name).Output()
		if err != nil {
			log.Fatal("removing tmp-dir for "+s.Name+" failed: ", err.Error())
			return
		}
		fmt.Println("deleted " + s.Name + " local mongodb dump in /tmp")
	}
	fmt.Println("finished, waiting for next cron time")
}

func main() {
	c := cron.New()
	c.AddFunc(os.Getenv("CRON_EXPRESSION"), dbdumper)
	c.Start()

	var port string
	if port = os.Getenv("PORT"); len(port) == 0 {
		port = "9000"
	}

	fmt.Println("Start " + port)
	http.ListenAndServe(":"+port, nil)
}
