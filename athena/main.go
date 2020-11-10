package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	a "github.com/aws/aws-sdk-go/service/athena"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	GithubActionInputPrefix = "INPUT" // Github actions send inputs as environment variables with prefix "INPUT"
	ViperPathParam          = "path"
	ViperDBParam            = "database"
	ViperWorkgroupParam     = "workgroup"
	ViperRegionParam        = "region"

	CmdName             = "athena-views"
	DefaultAWSCatalog   = "AwsDataCatalog"
	DefaultAWSWorkgroup = "default"
	DefaultAWSRegion    = "us-east-1"
	LogTimeFormat       = "2006-01-02 15:04:05"
	SqlExt              = ".sql"
)

var mainCmd = &cobra.Command{
	Use:   CmdName,
	Short: "Athena views generator",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logrus.WithField("cmd", CmdName)

		// parameters
		workgroup := viper.GetString(ViperWorkgroupParam)
		if workgroup == "" {
			log.Infof("Using default workgroup: %v", DefaultAWSWorkgroup)
			workgroup = DefaultAWSWorkgroup
		}
		region := viper.GetString(ViperRegionParam)
		if region == "" {
			log.Infof("Using default region: %v", DefaultAWSRegion)
			region = DefaultAWSRegion
		}
		path := viper.GetString(ViperPathParam)
		database := viper.GetString(ViperDBParam)
		if path == "" || database == "" {
			return fmt.Errorf("must supply all parameters: [%v, %v]", ViperPathParam, ViperDBParam)
		}

		// find files
		log.Infof("Scanning path: %v", path)
		files, err := listFiles(path, SqlExt)
		if err != nil {
			return err
		}
		log.Infof("Found %v SQL files", len(files))

		// create Athena queries
		sess := session.Must(session.NewSession())
		svc := a.New(sess, aws.NewConfig().WithRegion(region))

		var errs int
		for _, fl := range files {
			log.Infof("--- processing: %v in workgroup: %v", filepath.Base(fl), workgroup)
			sqlContent, err := ioutil.ReadFile(fl)
			if err != nil {
				errs++
				log.Errorf("  > got error reading file: %v", err)
			} else {
				result, err := svc.StartQueryExecution(&a.StartQueryExecutionInput{
					QueryExecutionContext: &a.QueryExecutionContext{
						Catalog:  aws.String(DefaultAWSCatalog),
						Database: aws.String(database),
					},
					WorkGroup:   aws.String(workgroup),
					QueryString: aws.String(string(sqlContent)),
				})

				if err != nil {
					errs++
					log.Errorf("  > got error during execution: %v", err)
				} else {
					log.Infof("  > query execution ID: %v", aws.StringValue(result.QueryExecutionId))
					timeout := time.After(5 * time.Second)
					ticker := time.Tick(1 * time.Second)
					for {
						select {
						case <-ticker:
							output, err := svc.GetQueryExecution(&a.GetQueryExecutionInput{QueryExecutionId: result.QueryExecutionId})
							if err != nil {
								errs++
								log.Errorf("  > got error getting execution: %v", err)
								goto exitLoop
							}

							state := aws.StringValue(output.QueryExecution.Status.State)
							if state == "FAILED" {
								log.Errorf("Failed to execute query in %v: \n\t%v", filepath.Base(fl),
									aws.StringValue(output.QueryExecution.Status.StateChangeReason))
								errs++
								goto exitLoop
							} else if state == "SUCCEEDED" {
								goto exitLoop
							}
						case <-timeout:
							log.Infof("Timed-out waiting to get query result")
							goto exitLoop
						}
					}
				}
			exitLoop:
			}
		}

		var finalErr error
		if errs > 0 {
			finalErr = fmt.Errorf("%v errors occurred during execution", errs)
		}
		return finalErr
	},
}

func main() {
	if err := mainCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// environment vars integration
	viper.AutomaticEnv()
	viper.SetEnvPrefix(GithubActionInputPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	customFormatter := new(logrus.TextFormatter)
	customFormatter.FullTimestamp = true
	customFormatter.TimestampFormat = LogTimeFormat
	logrus.SetFormatter(customFormatter)
}

// recursively lists files
func listFiles(directory string, ext string) ([]string, error) {
	var files []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path != "." && strings.HasSuffix(path, ext) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
