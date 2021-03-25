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
	limiter := util.GetRateLimiter(100*time.Millisecond, 1)
	var svc service.EstateService = service.NewEstateServiceWithRateLimiter(limiter, new(service.EstateServiceLive))
	//regions, err := svc.GetRegionList("1171000000")
	//if err != nil {
	//	panic(err)
	//}
	//for _, r := range regions {
	//	fmt.Printf("%s => %s\n", r.Cortarno, r.Cortarnm)
	//}

	//return

	result := []AResult{}

	result1 := GetResultList(&svc, 세곡동)
	result2 := GetResultList(&svc, 세곡동)
	result3 := GetResultList(&svc, 율현동)
	result4 := GetResultList(&svc, 자곡동)
	result5 := GetResultList(&svc, 문정동)
	result6 := GetResultList(&svc, 장지동)
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

func ComplexArticleListResultGetter(svc *service.EstateService, complexNo string, minPrice uint, maxPrice uint) (result []service.ComplexArticle, err error) {
	var page uint = 1
	result = make([]service.ComplexArticle, 0)
	for {
		var iResult service.ComplexArticleListResult
		iResult, err = (*svc).GetComplexArticleList(complexNo, minPrice, maxPrice, page)
		if err != nil {
			return
		}
		result = append(result, iResult.List...)
		if iResult.MoreDataYn == "Y" {
			page++
		} else {
			break
		}
	}
	return result, nil
}

func GetResultList(svc *service.EstateService, regionNo string) []AResult {
	complexes, err := (*svc).GetComplexList(regionNo)
	fmt.Fprintf(os.Stderr, "len: %d\n", len(complexes.Result))
	util.PanicError(err, "GetResultList Error")
	chs := make([]chan Result, len(complexes.Result))
	for i, r := range complexes.Result {
		chs[i] = make(chan Result)
		go func(no string, resultCh chan Result) {
			r, err := ComplexArticleListResultGetter(svc, no, 1100000, 9000000)
			resultCh <- Result{value: r, err: err}
		}(r.HscpNo, chs[i])
	}

	result := []AResult{}

	for _, c := range chs {
		r := <-c
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "%s", r.err)
		}
		for _, m := range r.value {
			if filter(&m) {
				price := parsePrice(m.PrcInfo)
				result = append(result, AResult{m.AtclNo, m.AtclNm, m.BildNm, m.FlrInfo, m.Spc2, price, m.CfmYmd})
			}
		}
	}

	return result
}

func filter(c *service.ComplexArticle) bool {
	return c.TradTpCd == "A1" &&
		c.RletTpCd == "A01" &&
		!strings.Contains(c.AtclNm, "세곡리엔파크") &&
		!strings.Contains(c.AtclNm, "강남데시앙파크") &&
		!strings.Contains(c.AtclNm, "송파파인타운") &&
		!strings.Contains(c.AtclNm, "강남신동아파밀리에")
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
