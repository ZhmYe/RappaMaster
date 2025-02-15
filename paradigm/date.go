package paradigm

import (
	"fmt"
	"time"
)

func GetDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}
func GetGenesisDate() time.Time {
	return GetDate(2025, time.February, 1)
}
func GetDateDuration(date time.Time) int {
	genesisDate := GetGenesisDate()
	duration := date.Sub(genesisDate)
	return int(duration.Hours() / 24) // 将小时转换为天数
}
func DateFormat(date time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d", date.Year(), date.Month(), date.Day())
}
func TimeFormat(date time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second())
}
func TimestampConvert(timestamp uint64) time.Time {
	// Convert to seconds and nanoseconds
	seconds := int64(timestamp) / 1000
	nanoseconds := (timestamp % 1000) * 1000000
	// Convert to time.Time
	return time.Unix(seconds, int64(nanoseconds))
}
