package stbsource

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type StbWeather struct {
	Temperature float64
	Humidity    float64
	Timestamp   time.Time
	Station     string
}

func GetStbSource() (*StbWeather, error) {
	respLen := 4
	resp, err := http.Get("http://weather.sun.ac.za/api/getlivedata.php?temperature&humidity&time&date")
	if err != nil {
		return nil, fmt.Errorf("could not get page: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body: %v", err)
	}
	valuesStr := strings.Split(string(body), "<br />")
	if len(valuesStr)-1 != respLen {
		return nil, fmt.Errorf("body did not contain expected result length: %v", len(valuesStr))
	}
	values := make([]float64, 2)
	for i := 0; i < 2; i++ {
		values[i], err = strconv.ParseFloat(valuesStr[i], 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse float: %v", err)
		}
	}
	tsStr := valuesStr[respLen-1] + " " + valuesStr[respLen-2]

	ts, err := time.Parse("2006-01-02 15:04", tsStr)
	ts = ts.Add(-2 * time.Hour)
	if err != nil {
		return nil, fmt.Errorf("could not parse time string: %v", err)
	}

	tmp := StbWeather{
		Temperature: values[0],
		Humidity:    values[1],
		Timestamp:   ts,
		Station:     "Sonbesie",
	}
	return &tmp, nil
}
