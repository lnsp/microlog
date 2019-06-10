package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/lnsp/microlog/common/logger"
)

var log = logger.New()

type specification struct {
	Addr         string `default:":8080" desc:"Address the service is listening on"`
	S3AccessKey  string `required:"true" desc:"Access key for S3-compatible object storage"`
	S3SecretKey  string `required:"true" desc:"Secret key for S3-compatible object storage"`
	S3BucketPath string `required:"true" desc:"S3-compatible bucket path string"`
	S3Bucket     string `required:"true" desc:"S3-compatible bucket name"`
	S3Endpoint   string `required:"true" desc:"S3-compatible endpoint URL"`
	Datasource   string `required:"true" desc:"gorm compatible datasource"`
}

func main() {
	var spec specification
	if err := envconfig.Process("profile", &spec); err != nil {
		envconfig.Usage("profile", &spec)
		return
	}
}
