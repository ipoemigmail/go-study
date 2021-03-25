package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// https://m.land.naver.com/map/getRegionList?cortarNo=1168000000&mycortarNo=
type RegionListResponse struct {
	Result RegionListResult `json:"result"`
}

type RegionListResult struct {
	List     []Region `json:"list"`
	DvsnInfo Region   `json:"dvsnInfo"`
	CityInfo Region   `json:"cityInfo"`
}

type Region struct {
	CortarNo   string `json:"CortarNo"`
	CortarNm   string `json:"CortarNm"`
	MapXCrdn   string `json:"MapXCrdn"`
	MapYCrdn   string `json:"MapYCrdn"`
	CortarType string `json:"CortarType"`
}

// https://m.land.naver.com/complex/ajax/complexListByCortarNo?cortarNo=1168010300
type ComplexListResult struct {
	Result   []Complex `json:"result"`
	SecInfo  Region    `json:"secInfo"`
	DvsnInfo Region    `json:"dvsnInfo"`
	LoginYN  string    `json:"loginYN"`
	CityInfo Region    `json:"cityInfo"`
}

type Complex struct {
	HscpNo      string `json:"hscpNo"`
	HscpNm      string `json:"hscpNm"`
	HscpTypeCd  string `json:"hscpTypeCd"`
	HscpTypeNm  string `json:"hscpTypeNm"`
	Lat         string `json:"lat"`
	Lng         string `json:"lng"`
	CortarNo    string `json:"cortarNo"`
	DealCnt     int    `json:"dealCnt"`
	LeaseCnt    int    `json:"leaseCnt"`
	RentCnt     int    `json:"rentCnt"`
	StrmrentCnt int    `json:"strmRentCnt"`
	HasBookMark int    `json:"hasBookMark"`
}

// https://m.land.naver.com/complex/getComplexArticleList?hscpNo=8928&tradTpCd=A1&order=price&showR0=N&page=1
// https://m.land.naver.com/cluster/clusterList?view=atcl&rletTpCd=OBYG%3AABYG%3AOPST%3AAPT&tradTpCd=A1&z=18&lat=37.47848&lon=127.054031&btm=37.4763855&lft=127.05052&top=37.4805745&rgt=127.057542&dprcMin=110000&dprcMax=900000000&pCortarNo=&addon=COMPLEX&bAddon=COMPLEX&isOnlyIsale=false
type ComplexArticleListResponse struct {
	Result ComplexArticleListResult `json:"result"`
}

type ComplexArticleListResult struct {
	List          []ComplexArticle `json:"list"`
	TotAtclCnt    int              `json:"totAtclCnt"`
	MoreDataYn    string           `json:"moreDataYn"`
	ShowGuarantee bool             `json:"showGuarantee"`
}

type ComplexArticle struct {
	RepimgUrl         string   `json:"repImgUrl"`
	AtclNo            string   `json:"atclNo"`
	RepImgTpCd        string   `json:"repImgTpCd"`
	VrfcTpCd          string   `json:"vrfcTpCd"`
	AtclNm            string   `json:"atclNm"`
	BildNm            string   `json:"bildNm"`
	TradTpCd          string   `json:"tradTpCd"`
	TradTpNm          string   `json:"tradTpNm"`
	RletTpCd          string   `json:"rletTpCd"`
	RletTpNm          string   `json:"rletTpNm"`
	Spc1              string   `json:"spc1"`
	Spc2              string   `json:"spc2"`
	FlrInfo           string   `json:"flrInfo"`
	CfmYmd            string   `json:"cfmYmd"`
	PrcInfo           string   `json:"prcInfo"`
	SameAddrCnt       int      `json:"sameAddrCnt"`
	SameAddrDirectCnt int      `json:"sameAddrDirectCnt"`
	SameAddrHash      string   `json:"sameAddrHash"`
	SameAddrMaxPrc    string   `json:"sameAddrMaxPrc"`
	SameAddrMinPrc    string   `json:"sameAddrMinPrc"`
	TradCmplYn        string   `json:"tradCmplYn"`
	TagList           []string `json:"tagList"`
	AtclStatCd        string   `json:"atclStatCd"`
	CpId              string   `json:"cpid"`
	CpNm              string   `json:"cpNm"`
	CpCnt             int      `json:"cpCnt"`
	CpLinkVO          struct {
		CpId                               string `json:"cpId"`
		MobileArticleLinkTypeCode          string `json:"mobileArticleLinkTypeCode"`
		MobileBmsInspectPassYn             string `json:"mobileBmsInspectPassYn"`
		PcArticleLinkUseAtArticleTitle     bool   `json:"pcArticleLinkUseAtArticleTitle"`
		PcArticleLinkUseAtCpName           bool   `json:"pcArticleLinkUseAtCpName"`
		MobileArticleLinkUseAtArticleTitle bool   `json:"mobileArticleLinkUseAtArticleTitle"`
		MobileArticleLinkUseAtCpName       bool   `json:"mobileArticleLinkUseAtCpName"`
	} `json:"cpLinkVO"`
	RltrNm              string  `json:"rltrNm"`
	DirectTradYn        string  `json:"directTradYn"`
	Direction           string  `json:"direction"`
	TradePriceHan       string  `json:"tradePriceHan"`
	TradeRentPrice      int     `json:"tradeRentPrice"`
	TradePriceInfo      string  `json:"tradePriceInfo"`
	TradeCheckedByOwner bool    `json:"tradeCheckedByOwner"`
	Point               int     `json:"point"`
	DtlAddr             string  `json:"dtlAddr"`
	DtlAddrYn           string  `json:"dtlAddrYn"`
	AtclFetrDesc        *string `json:"atclFetrDesc,omitempty"`
}

