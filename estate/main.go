package main

import (
	"fmt"
	"ipoemi/estate/service"
	"ipoemi/estate/util"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	value []service.ComplexArticle
	err   error
}

type AResult struct {
	No          string
	Name        string
	BuildName   string
	FloorInfo   string
	Spec2       string
	Price       uint64
	ConfirmDate string
}

const 개포동 = "1168010300"
const 세곡동 = "1168011100"
const 율현동 = "1168011300"
const 자곡동 = "1168011200"
const 문정동 = "1171010800"
const 장지동 = "1171010900"

func main() {
	//svc := service.EstateServiceLive{}
	//regions, err := svc.GetRegionList("1171000000")
	//if err != nil {
	//	panic(err)
	//}
	//for _, r := range regions {
	//	fmt.Printf("%s => %s\n", r.Cortarno, r.Cortarnm)
	//}

	//return

	result := []AResult{}

	limiter := util.GetRateLimiter(100*time.Millisecond, 1)

	result1 := GetResultList(limiter, 장지동)
	result2 := GetResultList(limiter, 세곡동)
	result3 := GetResultList(limiter, 율현동)
	result4 := GetResultList(limiter, 자곡동)
	result5 := GetResultList(limiter, 문정동)
	result6 := GetResultList(limiter, 장지동)
	result = append(result, result1...)
	result = append(result, result2...)
	result = append(result, result3...)
	result = append(result, result4...)
	result = append(result, result5...)
	result = append(result, result6...)

	grouped := make(map[string]AResult)

	for _, r := range result {
		key := fmt.Sprintf("%s|%s|%s", r.Name, r.BuildName, r.FloorInfo)
		if grouped[key].ConfirmDate <= r.ConfirmDate {
			grouped[key] = r
		}
	}

	groupedResult := []AResult{}
	for _, v := range grouped {
		groupedResult = append(groupedResult, v)
	}

	//groupedResult = result

	sort.Slice(groupedResult, func(i, j int) bool {
		return groupedResult[i].Price < groupedResult[j].Price
	})

	for _, k := range groupedResult {
		fmt.Printf("%s\t%s\t%s\t%s\t%d\t%s\n", k.Name, k.BuildName, k.Spec2, k.FloorInfo, k.Price, k.ConfirmDate)
	}
}

func GetResultList(limiter <-chan time.Time, regionNo string) []AResult {
	svc := service.EstateServiceLive{}
	complexes, err := svc.GetComplexList(regionNo)
	fmt.Fprintf(os.Stderr, "len: %d\n", len(complexes))
	util.PanicError(err, "GetResultList Error")
	chs := make([]chan Result, len(complexes))
	for i, r := range complexes {
		//fmt.Printf("%s => %s\n", r.Hscpno, r.Hscpnm)
		chs[i] = make(chan Result)
		go func(no string, resultCh chan Result) {
			<-limiter
			r, err := svc.GetComplexArticleListWithRateLimit(no, 1100000, 9000000, limiter)
			resultCh <- Result{value: r, err: err}
		}(r.Hscpno, chs[i])
	}

	result := []AResult{}

	for _, c := range chs {
		r := <-c
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "%s", r.err)
		}
		for _, m := range r.value {
			if filter(&m) {
				price := parsePrice(m.Prcinfo)
				result = append(result, AResult{m.Atclno, m.Atclnm, m.Bildnm, m.Flrinfo, m.Spc2, price, m.Cfmymd})
			}
		}
	}

	return result
}

func filter(c *service.ComplexArticle) bool {
	return c.Tradtpcd == "A1" &&
		c.Rlettpcd == "A01" &&
		!strings.Contains(c.Atclnm, "세곡리엔파크") &&
		!strings.Contains(c.Atclnm, "강남데시앙파크") &&
		!strings.Contains(c.Atclnm, "강남신동아파밀리에")
}

func parsePrice(s string) uint64 {
	v := strings.Split(s, "억")
	v1, err := strconv.ParseUint(strings.TrimSpace(v[0]), 10, 0)
	if err != nil {
		v1 = 0
	}
	v2, err := strconv.ParseUint(strings.ReplaceAll(strings.TrimSpace(v[1]), ",", ""), 10, 0)
	if err != nil {
		v2 = 0
	}
	return v1*100000 + v2*10
}
