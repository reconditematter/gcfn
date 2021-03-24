// Copyright (c) 2019-2021 Leonid Kneller. All rights reserved.
// Licensed under the MIT license.
// See the LICENSE file for full license information.

package randomnames

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// humanname -- the name and gender of a person.
type humanname struct {
	Family string `json:"family"`
	Given  string `json:"given"`
	Gender string `json:"gender"`
}

const (
	genboth = iota // specifies both genders
	genf           // specifies the female gender
	genm           // specifies the male gender
)

// gen -- generates `count` random names.
// `gengen` specifies the names gender.
// This function returns an error when count∉{1,...,1000}.
func gen(count int, gengen int) ([]humanname, error) {
	if !(1 <= count && count <= 1000) {
		return nil, errors.New("invalid count")
	}
	//
	names := make(map[[2]string]string)
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)
	//
	for len(names) < count {
		i := rng.Intn(1000)
		j := rng.Intn(1000)
		switch gengen {
		case genf:
			name := [2]string{family[i], givenf[j]}
			names[name] = "female"
		case genm:
			name := [2]string{family[i], givenm[j]}
			names[name] = "male"
		default:
			k := rng.Uint64() < math.MaxUint64/2
			if k {
				name := [2]string{family[i], givenf[j]}
				names[name] = "female"
			} else {
				name := [2]string{family[i], givenm[j]}
				names[name] = "male"
			}
		}
	}
	//
	hnames := make([]humanname, 0, count)
	for name, gender := range names {
		hn := humanname{Family: name[0], Given: name[1], Gender: gender}
		hnames = append(hnames, hn)
	}
	//
	return hnames, nil
}

// RandomNames?count=nnnn[&gender={F|M}] -- generates `count` random names.
//
//   count={1|2|...|1000}
//   gender={F|M}
func RandomNames(w http.ResponseWriter, r *http.Request) {
	qcount := r.URL.Query()["count"]
	if len(qcount) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request: invalid count"))
		return
	}
	count, err := strconv.Atoi(qcount[0])
	if err != nil || !(1 <= count && count <= 1000) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request: invalid count"))
		return
	}
	//
	qgender := r.URL.Query()["gender"]
	if len(qgender) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request: invalid gender"))
		return
	}
	gengen := genboth
	if len(qgender) != 0 {
		qgen := strings.ToUpper(qgender[0])
		if qgen == "F" {
			gengen = genf
		} else if qgen == "M" {
			gengen = genm
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 Bad Request: invalid gender"))
			return
		}
	}
	//
	result, err := gen(count, gengen)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 Internal Server Error"))
		return
	}
	//
	fcount := 0
	for _, hn := range result {
		if hn.Gender == "female" {
			fcount++
		}
	}
	resultx := struct {
		Count  int         `json:"count"`
		FCount int         `json:"fcount"`
		MCount int         `json:"mcount"`
		Names  []humanname `json:"names"`
	}{len(result), fcount, len(result) - fcount, result}
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
