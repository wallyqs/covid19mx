package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// Used to detect where the actual data is being located, this
	// seems to change day to day.
	sinaveURL = "https://covid19.sinave.gob.mx/mapa.aspx"

	// Latest data will be usually found in one of the following urls.
	sinaveURLA = "https://covid19.sinave.gob.mx/Mapa.aspx/Grafica22"
	sinaveURLB = "https://covid19.sinave.gob.mx/Mapa.aspx/Grafica23"

	// repoURL can be used to fetch previous days date.
	repoURL = "https://wallyqs.github.io/covid19mx/data/"

	// attackRateURL is the url with the info with the
	//
	// See: https://en.wikipedia.org/wiki/Attack_rate
	//      https://es.wikipedia.org/wiki/Incidencia
	//
	attackRateURL = "https://covid19.sinave.gob.mx/Log.aspx/Grafica22"

	// municipalURL is the url from where we can get the data at a
	// municipal level.
	municipalURL = "https://coronavirus.gob.mx/fHDMap/info/getInfoMun.php"
)

const (
	version     = "0.3.2"
	releaseDate = "April 20th, 2020"
)

var (
	ErrSourceNotFound = errors.New("Could not find datasource!")
)

func init() {
	log.SetFlags(0)
}

type State struct {
	Name          string  `json:"name"`
	PositiveCases int     `json:"positive"`
	NegativeCases int     `json:"negative"`
	SuspectCases  int     `json:"suspect"`
	Deaths        int     `json:"deaths"`
	AttackRate    float64 `json:"attack_rate"`
}

type Municipio struct {
	Name          string  `json:"name"`
	PositiveCases int     `json:"positive"`
	NegativeCases int     `json:"negative"`
	SuspectCases  int     `json:"suspect"`
	Deaths        int     `json:"deaths"`
	AttackRate    float64 `json:"attack_rate"`
}

type SinaveData struct {
	States []State `json:"states"`
	tpc    int
	tnc    int
	tsc    int
	td     int

	// ar is the attackRate
	ar float64
}

func (s *SinaveData) UnmarshalJSON(b []byte) error {
	// More JSON is embedded into the object...
	// {"d":"[[]]"}
	var all map[string]interface{}
	err := json.Unmarshal(b, &all)
	if err != nil {
		return err
	}
	data := all["d"].(string)
	var states [][]interface{}
	err = json.Unmarshal([]byte(data), &states)
	if err != nil {
		return err
	}

	s.States = make([]State, 0)
	for _, entry := range states {
		// e.g.
		// [1 Aguascalientes 1353758.409 01 24 243 74 0]
		name := entry[1].(string)
		if name == "NACIONAL" {
			continue
		}

		pos, err := strconv.Atoi(entry[4].(string))
		if err != nil {
			return err
		}
		neg, err := strconv.Atoi(entry[5].(string))
		if err != nil {
			return err
		}
		susp, err := strconv.Atoi(entry[6].(string))
		if err != nil {
			return err
		}
		deaths, err := strconv.Atoi(entry[7].(string))
		if err != nil {
			return err
		}
		attackRate, err := strconv.ParseFloat(entry[8].(string), 64)
		if err != nil {
			return err
		}
		state := State{
			Name:          name,
			PositiveCases: pos,
			NegativeCases: neg,
			SuspectCases:  susp,
			Deaths:        deaths,
			AttackRate:    attackRate,
		}
		s.States = append(s.States, state)
	}

	return nil
}

func (sdata *SinaveData) TotalPositiveCases() int {
	if sdata.tpc > 0 {
		return sdata.tpc
	}
	for _, state := range sdata.States {
		sdata.tpc += state.PositiveCases
	}
	return sdata.tpc
}

func (sdata *SinaveData) TotalNegativeCases() int {
	if sdata.tnc > 0 {
		return sdata.tnc
	}
	for _, state := range sdata.States {
		sdata.tnc += state.NegativeCases
	}
	return sdata.tnc
}

func (sdata *SinaveData) TotalSuspectCases() int {
	if sdata.tsc > 0 {
		return sdata.tsc
	}
	for _, state := range sdata.States {
		sdata.tsc += state.SuspectCases
	}
	return sdata.tsc
}

func (sdata *SinaveData) TotalDeaths() int {
	if sdata.td > 0 {
		return sdata.td
	}
	for _, state := range sdata.States {
		sdata.td += state.Deaths
	}
	return sdata.td
}

func (sdata *SinaveData) TestPositivityRate() float64 {
	return float64(sdata.TotalPositiveCases()) / float64((sdata.TotalPositiveCases() + sdata.TotalNegativeCases()))
}

func fetchData(endpoint string) (*SinaveData, error) {
	hc := &http.Client{}
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s", body)
	}

	var sdata *SinaveData
	err = json.Unmarshal(body, &sdata)
	if err != nil {
		return nil, err
	}
	return sdata, nil
}

