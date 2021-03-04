package timeutil

import (
	"errors"
	"math"
)

//Time : 走行記録内の時間を格納
type Time struct {
	Hour   uint
	Minute uint
	//秒には小数点以下も入るのでfloat64型
	Second float64
}

//NewTime : Timeのコンストラクタ
func NewTime(hour uint, minute uint, second float64) Time {
	return Time{Hour: hour, Minute: minute, Second: second}

}

//AddTime : 引数で指定しただけ加えた時刻を返す
func (time *Time) AddTime(hour uint, minute uint, second float64) Time {
	tmpHour := time.Hour + hour
	tmpMinute := time.Minute + minute
	tmpSecond := time.Second + second
	if tmpSecond >= 60 {
		tmpSecond -= 60
		tmpMinute++
	}
	if tmpMinute >= 60 {
		tmpMinute -= 60
		tmpHour++
	}

	return Time{Hour: tmpHour, Minute: tmpMinute, Second: tmpSecond}
}

//SubtractTime : 引数で指定しただけ引いた時刻を返す
//time.hour < hourの時はnanを返す
func (time *Time) SubtractTime(hour uint, minute uint, second float64) (Time, error) {
	//引く時間数の方が引かれる時間数より多い時はエラーを返す
	if time.Hour < hour {
		return Time{Hour: uint(math.NaN()), Minute: uint(math.NaN()), Second: math.NaN()}, errors.New("you try to subtract too much time")
	}
	if minute > 60 || second > 60 {
		return Time{Hour: uint(math.NaN()), Minute: uint(math.NaN()), Second: math.NaN()}, errors.New("minute or second are incorrect")
	}
	tmpHour := time.Hour - hour
	tmpMinute := time.Minute
	tmpSecond := time.Second
	if tmpSecond < second {
		if tmpMinute == 0 {
			tmpHour--
			tmpMinute = 59
		} else {
			tmpMinute--
		}

		tmpSecond = tmpSecond + 60 - second
	} else {
		tmpSecond -= second
	}
	if tmpMinute < minute {
		tmpHour--
		tmpMinute += 60
		tmpMinute -= minute
	} else {
		tmpMinute -= minute
	}

	return Time{Hour: tmpHour, Minute: tmpMinute, Second: tmpSecond}, nil
}

//RevisedTime : 時間を24時間表記に直す
func (time *Time) RevisedTime() Time {
	tmpHour := time.Hour % 24
	return Time{Hour: tmpHour, Minute: time.Minute, Second: time.Second}
}

//IsEarlier : timeAがtimeBより早い時刻の時true、そうでないならfalse
func IsEarlier(timeA Time, timeB Time) bool {
	if timeA.Hour < timeB.Hour {
		return true
	}
	if timeA.Minute < timeB.Minute {
		return true
	}
	if timeA.Second < timeB.Second {
		return true
	}

	return false
}

//IsLater : timeAがtimeBより遅い時刻の時true、そうでないならfalse
func IsLater(timeA Time, timeB Time) bool {
	if timeA.Hour > timeB.Hour {
		return true
	}
	if timeA.Minute > timeB.Minute {
		return true
	}
	if timeA.Second > timeB.Second {
		return true
	}

	return false
}

//IsStraddledNight : 夜間時間を跨いで通常時間に入っていたらtrueを返す
func IsStraddledNight(startTime Time, endTime Time) bool {
	for i := startTime.Hour; i <= endTime.Hour; i++ {
		if i%24 == 22 || i%24 == 23 || (0 <= i%24 && i%24 == 5) {
			if IsInNight(startTime) == true && IsInNight(endTime) == true {
				return true
			}
		}

	}

	return false

}

//IsStraddledDay : 通常時間を跨いで夜間時間に入っていたらtrueを返す
func IsStraddledDay(startTime Time, endTime Time) bool {
	for i := startTime.Hour; i <= endTime.Hour; i++ {
		if 6 <= i%24 && i%24 <= 21 {
			if IsInNight(startTime) == false && IsInNight(endTime) == false {
				return true
			}
		}
	}

	return false
}

//IsInNight : timeが夜間時間中ならtrue、通常時間ならfalse
func IsInNight(time Time) bool {
	//time.RevisedTime().hourが22,23,0,1,2,3,4の時は夜間時間
	revisedTime := time.RevisedTime().Hour
	if revisedTime == 22 || revisedTime == 23 || 0 <= revisedTime && revisedTime <= 4 {
		return true
	} else if revisedTime == 5 && time.Minute == 0 && time.Second == 0 {
		//time.RevisedTime().hourが5かつtime.minute = time.second = 0の時も夜間時間
		return true
	} else {
		//そうでないなら通常時間
		return false
	}

}

//ComputeTimeDifference : prevTimeとcurrentTimeの時間差を求める
func ComputeTimeDifference(prevTime Time, currentTime Time) float64 {
	var hourDifference = math.Abs((float64(currentTime.Hour - prevTime.Hour)))

	var minuteDifference = math.Abs(float64(currentTime.Minute - prevTime.Minute))

	var secondDifference = math.Abs(currentTime.Second - prevTime.Second)

	return 3600*hourDifference + 60*minuteDifference + secondDifference

}
