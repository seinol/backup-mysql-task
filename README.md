copied for the first time from https://github.com/making/backup-mysql-task, but now modified for our usecases...

## DB Dumper

Dump mariadb and mongodb service DB's bound to this app and send to s3 dynamic storage.

ATTENTION: Since a bug has occurred the buckets have to be created manually in advance on the s3!

1. set environment variable called `CRON_EXPRESSION` (written in [`manifest.yml`](manifest.yml)). Outcomment the call of the backup function(dbdumper) in the main.go file, otherwise your app will crash often after step 2 and you aren't able to create your s3 config in the container!
2. push your app to your Application Cloud (cf push).
3. create `.s3cfg` (for [`s3cmd`](http://s3tools.org/usage)) and put it on the top of this repository. Do this in your container Linux machine, because there is another version of s3cmd than on your personal computer! Access your container with "cf ssh <myapp>".
4. bind services to backup (written in [`manifest.yml`](manifest.yml))
5. revert the comment of the backup function call in the main.go file as done in step 1
6. push your app again to your Application Cloud
<br>

### ATTENTION
 - Only mariadbent and mongodbent are currently supported! It is possible to extend to mongodb-2, but it is not implemented yet!

### Troubleshooting in the container
run the following command first to set LIB and BIN Paths:
```sh
export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:/home/vcap/deps/0/lib/" && export PATH="$PATH:/home/vcap/deps/0/bin/"
```

#### s3cmd
```sh
s3cmd -c /home/vcap/app/.s3cfg <ls|la|mb...>
```

#### mysql dump
```sh
mysqldump --single-transaction --routines --triggers --no-create-db --skip-add-locks -u <username> --password=<password> -h <hostname> --databases <database>
```

#### mysql restore
Please use for mysql restore a tool like adminer or phpmyadmin.

#### mongodb dump
```sh
mongodump -u <username> -p <password> --host <replica_set>/<host1><:port1>,<host2><:port2>,<host3><:port3> -d <database> -o <path/to/folder>
```

#### mongodb restore
Please use in first case a tool which can restore mongodb archive dumps (p. e. Studio 3T). In second case you can use mongorestore(Attention: use the primary host of the mongodb replica).
```sh
mongorestore -u <username> -p <password> -h <host1>:<port1> -d <database> <path/to/folder>
```