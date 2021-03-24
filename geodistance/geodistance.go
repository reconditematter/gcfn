// Copyright (c) 2019-2021 Leonid Kneller. All rights reserved.
// Licensed under the MIT license.
// See the LICENSE file for full license information.

package geodistance

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
)

func andoyer(lat1, lon1, lat2, lon2 float64) (float64, error) {
	if !(math.Abs(lat1) <= 90 && math.Abs(lat2) <= 90 && math.Abs(lon1) <= 180 && math.Abs(lon2) <= 180) {
		return 0, errors.New("invalid coordinate")
	}
	//
	const (
		// WGS1984
		a = 6378137.0
		f = 1.0 / 298.257223563
	)
	//
	φ1, λ1 := lat1*(math.Pi/180), lon1*(math.Pi/180)
	φ2, λ2 := lat2*(math.Pi/180), lon2*(math.Pi/180)
	//
	F := (φ1 + φ2) / 2
	G := (φ1 - φ2) / 2
	L := (λ1 - λ2) / 2
	//
	sinF, cosF := math.Sincos(F)
	sinG, cosG := math.Sincos(G)
	sinL, cosL := math.Sincos(L)
	//
	S := math.Hypot(sinG*cosL, cosF*sinL)
	C := math.Hypot(cosG*cosL, sinF*sinL)
	ω := math.Atan2(S, C)
	D := 2 * a * ω
	//
	R := S * C / ω
	H1 := (3*R - 1) / (2 * C * C)
	H2 := (3*R + 1) / (2 * S * S)
	d := D * (1 + f*(H1*sq(sinF*cosG)-H2*sq(cosF*sinG)))
	//
	if fin(d) {
		return d, nil
	}
	// p1 antipodal p2
	if fin(R) {
		// use Ramanujan's celebrated formula for the perimeter of an ellipse
		t := 3 * sq(f/(2-f))
		return a * (1 - f/2) * (1 + t/(10+math.Sqrt(4-t))), nil
	}
	// p1=p2
	return 0, nil
}

func sq(x float64) float64 { return x * x }

func fin(x float64) bool { return (x - x) == 0 }

func GeoDistance(w http.ResponseWriter, r *http.Request) {
	qlat1 := r.URL.Query()["lat1"]
	qlon1 := r.URL.Query()["lon1"]
	qlat2 := r.URL.Query()["lat2"]
	qlon2 := r.URL.Query()["lon2"]
	if !(len(qlat1) == 1 && len(qlon1) == 1 && len(qlat2) == 1 && len(qlon2) == 1) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request: invalid coordinate"))
		return
	}
	//
	lat1, erra := strconv.ParseFloat(qlat1[0], 64)
	lon1, errb := strconv.ParseFloat(qlon1[0], 64)
	lat2, errc := strconv.ParseFloat(qlat2[0], 64)
	lon2, errd := strconv.ParseFloat(qlon2[0], 64)
	if !(erra == nil && errb == nil && errc == nil && errd == nil) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request: invalid coordinate"))
		return
	}
	//
	result, err := andoyer(lat1, lon1, lat2, lon2)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request: invalid coordinate"))
		return
	}
	//
	type geo struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
	geo1 := geo{math.Round(lat1*1e8) / 1e8, math.Round(lon1*1e8) / 1e8}
	geo2 := geo{math.Round(lat2*1e8) / 1e8, math.Round(lon2*1e8) / 1e8}
	resultx := struct {
		Source  geo     `json:"source"`
		Target  geo     `json:"target"`
		Dist_km float64 `json:"dist_km"`
		Dist_mi float64 `json:"dist_mi"`
	}{geo1, geo2, math.Round((result/1000)*100) / 100, math.Round((result/((1200.0/3937.0)*5280.0))*100) / 100}
	//
	resultj, err := json.Marshal(resultx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultj)
}
