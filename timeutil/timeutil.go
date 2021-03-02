package timeutil

import (
	"errors"
	"math"
)

//Time : 走行記録内の時間を格納
type Time struct {
	hour   uint
	minute uint
	//秒には小数点以下も入るのでfloat64型
	second float64
}

//AddTime : 引数で指定しただけ加えた時刻を返す
func (time *Time) AddTime(hour uint, minute uint, second float64) Time {
	tmpHour := time.hour + hour
	tmpMinute := time.minute + minute
	tmpSecond := time.second + second
	if tmpSecond >= 60 {
		tmpSecond -= 60
		tmpMinute++
	}
	if tmpMinute >= 60 {
		tmpMinute -= 60
		tmpHour++
	}

	return Time{hour: tmpHour, minute: tmpMinute, second: tmpSecond}
}

//SubtractTime : 引数で指定しただけ引いた時刻を返す
//time.hour < hourの時はnanを返す
func (time *Time) SubtractTime(hour uint, minute uint, second float64) (Time, error) {
	//引く時間数の方が引かれる時間数より多い時はエラーを返す
	if time.hour < hour {
		return Time{hour: uint(math.NaN()), minute: uint(math.NaN()), second: math.NaN()}, errors.New("you try to subtract too much time")
	}
	if minute > 60 || second > 60 {
		return Time{hour: uint(math.NaN()), minute: uint(math.NaN()), second: math.NaN()}, errors.New("minute or second are incorrect")
	}
	tmpHour := time.hour - hour
	tmpMinute := time.minute
	tmpSecond := time.second
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

	return Time{hour: tmpHour, minute: tmpMinute, second: tmpSecond}, nil
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

//prevTimeとcurrentTimeの時間差を求める
func computeTimeDifference(prevTime Time, currentTime Time) float64 {
	var hourDifference = math.Abs((float64(currentTime.hour - prevTime.hour)))

	var minuteDifference = math.Abs(float64(currentTime.minute - prevTime.minute))

	var secondDifference = math.Abs(currentTime.second - prevTime.second)

	return 3600*hourDifference + 60*minuteDifference + secondDifference

}
