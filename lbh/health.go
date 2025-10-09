package lbh

import (
	"fmt"
	"math"
	"strings"

	"kosis/pkg/kosis"
)

// 테이블 레퍼런스 (필요시 flag/환경변수로 치환 가능)
var (
	TableRegionEstEmp = kosis.TableRef{OrgID: "118", TblID: "DT_SAUP120"}       // 시군구별 사업체/종사자
	TableByEmpSize    = kosis.TableRef{OrgID: "213", TblID: "DT_21303_D000005"} // 시군×종사자규모별 사업체
)

type RegionMetric struct {
	RegionCode string  `json:"region_code"`
	RegionName string  `json:"region_name"`
	Year       int     `json:"year"`
	YoYEst     float64 `json:"yoy_establishments"`
	YoYEmp     float64 `json:"yoy_employment"`
	MicroShare float64 `json:"micro_share"`
	Score      float64 `json:"score"`
}

func RegionRanking(
	k *kosis.Client,
	year int,
	microLabel string,
	wEst, wEmp, wMic float64,
) ([]RegionMetric, error) {

	// 1) 메타 취득
	metaRegion, err := k.GetMetaITM(TableRegionEstEmp)
	if err != nil {
		return nil, err
	}
	dmRegion := kosis.DigestITM(metaRegion)

	idxArea, ok := kosis.FindClassIndexByName(dmRegion.Classes, "행정구역", "지역", "시도", "시군", "시군구")
	if !ok {
		return nil, fmt.Errorf("cannot find region classifier in %s/%s", TableRegionEstEmp.OrgID, TableRegionEstEmp.TblID)
	}

	itemEst, ok := kosis.FindItemIDByContains(dmRegion.Items, "사업체수")
	if !ok {
		return nil, fmt.Errorf("cannot find item '사업체수'")
	}
	itemEmp, ok := kosis.FindItemIDByContains(dmRegion.Items, "종사자수")
	if !ok {
		return nil, fmt.Errorf("cannot find item '종사자수'")
	}

	var regionCodes, regionNames []string
	for code, name := range dmRegion.Classes[idxArea].Values {
		if containsAny(name, "합계", "전국") {
			continue
		}
		regionCodes = append(regionCodes, code)
		regionNames = append(regionNames, name)
	}

	//type pair struct{ cur, prev float64 }
	//estVals := make([]pair, len(regionCodes))
	//empVals := make([]pair, len(regionCodes))

	// sem := make(chan struct{}, 8)
	// var wg sync.WaitGroup
	// var mu sync.Mutex
	// var firstErr error

	// let's take it one by one

	for i, code := range regionCodes {
		fmt.Printf("obj: %v, name: %s, est: %s, emp: %s\n", map[int]string{idxArea + 1: code}, regionNames[i], itemEst, itemEmp)
		rowsEst, err := k.ParamData(TableRegionEstEmp, "Y", fmt.Sprintf("%d", year-1), fmt.Sprintf("%d", year), itemEst, map[int]string{1: "ALL", 2: "ALL", 3: "ALL", 4: "ALL"})
		if err != nil {
			return nil, err
		}
		fmt.Printf("rowsEst: %v\n", rowsEst)

	}
	// i := 0
	// code := regionCodes[i]
	// name := regionNames[i]

	// rowsEmp, err := k.ParamData(TableRegionEstEmp, "Y", fmt.Sprintf("%d", year-1), fmt.Sprintf("%d", year), itemEmp, map[int]string{idxArea + 1: code})
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Printf("rowsEmp: %v\n", rowsEmp)

	return nil, nil

	/*
		for i, code := range regionCodes {
			i := i
			code := code
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer func() { <-sem; wg.Done() }()
				obj := map[int]string{idxArea + 1: code}
				rowsEst, err1 := k.ParamData(TableRegionEstEmp, "Y", fmt.Sprintf("%d", year-1), fmt.Sprintf("%d", year), itemEst, obj)
				rowsEmp, err2 := k.ParamData(TableRegionEstEmp, "Y", fmt.Sprintf("%d", year-1), fmt.Sprintf("%d", year), itemEmp, obj)
				if err1 != nil || err2 != nil {
					mu.Lock()
					if firstErr == nil {
						if err1 != nil {
							firstErr = err1
						} else {
							firstErr = err2
						}
					}
					mu.Unlock()
					return
				}
				valEst := map[string]float64{}
				for _, r := range rowsEst {
					if v, ok := kosis.ParseNumber(r.DT); ok {
						valEst[r.PRDDE] = v
					}
				}
				valEmp := map[string]float64{}
				for _, r := range rowsEmp {
					if v, ok := kosis.ParseNumber(r.DT); ok {
						valEmp[r.PRDDE] = v
					}
				}

				curEst, ok1 := valEst[fmt.Sprintf("%d", year)]
				prevEst, ok2 := valEst[fmt.Sprintf("%d", year-1)]
				curEmp, ok3 := valEmp[fmt.Sprintf("%d", year)]
				prevEmp, ok4 := valEmp[fmt.Sprintf("%d", year-1)]

				if !(ok1 && ok2 && ok3 && ok4) {
					mu.Lock()
					estVals[i] = pair{math.NaN(), math.NaN()}
					empVals[i] = pair{math.NaN(), math.NaN()}
					mu.Unlock()
					return
				}
				mu.Lock()
				estVals[i] = pair{curEst, prevEst}
				empVals[i] = pair{curEmp, prevEmp}
				mu.Unlock()
			}()
		}
		wg.Wait()
		if firstErr != nil {
			return nil, firstErr
		}

		// 2) 규모별 테이블에서 초소규모 비중 계산
		metaSize, err := k.GetMetaITM(TableByEmpSize)
		if err != nil {
			return nil, err
		}
		dmSize := kosis.DigestITM(metaSize)
		idxArea2, ok := kosis.FindClassIndexByName(dmSize.Classes, "행정구역")
		if !ok {
			return nil, fmt.Errorf("cannot find region classifier in size table")
		}
		idxScale, ok := kosis.FindClassIndexByName(dmSize.Classes, "규모")
		if !ok {
			return nil, fmt.Errorf("cannot find size classifier")
		}
		itemEst2, ok := kosis.FindItemIDByContains(dmSize.Items, "사업체수")
		if !ok {
			return nil, fmt.Errorf("cannot find item '사업체수' in size table")
		}

		var codeTotal, codeMicro string
		for code, nm := range dmSize.Classes[idxScale].Values {
			if stringsContains(nm, "합계") {
				codeTotal = code
			}
			if stringsContains(nm, microLabel) {
				codeMicro = code
			}
		}
		if codeTotal == "" || codeMicro == "" {
			return nil, fmt.Errorf("size codes not found (total:%s micro:%s)", codeTotal, codeMicro)
		}

		microShares := make([]float64, len(regionCodes))
		sem2 := make(chan struct{}, 8)
		var wg2 sync.WaitGroup
		var mu2 sync.Mutex
		var firstErr2 error

		for i, code := range regionCodes {
			i := i
			code := code
			wg2.Add(1)
			sem2 <- struct{}{}
			go func() {
				defer func() { <-sem2; wg2.Done() }()
				obj := map[int]string{idxArea2 + 1: code, idxScale + 1: codeTotal}
				rowsTot, err1 := k.ParamData(TableByEmpSize, "Y", fmt.Sprintf("%d", year), fmt.Sprintf("%d", year), itemEst2, obj)
				obj[idxScale+1] = codeMicro
				rowsMic, err2 := k.ParamData(TableByEmpSize, "Y", fmt.Sprintf("%d", year), fmt.Sprintf("%d", year), itemEst2, obj)
				if err1 != nil || err2 != nil {
					mu2.Lock()
					if firstErr2 == nil {
						if err1 != nil {
							firstErr2 = err1
						} else {
							firstErr2 = err2
						}
					}
					mu2.Unlock()
					return
				}
				tot := math.NaN()
				mic := math.NaN()
				if len(rowsTot) > 0 {
					if v, ok := kosis.ParseNumber(rowsTot[0].DT); ok {
						tot = v
					}
				}
				if len(rowsMic) > 0 {
					if v, ok := kosis.ParseNumber(rowsMic[0].DT); ok {
						mic = v
					}
				}
				share := math.NaN()
				if !math.IsNaN(tot) && tot > 0 && !math.IsNaN(mic) {
					share = mic / tot
				}
				mu2.Lock()
				microShares[i] = share
				mu2.Unlock()
			}()
		}
		wg2.Wait()
		if firstErr2 != nil {
			return nil, firstErr2
		}

		// 3) 합성 + 점수
		out := make([]RegionMetric, 0, len(regionCodes))
		for i := range regionCodes {
			yoYEst := math.NaN()
			yoYEmp := math.NaN()
			if !math.IsNaN(estVals[i].cur) && !math.IsNaN(estVals[i].prev) && estVals[i].prev != 0 {
				yoYEst = (estVals[i].cur - estVals[i].prev) / estVals[i].prev
			}
			if !math.IsNaN(empVals[i].cur) && !math.IsNaN(empVals[i].prev) && empVals[i].prev != 0 {
				yoYEmp = (empVals[i].cur - empVals[i].prev) / empVals[i].prev
			}
			out = append(out, RegionMetric{
				RegionCode: regionCodes[i],
				RegionName: regionNames[i],
				Year:       year,
				YoYEst:     round4(yoYEst),
				YoYEmp:     round4(yoYEmp),
				MicroShare: round4(microShares[i]),
			})
		}
		scored := computeScores(out, wEst, wEmp, wMic)
		sort.Slice(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
		return scored, nil
	*/
}