func fetchPastData(endpoint string) (*SinaveData, error) {
	hc := &http.Client{}
	resp, err := hc.Get(endpoint)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s", body)
	}

	type s struct {
		States []State `json:"states"`
	}
	var sd *s
	err = json.Unmarshal(body, &sd)
	if err != nil {
		log.Fatal(err)
	}

	sdata := &SinaveData{
		States: sd.States,
	}
	return sdata, nil
}

func detectLatestDataSource() (string, error) {
	hc := &http.Client{}
	resp, err := hc.Get(sinaveURL)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error: %s", body)
	}

	// ...
	if bytes.Contains(body, []byte("Grafica22")) {
		return sinaveURLA, nil
	}
	if bytes.Contains(body, []byte("Grafica23")) {
		return sinaveURLB, nil
	}
	return "", ErrSourceNotFound
}

func fetchMunicipalData(endpoint string, caseType string) (map[string]int, error) {
	vals := url.Values{"sPatType": {caseType}}
	resp, err := http.PostForm(endpoint, vals)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s", body)
	}

	muns, err := parseScript(string(body))
	if err != nil {
		return nil, err
	}

	return muns, nil
}

func parseScript(sample string) (map[string]int, error) {
	muns := map[string]int{}
	var (
		start, end   int
		vstart, vend int
		mun          string
	)
	for i, c := range sample {
		if start == 0 && c == 39 {
			start = i + 1
		} else if start > 0 && end == 0 && c == 39 {
			end = i
			mun = sample[start:end]
			if mun == "body" {
				break
			}
		} else if start > 0 && end > 0 && vstart == 0 && c == '=' {
			vstart = i + 1
		} else if start > 0 && end > 0 && vstart > 0 && vend == 0 && c == ';' {
			vend = i
			v, err := strconv.Atoi(sample[vstart:vend])
			if err != nil {
				return nil, err
			}
			muns[mun] = v

			// Reset everything
			start = 0
			end = 0
			vstart = 0
			vend = 0
		}
	}

	return muns, nil
}

func showTable(sdata *SinaveData) {
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|---------|-------------|------------|")
	fmt.Println("| Estado               | Casos Positivos | Casos Negativos | Casos Sospechosos | Decesos | Positividad | Incidencia |")
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|---------|-------------|------------|")
	var totalAttackRate float64
	for _, state := range sdata.States {
		if state.Name == "NACIONAL" {
			totalAttackRate = state.AttackRate
			continue
		}
		testPositivityRate := float64(state.PositiveCases) / (float64(state.PositiveCases) + float64(state.NegativeCases))
		fmt.Printf("| %-20s | %-15d | %-15d | %-17d | %-7d | %-8.4f    | %-8.2f   |\n",
			state.Name,
			state.PositiveCases,
			state.NegativeCases,
			state.SuspectCases,
			state.Deaths,
			testPositivityRate,
			state.AttackRate,
		)
	}
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|---------|-------------|------------|")
	fmt.Printf("| %-20s | %-15d | %-15d | %-17d | %-7d | %-8.4f    | %-8.4f   |\n",
		"TOTAL",
		sdata.TotalPositiveCases(),
		sdata.TotalNegativeCases(),
		sdata.TotalSuspectCases(),
		sdata.TotalDeaths(),
		sdata.TestPositivityRate(),
		totalAttackRate,
	)
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|---------|-------------|------------|")
}

func showTableDiff(sdata, pdata *SinaveData) {
	pmap := make(map[string]State)
	for _, state := range pdata.States {
		if state.Name == "NACIONAL" {
			continue
		}
		pmap[state.Name] = state
	}

	fmt.Println("|----------------------|-----------------|-----------------|-------------------|-------------|")
	fmt.Println("| Estado               | Casos Positivos | Casos Negativos | Casos Sospechosos | Decesos     |")
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|-------------|")
	for _, state := range sdata.States {
		pstate := pmap[state.Name]
		fmt.Printf("| %-20s | %-15s | %-15s | %-17s | %-11s |\n",
			state.Name,
			fmt.Sprintf("%-5d (%d)", state.PositiveCases-pstate.PositiveCases, state.PositiveCases),
			fmt.Sprintf("%-5d (%d)", state.NegativeCases-pstate.NegativeCases, state.NegativeCases),
			fmt.Sprintf("%-5d (%d)", state.SuspectCases-pstate.SuspectCases, state.SuspectCases),
			fmt.Sprintf("%-5d (%d)", state.Deaths-pstate.Deaths, state.Deaths),
		)
	}
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|-------------|")
	fmt.Printf("| %-20s | %-15d | %-15d | %-17d | %-11d |\n",
		"TOTAL",
		sdata.TotalPositiveCases()-pdata.TotalPositiveCases(),
		sdata.TotalNegativeCases()-pdata.TotalNegativeCases(),
		sdata.TotalSuspectCases()-pdata.TotalNegativeCases(),
		sdata.TotalDeaths()-pdata.TotalDeaths(),
	)
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|-------------|")
}

