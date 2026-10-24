package date

import (
	"time"

	"github.com/jalaali/go-jalaali"
)

func JalaliToGregorian(year, month, day int) (time.Time, error) {

	gy, gm, gd, err := jalaali.ToGregorian(
		year,
		jalaali.Month(month),
		day,
	)

	if err != nil {
		return time.Time{}, err
	}

	return time.Date(
		gy,
		gm,
		gd,
		0, 0, 0, 0,
		time.UTC,
	), nil
}

func GregorianToJalali(t time.Time) (int, int, int, error) {

	jy, jm, jd, err := jalaali.ToJalaali(
		t.Year(),
		t.Month(),
		t.Day(),
	)

	if err != nil {
		return 0, 0, 0, err
	}

	return jy, int(jm), jd, nil
}