package model

import (
	"fmt"
	"strconv"
	"time"
)

type User struct {
	ID                  string
	Firstname, Lastname string
	Birthday            time.Time
	PicturePath         string
}
type Users map[string]*User
type AverageData struct {
	Age float64
}

type ageFilter func(*User, uint64) bool
type picFilter func(*User, bool) bool
type UserFilter struct {
	AgeLE, AgeGE   ageFilter
	HasPicture     picFilter
	ageFrom, ageTo uint64
	hasPic         bool
}

func (uss Users) CalculateAverage() AverageData {
	var sumAge uint64
	for _, usr := range uss {
		sumAge += usr.GetAge()
	}
	return AverageData{Age: float64(sumAge) / float64(len(uss))}
}

func NewUserFilter() *UserFilter {
	ageLE := func(user *User, age uint64) bool {
		return true
	}
	ageGE := func(user *User, age uint64) bool {
		return true
	}
	hasPic := func(user *User, has bool) bool {
		return true
	}

	return &UserFilter{
		AgeLE:      ageLE,
		AgeGE:      ageGE,
		HasPicture: hasPic,
	}
}

func (uf *UserFilter) FitsToAll(user *User) bool {
	return uf.AgeLE(user, uf.ageTo) && uf.AgeGE(user, uf.ageFrom) && uf.HasPicture(user, uf.hasPic)
}

func (uf *UserFilter) AdjustFilters(filters map[string]string) error {
	var err error
	val, ok := filters["min_age"]
	if ok {
		uf.ageFrom, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("min_age conversion error %s to uint\t%w", val, err)
		}
		uf.AgeGE = func(user *User, u uint64) bool {
			return user.GetAge() >= u
		}
	}
	val, ok = filters["max_age"]
	if ok {
		uf.ageTo, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("max_age conversion error %s to uint\t%w", val, err)
		}

		uf.AgeLE = func(user *User, u uint64) bool {
			return user.GetAge() <= u
		}
	}
	val, ok = filters["is_image_exists"]
	if ok {
		uf.hasPic, err = strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("is_image_exists conversion error %s to uint\t%w", val, err)
		}

		uf.HasPicture = func(user *User, b bool) bool {
			return len(user.PicturePath) > 0 == b
		}
	}

	return nil
}

func (u *User) GetAge() uint64 {
	now := time.Now()
	timePassed := now.Year() - u.Birthday.Year()
	if now.YearDay() < u.Birthday.YearDay() {
		timePassed--
	}

	return uint64(timePassed)
}
