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
const 수서동 = "1168011500"
const 세곡동 = "1168011100"
const 율현동 = "1168011300"
const 자곡동 = "1168011200"
const 문정동 = "1171010800"
const 장지동 = "1171010900"

func main() {
	limiter := util.GetRateLimiter(200*time.Millisecond, 1)
	var svc service.EstateService = service.NewEstateServiceWithRateLimiter(limiter, new(service.EstateServiceLive))
	//regions, err := svc.GetRegionList("1168000000")
	//if err != nil {
	//	panic(err)
	//}
	//for _, r := range regions.List {
	//	fmt.Printf("%s => %s\n", r.CortarNo, r.CortarNm)
	//}
	//return

	result := []AResult{}

	regionNos := []string{개포동, 수서동, 세곡동, 율현동, 자곡동, 문정동, 장지동}

	complexNoList := make([]string, 0)

	for _, r := range regionNos {
		//result = append(result, GetResultList(&svc, t)...)
		complexListResult, err := svc.GetComplexList(r)
		util.PanicError(err, "GetComplexNoList Error")
		for _, c := range complexListResult.Result {
			complexNoList = append(complexNoList, c.HscpNo)
		}
	}

	time.Sleep(15 * time.Second)

	for i, no := range complexNoList {
		if i%10 == 0 {
			fmt.Fprintf(os.Stderr, "%d start \n", i)
			time.Sleep(1 * time.Minute)
		}
		articles, err := ComplexArticleListResultGetter(&svc, no, 1100000, 9000000)
		util.PanicError(err, "GetComplexNoList Error")
		for _, m := range articles {
			price := parsePrice(m.PrcInfo)
			result = append(result, AResult{m.AtclNo, m.AtclNm, m.BildNm, m.FlrInfo, m.Spc2, price, m.CfmYmd})
		}
	}

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
