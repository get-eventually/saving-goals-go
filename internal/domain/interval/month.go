package interval

import (
	"fmt"
	"time"
)

type MonthStarted struct {
	Month Month
}

type Month struct {
	Year  int
	Month time.Month
}

func MonthFromTime(t time.Time) Month {
	return Month{
		Year:  t.Year(),
		Month: t.Month(),
	}
}

func (m Month) String() string {
	return fmt.Sprintf("%d-%02d", m.Year, m.Month)
}
