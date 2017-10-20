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
	"math"
	"os"

	tsdb "github.com/prometheus/tsdb"
	labels "github.com/prometheus/tsdb/labels"
	log "github.com/sirupsen/logrus"
)

// Row is a Time Series value
type Row struct {
	Name   string
	Labels map[string]string
	Time   int64
	Value  float64
}

func openDB(path string) (*tsdb.DB, error) {
	options := tsdb.DefaultOptions
	options.NoLockfile = true
	return tsdb.Open(path, nil, nil, options)
}

// Filter the labels we are exporting
func intersectLabels(include []string, exclude []string, current labels.Labels) (string, map[string]string, error) {
	tmpResult := make(map[string]string)
	result := make(map[string]string)
	name := ""
	includeAll := false
	for _, label := range include {
		if includeAll = (label == "*"); includeAll {
			break
		}
	}
	for _, label := range current {
		tmpResult[label.Name] = label.Value
		if label.Name == "__name__" {
			name = label.Value
		}
	}
	// only include select labels, discard series where labels are not matched
	if !includeAll {
		for _, includeLabel := range include {
			if value, ok := tmpResult[includeLabel]; ok {
				result[includeLabel] = value
			} else {
				return "", nil, errors.New("labels not found")
			}
		}
	} else {
		result = tmpResult
	}
	for _, excludeLabel := range exclude {
		delete(result, excludeLabel)
	}
	return name, result, nil
}

func extract(querier tsdb.Querier, matchers []labels.Matcher, name string, include []string, exclude []string, rowsChan chan Row) {
	if name != "" {
		matchers = append(matchers, labels.NewEqualMatcher("__name__", name))
	}
	res := querier.Select(matchers...)
	for res.Next() {
		series := res.At()
		seriesLabels := series.Labels()
		name, localLabels, err := intersectLabels(include, exclude, seriesLabels)
		if err != nil {
			log.WithFields(log.Fields{"include": include, "exclude": exclude, "labels": seriesLabels}).Debug(err)
			continue
		}
		iter := series.Iterator()
		for iter.Next() {
			time, value := iter.At()
			if math.IsNaN(value) {
				continue
			}
			rowsChan <- Row{
				Name:   name,
				Value:  value,
				Time:   time,
				Labels: localLabels,
			}
		}
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func queryDB(parameters *Parameters) error {
	var (
		err          error
		otherFilters []labels.Matcher
	)
	db, err := openDB(parameters.DBPath)
	if err != nil {
		return err
	}
	querier, err := db.Querier(parameters.MinT, parameters.MaxT)
	if err != nil {
		return err
	}
	for _, matcher := range parameters.OtherFilters {
		otherFilters = append(otherFilters, labels.NewEqualMatcher(matcher[0], matcher[1]))
	}
	log.WithFields(log.Fields{"labels": parameters.LabelsToInclude}).Info("Including labels")
	log.WithFields(log.Fields{"labels": parameters.LabelsToExclude}).Info("Excluding labels")
	rowsChan := make(chan Row)
	go streamResults(parameters.OutputFile, parameters.Format, rowsChan)
	if len(parameters.MetricNames) != 0 {
		for _, name := range parameters.MetricNames {
			extract(querier,
				otherFilters,
				name,
				parameters.LabelsToInclude,
				parameters.LabelsToExclude,
				rowsChan)
		}
	} else if len(parameters.OtherFilters) != 0 {
		extract(querier,
			otherFilters,
			"",
			parameters.LabelsToInclude,
			parameters.LabelsToExclude,
			rowsChan)
	}
	return nil
}
