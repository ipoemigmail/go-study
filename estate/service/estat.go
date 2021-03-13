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
	Result struct {
		List     []Region `json:"list"`
		Dvsninfo Region   `json:"dvsnInfo"`
		Cityinfo Region   `json:"cityInfo"`
	} `json:"result"`
}

type Region struct {
	Cortarno   string `json:"CortarNo"`
	Cortarnm   string `json:"CortarNm"`
	Mapxcrdn   string `json:"MapXCrdn"`
	Mapycrdn   string `json:"MapYCrdn"`
	Cortartype string `json:"CortarType"`
}

// https://m.land.naver.com/complex/ajax/complexListByCortarNo?cortarNo=1168010300
type ComplexListResponse struct {
	Result  []Complex `json:"result"`
	Secinfo struct {
		Cortarno   string `json:"CortarNo"`
		Cortarnm   string `json:"CortarNm"`
		Mapxcrdn   string `json:"MapXCrdn"`
		Mapycrdn   string `json:"MapYCrdn"`
		Cortartype string `json:"CortarType"`
	} `json:"secInfo"`
	Dvsninfo Region `json:"dvsnInfo"`
	Loginyn  string `json:"loginYN"`
	Cityinfo Region `json:"cityInfo"`
}

type Complex struct {
	Hscpno      string `json:"hscpNo"`
	Hscpnm      string `json:"hscpNm"`
	Hscptypecd  string `json:"hscpTypeCd"`
	Hscptypenm  string `json:"hscpTypeNm"`
	Lat         string `json:"lat"`
	Lng         string `json:"lng"`
	Cortarno    string `json:"cortarNo"`
	Dealcnt     int    `json:"dealCnt"`
	Leasecnt    int    `json:"leaseCnt"`
	Rentcnt     int    `json:"rentCnt"`
	Strmrentcnt int    `json:"strmRentCnt"`
	Hasbookmark int    `json:"hasBookMark"`
}

// https://m.land.naver.com/complex/getComplexArticleList?hscpNo=8928&tradTpCd=A1&order=price&showR0=N&page=1
// https://m.land.naver.com/cluster/clusterList?view=atcl&rletTpCd=OBYG%3AABYG%3AOPST%3AAPT&tradTpCd=A1&z=18&lat=37.47848&lon=127.054031&btm=37.4763855&lft=127.05052&top=37.4805745&rgt=127.057542&dprcMin=110000&dprcMax=900000000&pCortarNo=&addon=COMPLEX&bAddon=COMPLEX&isOnlyIsale=false
type ComplexArticleListResponse struct {
	Result struct {
		List          []ComplexArticle `json:"list"`
		Totatclcnt    int              `json:"totAtclCnt"`
		Moredatayn    string           `json:"moreDataYn"`
		Showguarantee bool             `json:"showGuarantee"`
	} `json:"result"`
}

type ComplexArticle struct {
	Repimgurl         string   `json:"repImgUrl"`
	Atclno            string   `json:"atclNo"`
	Repimgtpcd        string   `json:"repImgTpCd"`
	Vrfctpcd          string   `json:"vrfcTpCd"`
	Atclnm            string   `json:"atclNm"`
	Bildnm            string   `json:"bildNm"`
	Tradtpcd          string   `json:"tradTpCd"`
	Tradtpnm          string   `json:"tradTpNm"`
	Rlettpcd          string   `json:"rletTpCd"`
	Rlettpnm          string   `json:"rletTpNm"`
	Spc1              string   `json:"spc1"`
	Spc2              string   `json:"spc2"`
	Flrinfo           string   `json:"flrInfo"`
	Cfmymd            string   `json:"cfmYmd"`
	Prcinfo           string   `json:"prcInfo"`
	Sameaddrcnt       int      `json:"sameAddrCnt"`
	Sameaddrdirectcnt int      `json:"sameAddrDirectCnt"`
	Sameaddrhash      string   `json:"sameAddrHash"`
	Sameaddrmaxprc    string   `json:"sameAddrMaxPrc"`
	Sameaddrminprc    string   `json:"sameAddrMinPrc"`
	Tradcmplyn        string   `json:"tradCmplYn"`
	Taglist           []string `json:"tagList"`
	Atclstatcd        string   `json:"atclStatCd"`
	Cpid              string   `json:"cpid"`
	Cpnm              string   `json:"cpNm"`
	Cpcnt             int      `json:"cpCnt"`
	Cplinkvo          struct {
		Cpid                               string `json:"cpId"`
		Mobilearticlelinktypecode          string `json:"mobileArticleLinkTypeCode"`
		Mobilebmsinspectpassyn             string `json:"mobileBmsInspectPassYn"`
		Pcarticlelinkuseatarticletitle     bool   `json:"pcArticleLinkUseAtArticleTitle"`
		Pcarticlelinkuseatcpname           bool   `json:"pcArticleLinkUseAtCpName"`
		Mobilearticlelinkuseatarticletitle bool   `json:"mobileArticleLinkUseAtArticleTitle"`
		Mobilearticlelinkuseatcpname       bool   `json:"mobileArticleLinkUseAtCpName"`
	} `json:"cpLinkVO"`
	Rltrnm              string `json:"rltrNm"`
	Directtradyn        string `json:"directTradYn"`
	Direction           string `json:"direction"`
	Tradepricehan       string `json:"tradePriceHan"`
	Traderentprice      int    `json:"tradeRentPrice"`
	Tradepriceinfo      string `json:"tradePriceInfo"`
	Tradecheckedbyowner bool   `json:"tradeCheckedByOwner"`
	Point               int    `json:"point"`
	Dtladdr             string `json:"dtlAddr"`
	Dtladdryn           string `json:"dtlAddrYn"`
	Atclfetrdesc        string `json:"atclFetrDesc,omitempty"`
}

