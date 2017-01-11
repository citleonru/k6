/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2016 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package v1

import (
	"github.com/julienschmidt/httprouter"
	"github.com/loadimpact/k6/api/common"
	"github.com/manyminds/api2go/jsonapi"
	"net/http"
)

func HandleGetMetrics(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	engine := common.GetEngine(r.Context())

	metrics := make([]Metric, 0)
	for m, _ := range engine.Metrics {
		metrics = append(metrics, NewMetric(*m))
	}

	data, err := jsonapi.Marshal(metrics)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = rw.Write(data)
}

func HandleGetMetric(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	engine := common.GetEngine(r.Context())

	var metric Metric
	var found bool
	for m, _ := range engine.Metrics {
		if m.Name == id {
			metric = NewMetric(*m)
			found = true
			break
		}
	}

	if !found {
		http.Error(rw, "No such metric", http.StatusNotFound)
		return
	}

	data, err := jsonapi.Marshal(metric)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = rw.Write(data)
}