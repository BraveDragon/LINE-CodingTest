package main

import (
	"bufio"
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
func (record *Record) TotimeForCompute() Time {
	parsedTime := strings.Split(record.time, ":")
	parsedHour, _ := strconv.Atoi(parsedTime[0])
	parsedMinute, _ := strconv.Atoi(parsedTime[1])
	parsedSecond, _ := strconv.ParseFloat(parsedTime[2], 64)
	return Time{hour: uint(parsedHour), minute: uint(parsedMinute), second: parsedSecond}
}

var errorRecord = Record{time: "error", mileage: "error"}

func main() {

	var input = readInput()

	if reflect.DeepEqual(input[0], errorRecord) == true {
		os.Exit(-1)
	} else {
		fmt.Println(computeFare(input))
		os.Exit(0)
	}

}

//走行記録を1行ずつ読み込む
func readInput() []Record {
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
				if isEarlier(prevTime, currentTime) == false {
					return []Record{errorRecord}
				}

			}
			if len(records) < 2 {
				return []Record{errorRecord}
			}
			//全て正しく入力されていた場合はrecordsを返して終了する
			return records
		}
		//時刻の入力が正しいかチェックし、間違っていたらエラーを示すRecordを返す
		if regexp.MustCompile(`[0-9]{2}:[0-9]{2}:[0-9]{2}.[0-9]{3}`).Match([]byte(sliced[0])) == false {
			return []Record{errorRecord}
		}
		//走行距離の入力が正しいかチェックし、間違っていたらエラーを示すRecordを返す
		if regexp.MustCompile(`[0-9]+.[0-9]`).Match([]byte(sliced[1])) == false {
			return []Record{errorRecord}
		}
		//1行目の走行距離の記録が"0.0"でない時は異常なのでエラーを示すRecordを返す
		if len(records) == 0 {
			if sliced[1] != "0.0" {
				return []Record{errorRecord}
			}
		}

		//ここから正常に入力された時の処理
		//時刻にsliced[0]、走行距離にsliced[1]を代入し、recordsに追加
		records = append(records, Record{time: sliced[0], mileage: sliced[1]})
		//入力が時系列順になっていない時もエラーなのでエラーを示すRecordを返す
		if len(records) >= 2 {
			prevTime := records[len(records)-2].TotimeForCompute()
			currentTime := records[len(records)-1].TotimeForCompute()
			if isEarlier(prevTime, currentTime) == false {
				return []Record{errorRecord}
			}

		}

	}

	//正常にEOFに到達できないときはエラーを示すRecordを返して終了する
	if scanner.Err() != nil {
		return []Record{errorRecord}
	}
	//何らかの理由でforループを抜けてしまった時は、これも異常なのでエラーを示すRecordを返して終了する
	return []Record{errorRecord}
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

//Time : 走行記録内の時間を格納
type Time struct {
	hour   uint
	minute uint
	//秒には小数点以下も入るのでfloat64型
	second float64
}

//AddTime : 引数で指定しただけ加えた時刻を返す
func (time *Time) AddTime(hour uint, minute uint, second float64) Time {
	tmphour := time.hour + hour
	tmpminute := time.minute + minute
	tmpsecond := time.second + second
	if tmpsecond >= 60 {
		tmpsecond -= 60
		tmpminute++
	}
	if tmpminute >= 60 {
		tmpminute -= 60
		tmphour++
	}

	return Time{hour: tmphour, minute: tmpminute, second: tmpsecond}
}

//SubtractTime : 引数で指定しただけ引いた時刻を返す
//time.hour < hourの時はnanを返す
func (time *Time) SubtractTime(hour uint, minute uint, second float64) Time {

	if time.hour < hour {
		return Time{hour: uint(math.NaN()), minute: uint(math.NaN()), second: math.NaN()}
	}
	if minute > 60 || second > 60 {
		return Time{hour: uint(math.NaN()), minute: uint(math.NaN()), second: math.NaN()}
	}
	tmpHour := time.hour - hour
	tmpminute := time.minute
	tmpsecond := time.second
	if tmpsecond < second {
		if tmpminute == 0 {
			tmpHour--
			tmpminute = 59
		} else {
			tmpminute--
		}

		tmpsecond = tmpsecond + 60 - second
	} else {
		tmpsecond -= second
	}
	if tmpminute < minute {
		tmpHour--
		tmpminute += 60
		tmpminute -= minute
	} else {
		tmpminute -= minute
	}

	return Time{hour: tmpHour, minute: tmpminute, second: tmpsecond}
}

