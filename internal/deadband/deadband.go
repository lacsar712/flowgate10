package deadband

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Rec struct {
	Title, Body string
	Tags        []string
}

func Sample() Rec {
	return Rec{Title: "Kiln.Temp", Body: "eu=842.0 deadband=2.5", Tags: []string{"Kiln.Temp"}}
}

func Seed() []Rec {
	return []Rec{
		Sample(),
		{Title: "Mill.RPM", Body: "eu=1200 deadband=5", Tags: []string{"Mill.RPM"}},
	}
}

func AfterWrite(getMin func() (string, error), setMin func(string) error, body string) error {
	eu, _, err := parse(body)
	if err != nil {
		return err
	}
	cur, err := getMin()
	if err == nil && strings.TrimSpace(cur) != "" {
		last, conv := strconv.ParseFloat(strings.TrimSpace(cur), 64)
		if conv == nil && eu+1e-9 < last {
			return fmt.Errorf("EU %v is below last committed %v", eu, last)
		}
	}
	return setMin(strconv.FormatFloat(eu, 'f', -1, 64))
}

func Steps() []string { return []string{"deadband-check", "index-nodes", "export-items"} }

func Enforce(title, body string, tags []string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("node title required")
	}
	eu, db, err := parse(body)
	if err != nil {
		return err
	}
	if db < 0 {
		return fmt.Errorf("deadband must be >= 0")
	}
	limit := math.Abs(eu) * 0.10
	if math.Abs(eu) < 1e-9 {
		if db != 0 {
			return fmt.Errorf("zero EU requires deadband 0")
		}
		return nil
	}
	if db > limit {
		return fmt.Errorf("deadband %v exceeds 10%% of |eu| (%v)", db, limit)
	}
	if len(tags) == 0 {
		return fmt.Errorf("browse name tag required")
	}
	return nil
}

func parse(body string) (eu, db float64, err error) {
	gotE, gotD := false, false
	for _, part := range strings.Fields(body) {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		n, conv := strconv.ParseFloat(v, 64)
		if conv != nil {
			return 0, 0, fmt.Errorf("bad %s", part)
		}
		switch k {
		case "eu":
			eu, gotE = n, true
		case "deadband":
			db, gotD = n, true
		}
	}
	if !gotE || !gotD {
		return 0, 0, fmt.Errorf("body must contain eu= and deadband=")
	}
	return eu, db, nil
}
