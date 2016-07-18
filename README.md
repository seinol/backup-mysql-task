## Backup MySQL on Cloud Foundry

Dump mysql bound to this app and send to S3.

1. create `.s3cfg` (for [`s3cmd`](http://s3tools.org/usage)) on the top of this repository.
1. bind mysql services to backup (written in [`manifest.yml`](manifest.yml))
1. set environment variable called `BUCKET_NAME` for S3 (written in [`manifest.yml`](manifest.yml))