type EstateError struct {
	Msg  string
	From error
}

func newEstateError(msg string, from error) EstateError {
	return EstateError{Msg: msg, From: from}
}

func (e EstateError) Error() string {
	return e.Msg
}

type EstateService interface {
	GetRegionList(parentRegionNo string) ([]Region, error)
	GetComplexList(regionNo string) ([]Complex, error)
	GetComplexArticleList(complexNo string) ([]ComplexArticle, error)
}

type EstateServiceLive struct{}

const baseURL = "https://m.land.naver.com"

func (s *EstateServiceLive) GetRegionList(parentRegionNo string) ([]Region, error) {
	url := fmt.Sprintf("%s/map/getRegionList?cortarNo=%s", baseURL, parentRegionNo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, newEstateError("GetRegionList(http call error)", err)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newEstateError("GetRegionList(http body get error)", err)
	}
	defer resp.Body.Close()
	var result RegionListResponse
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, newEstateError("GetRegionList(http body parse error)", err)
	}
	return result.Result.List, nil
}

func (s *EstateServiceLive) GetComplexList(regionNo string) ([]Complex, error) {
	url := fmt.Sprintf("%s/complex/ajax/complexListByCortarNo?cortarNo=%s", baseURL, regionNo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, newEstateError("GetComplexList(http call error)", err)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newEstateError("GetComplexList(http body get error)", err)
	}
	defer resp.Body.Close()
	var result ComplexListResponse
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, newEstateError("GetComplexList(http body parse error)", err)
	}
	return result.Result, nil
}

func (s *EstateServiceLive) getComplexArticleListByPage(complexNo string, minPrice int, maxPrice int, page int) (ComplexArticleListResponse, error) {
	result := ComplexArticleListResponse{}
	url := fmt.Sprintf("%s/complex/getComplexArticleList?hscpNo=%s&tradTpCd=A1&order=price&dprcMin=%d&dprcMax=%d&showR0=N&page=%d", baseURL, complexNo, minPrice, maxPrice, page)
	resp, err := http.Get(url)
	if err != nil {
		return result, newEstateError("getComplexArticleListByPage(http call error)", err)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, newEstateError("getComplexArticleListByPage(http body get error)", err)
	}
	defer resp.Body.Close()
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return result, newEstateError("getComplexArticleListByPage(http body parse error)", err)
	}
	return result, nil
}

func (s *EstateServiceLive) GetComplexArticleList(complexNo string, minPrice int, maxPrice int) ([]ComplexArticle, error) {
	page := 1
	result := make([]ComplexArticle, 0)
	for {
		response, err := s.getComplexArticleListByPage(complexNo, minPrice, maxPrice, page)
		if err != nil {
			return nil, newEstateError("GetComplexArticleList(getComplexArticleListByPage call error)", err)
		}
		result = append(result, response.Result.List...)
		if response.Result.Moredatayn == "Y" {
			page++
		} else {
			break
		}
	}
	return result, nil
}

func (s *EstateServiceLive) GetComplexArticleListWithRateLimit(complexNo string, minPrice int, maxPrice int, rateLimiter <-chan time.Time) ([]ComplexArticle, error) {
	page := 1
	result := make([]ComplexArticle, 0)
	for {
		<-rateLimiter
		response, err := s.getComplexArticleListByPage(complexNo, minPrice, maxPrice, page)
		if err != nil {
			return nil, newEstateError("GetComplexArticleList(getComplexArticleListByPage call error)", err)
		}
		result = append(result, response.Result.List...)
		if response.Result.Moredatayn == "Y" {
			page++
		} else {
			break
		}
	}
	return result, nil
}
