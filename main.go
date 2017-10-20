// Copyright 2017 Alter Way
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

// Parameters passed through the commandline
type Parameters struct {
	DBPath          string
	Format          string
	OutputFile      string
	MetricNames     []string
	MinT            int64
	MaxT            int64
	LabelsToInclude []string
	LabelsToExclude []string
	OtherFilters    [][]string
}

func processParameters(c *cli.Context) (*Parameters, error) {
	var (
		includeAll bool
	)
	filters := make([][]string, 0)
	for _, filter := range c.StringSlice("select") {
		filterList := strings.SplitN(filter, ":", 2)
		if filterList == nil {
			return nil, errors.New("Wrong filter value: " + filter)
		}
		filters = append(filters, filterList)
	}
	mint, e1 := time.Parse(time.RFC3339, c.String("mint"))
	if e1 != nil {
		cli.ShowAppHelp(c)
		return nil, errors.New("Bad value for MinT (RFC 3339): " + c.String("mint"))
	}
	maxt, e2 := time.Parse(time.RFC3339, c.String("maxt"))
	if e2 != nil {
		cli.ShowAppHelp(c)
		return nil, errors.New("Bad value for MaxT (RFC 3339): " + c.String("mint"))
	}

	parameters := Parameters{
		DBPath:          c.String("path"),
		Format:          c.String("format"),
		OutputFile:      c.String("out"),
		MetricNames:     c.StringSlice("name"),
		MinT:            mint.Unix() * 1000,
		MaxT:            maxt.Unix() * 1000,
		LabelsToInclude: c.StringSlice("include-label"),
		LabelsToExclude: c.StringSlice("exclude-label"),
		OtherFilters:    filters,
	}
	if parameters.MinT > parameters.MaxT {
		cli.ShowAppHelp(c)
		return &parameters, errors.New("MinT value cannot be superior to MaxT")
	}
	if len(parameters.LabelsToInclude) == 0 {
		parameters.LabelsToInclude = append(parameters.LabelsToInclude, "*")
	}
	for _, label := range parameters.LabelsToInclude {
		if label == "*" {
			includeAll = true
			break
		}
	}
	if Writers[parameters.Format] == nil {
		cli.ShowAppHelp(c)
		return &parameters, errors.New("Invalid output format: " + parameters.Format)
	}
	if parameters.Format == "csv" && includeAll {
		cli.ShowAppHelp(c)
		return &parameters, errors.New("Cannot use csv format with wildcard label export")
	}
	if !exists(parameters.DBPath) {
		return &parameters, errors.New("The database path cannot be accessed")
	}
	return &parameters, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "prometheus-export"
	app.Version = "1.0"
	app.Usage = "Prometheus v2 DB exporter"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "format, f",
			Usage: "export format; one of " + strings.Join(WriterList(), ", "),
			Value: "json",
		},
		cli.StringFlag{
			Name:  "mint, min",
			Usage: "lower bound of the observed time period",
			Value: "1970-01-01T00:00:00Z",
		},
		cli.StringFlag{
			Name:  "maxt, max",
			Usage: "upper bound of the observed time period",
			Value: "2050-01-01T00:00:00Z",
		},
		cli.StringSliceFlag{
			Name:  "name, N",
			Usage: "time series names to export",
		},
		cli.StringSliceFlag{
			Name:  "select, S",
			Usage: "additional filters with -S \"label:value\"",
		},
		cli.StringFlag{
			Name:  "path, p",
			Usage: "path to the directory of the Prometheus Database",
			Value: "/data/prometheus",
		},
		cli.StringFlag{
			Name:  "out, o",
			Usage: "output file path (- for stdout)",
			Value: "-",
		},
		cli.StringSliceFlag{
			Name:  "include-label, I",
			Usage: "labels to include in the export (series without those labels will not be kept)",
		},
		cli.StringSliceFlag{
			Name:  "exclude-label, E",
			Usage: "labels to exclude from the export",
		},
	}
	app.Action = func(c *cli.Context) error {
		parameters, err := processParameters(c)
		if err != nil {
			return err
		}
		err = queryDB(parameters)
		return err
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