type EstateError struct {
	Msg  string
	From error
}

func newEstateError(msg string, from error) EstateError {
	return EstateError{Msg: msg, From: from}
}

func (e EstateError) Error() string {
	return fmt.Sprintf("%s\nfrom %s", e.Msg, e.From)
}

type EstateService interface {
	GetRegionList(parentRegionNo string) (RegionListResult, error)
	GetComplexList(regionNo string) (ComplexListResult, error)
	GetComplexArticleList(complexNo string, minPrice uint, maxPrice uint, page uint) (ComplexArticleListResult, error)
}

type EstateServiceLive struct{}

const baseURL = "https://m.land.naver.com"

func (s *EstateServiceLive) GetRegionList(parentRegionNo string) (result RegionListResult, err error) {
	url := fmt.Sprintf("%s/map/getRegionList?cortarNo=%s", baseURL, parentRegionNo)
	resp, err := http.Get(url)
	if err != nil {
		err = newEstateError("GetRegionList(http call error)", err)
		return
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = newEstateError("GetRegionList(http body get error)", err)
		return
	}
	defer resp.Body.Close()
	var response RegionListResponse
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		err = newEstateError("GetRegionList(http body parse error)", err)
		return
	}
	result = response.Result
	return
}

func (s *EstateServiceLive) GetComplexList(regionNo string) (result ComplexListResult, err error) {
	url := fmt.Sprintf("%s/complex/ajax/complexListByCortarNo?cortarNo=%s", baseURL, regionNo)
	resp, err := http.Get(url)
	if err != nil {
		err = newEstateError("GetComplexList(http call error)", err)
		return
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = newEstateError("GetComplexList(http body get error)", err)
		return
	}
	defer resp.Body.Close()
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		err = newEstateError("GetComplexList(http body parse error)", err)
		return
	}
	return
}

func (s *EstateServiceLive) GetComplexArticleList(complexNo string, minPrice uint, maxPrice uint, page uint) (result ComplexArticleListResult, err error) {
	url := fmt.Sprintf("%s/complex/getComplexArticleList?hscpNo=%s&tradTpCd=A1&order=price&dprcMin=%d&dprcMax=%d&showR0=N&page=%d", baseURL, complexNo, minPrice, maxPrice, page)
	resp, err := http.Get(url)
	if err != nil {
		err = newEstateError("GetComplexArticleListByPage(http call error)", err)
		return
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = newEstateError("GetComplexArticleListByPage(http body get error)", err)
		return
	}
	defer resp.Body.Close()
	var response ComplexArticleListResponse
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		err = newEstateError("GetComplexArticleListByPage(http body parse error)", err)
		return
	}
	result = response.Result
	return
}

type EstateServiceWithRateLimiter struct {
	rateLimiter   <-chan time.Time
	estateService EstateService
}

func NewEstateServiceWithRateLimiter(rateLimiter <-chan time.Time, estateService EstateService) *EstateServiceWithRateLimiter {
	return &EstateServiceWithRateLimiter{
		rateLimiter,
		estateService,
	}
}

func (s *EstateServiceWithRateLimiter) GetRegionList(parentRegionNo string) (result RegionListResult, err error) {
	<-s.rateLimiter
	return s.estateService.GetRegionList(parentRegionNo)
}

func (s *EstateServiceWithRateLimiter) GetComplexList(regionNo string) (result ComplexListResult, err error) {
	<-s.rateLimiter
	return s.estateService.GetComplexList(regionNo)
}

func (s *EstateServiceWithRateLimiter) GetComplexArticleList(complexNo string, minPrice uint, maxPrice uint, page uint) (result ComplexArticleListResult, err error) {
	<-s.rateLimiter
	return s.estateService.GetComplexArticleList(complexNo, minPrice, maxPrice, page)
}