//RevisedTime : 時間を24時間表記に直す
func (time *Time) RevisedTime() Time {
	tmpHour := time.hour % 24
	return Time{hour: tmpHour, minute: time.minute, second: time.second}
}

//timeAがtimeBより早い時刻の時true、そうでないならfalse
func isEarlier(timeA Time, timeB Time) bool {
	if timeA.hour < timeB.hour {
		return true
	}
	if timeA.minute < timeB.minute {
		return true
	}
	if timeA.second < timeB.second {
		return true
	}

	return false
}

//timeAがtimeBより遅い時刻の時true、そうでないならfalse
func isLater(timeA Time, timeB Time) bool {
	if timeA.hour > timeB.hour {
		return true
	}
	if timeA.minute > timeB.minute {
		return true
	}
	if timeA.second > timeB.second {
		return true
	}

	return false
}

//夜間時間を跨いで通常時間に入っていたらtrueを返す
func isStraddledNight(startTime Time, endTime Time) bool {
	for i := startTime.hour; i <= endTime.hour; i++ {
		if i%24 == 22 || i%24 == 23 || (0 <= i%24 && i%24 == 5) {
			if isInNight(startTime) == true && isInNight(endTime) == true {
				return true
			}
		}

	}

	return false

}

//通常時間を跨いで夜間時間に入っていたらtrueを返す
func isStraddledDay(startTime Time, endTime Time) bool {
	for i := startTime.hour; i <= endTime.hour; i++ {
		if 6 <= i%24 && i%24 <= 21 {
			if isInNight(startTime) == false && isInNight(endTime) == false {
				return true
			}
		}
	}

	return false
}

//isInNight : timeが夜間時間中ならtrue、通常時間ならfalse
func isInNight(time Time) bool {
	//time.RevisedTime().hourが22,23,0,1,2,3,4の時は夜間時間
	revisedTime := time.RevisedTime().hour
	if revisedTime == 22 || revisedTime == 23 || 0 <= revisedTime && revisedTime <= 4 {
		return true
	} else if revisedTime == 5 && time.minute == 0 && time.second == 0 {
		//time.RevisedTime().hourが5かつtime.minute = time.second = 0の時も夜間時間
		return true
	} else {
		//そうでないなら通常時間
		return false
	}

}

//NightStart : 夜間時間開始(Time{hour: 22, minute: 0, second: 0})
var NightStart = Time{hour: 22, minute: 0, second: 0}

//NightEnd : 夜間時間終了(Time{hour: 5, minute: 0, second: 0})
var NightEnd = Time{hour: 5, minute: 0, second: 0}

//nightTimeRate : 夜間時の距離、走行時間の倍率(1.25倍)
const nightTimeRate float64 = 1.25

//prevTimeとcurrentTimeの時間差を求める
func computeTimeDifference(prevTime Time, currentTime Time) float64 {
	var hourDifference = math.Abs((float64(currentTime.hour - prevTime.hour)))
	var minuteDifference = math.Abs(float64(currentTime.minute - prevTime.minute))
	var secondDifference = math.Abs(currentTime.second - prevTime.second)
	return 3600*hourDifference + 60*minuteDifference + secondDifference

}