func RegionByName(k *kosis.Client, name string, year int, microLabel string, wEst, wEmp, wMic float64) (*RegionMetric, error) {
	list, err := RegionRanking(k, year, microLabel, wEst, wEmp, wMic)
	if err != nil {
		return nil, err
	}
	for _, r := range list {
		if r.RegionName == name {
			return &r, nil
		}
	}
	for _, r := range list {
		if stringsContains(r.RegionName, name) || stringsContains(name, r.RegionName) {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("region not found: %s", name)
}

// ---- internal helpers ----

func zscores(vals []float64) []float64 {
	n := len(vals)
	if n == 0 {
		return nil
	}
	mu := 0.0
	for _, v := range vals {
		mu += v
	}
	mu /= float64(n)
	sd := 0.0
	for _, v := range vals {
		d := v - mu
		sd += d * d
	}
	sd = math.Sqrt(sd / float64(n))
	out := make([]float64, n)
	for i, v := range vals {
		if sd == 0 {
			out[i] = 0
		} else {
			out[i] = (v - mu) / sd
		}
	}
	return out
}

func computeScores(list []RegionMetric, wEst, wEmp, wMic float64) []RegionMetric {
	n := len(list)
	v1, v2, v3 := make([]float64, n), make([]float64, n), make([]float64, n)
	for i, r := range list {
		v1[i] = r.YoYEst
		v2[i] = r.YoYEmp
		v3[i] = r.MicroShare
	}
	z1, z2, z3 := zscores(v1), zscores(v2), zscores(v3)
	out := make([]RegionMetric, n)
	for i, r := range list {
		score := wEst*z1[i] + wEmp*z2[i] - wMic*z3[i]
		s := 50.0 + score*10.0 // 0..100 rough scaling
		if s < 0 {
			s = 0
		}
		if s > 100 {
			s = 100
		}
		r.Score = math.Round(s*10) / 10
		out[i] = r
	}
	return out
}

func round4(f float64) float64 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return f
	}
	return math.Round(f*10000) / 10000
}

func stringsContains(s, sub string) bool {
	return strings.Contains(s, sub)
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