func showTableAwkFriendly(sdata *SinaveData) {
	for _, state := range sdata.States {
		if state.Name == "NACIONAL" {
			continue
		}
		var name string
		if state.Name == "Ciudad de México" {
			name = "CDMX"
		} else {
			name = strings.Join(strings.Fields(state.Name), "1d")
		}
		fmt.Printf("%-20s\t%-15d\t%-15d\t%-17d\t%-7d\n",
			name, state.PositiveCases, state.NegativeCases, state.SuspectCases, state.Deaths)
	}
}

func showJSON(sdata *SinaveData) {
	result, err := json.MarshalIndent(sdata, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result))
}

func showCSV(sdata *SinaveData) {
	fmt.Println("\"Estado\"               , \"Casos Positivos\" , \"Casos Negativos\" , \"Casos Sospechosos\" , \"Decesos\"")
	for _, state := range sdata.States {
		if state.Name == "NACIONAL" {
			continue
		}
		fmt.Printf("  %-20s , %-15d , %-15d , %-17d , %-7d \n",
			state.Name, state.PositiveCases, state.NegativeCases, state.SuspectCases, state.Deaths)
	}
}

type CliConfig struct {
	showVersion  bool
	showHelp     bool
	exportFormat string
	source       string
	since        string
	municipio    string
}

func main() {
	fs := flag.NewFlagSet("covid19mx", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Printf("Usage: covid19mx [options...]\n\n")
		fs.PrintDefaults()
		fmt.Println()
	}

	// Top level global config
	config := &CliConfig{}
	fs.BoolVar(&config.showHelp, "h", false, "Show help")
	fs.BoolVar(&config.showHelp, "help", false, "Show help")
	fs.BoolVar(&config.showVersion, "version", false, "Show version")
	fs.BoolVar(&config.showVersion, "v", false, "Show version")
	fs.StringVar(&config.exportFormat, "o", "", "Export format (options: json, csv, table)")
	fs.StringVar(&config.source, "source", "", "Source of the data")
	fs.StringVar(&config.since, "since", "", "Date against which to compare the data")
	fs.StringVar(&config.municipio, "municipio", "", "Municipio used to narrow down data")
	fs.StringVar(&config.municipio, "mun", "", "Municipio used to narrow down data")
	fs.Parse(os.Args[1:])

	switch {
	case config.showHelp:
		flag.Usage()
		os.Exit(0)
	case config.showVersion:
		fmt.Printf("covid19mx v%s\n", version)
		fmt.Printf("Release-Date %s\n", releaseDate)
		os.Exit(0)
	}

	if config.municipio != "" {
		// Try to fetch by municipal data instead.
		pCases, err := fetchMunicipalData(municipalURL, "Confirmados")
		if err != nil {
			log.Fatal(err)
		}
		nCases, err := fetchMunicipalData(municipalURL, "Negativos")
		if err != nil {
			log.Fatal(err)
		}
		sCases, err := fetchMunicipalData(municipalURL, "Sospechosos")
		if err != nil {
			log.Fatal(err)
		}
		log.Println(len(pCases), len(nCases), len(sCases))
		muns := make(map[string]Municipio)
		for k, v := range pCases {
			muns[k] = Municipio{
				PositiveCases: v,
			}
		}
		for k, v := range nCases {
			m := muns[k]
			m.NegativeCases = v
			muns[k] = m
		}
		for k, v := range sCases {
			m := muns[k]
			m.SuspectCases = v
			muns[k] = m
		}
		for s, m := range muns {
			fmt.Println(s, m)
		}
	}

	var (
		sdata *SinaveData
		err   error
	)
	if strings.Contains(config.source, ".json") {
		// Use a local file as the source
		data, err := ioutil.ReadFile(config.source)
		if err != nil {
			log.Fatal(err)
		}
		type s struct {
			States []State `json:"states"`
		}
		var sd *s
		err = json.Unmarshal(data, &sd)
		if err != nil {
			log.Fatal(err)
		}
		sdata = &SinaveData{
			States: sd.States,
		}
	} else {
		// Get latest sinave data by default.  Can also use a local checked
		// version for the data or an explicit http endpoint.
		if config.source == "" {
			config.source = attackRateURL
		}
		sdata, err = fetchData(config.source)
		if err != nil {
			log.Fatal(err)
		}
	}

	if config.since != "" {
		var date time.Time
		switch config.since {
		case "-1d", "1d", "yesterday":
			date = time.Now().AddDate(0, 0, -1)
		case "-2d", "2d", "2 days ago":
			date = time.Now().AddDate(0, 0, -2)
		default:
			days, err := strconv.Atoi(config.since)
			if err != nil {
				log.Fatal(err)
			}
			date = time.Now().AddDate(0, 0, days*-1)
		}
		pdata, err := fetchPastData(repoURL + date.Format("2006-01-02") + ".json")
		if err != nil {
			log.Fatal(err)
		}
		showTableDiff(sdata, pdata)
	} else {
		switch config.exportFormat {
		case "csv":
			showCSV(sdata)
		case "json":
			showJSON(sdata)
		case "table":
			showTable(sdata)
		case "awk":
			showTableAwkFriendly(sdata)
		default:
			showTable(sdata)
		}
	}
}
