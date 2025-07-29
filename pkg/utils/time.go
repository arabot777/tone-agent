package utils

import (
	"errors"
	"time"
	"tone/agent/pkg/common/gin/code"
	"tone/agent/pkg/common/pkgerror"
)

const TimeFormatLong = "2006-01-02 15:04:05"

func ParseReqDay(startDayStr, endDayStr string) (time.Time, time.Time, error) {
	// req.StartDay 2024-02-03 转为 2024-02-03 00:00:00
	startDay, err := time.ParseInLocation("2006-01-02", startDayStr, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, pkgerror.WithCode(code.ErrBadParams, "startDay is not a valid date")
	}
	endDay, err := time.ParseInLocation("2006-01-02", endDayStr, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, pkgerror.WithCode(code.ErrBadParams, "endDay is not a valid date")
	}

	if endDay.Before(startDay) {
		return time.Time{}, time.Time{}, pkgerror.WithCode(code.ErrBadParams, "endDay should not before startDay")
	}
	return startDay, endDay, nil
}

func ParseDay(timeStr string) (time.Time, error) {
	//  2024-02-03 转为 2024-02-03 00:00:00

	day, err := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	if err != nil {
		return time.Time{}, pkgerror.WithCode(code.ErrBadParams, "day is not a valid date")
	}
	return day, nil
}

func FormatToPreviousDayEnd(t time.Time) time.Time {
	year, month, day := t.Date()
	previousDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location()).AddDate(0, 0, -1)
	return time.Date(previousDay.Year(), previousDay.Month(), previousDay.Day(), 23, 59, 59, 0, t.Location())
}

func FormatToAfterDayStart(t time.Time) time.Time {
	year, month, day := t.Date()
	afterDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location()).AddDate(0, 0, 1)
	return time.Date(afterDay.Year(), afterDay.Month(), afterDay.Day(), 0, 0, 0, 0, t.Location())

}

func FormatTo0000Day(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func FormatTo2359Day(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, t.Location())
}

func TimestampNowMs() int64 {
	return time.Now().UTC().UnixNano() / 1000000
}

func TimeToStr(time time.Time) string {
	return time.Format(TimeFormatLong)
}

func StrToTime(timeStr string) (time.Time, error) {
	beijingLoc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, errors.New("日期转换异常")
	}
	parse, err := time.ParseInLocation(TimeFormatLong, timeStr, beijingLoc)
	if err != nil {
		return time.Time{}, errors.New("日期转换异常")
	}
	return parse.In(time.UTC), nil
}