//computeComplicatedCase :  通常時間と夜間時間が混ざっている場合の計算を行う
func computeComplicatedCase(startRecord Record, endRecord Record) (float64, float64) {

	var startTime = startRecord.TotimeForCompute()
	var endTime = endRecord.TotimeForCompute()
	Mileage, _ := strconv.ParseFloat(endRecord.mileage, 64)
	if isInNight(startTime) == false {
		timeDifference := computeTimeDifference(startTime, endTime)
		nightTime := math.Floor(float64(float64(endTime.hour)-22)/24) * 7 * 3600
		if 0 <= ((int(endTime.hour)-22)%24) && ((int(endTime.hour)-22)%24) <= 6 ||
			((int(endTime.hour)-22)%24) == 7 && endTime.minute == 0 && endTime.second == 0 {
			nightTime += float64(int(endTime.hour)-22%24)*3600 + float64(endTime.minute)*60 + endTime.second
		}
		dayTime := timeDifference - nightTime

		dayMileage := Mileage * (dayTime / timeDifference)
		nightMileage := Mileage * (nightTime / timeDifference)

		nightTime *= nightTimeRate
		nightMileage *= nightTimeRate

		return dayTime + nightTime, dayMileage + nightMileage

	} else {
		//翌日の5時
		nightendTomorrow := Time{hour: 29, minute: 0, second: 0}
		//翌5時までの時間を夜間時間に加える
		nightTime := computeTimeDifference(startTime.RevisedTime(), nightendTomorrow)
		//基準となる時間を求めるために24時間表記に直す
		revisedStartTime := startTime.RevisedTime()
		//基準となる時間
		criterionTime := nightendTomorrow.SubtractTime(revisedStartTime.hour, revisedStartTime.minute, revisedStartTime.second)
		criterionTime = criterionTime.AddTime(startTime.hour, startTime.minute, startTime.second)

		timeDifference := computeTimeDifference(criterionTime, endTime)
		nightTime += math.Floor(float64(float64(endTime.hour)-22)/24) * 7 * 3600
		if 0 <= ((int(endTime.hour)-22)%24) && ((int(endTime.hour)-22)%24) <= 6 ||
			((int(endTime.hour)-22)%24) == 7 && endTime.minute == 0 && endTime.second == 0 {
			nightTime += float64((int(endTime.hour)-22)%24)*3600 + float64(endTime.minute)*60 + endTime.second
		}
		dayTime := timeDifference - nightTime

		dayMileage := Mileage * (dayTime / timeDifference)
		nightMileage := Mileage * (nightTime / timeDifference)

		nightTime *= nightTimeRate
		nightMileage *= nightTimeRate

		return dayTime + nightTime, dayMileage + nightMileage

	}

}

//走行時間(秒)と走行距離(m)を求める
//夜間割増もここで処理する
func computeTimeAndMileage(startRecord Record, endRecord Record) (float64, float64) {

	var startTime = startRecord.TotimeForCompute()
	var endTime = endRecord.TotimeForCompute()
	//startTimeが通常時間で、かつendTimeが通常時間で夜間時間を跨がない場合は走行時間と走行距離をそのまま返す
	//5時から始まり、22時に終わる場合も同様なので走行時間と走行距離をそのまま返す
	if (isInNight(startTime) == false && isInNight(endTime) == false) &&
		isStraddledNight(startTime, endTime) == false ||
		(reflect.DeepEqual(startTime.RevisedTime(), NightEnd) == true && reflect.DeepEqual(endTime.RevisedTime(), NightStart) == true) {
		runningTime := computeTimeDifference(startTime, endTime)
		runingMileage, _ := strconv.ParseFloat(endRecord.mileage, 64)
		return runningTime, runingMileage

	} else if isInNight(startTime) == true && isInNight(endTime) == true &&
		isStraddledDay(startTime, endTime) == false ||
		(reflect.DeepEqual(startTime.RevisedTime(), NightStart) == true && reflect.DeepEqual(endTime.RevisedTime(), NightEnd) == true) {
		//startTimeがNightStartより遅く、かつendTimeがNightEndより早くて通常時間を跨がない場合は走行時間と走行距離を1.25倍して返す
		//startTimeがNightStartと一致し、かつendTimeがNightEndと一致する場合も同様なのでその時も同様に処理する

		var nightTime = computeTimeDifference(startTime, endTime) * nightTimeRate
		var nightMileage, _ = strconv.ParseFloat(endRecord.mileage, 64)
		nightMileage *= nightTimeRate
		return nightTime, nightMileage

	} else {
		//どこかで通常時間と夜間時間を跨ぐ時は夜間走行時間と夜間走行距離を別途で求め、通常時間の走行時間と走行距離に加える
		runningTime, runingMileage := computeComplicatedCase(startRecord, endRecord)
		return runningTime, runingMileage

	}

}