package main

import (
	"LINECodingTest/timeutil"
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

//Record : 走行記録を格納する構造体
type Record struct {
	time    string
	mileage string
}

//TotimeForCompute : Record内のtimeを計算できるように変換
func (record *Record) TotimeForCompute() timeutil.Time {
	parsedTime := strings.Split(record.time, ":")
	parsedHour, _ := strconv.Atoi(parsedTime[0])
	parsedMinute, _ := strconv.Atoi(parsedTime[1])
	parsedSecond, _ := strconv.ParseFloat(parsedTime[2], 64)
	return timeutil.NewTime(uint(parsedHour), uint(parsedMinute), parsedSecond)
}

var errorRecord = Record{time: "error", mileage: "error"}

func main() {

	var input, err = readInput()

	if err != nil {
		os.Exit(-1)
	}

	fmt.Println(computeFare(input))
	os.Exit(0)

}

//走行記録を1行ずつ読み込む
func readInput() ([]Record, error) {
	var scanner = bufio.NewScanner(os.Stdin)
	var sliced = []string{}
	var records = []Record{}

	for scanner.Scan() {
		sliced = strings.Split(scanner.Text(), " ")

		//入力内容のチェック

		//sliced[0]が空白の場合、入力終了とみなしてこれまでの走行記録を返して終了する
		//ただし、走行記録が2行以上ない時は異常なのでエラーを示すRecordを返す
		if sliced[0] == "" {
			//入力が時系列順になっていない時もエラーなのでエラーを示すRecordを返す
			if len(records) >= 2 {
				prevTime := records[len(records)-2].TotimeForCompute()
				currentTime := records[len(records)-1].TotimeForCompute()
				if timeutil.IsEarlier(prevTime, currentTime) == false {
					return []Record{errorRecord}, errors.New("inputs must be in chronological order")
				}

			}
			if len(records) < 2 {
				return []Record{errorRecord}, errors.New("length of input must be 2 or lines")
			}
			//全て正しく入力されていた場合はrecordsを返して終了する
			return records, nil
		}
		//時刻の入力が正しいかチェックし、間違っていたらエラーを示すRecordを返す
		if regexp.MustCompile(`[0-9]{2}:[0-9]{2}:[0-9]{2}.[0-9]{3}`).Match([]byte(sliced[0])) == false {
			return []Record{errorRecord}, errors.New("input of the time is incorrect")
		}
		//走行距離の入力が正しいかチェックし、間違っていたらエラーを示すRecordを返す
		if regexp.MustCompile(`[0-9]+.[0-9]`).Match([]byte(sliced[1])) == false {
			return []Record{errorRecord}, errors.New("input of mileage is incorrect")
		}
		//1行目の走行距離の記録が"0.0"でない時は異常なのでエラーを示すRecordを返す
		if len(records) == 0 {
			if sliced[1] != "0.0" {
				return []Record{errorRecord}, errors.New("mileage of the first line must be '0.0'")
			}
		}

		//ここから正常に入力された時の処理
		//時刻にsliced[0]、走行距離にsliced[1]を代入し、recordsに追加
		records = append(records, Record{time: sliced[0], mileage: sliced[1]})

		if len(records) >= 2 {
			prevTime := records[len(records)-2].TotimeForCompute()
			currentTime := records[len(records)-1].TotimeForCompute()
			//入力が時系列順になっていない時もエラーなのでエラーを示すRecordを返す
			if timeutil.IsEarlier(prevTime, currentTime) == false {
				return []Record{errorRecord}, errors.New("inputs must be in chronological order")
			}

		}

	}

	//正常にEOFに到達できないときはエラーを示すRecordを返して終了する
	if scanner.Err() != nil {
		return []Record{errorRecord}, errors.New("unknown error has occurred")
	}
	//何らかの理由でforループを抜けてしまった時も異常なのでエラーを示すRecordを返して終了する
	return []Record{errorRecord}, errors.New("unknown error has occurred")
}

//距離から運賃を計算する
func computeFare(records []Record) uint {
	//初乗り運賃
	const baseFare uint = 410
	//加算運賃
	const additionalFare uint = 80
	//低速運賃
	const additionalLowSpeedFare uint = 80
	//運賃が加算される距離
	const additionalFareSection uint = 237
	//初乗り区間(m)
	const baseFareSection float64 = 1052
	//低速走行の判断基準(m/秒):10000 / 3600 = 2.7777....
	const LowSpeedLimit = 10000.0 / 3600.0

	//走行距離
	var mileage float64 = 0
	//総低速走行時間(秒)
	var lowSpeedRunningTime float64 = 0

	//運賃
	var fare uint = baseFare

	//総走行距離を求める
	//iは1から始まる
	//理由:1. 1行目は0mであることが保証されているので加算する意味がない
	//     2. [i-1]と[i]で時間差を求めるため
	for i := 1; i < len(records); i++ {
		runningTime, singleMileage := computeTimeAndMileage(records[i-1], records[i])
		//速度が低速走行の判断基準以下の場合、低速走行時間に走行時間を加える
		//移動距離が0 = 速度が0の場合も、同様に低速走行時間だと判断する
		if float64(singleMileage)/runningTime <= LowSpeedLimit ||
			singleMileage == 0 {
			lowSpeedRunningTime += runningTime
		}

		mileage += float64(singleMileage)

	}
	//低速運賃を加算
	fare += (uint(lowSpeedRunningTime / 90)) * 80

	//初乗り区間内はfareの初期値+低速走行運賃を運賃とする
	if mileage <= float64(baseFareSection) {

		return fare
	}
	//初乗り区間外の場合、走行距離から初乗り区間を引き、運賃が加算される距離での商(小数点以下切り上げ) * 加算運賃を運賃に足す
	fareAdded := (math.Ceil(math.Ceil(mileage-baseFareSection) / float64(additionalFareSection))) * float64(additionalFare)
	fare += uint(fareAdded)

	return fare

}

//NightStart : 夜間時間開始(Time{hour: 22, minute: 0, second: 0})
var NightStart = timeutil.NewTime(22, 0, 0)

//NightEnd : 夜間時間終了(Time{hour: 5, minute: 0, second: 0})
var NightEnd = timeutil.NewTime(5, 0, 0)

//nightTimeRate : 夜間時の距離、走行時間の倍率(1.25倍)
const nightTimeRate float64 = 1.25

//computeComplicatedCase :  通常時間と夜間時間が混ざっている場合の計算を行う
func computeComplicatedCase(startRecord Record, endRecord Record) (float64, float64) {

	var startTime = startRecord.TotimeForCompute()
	var endTime = endRecord.TotimeForCompute()
	Mileage, _ := strconv.ParseFloat(endRecord.mileage, 64)

	if timeutil.IsInNight(startTime) == false {
		timeDifference := timeutil.ComputeTimeDifference(startTime, endTime)
		//(endTime.Hour)-22)/24を小数点以下切り捨てし、7*3600をかけて秒数にする
		nightTime := math.Floor(float64(float64(endTime.Hour)-22)/24) * 7 * 3600
		//余りの時間をnightTimeに加算
		if 0 <= ((int(endTime.Hour)-22)%24) && ((int(endTime.Hour)-22)%24) <= 6 {
			nightTime += float64((int(endTime.Hour)-22)%24)*3600 + float64(endTime.Minute)*60 + endTime.Second
		}
		if ((int(endTime.Hour)-22)%24) == 7 && endTime.Minute == 0 && endTime.Second == 0 {
			nightTime += float64((int(endTime.Hour)-22)%24)*3600 + float64(endTime.Minute)*60 + endTime.Second
		}
		dayTime := timeDifference - nightTime
		//全体の時間の割合から距離を求める
		dayMileage := Mileage * (dayTime / timeDifference)
		nightMileage := Mileage * (nightTime / timeDifference)
		//夜間時間・距離を1.25倍
		nightTime *= nightTimeRate
		nightMileage *= nightTimeRate
		//距離・時間の合計を返す
		return dayTime + nightTime, dayMileage + nightMileage

	}
	//翌日の5時
	nightendTomorrow := timeutil.NewTime(29, 0, 0)
	//翌5時までの時間を夜間時間に加える
	nightTime := timeutil.ComputeTimeDifference(startTime.RevisedTime(), nightendTomorrow)
	//基準となる時間を求めるために24時間表記に直す
	revisedStartTime := startTime.RevisedTime()
	//基準となる時間
	criterionTime, err := nightendTomorrow.SubtractTime(revisedStartTime.Hour, revisedStartTime.Minute, revisedStartTime.Second)
	if err != nil {
		os.Exit(-1)
	}
	//始点時刻を加算
	criterionTime = criterionTime.AddTime(startTime.Hour, startTime.Minute, startTime.Second)

	//全体の時間差を求める
	timeDifference := timeutil.ComputeTimeDifference(criterionTime, endTime)
	//(endTime.Hour)-22)/24を小数点以下切り捨てし、7*3600をかけて秒数にする
	nightTime += math.Floor(float64(float64(endTime.Hour)-22)/24) * 7 * 3600
	//余りの時間をnightTimeに加算
	if 0 <= ((int(endTime.Hour)-22)%24) && ((int(endTime.Hour)-22)%24) <= 6 {
		nightTime += float64((int(endTime.Hour)-22)%24)*3600 + float64(endTime.Minute)*60 + endTime.Second
	}
	if ((int(endTime.Hour)-22)%24) == 7 && endTime.Minute == 0 && endTime.Second == 0 {
		nightTime += float64((int(endTime.Hour)-22)%24)*3600 + float64(endTime.Minute)*60 + endTime.Second
	}
	//全体の時間差-夜間時間で通常時間が分かる
	dayTime := timeDifference - nightTime
	//全体の時間の割合から距離を求める
	dayMileage := Mileage * (dayTime / timeDifference)
	nightMileage := Mileage * (nightTime / timeDifference)
	//夜間時間・距離を1.25倍
	nightTime *= nightTimeRate
	nightMileage *= nightTimeRate
	//距離・時間の合計を返す
	return dayTime + nightTime, dayMileage + nightMileage

}

//走行時間(秒)と走行距離(m)を求める
//夜間割増もここで処理する
func computeTimeAndMileage(startRecord Record, endRecord Record) (float64, float64) {

	var startTime = startRecord.TotimeForCompute()
	var endTime = endRecord.TotimeForCompute()
	//startTimeが通常時間で、かつendTimeが通常時間で夜間時間を跨がない場合は走行時間と走行距離をそのまま返す
	//5時から始まり、22時に終わる場合も同様なので走行時間と走行距離をそのまま返す
	if (timeutil.IsInNight(startTime) == false && timeutil.IsInNight(endTime) == false) &&
		timeutil.IsStraddledNight(startTime, endTime) == false ||
		(reflect.DeepEqual(startTime.RevisedTime(), NightEnd) == true && reflect.DeepEqual(endTime.RevisedTime(), NightStart) == true) {
		runningTime := timeutil.ComputeTimeDifference(startTime, endTime)
		runingMileage, _ := strconv.ParseFloat(endRecord.mileage, 64)
		return runningTime, runingMileage

	} else if timeutil.IsInNight(startTime) == true && timeutil.IsInNight(endTime) == true &&
		timeutil.IsStraddledDay(startTime, endTime) == false ||
		(reflect.DeepEqual(startTime.RevisedTime(), NightStart) == true && reflect.DeepEqual(endTime.RevisedTime(), NightEnd) == true) {
		//startTimeがNightStartより遅く、かつendTimeがNightEndより早くて通常時間を跨がない場合は走行時間と走行距離を1.25倍して返す
		//startTimeがNightStartと一致し、かつendTimeがNightEndと一致する場合も同様なのでその時も同様に処理する

		var nightTime = timeutil.ComputeTimeDifference(startTime, endTime) * nightTimeRate
		var nightMileage, _ = strconv.ParseFloat(endRecord.mileage, 64)
		nightMileage *= nightTimeRate
		return nightTime, nightMileage

	} else {
		//どこかで通常時間と夜間時間を跨ぐ時は夜間走行時間と夜間走行距離を別途で求め、通常時間の走行時間と走行距離に加える
		runningTime, runingMileage := computeComplicatedCase(startRecord, endRecord)
		return runningTime, runingMileage

	}

}
